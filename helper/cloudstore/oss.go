package CloudStore

import (
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OSS struct {
	Key       string
	Secret    string
	Endpoint  string
	Bucket    string
	bucketObj *oss.Bucket
}

// New OSS
func NewOSS(key, secret, endpoint, bucket string) (o *OSS, err error) {
	var client *oss.Client
	o = &OSS{
		Key:      key,
		Secret:   secret,
		Endpoint: endpoint,
		Bucket:   bucket,
	}
	client, err = oss.New(endpoint, key, secret)
	if err != nil {
		return
	}
	o.bucketObj, err = client.Bucket(bucket)
	return
}

// put object
func (o *OSS) PutObject(local, object string, header map[string]string) (err error) {
	var opts []oss.Option
	if len(header) > 0 {
		for key, val := range header {
			// TODO: more headers to set
			switch strings.ToLower(key) {
			case "content-encoding":
				opts = append(opts, oss.ContentEncoding(val))
			case "content-disposition":
				opts = append(opts, oss.ContentDisposition(val))
			case "content-type":
				opts = append(opts, oss.ContentType(val))
			}
		}
	}
	return o.bucketObj.PutObjectFromFile(object, local, opts...)
}

// delete objects
func (o *OSS) DeleteObjects(objects []string) (err error) {
	if len(objects) > 0 {
		_, err = o.bucketObj.DeleteObjects(objects)
	}
	return
}

// get objects url
func (o *OSS) GetObjectURL(object string, expire int64) (urlStr string, err error) {
	//TODO:
}
