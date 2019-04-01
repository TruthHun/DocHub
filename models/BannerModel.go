package models

import (
	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego/orm"
)

//横幅
type Banner struct {
	Id         int    `orm:"column(Id)"`
	Picture    string `orm:"column(Picture);default();size(50)"` //图片
	Title      string `orm:"column(Title);default()"`            //标题
	Link       string `orm:"column(Link);default()"`             //链接
	Sort       int    `orm:"column(Sort);default(0)"`            //排序
	Status     bool   `orm:"column(Status);default(true)"`       //状态，false表示关闭，true表示正常显示
	TimeCreate int    `orm:"column(TimeCreate);default(0)"`      //横幅添加时间
}

func NewBanner() *Banner {
	return &Banner{}
}

func GetTableBanner() string {
	return getTable("banner")
}

//删除横幅：删除横幅记录，同时删除横幅图片
//@param            id              横幅id
//@return           affected        影响的记录数
//@return           err             错误
func (this *Banner) Del(id ...interface{}) (affected int64, err error) {

	if len(id) == 0 {
		return
	}

	var banners []Banner
	var objects []string

	defer func() {
		go func() {
			if len(objects) > 0 {
				cs, errCS := NewCloudStore(false)
				if errCS != nil {
					helper.Logger.Error(errCS.Error())
					return
				}
				if errDelete := cs.Delete(objects...); errDelete != nil {
					helper.Logger.Error(errDelete.Error())
				}
			}
		}()
	}()

	qs := orm.NewOrm().QueryTable(GetTableBanner()).Filter("Id__in", id...)
	qs.All(&banners)
	for _, banner := range banners {
		if len(banner.Picture) > 0 {
			objects = append(objects, banner.Picture)
		}
	}
	return qs.Delete()
}

//获取横幅列表
//@param            p               页码
//@param            listRows        每页记录数
//@param            status          横幅状态，0表示关闭，1表示正常，不传值则获取全部
//@return           banners         返回列表
//@return           rows            返回记录数
//@return           err             错误
func (this *Banner) List(p, listRows int, status ...int) (banners []Banner, rows int64, err error) {
	qs := orm.NewOrm().QueryTable(GetTableBanner()).OrderBy("Sort", "-Status", "-Id").Limit(listRows).Offset((p - 1) * listRows)
	if len(status) > 0 {
		qs.Filter("Status__in", status)
	}
	rows, err = qs.All(&banners)
	return
}
