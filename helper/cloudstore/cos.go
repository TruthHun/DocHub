package CloudStore

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mozillazg/go-cos"
)

type COS struct {
	SecretId  string
	SecretKey string
	Bucket    string
	AppID     string
	Region    string
	client    *cos.Client
}

func NewCOS(secretId, secretKey, bucket, appId, region string) (c *COS, err error) {
	var baseURL *cos.BaseURL

	c = &COS{
		SecretId:  secretId,
		SecretKey: secretKey,
		Bucket:    bucket,
		AppID:     appId,
		Region:    region,
	}

	//	https://wafer-1251298948.cos.ap-guangzhou.myqcloud.com/
	baseURL, err = cos.NewBaseURL(fmt.Sprintf("https://%v-%v.cos.%v.myqcloud.com", bucket, appId, region))
	if err != nil {
		return
	}
	c.client = cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretId,
			SecretKey: secretKey,
		}})
	return
}

func (c *COS) PutObject(local, object string, header map[string]string) (err error) {
	var b []byte
	b, err = ioutil.ReadFile(local)
	if err != nil {
		return
	}
	var opt *cos.ObjectPutOptions
	if len(header) > 0 {
		for key, val := range header {
			// TODO: more headers to set
			switch strings.ToLower(key) {
			case "content-encoding":
				opt.ContentEncoding = val
			case "content-disposition":
				opt.ContentDisposition = val
			case "content-type":
				opt.ContentType = val
			}
		}
	}
	_, err = c.client.Object.Put(context.Background(), object, bytes.NewReader(b), opt)
	return
}

func (c *COS) DeleteObjects(objects []string) (err error) {
	if len(objects) > 0 {
		var opt = &cos.ObjectDeleteMultiOptions{}
		var objs []cos.Object
		for _, obj := range objects {
			objs = append(objs, cos.Object{Key: obj})
		}
		opt.Objects = objs
		_, _, err = c.client.Object.DeleteMulti(context.Background(), opt)
	}
	return
}

func (c *COS) GetObjectURL(object string, expire int64) (urlStr string, err error) {
	var uri *url.URL
	auth := cos.Auth{
		SecretID:  c.SecretId,
		SecretKey: c.SecretKey,
		Expire:    time.Duration(expire) * time.Second,
	}
	uri, err = c.client.Object.PresignedURL(context.Background(), http.MethodGet, "oss.go", auth, nil)
	if err == nil {
		urlStr = uri.String()
	}
	return
}
