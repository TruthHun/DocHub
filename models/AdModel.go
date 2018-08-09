package models

//广告位表【程序开发完成之后再设立广告位】
type AdPosition struct {
	Id       int    `orm:"column(Id)"`                      //主键
	Title    string `orm:"column(Title);default()"`         //广告位名称
	Alias    string `orm:"column(Alias);default();unique"`  //广告位别称
	IsMobile bool   `orm:"column(IsMobile);default(false)"` //是否是手机广告位
	Width    string `orm:"column(Width);default()"`         //广告位的最大宽度
}

//广告表
type Ad struct {
	Id         int    `orm:"column(Id)"`
	Title      string `orm:"column(Title);size(100);default()"` //广告名称
	Pid        int    `orm:"column(Pid);default(0);index"`      //广告位id
	Code       string `orm:"column(Code);default();size(1024)"` //广告代码
	Status     bool   `orm:"column(Status);default(true)"`      //广告状态，0表示广告关闭，否则为开启
	TimeStart  int    `orm:"column(TimeStart);default(0)"`      //广告开始时间
	TimeEnd    int    `orm:"column(TimeEnd);default(0)"`        //广告截止时间
	TimeCreate int    `orm:"column(TimeCreate);default(0)"`     //广告添加时间
}

func NewAdPosition() *AdPosition {
	return &AdPosition{}
}

func NewAd() *Ad {
	return &Ad{}
}

func GetTableAdPosition() string {
	return getTable("ad_position")
}
