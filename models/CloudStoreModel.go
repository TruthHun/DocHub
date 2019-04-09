package models

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/PuerkitoBio/goquery"

	CloudStore2 "github.com/TruthHun/CloudStore"
	"github.com/TruthHun/DocHub/helper"
)

type CloudStore struct {
	Private       bool
	StoreType     helper.ConfigCate
	CanGZIP       bool
	client        interface{}
	config        interface{}
	expire        int64
	publicDomain  string
	privateDomain string
}

// 创建云存储
func NewCloudStore(private bool) (cs *CloudStore, err error) {
	storeType := helper.ConfigCate(GlobalSys.StoreType)
	modelConfig := NewConfig()
	config := modelConfig.GetGlobalConfigWithStruct(storeType)
	return NewCloudStoreWithConfig(config, storeType, private)
}

var errWithoutConfig = errors.New("云存储配置不正确")

func NewCloudStoreWithConfig(storeConfig interface{}, storeType helper.ConfigCate, private bool) (cs *CloudStore, err error) {
	cs = &CloudStore{
		StoreType: storeType,
		config:    storeConfig,
	}
	cs.Private = private
	switch cs.StoreType {
	case StoreOss:
		cfg := cs.config.(*ConfigOss)
		bucket := cfg.PublicBucket
		domain := cfg.PublicBucketDomain
		if cs.Private {
			bucket = cfg.PrivateBucket
			domain = cfg.PrivateBucketDomain
			if cfg.Expire <= 0 {
				cfg.Expire = 1800
			}
			cs.expire = cfg.Expire
		}
		cs.privateDomain = cfg.PrivateBucketDomain
		cs.publicDomain = cfg.PublicBucketDomain
		if cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Endpoint == "" || bucket == "" {
			err = errWithoutConfig
			return
		}
		cs.client, err = CloudStore2.NewOSS(cfg.AccessKey, cfg.SecretKey, cfg.Endpoint, bucket, domain)
		cs.CanGZIP = true
	case StoreObs:
		cfg := cs.config.(*ConfigObs)
		bucket := cfg.PublicBucket
		domain := cfg.PublicBucketDomain
		if cs.Private {
			bucket = cfg.PrivateBucket
			domain = cfg.PrivateBucketDomain
			if cfg.Expire <= 0 {
				cfg.Expire = 1800
			}
			cs.expire = cfg.Expire
		}
		cs.privateDomain = cfg.PrivateBucketDomain
		cs.publicDomain = cfg.PublicBucketDomain
		if cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Endpoint == "" || bucket == "" {
			err = errWithoutConfig
			return
		}
		cs.client, err = CloudStore2.NewOBS(cfg.AccessKey, cfg.SecretKey, bucket, cfg.Endpoint, domain)
	case StoreQiniu:
		cfg := cs.config.(*ConfigQiniu)
		bucket := cfg.PublicBucket
		domain := cfg.PublicBucketDomain
		if cs.Private {
			bucket = cfg.PrivateBucket
			domain = cfg.PrivateBucketDomain
			if cfg.Expire <= 0 {
				cfg.Expire = 1800
			}
			cs.expire = cfg.Expire
		}
		cs.privateDomain = cfg.PrivateBucketDomain
		cs.publicDomain = cfg.PublicBucketDomain
		if cfg.AccessKey == "" || cfg.SecretKey == "" || bucket == "" {
			err = errWithoutConfig
			return
		}
		cs.client, err = CloudStore2.NewQINIU(cfg.AccessKey, cfg.SecretKey, bucket, domain)
	case StoreUpyun:
		cfg := cs.config.(*ConfigUpYun)
		bucket := cfg.PublicBucket
		domain := cfg.PublicBucketDomain
		if cs.Private {
			bucket = cfg.PrivateBucket
			domain = cfg.PrivateBucketDomain
			if cfg.Expire <= 0 {
				cfg.Expire = 1800
			}
			cs.expire = cfg.Expire
		}
		cs.privateDomain = cfg.PrivateBucketDomain
		cs.publicDomain = cfg.PublicBucketDomain
		if cfg.Operator == "" || cfg.Password == "" || bucket == "" {
			err = errWithoutConfig
			return
		}
		cs.client = CloudStore2.NewUpYun(bucket, cfg.Operator, cfg.Password, domain, cfg.Secret)
	case StoreMinio:
		cfg := cs.config.(*ConfigMinio)
		bucket := cfg.PublicBucket
		domain := cfg.PublicBucketDomain
		if cs.Private {
			bucket = cfg.PrivateBucket
			domain = cfg.PrivateBucketDomain
			if cfg.Expire <= 0 {
				cfg.Expire = 1800
			}
			cs.expire = cfg.Expire
		}
		cs.privateDomain = cfg.PrivateBucketDomain
		cs.publicDomain = cfg.PublicBucketDomain
		if cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Endpoint == "" || bucket == "" {
			err = errWithoutConfig
			return
		}
		cs.client, err = CloudStore2.NewMinIO(cfg.AccessKey, cfg.SecretKey, bucket, cfg.Endpoint, domain)
		cs.CanGZIP = true
	case StoreBos:
		cfg := cs.config.(*ConfigBos)
		bucket := cfg.PublicBucket
		domain := cfg.PublicBucketDomain
		if cs.Private {
			bucket = cfg.PrivateBucket
			domain = cfg.PrivateBucketDomain
			if cfg.Expire <= 0 {
				cfg.Expire = 1800
			}
			cs.expire = cfg.Expire
		}
		cs.privateDomain = cfg.PrivateBucketDomain
		cs.publicDomain = cfg.PublicBucketDomain
		if cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Endpoint == "" || bucket == "" {
			err = errWithoutConfig
			return
		}
		cs.client, err = CloudStore2.NewBOS(cfg.AccessKey, cfg.SecretKey, bucket, cfg.Endpoint, domain)
		cs.CanGZIP = true
	case StoreCos:
		cfg := cs.config.(*ConfigCos)
		bucket := cfg.PublicBucket
		domain := cfg.PublicBucketDomain
		if cs.Private {
			bucket = cfg.PrivateBucket
			domain = cfg.PrivateBucketDomain
			if cfg.Expire <= 0 {
				cfg.Expire = 1800
			}
			cs.expire = cfg.Expire
		}
		cs.privateDomain = cfg.PrivateBucketDomain
		cs.publicDomain = cfg.PublicBucketDomain
		if cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.AppId == "" || bucket == "" || cfg.Region == "" {
			err = errWithoutConfig
			return
		}
		cs.client, err = CloudStore2.NewCOS(cfg.AccessKey, cfg.SecretKey, bucket, cfg.AppId, cfg.Region, domain)
		cs.CanGZIP = true
	}
	return
}

func (c *CloudStore) Upload(tmpFile, saveFile string, headers ...map[string]string) (err error) {
	switch c.StoreType {
	case StoreCos:
		err = c.client.(*CloudStore2.COS).Upload(tmpFile, saveFile, headers...)
	case StoreOss:
		err = c.client.(*CloudStore2.OSS).Upload(tmpFile, saveFile, headers...)
	case StoreBos:
		err = c.client.(*CloudStore2.BOS).Upload(tmpFile, saveFile, headers...)
	case StoreObs:
		err = c.client.(*CloudStore2.OBS).Upload(tmpFile, saveFile, headers...)
	case StoreUpyun:
		err = c.client.(*CloudStore2.UpYun).Upload(tmpFile, saveFile, headers...)
	case StoreMinio:
		err = c.client.(*CloudStore2.MinIO).Upload(tmpFile, saveFile, headers...)
	case StoreQiniu:
		err = c.client.(*CloudStore2.QINIU).Upload(tmpFile, saveFile, headers...)
	}
	return
}

func (c *CloudStore) Delete(objects ...string) (err error) {
	switch c.StoreType {
	case StoreCos:
		err = c.client.(*CloudStore2.COS).Delete(objects...)
	case StoreOss:
		err = c.client.(*CloudStore2.OSS).Delete(objects...)
	case StoreBos:
		err = c.client.(*CloudStore2.BOS).Delete(objects...)
	case StoreObs:
		err = c.client.(*CloudStore2.OBS).Delete(objects...)
	case StoreUpyun:
		err = c.client.(*CloudStore2.UpYun).Delete(objects...)
	case StoreMinio:
		err = c.client.(*CloudStore2.MinIO).Delete(objects...)
	case StoreQiniu:
		err = c.client.(*CloudStore2.QINIU).Delete(objects...)
	}
	return
}

// err 返回 nil，表示文件存在，否则表示文件不存在
func (c *CloudStore) IsExist(object string) (err error) {
	switch c.StoreType {
	case StoreCos:
		err = c.client.(*CloudStore2.COS).IsExist(object)
	case StoreOss:
		err = c.client.(*CloudStore2.OSS).IsExist(object)
	case StoreBos:
		err = c.client.(*CloudStore2.BOS).IsExist(object)
	case StoreObs:
		err = c.client.(*CloudStore2.OBS).IsExist(object)
	case StoreUpyun:
		err = c.client.(*CloudStore2.UpYun).IsExist(object)
	case StoreMinio:
		err = c.client.(*CloudStore2.MinIO).IsExist(object)
	case StoreQiniu:
		err = c.client.(*CloudStore2.QINIU).IsExist(object)
	}
	return
}

func GetImageFromCloudStore(picture string, ext ...string) (link string) {
	cs, err := NewCloudStore(false)
	if err != nil {
		helper.Logger.Error(err.Error())
	}
	return cs.getImageFromCloudStore(picture, ext...)
}

//设置默认图片
//@param                picture             图片文件或者图片文件md5等
//@param                ext                 图片扩展名，如果图片文件参数(picture)的值为md5时，需要加上后缀扩展名
//@return               link                图片url链接
func (c *CloudStore) getImageFromCloudStore(picture string, ext ...string) (link string) {
	if len(ext) > 0 {
		picture = picture + "." + ext[0]
	} else if !strings.Contains(picture, ".") && len(picture) > 0 {
		picture = picture + ".jpg"
	}
	if c == nil || c.client == nil {
		return
	}

	return c.GetSignURL(picture)
}

func (c *CloudStore) GetSignURL(object string) (link string) {
	var err error
	switch c.StoreType {
	case StoreCos:
		link, err = c.client.(*CloudStore2.COS).GetSignURL(object, c.expire)
	case StoreOss:
		link, err = c.client.(*CloudStore2.OSS).GetSignURL(object, c.expire)
	case StoreBos:
		link, err = c.client.(*CloudStore2.BOS).GetSignURL(object, c.expire)
	case StoreObs:
		link, err = c.client.(*CloudStore2.OBS).GetSignURL(object, c.expire)
	case StoreUpyun:
		link, err = c.client.(*CloudStore2.UpYun).GetSignURL(object, c.expire)
	case StoreMinio:
		link, err = c.client.(*CloudStore2.MinIO).GetSignURL(object, c.expire)
	case StoreQiniu:
		link, err = c.client.(*CloudStore2.QINIU).GetSignURL(object, c.expire)
	}
	if err != nil {
		helper.Logger.Error("GetSignURL:%v", err.Error())
	}
	return
}

func (c *CloudStore) ImageWithDomain(htmlOld string) (htmlNew string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlOld))
	if err != nil {
		helper.Logger.Error(err.Error())
		return htmlOld
	}
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		if src, exist := s.Attr("src"); exist {
			if !(strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://")) {
				s.SetAttr("src", c.GetSignURL(src))
			}
		}

	})
	htmlNew, _ = doc.Find("body").Html()
	return
}

func (c *CloudStore) ImageWithoutDomain(htmlOld string) (htmlNew string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlOld))
	if err != nil {
		helper.Logger.Error(err.Error())
		return htmlOld
	}
	domain := c.GetPublicDomain()

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		if src, exist := s.Attr("src"); exist {
			//不存在http开头的图片链接，则更新为绝对链接
			if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
				src = strings.TrimPrefix(src, domain)
				s.SetAttr("src", src)
			}
		}
	})
	htmlNew, _ = doc.Find("body").Html()
	return
}

//从HTML中提取图片文件，并删除
func (c *CloudStore) DeleteImageFromHtml(htmlStr string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	var objects []string

	domain := c.GetPublicDomain()

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		if src, exist := s.Attr("src"); exist {
			//不存在http开头的图片链接，则更新为绝对链接
			if !(strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://")) {
				objects = append(objects, src)
			} else {
				src = strings.TrimPrefix(src, domain)
				objects = append(objects, src)
			}
		}
	})
	if err = c.Delete(objects...); err != nil {
		helper.Logger.Error(err.Error())
	}
}

func (c *CloudStore) PingTest() (err error) {
	tmpFile := "dochub-test-file.txt"
	saveFile := "dochub-test-file.txt"
	text := "hello world"

	defer func() {
		if err != nil {
			err = fmt.Errorf("Bucket是否私有：%v，错误信息：%v", c.Private, err.Error())
		}
	}()

	err = ioutil.WriteFile(tmpFile, []byte(text), os.ModePerm)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	if err = c.Upload(tmpFile, saveFile); err != nil {
		return
	}
	if err = c.IsExist(saveFile); err != nil {
		return
	}
	if !helper.Debug {
		if err = c.Delete(saveFile); err != nil {
			return
		}
	}

	return
}

func (c *CloudStore) GetPublicDomain() (domain string) {
	object := "test.dochub.test"
	link := c.GetSignURL(object)
	return strings.TrimRight(strings.Split(link, object)[0], "/")
}
