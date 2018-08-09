package models

import (
	"strings"
	"time"

	"os"

	"strconv"

	"fmt"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/gotil/sitemap"
	"github.com/astaxie/beego/orm"
)

//SEO配置表
type Seo struct {
	Id          int    `orm:"column(Id)"`                                 //主键
	Name        string `orm:"column(Name);default()"`                     //SEO页面名称
	Page        string `orm:"column(Page);default();size(30);unique"`     //SEO页面【英文】
	IsMobile    bool   `orm:"column(IsMobile);default(false)"`            //是否是手机页面
	Title       string `orm:"column(Title);default({title})"`             //SEO标题
	Keywords    string `orm:"column(Keywords);default({keywords})"`       //SEO关键字
	Description string `orm:"column(Description);default({description})"` //SEO摘要
}

func NewSeo() *Seo {
	return &Seo{}
}

func GetTableSeo() string {
	return getTable("seo")
}

//获取SEO
//@param                page                页面
//@param                defaultTitle        默认标题
//@param                defaultKeywords     默认关键字
//@param                defaultDescription  默认摘要
//@param                Sitename            站点名称
//@return               seo                 SEO数据
func (this *Seo) GetByPage(page string, defaultTitle, defaultKeywords, defaultDescription, Sitename string) (seo map[string]string) {
	var seoStruct Seo
	seo = make(map[string]string)
	defSeo := map[string]string{
		"Title":       defaultTitle,
		"Keywords":    defaultKeywords,
		"Description": defaultDescription,
		"Sitename":    Sitename,
	}
	replace := map[string]string{
		"Title":       "{title}",
		"Keywords":    "{keywords}",
		"Description": "{description}",
		"Sitename":    "{sitename}",
	}
	replaceFunc := func(item string, defSeo map[string]string) string {
		for k, v := range replace {
			item = strings.Replace(item, v, defSeo[k], -1)
		}
		return item
	}
	orm.NewOrm().QueryTable(GetTableSeo()).Filter("Page", page).One(&seoStruct)
	if seoStruct.Id > 0 {
		seo["Title"] = replaceFunc(seoStruct.Title, defSeo)
		seo["Keywords"] = replaceFunc(seoStruct.Keywords, defSeo)
		seo["Description"] = replaceFunc(seoStruct.Description, defSeo)
	}
	return seo
}

//baseUrl := this.Ctx.Input.Scheme() + "://" + this.Ctx.Request.Host
//if host := beego.AppConfig.String("sitemap_host"); len(host) > 0 {
//	baseUrl = this.Ctx.Input.Scheme() + "://" + host
//}
//生成站点地图
func (this *Seo) BuildSitemap() {
	//更新站点地图
	helper.Logger.Info(fmt.Sprintf("[%v]更新站点地图[start]", time.Now().Format("2006-01-02 15:04:05")))
	var (
		files   []string
		fileNum int
		Sitemap = sitemap.NewSitemap("1.0", "utf-8")
		si      []sitemap.SitemapIndex
		count   int64
		limit   = 10000 //每个sitemap文件，限制10000个链接
		domain  = strings.ToLower(NewSys().GetByField("DomainPc").DomainPc)
		o       = orm.NewOrm()
	)
	if !(strings.HasPrefix(domain, "https://") || strings.HasPrefix(domain, "http://")) {
		domain = "http://" + domain
	}
	domain = strings.TrimRight(domain, "/")
	//文档总数
	count, _ = o.QueryTable(GetTableDocumentInfo()).Filter("Status__gt", -1).Count()
	cnt := int(count)
	if fileNum = cnt / limit; cnt%limit > 0 {
		fileNum = fileNum + 1
	}
	//创建文件夹
	os.MkdirAll("sitemap", os.ModePerm)
	for i := 0; i < fileNum; i++ {
		var docs []DocumentInfo
		o.QueryTable(GetTableDocumentInfo()).Filter("Status__gt", -1).Limit(limit).Offset(i*limit).All(&docs, "Id", "TimeCreate")
		if len(docs) > 0 {
			//文件
			file := "sitemap/doc-" + strconv.Itoa(i) + ".xml"
			files = append(files, file)
			var su []sitemap.SitemapUrl
			for _, doc := range docs {
				su = append(su, sitemap.SitemapUrl{
					Loc:        domain + "/view/" + strconv.Itoa(doc.Id),
					Lastmod:    time.Unix(int64(doc.TimeCreate), 0).Format("2006-01-02 15:04:05"),
					ChangeFreq: sitemap.WEEKLY,
					Priority:   0.9,
				})
			}
			if err := Sitemap.CreateSitemapContent(su, file); err != nil {
				helper.Logger.Error("sitemap生成失败：" + err.Error())
			}
		}
	}
	if len(files) > 0 {
		for _, f := range files {
			si = append(si, sitemap.SitemapIndex{
				Loc:     domain + "/" + f,
				Lastmod: time.Now().Format("2006-01-02 15:04:05"),
			})
		}
	}
	if err := Sitemap.CreateSitemapIndex(si, "sitemap.xml"); err != nil {
		helper.Logger.Error("sitemap生成失败：" + err.Error())
	}
	helper.Logger.Info(fmt.Sprintf("[%v]更新站点地图[end]", time.Now().Format("2006-01-02 15:04:05")))
}
