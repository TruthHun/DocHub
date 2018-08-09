package models

import "github.com/astaxie/beego/orm"

//不良信息举报
type Report struct {
	Id         int  `orm:"column(Id)"`
	Uid        int  `orm:"column(Uid)"`                   //用户ID
	Did        int  `orm:"column(Did)"`                   //文档ID
	Reason     int  `orm:"column(Reason);default(1)"`     //举报理由：  1、垃圾广告  2、淫秽色情  3、虚假中奖  4、敏感信息  5、人身攻击  6、骚扰他人
	Status     bool `orm:"column(Status);default(false)"` //是否已处理
	TimeCreate int  `orm:"column(TimeCreate)"`            //举报时间
	TimeUpdate int  `orm:"column(TimeUpdate);default(0)"` //举报处理时间
}

func NewReport() *Report {
	return &Report{}
}

func GetTableReport() string {
	return getTable("report")
}

// 不良信息举报多字段唯一索引
func (this *Report) TableUnique() [][]string {
	return [][]string{
		[]string{"Uid", "Did"},
	}
}

//获取未删除的举报数据列表
func (this *Report) Lists(p, listRows int) (params []orm.Params, rows int64, err error) {
	var sql string
	tables := []string{GetTableReport() + " r", GetTableUser() + " u", GetTableDocument() + " d"}
	on := []map[string]string{
		{"r.Did": "d.Id"},
		{"r.Uid": "u.Id"},
	}
	fields := map[string][]string{
		"r": {"*"},
		"d": {"Title"},
		"u": {"Username"},
	}
	if sql, err = LeftJoinSqlBuild(tables, on, fields, p, listRows, []string{"r.Status asc", "r.Id desc"}, nil, "r.Status>-1"); err == nil {
		rows, err = orm.NewOrm().Raw(sql).Values(&params)
	}
	return
}
