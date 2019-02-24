package storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Base64Uploader 表示一个Base64上传对象
type Base64Uploader struct {
	client *Client
	cfg    *Config
}

// NewBase64Uploader 用来构建一个Base64上传的对象
func NewBase64Uploader(cfg *Config) *Base64Uploader {
	if cfg == nil {
		cfg = &Config{}
	}

	return &Base64Uploader{
		client: &DefaultClient,
		cfg:    cfg,
	}
}

// NewBase64UploaderEx 用来构建一个Base64上传的对象
func NewBase64UploaderEx(cfg *Config, client *Client) *Base64Uploader {
	if cfg == nil {
		cfg = &Config{}
	}

	if client == nil {
		client = &DefaultClient
	}

	return &Base64Uploader{
		client: client,
		cfg:    cfg,
	}
}

// Base64PutExtra 为Base64上传的额外可选项
type Base64PutExtra struct {
	// 可选，用户自定义参数，必须以 "x:" 开头。若不以x:开头，则忽略。
	Params map[string]string

	// 可选，当为 "" 时候，服务端自动判断。
	MimeType string
}

// Put 用来以Base64方式上传一个文件
//
// ctx        是请求的上下文。
// ret        是上传成功后返回的数据。如果 uptoken 中没有设置 callbackUrl 或 returnBody，那么返回的数据结构是 PutRet 结构。
// uptoken    是由业务服务器颁发的上传凭证。
// key        是要上传的文件访问路径。比如："foo/bar.jpg"。注意我们建议 key 不要以 '/' 开头。另外，key 为空字符串是合法的。
// base64Data 是要上传的Base64数据，一般为图片数据的Base64编码字符串
// extra      是上传的一些可选项，可以指定为nil。详细见 Base64PutExtra 结构的描述。
//
func (p *Base64Uploader) Put(
	ctx context.Context, ret interface{}, uptoken, key string, base64Data []byte, extra *Base64PutExtra) (err error) {
	return p.put(ctx, ret, uptoken, key, true, base64Data, extra)
}

// PutWithoutKey 用来以Base64方式上传一个文件，保存的文件名以文件的内容hash作为文件名
func (p *Base64Uploader) PutWithoutKey(
	ctx context.Context, ret interface{}, uptoken string, base64Data []byte, extra *Base64PutExtra) (err error) {
	return p.put(ctx, ret, uptoken, "", false, base64Data, extra)
}

func (p *Base64Uploader) put(
	ctx context.Context, ret interface{}, uptoken, key string, hasKey bool, base64Data []byte, extra *Base64PutExtra) (err error) {
	//get up host
	ak, bucket, gErr := getAkBucketFromUploadToken(uptoken)
	if gErr != nil {
		err = gErr
		return
	}

	var upHost string
	upHost, err = p.upHost(ak, bucket)
	if err != nil {
		return
	}

	//set default extra
	if extra == nil {
		extra = &Base64PutExtra{}
	}

	//calc crc32
	h := crc32.NewIEEE()
	rawReader := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(base64Data))
	fsize, decodeErr := io.Copy(h, rawReader)
	if decodeErr != nil {
		err = fmt.Errorf("invalid base64 data, %s", decodeErr.Error())
		return
	}
	fCrc32 := h.Sum32()

	postPath := bytes.NewBufferString("/putb64")
	//add fsize
	postPath.WriteString("/")
	postPath.WriteString(strconv.Itoa(int(fsize)))

	//add key
	if hasKey {
		postPath.WriteString("/key/")
		postPath.WriteString(base64.URLEncoding.EncodeToString([]byte(key)))
	}
	//add mimeType
	if extra.MimeType != "" {
		postPath.WriteString("/mimeType/")
		postPath.WriteString(base64.URLEncoding.EncodeToString([]byte(extra.MimeType)))
	}

	//add crc32
	postPath.WriteString("/crc32/")
	postPath.WriteString(fmt.Sprintf("%d", fCrc32))

	//add extra params
	if len(extra.Params) > 0 {
		for k, v := range extra.Params {
			if strings.HasPrefix(k, "x:") && v != "" {
				postPath.WriteString("/")
				postPath.WriteString(k)
				postPath.WriteString("/")
				postPath.WriteString(base64.URLEncoding.EncodeToString([]byte(v)))
			}
		}
	}

	postURL := fmt.Sprintf("%s%s", upHost, postPath.String())
	headers := http.Header{}
	headers.Add("Content-Type", "application/octet-stream")
	headers.Add("Authorization", "UpToken "+uptoken)

	return p.client.CallWith(ctx, ret, "POST", postURL, headers, bytes.NewReader(base64Data), len(base64Data))
}

func (p *Base64Uploader) upHost(ak, bucket string) (upHost string, err error) {
	var zone *Zone
	if p.cfg.Zone != nil {
		zone = p.cfg.Zone
	} else {
		if v, zoneErr := GetZone(ak, bucket); zoneErr != nil {
			err = zoneErr
			return
		} else {
			zone = v
		}
	}

	scheme := "http://"
	if p.cfg.UseHTTPS {
		scheme = "https://"
	}

	host := zone.SrcUpHosts[0]
	if p.cfg.UseCdnDomains {
		host = zone.CdnUpHosts[0]
	}

	upHost = fmt.Sprintf("%s%s", scheme, host)
	return
}
