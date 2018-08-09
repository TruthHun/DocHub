package models

import (
	"fmt"
	"io/ioutil"

	"time"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego/orm"
)

//文档备注，用于侵权文档等的部分内容的预览展示，并在文档预览页面挂上跳转购买正版的导购链接；同时对于一些开源书籍，也可以一面提供站内文档的下载，一面引导用户购买正版。
type DocumentRemark struct {
	Id            int    `orm:"column(Id)"`                           //document_store表中的ID，当前台页面查询文档的备注的时候，以DsId作为document_remark表中的主键进行查询
	Content       string `orm:"column(Content);size(5120);default()"` //备注内容内容
	AllowDownload bool   `orm:"column(AllowDownload);default(true)"`  //是否允许下载文档
	Status        bool   `orm:"column(Status);default(true)"`         //备注状态，true表示显示，false表示隐藏
	TimeCreate    int    `orm:"column(TimeCreate);default(0)"`        //创建时间
}

func NewDocumentRemark() *DocumentRemark {
	return &DocumentRemark{}
}

func GetTableDocumentRemark() string {
	return getTable("document_remark")
}

//注意：这里处理GetParseContentByDocId，其他基本上是模板内容的形式
//{$title}：文档的标题
//{$md5}:文档的md5
//{$share}:文档的分享时间
//{$uid}:文档的上传用户
//{$username}:文档分享用户名
//{$sitename}:站点名称

//根据文档id获取文档的备注
//@param            id              document_store表中的ID[TODO:注意！！！！(这里加TODO，主要是为了在IDE上更显眼)]
//@return           rm              文档备注
//@return           err             文档错误
func (this *DocumentRemark) GetParseContentByDocId(docid interface{}) (rm DocumentRemark, err error) {
	err = orm.NewOrm().QueryTable(GetTableDocumentRemark()).Filter("Id", docid).One(&rm)
	return
}

//获取内容模板
//@param            DsId            文档DsId
//@return           rm              生成的文档备注模板
func (this *DocumentRemark) GetContentTplByDsId(DsId int) (rm DocumentRemark) {
	rm.Id = DsId
	if err := orm.NewOrm().Read(&rm); err != nil || rm.Id == 0 {
		return this.GetDefaultTpl(DsId)
	}
	return rm
}

//获取文档备注模板
//@param            DsId            文档DsId
//@return           rm              生成的文档备注模板
func (this *DocumentRemark) GetDefaultTpl(DsId int) (rm DocumentRemark) {
	rm.Id = DsId
	rm.Status = false
	rm.AllowDownload = true
	rm.TimeCreate = 0
	if bytes, err := ioutil.ReadFile("./conf/remarktpl.html"); err != nil {
		rm.Content = fmt.Sprintf("模板文件打开失败：%v", err.Error())
	} else {
		rm.Content = string(bytes)
	}
	return
}

//根据dsid判断文档是否已存在备注
func (this *DocumentRemark) IsRemark(dsid interface{}) bool {
	var rm = DocumentRemark{Id: helper.Interface2Int(dsid)}
	if rm.Id > 0 {
		if orm.NewOrm().Read(&rm); rm.TimeCreate > 0 {
			return true
		}
	}
	return false
}

//新增或更改内容，如果TimeCreate为0，表示新增，否则表示更新
//@param                rm              备注内容
//@return               err             返回错误，nil表示成功
func (this *DocumentRemark) Insert(rm DocumentRemark) (err error) {
	if rm.TimeCreate == 0 {
		rm.TimeCreate = int(time.Now().Unix())
		_, err = orm.NewOrm().Insert(&rm)
	} else {
		_, err = orm.NewOrm().Update(&rm)
	}
	return err
}
