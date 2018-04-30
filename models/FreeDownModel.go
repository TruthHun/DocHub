package models

//免费下载
type FreeDown struct {
	Id         int `orm:"column(Id)"`
	Uid        int `orm:"column(Uid)"`                   //用户id
	Did        int `orm:"column(Did)"`                   //文档id
	TimeCreate int `orm:"column(TimeCreate);default(0)"` //文档上次下载时间
}

//是否可以免费下载，如果之前下载过而且未过免费下载期，可以继续免费下载【注意时间校验，这里只是返回值】
func (this *FreeDown) IsFreeDown(uid, did interface{}) (free FreeDown) {
	O.QueryTable(TableFreeDown).Filter("Uid", uid).Filter("Did", did).One(&free)
	return
}
