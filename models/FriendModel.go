package models

import "github.com/astaxie/beego/orm"

//友情链接表
type Friend struct {
	Id         int    `orm:"column(Id)"`
	Title      string `orm:"column(Title);size(100)"`       //友链名称
	Link       string `orm:"column(Link);size(100);unique"` //友链地址
	Status     bool   `orm:"column(Status);default(true)"`  //友链状态
	Sort       int    `orm:"column(Sort);default(0)"`       //友链排序，值越小越靠前
	TimeCreate int    `orm:"column(TimeCreate);default(0)"` //友链添加时间
}

func NewFriend() *Friend {
	return &Friend{}
}

func GetTableFriend() string {
	return getTable("friend")
}

//获取指定状态的友链
//@param            status          友链状态，-1表示全部状态的友链，0表示关闭状态的友链，1表示正常状态的友链
//@return           links           友链数组
//@return           rows            记录数
//@return           err             错误
func (this *Friend) GetListByStatus(status int) (links []Friend, rows int64, err error) {
	qs := orm.NewOrm().QueryTable(GetTableFriend())
	if status == 0 || status == 1 {
		qs = qs.Filter("Status", status)
	}
	rows, err = qs.OrderBy("sort", "-status", "-id").All(&links)
	return
}

//获取友链
func (this *Friend) Friends() (links []Friend) {
	links, _, _ = this.GetListByStatus(1)
	return
}
