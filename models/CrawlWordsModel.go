package models

//采集关键字

type CrawlWords struct {
	Id         int    `orm:"column(Id)"`                             //自增id
	Word       string `orm:"column(Word);default();unique;size(50)"` //关键字
	Baidu      bool   `orm:"column(Baidu);default(false)"`           //是否百度采集了
	Bing       bool   `orm:"column(Bing);default(false)"`            //是否必应采集了
	Sogou      bool   `orm:"column(Sogou);default(false)"`           //是否搜狗采集了
	Google     bool   `orm:"column(Google);default(false)"`          //是否谷歌采集了
	TimeCreate int    `orm:"column(TimeCreate);default(0)"`          //关键字添加时间
}
