package models

import "strings"

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
	O.QueryTable(TableSeo).Filter("Page", page).One(&seoStruct)
	if seoStruct.Id > 0 {
		seo["Title"] = replaceFunc(seoStruct.Title, defSeo)
		seo["Keywords"] = replaceFunc(seoStruct.Keywords, defSeo)
		seo["Description"] = replaceFunc(seoStruct.Description, defSeo)
	}
	return seo
}
