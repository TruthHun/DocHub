package models

//TODO
//意见和建议表，用于收集用户的意见和建议
type Suggest struct {
	Id         int    `orm:"column(Id)"`
	Uid        int    `orm:"column(Uid);default(0)"`          //意见和建议提交人
	Content    string `orm:"column(Content);size(512);"`      //内容
	Email      string `orm:"column(Email);size(50)"`          //邮箱
	Name       string `orm:"column(Name);size(20);default()"` //称呼
	TimeCreate int    `orm:"column(TimeCreate)"`              //意见建议提交时间
	TimeUpdate int    `orm:"column(TimeUpdate)"`              //意见建议查看时间
	Status     bool   `orm:"column(Status);default(false)"`   //是否已查看
}

func NewSuggest() *Suggest {
	return &Suggest{}
}
func GetTableSuggest() string {
	return getTable("suggest")
}
