package models

import "github.com/astaxie/beego/orm"

//单页
type Pages struct {
	Id          int    `orm:"column(Id)"`
	Name        string `orm:"column(Name);size(100)"`        //单页名称，如关于我们
	Alias       string `orm:"column(Alias);unique;size(30)"` //单页别名，数字和字母
	Title       string `orm:"column(Title)"`                 //单页标题，如关于IT文库
	Keywords    string `orm:"column(Keywords)"`              //单页关键字
	Description string `orm:"column(Description)"`           //单页摘要
	Content     string `orm:"column(Content);size(5120)"`    //单页文章内容
	TimeCreate  int    `orm:"column(TimeCreate)"`            //单页创建时间
	Sort        int    `orm:"column(Sort);default(100)"`     //单页排序，值越小越靠前
	Vcnt        int    `orm:"column(Vcnt);default(0)"`       //单页浏览记录
	Status      bool   `orm:"column(Status);default(true)"`  //单页状态，true显示，false不显示
}

func NewPages() *Pages {
	return &Pages{}
}

func GetTablePages() string {
	return getTable("pages")
}

//查询单页列表
func (this *Pages) List(listRows int, status ...int) (pages []Pages, rows int64, err error) {
	qs := orm.NewOrm().QueryTable(GetTablePages())
	if len(status) > 0 {
		qs = qs.Filter("Status", status[0])
	}
	cols := []string{"Name", "Alias", "Title", "Keywords", "Description", "Id", "TimeCreate", "Sort", "Vcnt", "Status"}
	rows, err = qs.OrderBy("sort").Limit(listRows).All(&pages, cols...)
	return
}

//查询单页内容
func (this *Pages) One(alias string) (page Pages, err error) {
	err = orm.NewOrm().QueryTable(GetTablePages()).Filter("Alias", alias).One(&page)
	return page, err
}
