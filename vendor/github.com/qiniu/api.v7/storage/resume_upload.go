package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/qiniu/x/xlog.v7"
)

// 分片上传过程中可能遇到的错误
var (
	ErrInvalidPutProgress = errors.New("invalid put progress")
	ErrPutFailed          = errors.New("resumable put failed")
	ErrUnmatchedChecksum  = errors.New("unmatched checksum")
	ErrBadToken           = errors.New("invalid token")
)

// 上传进度过期错误
const (
	InvalidCtx = 701 // UP: 无效的上下文(bput)，可能情况：Ctx非法或者已经被淘汰（太久未使用）
)

// 分片上传默认参数设置
const (
	defaultWorkers   = 4               // 默认的并发上传的块数量
	defaultChunkSize = 4 * 1024 * 1024 // 默认的分片大小，4MB
	defaultTryTimes  = 3               // bput 失败重试次数
)

// Settings 为分片上传设置
type Settings struct {
	TaskQsize int // 可选。任务队列大小。为 0 表示取 Workers * 4。
	Workers   int // 并行 Goroutine 数目。
	ChunkSize int // 默认的Chunk大小，不设定则为4M
	TryTimes  int // 默认的尝试次数，不设定则为3
}

// 分片上传的默认设置
var settings = Settings{
	TaskQsize: defaultWorkers * 4,
	Workers:   defaultWorkers,
	ChunkSize: defaultChunkSize,
	TryTimes:  defaultTryTimes,
}

// SetSettings 可以用来设置分片上传参数
func SetSettings(v *Settings) {
	settings = *v
	if settings.Workers == 0 {
		settings.Workers = defaultWorkers
	}
	if settings.TaskQsize == 0 {
		settings.TaskQsize = settings.Workers * 4
	}
	if settings.ChunkSize == 0 {
		settings.ChunkSize = defaultChunkSize
	}
	if settings.TryTimes == 0 {
		settings.TryTimes = defaultTryTimes
	}
}

var tasks chan func()

func worker(tasks chan func()) {
	for {
		task := <-tasks
		task()
	}
}
func initWorkers() {
	tasks = make(chan func(), settings.TaskQsize)
	for i := 0; i < settings.Workers; i++ {
		go worker(tasks)
	}
}

// 上传完毕块之后的回调
func notifyNil(blkIdx int, blkSize int, ret *BlkputRet) {}
func notifyErrNil(blkIdx int, blkSize int, err error)   {}

const (
	blockBits = 22
	blockMask = (1 << blockBits) - 1
)

// BlockCount 用来计算文件的分块数量
func BlockCount(fsize int64) int {
	return int((fsize + blockMask) >> blockBits)
}

// BlkputRet 表示分片上传每个片上传完毕的返回值
type BlkputRet struct {
	Ctx       string `json:"ctx"`
	Checksum  string `json:"checksum"`
	Crc32     uint32 `json:"crc32"`
	Offset    uint32 `json:"offset"`
	Host      string `json:"host"`
	ExpiredAt int64  `json:"expired_at"`
}

// RputExtra 表示分片上传额外可以指定的参数
type RputExtra struct {
	Params     map[string]string // 可选。用户自定义参数，以"x:"开头，而且值不能为空，否则忽略
	UpHost     string
	MimeType   string                                        // 可选。
	ChunkSize  int                                           // 可选。每次上传的Chunk大小
	TryTimes   int                                           // 可选。尝试次数
	Progresses []BlkputRet                                   // 可选。上传进度
	Notify     func(blkIdx int, blkSize int, ret *BlkputRet) // 可选。进度提示（注意多个block是并行传输的）
	NotifyErr  func(blkIdx int, blkSize int, err error)
}

var once sync.Once

// Put 方法用来上传一个文件，支持断点续传和分块上传。
//
// ctx     是请求的上下文。
// ret     是上传成功后返回的数据。如果 upToken 中没有设置 CallbackUrl 或 ReturnBody，那么返回的数据结构是 PutRet 结构。
// upToken 是由业务服务器颁发的上传凭证。
// key     是要上传的文件访问路径。比如："foo/bar.jpg"。注意我们建议 key 不要以 '/' 开头。另外，key 为空字符串是合法的。
// f       是文件内容的访问接口。考虑到需要支持分块上传和断点续传，要的是 io.ReaderAt 接口，而不是 io.Reader。
// fsize   是要上传的文件大小。
// extra   是上传的一些可选项。详细见 RputExtra 结构的描述。
//
func (p *ResumeUploader) Put(ctx context.Context, ret interface{}, upToken string, key string, f io.ReaderAt,
	fsize int64, extra *RputExtra) (err error) {
	err = p.rput(ctx, ret, upToken, key, true, f, fsize, extra)
	return
}

// PutWithoutKey 方法用来上传一个文件，支持断点续传和分块上传。文件命名方式首先看看
// upToken 中是否设置了 saveKey，如果设置了 saveKey，那么按 saveKey 要求的规则生成 key，否则自动以文件的 hash 做 key。
//
// ctx     是请求的上下文。
// ret     是上传成功后返回的数据。如果 upToken 中没有设置 CallbackUrl 或 ReturnBody，那么返回的数据结构是 PutRet 结构。
// upToken 是由业务服务器颁发的上传凭证。
// f       是文件内容的访问接口。考虑到需要支持分块上传和断点续传，要的是 io.ReaderAt 接口，而不是 io.Reader。
// fsize   是要上传的文件大小。
// extra   是上传的一些可选项。详细见 RputExtra 结构的描述。
//
func (p *ResumeUploader) PutWithoutKey(
	ctx context.Context, ret interface{}, upToken string, f io.ReaderAt, fsize int64, extra *RputExtra) (err error) {
	err = p.rput(ctx, ret, upToken, "", false, f, fsize, extra)
	return
}

// PutFile 用来上传一个文件，支持断点续传和分块上传。
// 和 Put 不同的只是一个通过提供文件路径来访问文件内容，一个通过 io.ReaderAt 来访问。
//
// ctx       是请求的上下文。
// ret       是上传成功后返回的数据。如果 upToken 中没有设置 CallbackUrl 或 ReturnBody，那么返回的数据结构是 PutRet 结构。
// upToken   是由业务服务器颁发的上传凭证。
// key       是要上传的文件访问路径。比如："foo/bar.jpg"。注意我们建议 key 不要以 '/' 开头。另外，key 为空字符串是合法的。
// localFile 是要上传的文件的本地路径。
// extra     是上传的一些可选项。详细见 RputExtra 结构的描述。
//
func (p *ResumeUploader) PutFile(
	ctx context.Context, ret interface{}, upToken, key, localFile string, extra *RputExtra) (err error) {
	err = p.rputFile(ctx, ret, upToken, key, true, localFile, extra)
	return
}

// PutFileWithoutKey 上传一个文件，支持断点续传和分块上传。文件命名方式首先看看
// upToken 中是否设置了 saveKey，如果设置了 saveKey，那么按 saveKey 要求的规则生成 key，否则自动以文件的 hash 做 key。
// 和 PutWithoutKey 不同的只是一个通过提供文件路径来访问文件内容，一个通过 io.ReaderAt 来访问。
//
// ctx       是请求的上下文。
// ret       是上传成功后返回的数据。如果 upToken 中没有设置 CallbackUrl 或 ReturnBody，那么返回的数据结构是 PutRet 结构。
// upToken   是由业务服务器颁发的上传凭证。
// localFile 是要上传的文件的本地路径。
// extra     是上传的一些可选项。详细见 RputExtra 结构的描述。
//
func (p *ResumeUploader) PutFileWithoutKey(
	ctx context.Context, ret interface{}, upToken, localFile string, extra *RputExtra) (err error) {
	return p.rputFile(ctx, ret, upToken, "", false, localFile, extra)
}

func (p *ResumeUploader) rput(
	ctx context.Context, ret interface{}, upToken string,
	key string, hasKey bool, f io.ReaderAt, fsize int64, extra *RputExtra) (err error) {

	once.Do(initWorkers)

	log := xlog.NewWith(ctx)
	blockCnt := BlockCount(fsize)

	if extra == nil {
		extra = new(RputExtra)
	}
	if extra.Progresses == nil {
		extra.Progresses = make([]BlkputRet, blockCnt)
	} else if len(extra.Progresses) != blockCnt {
		return ErrInvalidPutProgress
	}

	if extra.ChunkSize == 0 {
		extra.ChunkSize = settings.ChunkSize
	}
	if extra.TryTimes == 0 {
		extra.TryTimes = settings.TryTimes
	}
	if extra.Notify == nil {
		extra.Notify = notifyNil
	}
	if extra.NotifyErr == nil {
		extra.NotifyErr = notifyErrNil
	}
	//get up host

	var upHost string
	if extra.UpHost != "" {
		upHost = extra.UpHost
	} else {
		ak, bucket, gErr := getAkBucketFromUploadToken(upToken)
		if gErr != nil {
			err = gErr
			return
		}

		upHost, gErr = p.UpHost(ak, bucket)
		if gErr != nil {
			err = gErr
			return
		}
	}

	var wg sync.WaitGroup
	wg.Add(blockCnt)

	last := blockCnt - 1
	blkSize := 1 << blockBits
	nfails := 0

	for i := 0; i < blockCnt; i++ {
		blkIdx := i
		blkSize1 := blkSize
		if i == last {
			offbase := int64(blkIdx) << blockBits
			blkSize1 = int(fsize - offbase)
		}
		task := func() {
			defer wg.Done()
			tryTimes := extra.TryTimes
		lzRetry:
			err := p.resumableBput(ctx, upToken, upHost, &extra.Progresses[blkIdx], f, blkIdx, blkSize1, extra)
			if err != nil {
				if tryTimes > 1 {
					tryTimes--
					log.Info("resumable.Put retrying ...", blkIdx, "reason:", err)
					goto lzRetry
				}
				log.Warn("resumable.Put", blkIdx, "failed:", err)
				extra.NotifyErr(blkIdx, blkSize1, err)
				nfails++
			}
		}
		tasks <- task
	}

	wg.Wait()
	if nfails != 0 {
		return ErrPutFailed
	}

	return p.Mkfile(ctx, upToken, upHost, ret, key, hasKey, fsize, extra)
}

func (p *ResumeUploader) rputFile(
	ctx context.Context, ret interface{}, upToken string,
	key string, hasKey bool, localFile string, extra *RputExtra) (err error) {

	f, err := os.Open(localFile)
	if err != nil {
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return
	}

	return p.rput(ctx, ret, upToken, key, hasKey, f, fi.Size(), extra)
}

func (p *ResumeUploader) UpHost(ak, bucket string) (upHost string, err error) {
	var zone *Zone
	if p.Cfg.Zone != nil {
		zone = p.Cfg.Zone
	} else {
		if v, zoneErr := GetZone(ak, bucket); zoneErr != nil {
			err = zoneErr
			return
		} else {
			zone = v
		}
	}

	scheme := "http://"
	if p.Cfg.UseHTTPS {
		scheme = "https://"
	}

	host := zone.SrcUpHosts[0]
	if p.Cfg.UseCdnDomains {
		host = zone.CdnUpHosts[0]
	}

	upHost = fmt.Sprintf("%s%s", scheme, host)
	return
}
