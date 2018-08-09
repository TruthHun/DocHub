package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

//金币变更日志
type CoinLog struct {
	Id         int    `orm:"column(Id)"`
	Uid        int    `orm:"column(Uid);index"`               //用户id
	Coin       int    `orm:"column(Coin);default(0)"`         //金币变更，正表示加，负表示减
	Log        string `orm:"column(Log);size(512);default()"` //记录说明
	TimeCreate int    `orm:"column(TimeCreate)"`              //记录变更时间
}

func NewCoinLog() *CoinLog {
	return &CoinLog{}
}

func GetTableCoinLog() string {
	return getTable("coin_log")
}

//记录金币记录变更情况，会自动对用户的金币做变更
//@param                log             日志对象
//@return               err             错误，nil表示true，否则表示false
func (this *CoinLog) LogRecord(log CoinLog) (err error) {
	log.TimeCreate = int(time.Now().Unix())
	_, err = orm.NewOrm().Insert(&log)
	return
}
