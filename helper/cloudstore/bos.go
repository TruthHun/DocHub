package CloudStore

import (
	"strings"

	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
)

type BOS struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Endpoint  string
	client    *bos.Client
}

// new bos
func NewBOS(accessKey, secretKey, bucket, endpoint string) (b *BOS, err error) {
	b = &BOS{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Bucket:    bucket,
		Endpoint:  endpoint,
	}
	b.client, err = bos.NewClient(accessKey, secretKey, endpoint)
	if err != nil {
		return
	}
	return
}

// put object
func (b *BOS) PutObject(local, object string, header map[string]string) (err error) {
	args := &api.PutObjectArgs{UserMeta: make(map[string]string)}
	if len(header) > 0 {
		for key, val := range header {
			// TODO: more headers to set
			switch strings.ToLower(key) {
			case "content-type":
				args.ContentType = val
			case "content-disposition":
				args.ContentDisposition = val
				//case "content-encoding":
				//	args.ContentEncoding = val
				//default:
				//	args.UserMeta[key] = val
			}
		}
	}
	_, err = b.client.PutObjectFromFile(b.Bucket, object, local, args)
	return
}

// delete objects
func (b *BOS) DeleteObjects(objects []string) (err error) {
	if len(objects) > 0 {
		b.client.DeleteMultipleObjectsFromKeyList(b.Bucket, objects)
	}
	return
}

// get objects url
func (b *BOS) GetObjectURL(object string, expire int64) (urlStr string, err error) {
	urlStr = b.client.BasicGeneratePresignedUrl(b.Bucket, object, int(expire))
	return
}
