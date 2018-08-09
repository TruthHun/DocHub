package models

//相关文档记录存储表
type Relate struct {
	Id         int    `orm:"column(Id)"`            //文档的ID
	Ids        string `orm:"column(Ids);size(512)"` //相关文档的ID字符串
	TimeCreate int    `orm:"column(TimeCreate)"`    //记录的创建时间
}

func NewRelate() *Relate {
	return &Relate{}
}

func GetTableRelate() string {
	return getTable("relate")
}
