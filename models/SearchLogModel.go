package models

//搜索记录表
type SearchLog struct {
	Id int    `orm:"column(Id)"`                 //自增主键
	Wd string `orm:"column(Wd);size(20);unique"` //用户搜索的关键字
}

//获取最新的搜索关键字
func (this *SearchLog) List(p, listRows int) (rows []SearchLog) {
	O.QueryTable(TableSearchLog).Limit(listRows).Offset((p - 1) * listRows).OrderBy("-Id").All(&rows)
	return
}
