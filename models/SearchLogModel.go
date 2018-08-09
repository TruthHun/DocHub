package models

import "github.com/astaxie/beego/orm"

//搜索记录表
type SearchLog struct {
	Id int    `orm:"column(Id)"`                 //自增主键
	Wd string `orm:"column(Wd);size(20);unique"` //用户搜索的关键字
}

func NewSearchLog() *SearchLog {
	return &SearchLog{}
}

func GetTableSearchLog() string {
	return getTable("search_log")
}

//获取最新的搜索关键字。这里的查询error可以忽略
//@param            p           页面
//@param            listRows    每页显示记录
//@return           rows        查询到的记录
func (this *SearchLog) List(p, listRows int) (rows []SearchLog) {
	orm.NewOrm().QueryTable(GetTableSearchLog()).Limit(listRows).Offset((p - 1) * listRows).OrderBy("-Id").All(&rows)
	return
}
