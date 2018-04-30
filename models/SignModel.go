package models

//签到表表结构
type Sign struct {
	Id   int    `orm:"column(Id);"`  //主键
	Uid  int    `orm:"column(Uid)"`  //用户ID
	Date string `orm:"column(Date)"` //签到日期，格式如20170322
}

// 签到表多字段唯一索引
func (s *Sign) TableUnique() [][]string {
	return [][]string{
		[]string{"Uid", "Date"},
	}
}
