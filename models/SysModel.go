package models

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/orm"
)

//系统设置表
type Sys struct {
	Id                int    `orm:"column(Id)"`
	Site              string `orm:"size(100);default();column(Site)"`                    //站点名称
	Tongji            string `orm:"size(2048);default();column(Tongji)"`                 //统计代码
	CntDoc            int    `orm:"default(0);column(CntDoc)"`                           //文档数量
	CntUser           int    `orm:"default(0);column(CntUser)"`                          //注册用户数量
	Reward            int    `orm:"column(Reward);default(5)"`                           //上传一篇未被上传过的文档可以获得的金币奖励
	MaxFile           int    `orm:"column(MaxFile);default(52428800)"`                   //允许上传的文件大小(字节)，默认50M
	Sign              int    `orm:"column(Sign);default(1)"`                             //每日签到获得的金币奖励
	ListRows          int    `orm:"default(10);column(ListRows)"`                        //每页记录数
	Icp               string `orm:"default();column(Icp)"`                               //ICP备案
	DirtyWord         string `orm:"size(2048);default();column(DirtyWord)"`              //不良信息关键字
	TimeExpireRelate  int    `orm:"default(604800);column(TimeExpireRelate)"`            //相关资源过期更新的时间周期，0表示关闭相关资源
	TimeExpireHotspot int    `orm:"default(604800);column(TimeExpireHotspot)"`           //热门文档的时间范围
	MobileOn          bool   `orm:"default(true);column(MobileOn)"`                      //是否开启移动端
	TplMobile         string `orm:"default(default);column(TplMobile)"`                  //手机模板
	TplComputer       string `orm:"default(default);column(TplComputer)"`                //电脑端模板
	TplAdmin          string `orm:"default(default);column(TplAdmin)"`                   //后台模板
	TplEmailReg       string `orm:"size(2048);default();column(TplEmailReg)"`            //会员注册邮件验证码发送模板
	TplEmailFindPwd   string `orm:"size(2048);default();column(TplEmailFindPwd)"`        //会员找回密码邮件验证码发送模板
	DomainPc          string `orm:"size(100);default(dochub.me);column(DomainPc)"`       //PC域名
	DomainMobile      string `orm:"size(100);default(m.dochub.me);column(DomainMobile)"` //移动端域名
	PreviewPage       int    `orm:"default(50);column(PreviewPage)"`                     //文档共预览的最大页数，0表示不限制
	Trends            string `orm:"default();column(Trends)"`                            //文库动态，填写文档的id
	FreeDay           int    `orm:"default(7);column(FreeDay)"`                          //文档免费下载时长。即上次下载扣除金币后多长时间后下载需要收费。时间单位为天
	Question          string `orm:"default(DocHub文库的中文名是？);column(Question)"`            //评论问答问题
	Answer            string `orm:"default(多哈);column(Answer)"`                          //评论问答的问题
	CoinReg           int    `orm:"column(CoinReg);default(10)"`                         //用户注册奖励金币
	Watermark         string `orm:"column(Watermark);default()"`                         //水印文案
	ReportReasons     string `orm:"column(ReportReasons);default();size(2048)"`          //举报原因
	IsCloseReg        bool   `orm:"default(false);column(IsCloseReg)"`                   //是否关闭注册
	StoreType         string `orm:"default(cs-oss);column(StoreType);size(15)"`          //文档存储方式
	CheckRegEmail     bool   `orm:"default(true);column(CheckRegEmail);"`                //是否需要验证注册邮箱，如果需要验证注册邮箱，提要求发送注册验证码
}

func NewSys() *Sys {
	return &Sys{}
}

func GetTableSys() string {
	return getTable("sys")
}

//获取系统配置信息。注意：系统配置信息的记录只有一条，而且id主键为1
//@return           sys         返回的系统信息
//@return           err         错误
func (this *Sys) Get() (sys Sys, err error) {
	sys.Id = 1
	err = orm.NewOrm().Read(&sys)
	return
}

//更新系统全局变量
//@return           sys         返回的系统信息
//@return           err         错误
func (this *Sys) UpdateGlobalConfig() {
	GlobalSys, _ = this.Get()
}

//获取指定指端内容
//@param			field			需要查询的字段
//@return			sys				系统配置信息
func (this *Sys) GetByField(field string) (sys Sys) {
	orm.NewOrm().QueryTable(GetTableSys()).Filter("Id", 1).One(&sys, field)
	return
}

//获取举报原因
func (this *Sys) GetReportReasons() (reasons map[string]string) {
	reasons = make(map[string]string)
	reasonStr := this.GetByField("ReportReasons").ReportReasons
	if slice := strings.Split(reasonStr, "\n"); len(slice) > 0 {
		for _, item := range slice {
			if arr := strings.Split(item, ":"); len(arr) > 1 {
				reasons[arr[0]] = strings.Join(arr[1:], ":")
			}
		}
	}
	return
}

//根据序号获取举报原因
func (this *Sys) GetReportReason(num interface{}) (reason string) {
	reasons := this.GetReportReasons()
	reason, _ = reasons[fmt.Sprintf("%v", num)]
	return
}
