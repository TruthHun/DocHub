package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

//文档表
type DocText struct {
	Id      int    `orm:"Column(Id)"`
	Md5     string `orm:"size(32);default();column(Md5)"`       //文档的md5，之所以存储md5而不是文档的存档id，主要是因为文档在这之前没录入数据库
	Content string `orm:"size(5000);default();column(Content)"` //文档提取到的文档内容
}

func NewDocText() *DocText {
	return &DocText{}
}

func GetTableDocText() string {
	return getTable("doc_text")
}

//根据md5获取文档摘要，默认获取255个字符长度
//@param            md5str              文档md5
//@param            length              需要获取的长度
//@return           desc                返回摘要内容
func (this *DocText) GetDescByMd5(md5str interface{}, length ...int) (desc string) {
	orm.NewOrm().QueryTable(GetTableDocText()).Filter("Md5", md5str).One(this)
	l := 255
	if len(length) > 0 {
		l = length[0]
	}
	return beego.Substr(this.Content, 0, l)
}

//根据存档表的id获取文档摘要，默认获取255个字符长度
//@param            dsid                document_store的id
//@param            length              需要获取的长度
//@return           desc                返回摘要内容
func (this *DocText) GetDescByDsId(dsid interface{}, length ...int) (desc string) {
	if dsinfo, rows, _ := NewDocument().GetDocStoreByDsId(dsid); rows > 0 {
		return this.GetDescByMd5(dsinfo[0].Md5, length...)
	}
	return
}

//根据文档表的文档id获取文档摘要，默认获取255个字符长度
//@param            did                document_store的id
//@param            length              需要获取的长度
//@return           desc                返回摘要内容
func (this *DocText) GetDescByDid(did interface{}, length ...int) (desc string) {
	var dsid = 0
	if docinfo, rows, _ := NewDocument().GetDocInfoById(did); rows > 0 {
		dsid = docinfo[0].DsId
	}
	if dsinfo, rows, _ := NewDocument().GetDocStoreByDsId(dsid); rows > 0 {
		return this.GetDescByMd5(dsinfo[0].Md5, length...)
	}
	return
}
