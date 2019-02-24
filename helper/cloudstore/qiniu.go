package CloudStore

import (
	"context"
	"time"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

type QINIU struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
}

func NewQINIU(accessKey, secretKey, bucket, domain string) (q *QINIU) {
	return &QINIU{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Bucket:    bucket,
		Domain:    domain,
	}
}

func (q *QINIU) PutObject(local, object string, header map[string]string) (err error) {
	// TODO: set headers
	putPolicy := storage.PutPolicy{
		Scope: q.Bucket,
	}
	mac := qbox.NewMac(q.AccessKey, q.SecretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{
		UseHTTPS:      false,
		UseCdnDomains: false,
	}
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{} // 可选配置
	err = formUploader.PutFile(context.Background(), &ret, upToken, object, local, &putExtra)
	return
}

func (q *QINIU) DeleteObjects(objects []string) (err error) {
	//每个batch的操作数量不可以超过1000个，如果总数量超过1000，需要分批发送

	deleteOps := make([]string, 0, len(objects))
	for _, key := range objects {
		deleteOps = append(deleteOps, storage.URIDelete(q.Bucket, key))
	}

	mac := qbox.NewMac(q.AccessKey, q.SecretKey)
	cfg := storage.Config{
		UseHTTPS:      false,
		UseCdnDomains: false,
	}

	// 指定空间所在的区域，如果不指定将自动探测
	// 如果没有特殊需求，默认不需要指定
	//cfg.Zone=&storage.ZoneHuabei
	bucketManager := storage.NewBucketManager(mac, &cfg)

	_, err = bucketManager.Batch(deleteOps)

	return
}

func (q *QINIU) GetObjectURL(object string, expire int64) (urlStr string, err error) {
	mac := qbox.NewMac(q.AccessKey, q.SecretKey)
	deadline := time.Now().Add(time.Second * time.Duration(expire)).Unix() //1小时有效期
	urlStr = storage.MakePrivateURL(mac, q.Domain, object, deadline)
	return
}
