package models

//系统设置表
type Sys struct {
	Id                int    `orm:"column(Id)"`
	Site              string `orm:"size(100);default();column(Site)"`                   //站点名称
	Tongji            string `orm:"size(2048);default();column(Tongji)"`                //统计代码
	CntDoc            int    `orm:"default(0);column(CntDoc)"`                          //文档数量
	CntUser           int    `orm:"default(0);column(CntUser)"`                         //注册用户数量
	Reward            int    `orm:"column(Reward);default(5)"`                          //上传一篇未被上传过的文档可以获得的金币奖励
	Price             int    `orm:"default(1);column(Price)"`                           //会员下载一篇文档需要的最大金币【会员在上传分享文档时允许设置的最大金币上限】
	Sign              int    `orm:"column(Sign);default(1)"`                            //每日签到获得的金币奖励
	ListRows          int    `orm:"default(10);column(ListRows)"`                       //每页记录数
	Statement         string `orm:"size(512);default();column(Statement)"`              //站点声明
	Icp               string `orm:"default();column(Icp)"`                              //ICP备案
	DirtyWord         string `orm:"size(2048);default();column(DirtyWord)"`             //不良信息关键字
	TimeExpireRelate  int    `orm:"default(604800);column(TimeExpireRelate)"`           //相关资源过期更新的时间周期，0表示关闭相关资源
	TimeExpireHotspot int    `orm:"default(604800);column(TimeExpireHotspot)"`          //热门文档的时间范围
	TplMobile         string `orm:"default(default);column(TplMobile)"`                 //手机模板
	TplComputer       string `orm:"default(default);column(TplComputer)"`               //电脑端模板
	TplAdmin          string `orm:"default(default);column(TplAdmin)"`                  //后台模板
	TplEmailReg       string `orm:"size(2048);default();column(TplEmailReg)"`           //会员注册邮件验证码发送模板
	TplEmailFindPwd   string `orm:"size(2048);default();column(TplEmailFindPwd)"`       //会员找回密码邮件验证码发送模板
	DomainPc          string `orm:"size(100);default(wenku.it);column(DomainPc)"`       //PC域名
	DomainMobile      string `orm:"size(100);default(m.wenku.it);column(DomainMobile)"` //移动端域名
	PreviewPage       int    `orm:"default(50);column(PreviewPage)"`                    //文档共预览的最大页数，0表示不限制
	Trends            string `orm:"default();column(Trends)"`                           //文库动态，填写文档的id
	HomeCates         string `orm:"default();column(HomeCates);size(50)"`               //首页分类，填写频道ids
	IsCloseReg        bool   `orm:"default(false);column(IsCloseReg)"`                  //是否关闭注册
	FreeDay           int    `orm:"default(7);column(FreeDay)"`                         //文档免费下载时长。即上次下载扣除金币后多长时间后下载需要收费。时间单位为天
	Question          string `orm:"default(IT文库的网址是多少？);column(Question)"`              //评论问答问题
	Answer            string `orm:"default(wenku.it);column(Answer)"`                   //评论问答的问题
}

//获取系统配置信息。注意：系统配置信息的记录只有一条，而且id主键为1
//@return           sys         返回的系统信息
//@return           err         错误
func (this *Sys) Get() (sys Sys, err error) {
	sys.Id = 1
	err = O.Read(&sys)
	return
}

//更新系统全局变量
//@return           sys         返回的系统信息
//@return           err         错误
func (this *Sys) UpdateGlobal() {
	GlobalSys, _ = this.Get()
}
