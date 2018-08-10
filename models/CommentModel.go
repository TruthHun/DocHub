package models

import "github.com/astaxie/beego/orm"

//文档评分记录表
type DocumentComment struct {
	Id         int    `orm:"column(Id)"`
	Did        int    `orm:"column(Did);index"`                   //文档ID
	Uid        int    `orm:"column(Uid);"`                        //评分的用户Id
	Score      int    `orm:"column(Score);default(30000)"`        //评分分数
	Content    string `orm:"column(Content);size(256);default()"` //评论内容
	TimeCreate int    `orm:"column(TimeCreate);default(0)"`       //评论发表时间
	Status     bool   `orm:"column(Status);default(true)"`        //评论是否正常
}

func NewDocumentComment() *DocumentComment {
	return &DocumentComment{}
}

func GetTableDocumentComment() string {
	return getTable("document_comment")
}

// 文档评分记录表多字段唯一索引
func (this *DocumentComment) TableUnique() [][]string {
	return [][]string{
		[]string{"Did", "Uid"},
	}
}

//获取文档评论列表
//@param            did             文档ID
//@param            p               页码
//@param            listRows        每页记录数
//@return           params          返回的数据
//@return           rows            返回的数据记录数
//@return           err             返回错误
func (this *DocumentComment) GetCommentList(did, p, listRows int) (params []orm.Params, rows int64, err error) {
	tables := []string{GetTableDocumentComment() + " c", GetTableUser() + " u"}
	on := []map[string]string{
		{"c.Uid": "u.Id"},
	}
	fields := map[string][]string{
		"c": GetFields(NewDocumentComment()),
		"u": {"Username", "Avatar"},
	}
	if sql, err := LeftJoinSqlBuild(tables, on, fields, p, listRows, []string{"c.Id desc"}, nil, "c.Did=?"); err == nil {
		rows, err = orm.NewOrm().Raw(sql, did).Values(&params)
	}
	return params, rows, err
}

//根据文档ID删除文档评论
//@param                ids             文档id
//@return               err             错误，nil表示删除成功
func (this *DocumentComment) DelCommentByDocId(ids ...interface{}) (err error) {
	if len(ids) > 0 {
		_, err = orm.NewOrm().QueryTable(GetTableDocumentComment()).Filter("Did__in", ids...).Delete()
	}
	return err
}
