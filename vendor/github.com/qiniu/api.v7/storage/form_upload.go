package storage

import (
	"bytes"
	"context"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// PutExtra 为表单上传的额外可选项
type PutExtra struct {
	// 可选，用户自定义参数，必须以 "x:" 开头。若不以x:开头，则忽略。
	Params map[string]string

	UpHost string

	// 可选，当为 "" 时候，服务端自动判断。
	MimeType string

	// 上传事件：进度通知。这个事件的回调函数应该尽可能快地结束。
	OnProgress func(fsize, uploaded int64)
}

// PutRet 为七牛标准的上传回复内容。
// 如果使用了上传回调或者自定义了returnBody，那么需要根据实际情况，自己自定义一个返回值结构体
type PutRet struct {
	Hash         string `json:"hash"`
	PersistentID string `json:"persistentId"`
	Key          string `json:"key"`
}

// FormUploader 表示一个表单上传的对象
type FormUploader struct {
	Client *Client
	Cfg    *Config
}

// NewFormUploader 用来构建一个表单上传的对象
func NewFormUploader(cfg *Config) *FormUploader {
	if cfg == nil {
		cfg = &Config{}
	}

	return &FormUploader{
		Client: &DefaultClient,
		Cfg:    cfg,
	}
}

// NewFormUploaderEx 用来构建一个表单上传的对象
func NewFormUploaderEx(cfg *Config, client *Client) *FormUploader {
	if cfg == nil {
		cfg = &Config{}
	}

	if client == nil {
		client = &DefaultClient
	}

	return &FormUploader{
		Client: client,
		Cfg:    cfg,
	}
}

// PutFile 用来以表单方式上传一个文件，和 Put 不同的只是一个通过提供文件路径来访问文件内容，一个通过 io.Reader 来访问。
//
// ctx       是请求的上下文。
// ret       是上传成功后返回的数据。如果 uptoken 中没有设置 callbackUrl 或 returnBody，那么返回的数据结构是 PutRet 结构。
// uptoken   是由业务服务器颁发的上传凭证。
// key       是要上传的文件访问路径。比如："foo/bar.jpg"。注意我们建议 key 不要以 '/' 开头。另外，key 为空字符串是合法的。
// localFile 是要上传的文件的本地路径。
// extra     是上传的一些可选项，可以指定为nil。详细见 PutExtra 结构的描述。
//
func (p *FormUploader) PutFile(
	ctx context.Context, ret interface{}, uptoken, key, localFile string, extra *PutExtra) (err error) {
	return p.putFile(ctx, ret, uptoken, key, true, localFile, extra)
}

// PutFileWithoutKey 用来以表单方式上传一个文件。不指定文件上传后保存的key的情况下，文件命名方式首先看看
// uptoken 中是否设置了 saveKey，如果设置了 saveKey，那么按 saveKey 要求的规则生成 key，否则自动以文件的 hash 做 key。
// 和 Put 不同的只是一个通过提供文件路径来访问文件内容，一个通过 io.Reader 来访问。
//
// ctx       是请求的上下文。
// ret       是上传成功后返回的数据。如果 uptoken 中没有设置 CallbackUrl 或 ReturnBody，那么返回的数据结构是 PutRet 结构。
// uptoken   是由业务服务器颁发的上传凭证。
// localFile 是要上传的文件的本地路径。
// extra     是上传的一些可选项。可以指定为nil。详细见 PutExtra 结构的描述。
//
func (p *FormUploader) PutFileWithoutKey(
	ctx context.Context, ret interface{}, uptoken, localFile string, extra *PutExtra) (err error) {
	return p.putFile(ctx, ret, uptoken, "", false, localFile, extra)
}

func (p *FormUploader) putFile(
	ctx context.Context, ret interface{}, uptoken string,
	key string, hasKey bool, localFile string, extra *PutExtra) (err error) {

	f, err := os.Open(localFile)
	if err != nil {
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return
	}
	fsize := fi.Size()

	if extra == nil {
		extra = &PutExtra{}
	}

	return p.put(ctx, ret, uptoken, key, hasKey, f, fsize, extra, filepath.Base(localFile))
}

// Put 用来以表单方式上传一个文件。
//
// ctx     是请求的上下文。
// ret     是上传成功后返回的数据。如果 uptoken 中没有设置 callbackUrl 或 returnBody，那么返回的数据结构是 PutRet 结构。
// uptoken 是由业务服务器颁发的上传凭证。
// key     是要上传的文件访问路径。比如："foo/bar.jpg"。注意我们建议 key 不要以 '/' 开头。另外，key 为空字符串是合法的。
// data    是文件内容的访问接口（io.Reader）。
// fsize   是要上传的文件大小。
// extra   是上传的一些可选项。可以指定为nil。详细见 PutExtra 结构的描述。
//
func (p *FormUploader) Put(
	ctx context.Context, ret interface{}, uptoken, key string, data io.Reader, size int64, extra *PutExtra) (err error) {
	err = p.put(ctx, ret, uptoken, key, true, data, size, extra, path.Base(key))
	return
}

// PutWithoutKey 用来以表单方式上传一个文件。不指定文件上传后保存的key的情况下，文件命名方式首先看看 uptoken 中是否设置了 saveKey，
// 如果设置了 saveKey，那么按 saveKey 要求的规则生成 key，否则自动以文件的 hash 做 key。
//
// ctx     是请求的上下文。
// ret     是上传成功后返回的数据。如果 uptoken 中没有设置 CallbackUrl 或 ReturnBody，那么返回的数据结构是 PutRet 结构。
// uptoken 是由业务服务器颁发的上传凭证。
// data    是文件内容的访问接口（io.Reader）。
// fsize   是要上传的文件大小。
// extra   是上传的一些可选项。详细见 PutExtra 结构的描述。
//
func (p *FormUploader) PutWithoutKey(
	ctx context.Context, ret interface{}, uptoken string, data io.Reader, size int64, extra *PutExtra) (err error) {
	err = p.put(ctx, ret, uptoken, "", false, data, size, extra, "filename")
	return err
}

func (p *FormUploader) put(
	ctx context.Context, ret interface{}, uptoken string,
	key string, hasKey bool, data io.Reader, size int64, extra *PutExtra, fileName string) (err error) {

	var upHost string
	if extra.UpHost != "" {
		upHost = extra.UpHost
	} else {
		ak, bucket, gErr := getAkBucketFromUploadToken(uptoken)
		if gErr != nil {
			err = gErr
			return
		}

		upHost, err = p.UpHost(ak, bucket)
		if err != nil {
			return
		}
	}

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	if extra == nil {
		extra = &PutExtra{}
	}

	if extra.OnProgress != nil {
		data = &readerWithProgress{reader: data, fsize: size, onProgress: extra.OnProgress}
	}

	err = writeMultipart(writer, uptoken, key, hasKey, extra, fileName)
	if err != nil {
		return
	}

	var dataReader io.Reader

	h := crc32.NewIEEE()
	dataReader = io.TeeReader(data, h)
	crcReader := newCrc32Reader(writer.Boundary(), h)
	//write file
	head := make(textproto.MIMEHeader)
	head.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`,
		escapeQuotes(fileName)))
	if extra.MimeType != "" {
		head.Set("Content-Type", extra.MimeType)
	}

	_, err = writer.CreatePart(head)
	if err != nil {
		return
	}

	lastLine := fmt.Sprintf("\r\n--%s--\r\n", writer.Boundary())
	r := strings.NewReader(lastLine)

	bodyLen := int64(-1)
	if size >= 0 {
		bodyLen = int64(b.Len()) + size + int64(len(lastLine))
		bodyLen += crcReader.length()
	}

	mr := io.MultiReader(&b, dataReader, crcReader, r)

	contentType := writer.FormDataContentType()
	headers := http.Header{}
	headers.Add("Content-Type", contentType)
	err = p.Client.CallWith64(ctx, ret, "POST", upHost, headers, mr, bodyLen)
	if err != nil {
		return
	}
	if extra.OnProgress != nil {
		extra.OnProgress(size, size)
	}

	return
}

type crc32Reader struct {
	h                hash.Hash32
	boundary         string
	r                io.Reader
	flag             bool
	nlDashBoundaryNl string
	header           string
	crc32PadLen      int64
}

func newCrc32Reader(boundary string, h hash.Hash32) *crc32Reader {
	nlDashBoundaryNl := fmt.Sprintf("\r\n--%s\r\n", boundary)
	header := `Content-Disposition: form-data; name="crc32"` + "\r\n\r\n"
	return &crc32Reader{
		h:                h,
		boundary:         boundary,
		nlDashBoundaryNl: nlDashBoundaryNl,
		header:           header,
		crc32PadLen:      10,
	}
}

func (r *crc32Reader) Read(p []byte) (int, error) {
	if r.flag == false {
		crc32Sum := r.h.Sum32()
		crc32Line := r.nlDashBoundaryNl + r.header + fmt.Sprintf("%010d", crc32Sum) //padding crc32 results to 10 digits
		r.r = strings.NewReader(crc32Line)
		r.flag = true
	}
	return r.r.Read(p)
}

func (r crc32Reader) length() (length int64) {
	return int64(len(r.nlDashBoundaryNl+r.header)) + r.crc32PadLen
}

func (p *FormUploader) UpHost(ak, bucket string) (upHost string, err error) {
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

type readerWithProgress struct {
	reader     io.Reader
	uploaded   int64
	fsize      int64
	onProgress func(fsize, uploaded int64)
}

func (p *readerWithProgress) Read(b []byte) (n int, err error) {
	if p.uploaded > 0 {
		p.onProgress(p.fsize, p.uploaded)
	}

	n, err = p.reader.Read(b)
	p.uploaded += int64(n)
	return
}

func writeMultipart(writer *multipart.Writer, uptoken, key string, hasKey bool,
	extra *PutExtra, fileName string) (err error) {

	//token
	if err = writer.WriteField("token", uptoken); err != nil {
		return
	}

	//key
	if hasKey {
		if err = writer.WriteField("key", key); err != nil {
			return
		}
	}

	//extra.Params
	if extra.Params != nil {
		for k, v := range extra.Params {
			if (strings.HasPrefix(k, "x:") || strings.HasPrefix(k, "x-qn-meta-")) && v != "" {
				err = writer.WriteField(k, v)
				if err != nil {
					return
				}
			}
		}
	}

	return err
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
