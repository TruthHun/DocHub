package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

//免费下载，如果用户花费金币下载了一次文档，下次在下载则免费
type FreeDown struct {
	Id         int `orm:"column(Id)"`
	Uid        int `orm:"column(Uid)"`                   //用户id
	Did        int `orm:"column(Did)"`                   //文档id
	TimeCreate int `orm:"column(TimeCreate);default(0)"` //文档上次下载时间
}

func NewFreeDown() *FreeDown {
	return &FreeDown{}
}

func GetTableFreeDown() string {
	return getTable("free_down")
}

//是否可以免费下载，如果之前下载过而且未过免费下载期，可以继续免费下载【注意时间校验，这里只是返回值】
//@param            uid         用户id
//@param            did         文档id，document id
//@return           isFree      是否免费
func (this *FreeDown) IsFreeDown(uid, did interface{}) (isFree bool) {
	var free FreeDown
	orm.NewOrm().QueryTable(GetTableFreeDown()).Filter("Uid", uid).Filter("Did", did).One(&free)
	if free.Id > 0 && free.TimeCreate+NewSys().GetByField("FreeDay").FreeDay > int(time.Now().Unix()) {
		return true
	}
	return
}
