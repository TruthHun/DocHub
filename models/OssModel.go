package models

import (
	"errors"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego"

	"strings"

	"os"

	"time"

	"fmt"

	"bytes"
	"io/ioutil"

	"compress/gzip"

	"github.com/PuerkitoBio/goquery"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	oss2 "github.com/denverdino/aliyungo/oss"
)

//OSS配置【这个不再作为数据库表，直接在oss.conf文件中进行配置】
type Oss struct {
	Id               int    //主键
	EndpointInternal string //内网的endpoint
	EndpointOuter    string //外网的endpoint
	AccessKeyId      string //key
	AccessKeySecret  string //secret
	BucketPreview    string //供文档预览的bucket
	BucketStore      string //供文档存储的bucket
	IsInternal       bool   //是否内网，内网则启用内网endpoint，否则启用外网endpoint
	PreviewUrl       string //预览链接
	DownloadUrl      string //下载链接
	UrlExpire        int    //签名链接默认有效期时间，单位为秒
}

//获取oss的配置
//@return               oss             Oss配置信息
func NewOss() (oss *Oss) {
	oss = &Oss{
		IsInternal:       beego.AppConfig.DefaultBool("oss::IsInternal", false),
		EndpointInternal: beego.AppConfig.String("oss::EndpointInternal"),
		EndpointOuter:    beego.AppConfig.String("oss::EndpointOuter"),
		AccessKeyId:      beego.AppConfig.String("oss::AccessKeyId"),
		AccessKeySecret:  beego.AppConfig.String("oss::AccessKeySecret"),
		BucketPreview:    beego.AppConfig.String("oss::BucketPreview"),
		BucketStore:      beego.AppConfig.String("oss::BucketStore"),
		UrlExpire:        beego.AppConfig.DefaultInt("oss::UrlExpire", 60),
		PreviewUrl:       strings.TrimRight(beego.AppConfig.String("oss::PreviewUrl"), "/") + "/",
		DownloadUrl:      strings.TrimRight(beego.AppConfig.String("oss::DownloadUrl"), "/") + "/",
	}

	return oss
}

//创建OSS Bucket
//@param		IsPreview			是否是公共读的Bucket。false表示私有Bucket
func (this *Oss) NewBucket(IsPreview bool) (bucket *oss.Bucket, err error) {
	var (
		client   *oss.Client
		endpoint string
	)
	if this.IsInternal {
		endpoint = this.EndpointInternal
	} else {
		endpoint = this.EndpointOuter
	}
	if client, err = oss.New(endpoint, this.AccessKeyId, this.AccessKeySecret); err == nil {
		if IsPreview {
			bucket, err = client.Bucket(this.BucketPreview)
		} else {
			bucket, err = client.Bucket(this.BucketStore)
		}
	}
	return
}

//判断文件对象是否存在
//@param                object              文件对象
//@param                isBucketPreview     是否是供预览的bucket，true表示预览bucket，false表示存储bucket
//@return               err                 错误，nil表示文件存在，否则表示文件不存在，并告知错误信息
func (this *Oss) IsObjectExist(object string, isBucketPreview bool) (err error) {
	bucket, err := this.NewBucket(isBucketPreview)
	if err == nil {
		if ok, err := bucket.IsObjectExist(object); err != nil {
			helper.Logger.Error(err.Error())
		} else if !ok {
			err = errors.New(fmt.Sprintf("文件 %v 不存在", object))
		}
	}
	return
}

//设置默认图片
//@param                picture             图片文件
//@param                style               图片处理风格
//@param                ext                 图片扩展名，如果图片文件参数(picture)的值为md5时，需要加上后缀扩展名
//@return               url                 图片url链接
func (this *Oss) DefaultPicture(picture, style string, ext ...string) (url string) {
	beego.Error(picture)
	if len(ext) > 0 {
		picture = picture + "." + ext[0]
	} else if !strings.Contains(picture, ".") && len(picture) > 0 {
		picture = picture + ".jpg"
	}
	picture = strings.Trim(picture, "/")
	style = strings.ToLower(style)
	switch style {
	case "avatar", "cover", "banner":
		return this.PreviewUrl + picture + "/" + style
	}
	return this.PreviewUrl + picture //返回默认图片
}

//文件移动到OSS进行存储
//@param            local            本地文件
//@param            save             存储到OSS的文件
//@param            IsPreview        是否是预览，是预览的话，则上传到预览的OSS，否则上传到存储的OSS。存储的OSS，只作为文档的存储，以供下载，但不提供预览等访问，为私有
//@param            IsDel            文件上传后，是否删除本地文件
//@param            IsGzip           是否做gzip压缩，做gzip压缩的话，需要修改oss中对象的响应头，设置gzip响应
func (this *Oss) MoveToOss(local, save string, IsPreview, IsDel bool, IsGzip ...bool) error {

	bucket, err := this.NewBucket(IsPreview)
	if err != nil {
		helper.Logger.Error("OSS Bucket初始化错误：%v", err.Error())
		return err
	}

	isgzip := false
	//如果是开启了gzip，则需要设置文件对象的响应头
	if len(IsGzip) > 0 && IsGzip[0] == true {
		isgzip = true
	}

	//在移动文件到OSS之前，先压缩文件
	if isgzip {
		if bs, err := ioutil.ReadFile(local); err != nil {
			helper.Logger.Error(err.Error())
			isgzip = false //设置为false
		} else {
			var by bytes.Buffer
			w := gzip.NewWriter(&by)
			defer w.Close()
			w.Write(bs)
			w.Flush()
			ioutil.WriteFile(local, by.Bytes(), 0777)
		}
	}
	err = bucket.PutObjectFromFile(save, local)
	if err != nil {
		helper.Logger.Error("文件移动到OSS失败：%v", err.Error())
	}
	//如果是开启了gzip，则需要设置文件对象的响应头
	if isgzip {
		bucket.SetObjectMeta(save, oss.ContentEncoding("gzip")) //设置gzip响应头
	}

	if err == nil && IsDel {
		err = os.Remove(local)
	}

	return err
}

//从OSS中删除文件
//@param           object                     文件对象
//@param           IsPreview                  是否是预览的OSS
func (this *Oss) DelFromOss(IsPreview bool, object ...string) error {
	bucket, err := this.NewBucket(IsPreview)
	if err == nil {
		_, err = bucket.DeleteObjects(object)
	}
	return err
}

//OSS文档访问链接签名
//@param                object              文档存储对象
//@param                expire              文档过期时间，不传递，则使用配置文件中的默认时间
//@return               url                 url签名链接
func (this *Oss) BuildSign(object string, expire ...int) (url string) {
	slice := strings.Split(this.EndpointOuter, ".")
	client := oss2.NewOSSClient(oss2.Region(slice[0]), false, this.AccessKeyId, this.AccessKeySecret, true)
	bucket := client.Bucket(this.BucketStore)
	if len(expire) > 0 {
		this.UrlExpire = expire[0]
	}
	if slice := strings.Split(bucket.SignedURL(object, time.Now().Add(time.Duration(this.UrlExpire)*time.Second)), "aliyuncs.com/"); len(slice) == 2 {
		url = this.DownloadUrl + slice[1]
	}
	return
}

//设置文件的下载名
//@param            obj             文档对象
//@param            filename        文件名
func (this *Oss) SetObjectMeta(obj, filename string) {
	if client, err := oss.New(this.EndpointOuter, this.AccessKeyId, this.AccessKeySecret); err == nil {
		if bucket, err := client.Bucket(this.BucketStore); err == nil {
			bucket.SetObjectMeta(obj, oss.ContentDisposition(fmt.Sprintf("attachment; filename=%v", filename)))
		}
	}
}

//OSS文档访问链接签名，有效期为当天
//@param                object              文档存储对象
//@return               url                 url签名链接
func (this *Oss) BuildSignDaily(object string) (url string) {
	slice := strings.Split(this.EndpointOuter, ".")
	client := oss2.NewOSSClient(oss2.Region(slice[0]), false, this.AccessKeyId, this.AccessKeySecret, true)
	bucket := client.Bucket(this.BucketStore)
	time_format := "2006-01-02 00:00:00"
	t, _ := time.Parse(time_format, time.Now().Format(time_format))
	//创建有效期为
	return bucket.SignedURL(object, t.Add(24*time.Hour))
}

//处理html中的OSS数据：如果是用于预览的内容，则把img等的链接的相对路径转成绝对路径，否则反之
//@param            htmlstr             html字符串
//@param            forPreview          是否是供浏览的页面需求
//@return           str                 处理后返回的字符串
func (this *Oss) HandleContent(htmlstr string, forPreview bool) (str string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlstr))
	if err == nil {
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the band and title
			if src, exist := s.Attr("src"); exist {
				//预览
				if forPreview {
					//不存在http开头的图片链接，则更新为绝对链接
					if !(strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://")) {
						s.SetAttr("src", this.PreviewUrl+strings.TrimLeft(src, "./"))
					}
				} else {
					s.SetAttr("src", strings.TrimPrefix(src, this.PreviewUrl))
				}
			}

		})
		str, _ = doc.Find("body").Html()
	}
	return
}

//从HTML中提取图片文件，并删除
func (this *Oss) DelByHtmlPics(htmlstr string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlstr))
	if err == nil {
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the band and title
			if src, exist := s.Attr("src"); exist {
				//不存在http开头的图片链接，则更新为绝对链接
				if !(strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://")) {
					this.DelFromOss(true, strings.TrimLeft(src, "./")) //删除
				} else if strings.HasPrefix(src, this.PreviewUrl) {
					this.DelFromOss(true, strings.TrimPrefix(src, this.PreviewUrl)) //删除
				}
			}
		})
	}
	return
}
