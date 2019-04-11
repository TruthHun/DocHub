package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

//签到表表结构
type Sign struct {
	Id   int    `orm:"column(Id);"`          //主键
	Uid  int    `orm:"column(Uid)"`          //用户ID
	Date string `orm:"column(Date);size(8)"` //签到日期，格式如20170322
}

func NewSign() *Sign {
	return &Sign{}
}

func GetTableSign() string {
	return getTable("sign")
}

// 签到表多字段唯一索引
func (s *Sign) TableUnique() [][]string {
	return [][]string{
		[]string{"Uid", "Date"},
	}
}

//检测用户今天是否已签到
func (this *Sign) DoesSign(uid int) bool {
	if _, rows, err := GetList(GetTableSign(), 1, 1, orm.NewCondition().And("Date", time.Now().Format("20060102")).And("Uid", uid)); err == nil && rows > 0 {
		return true
	}
	return false
}
