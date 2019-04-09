package HomeControllers

import (
	"html/template"
	"strings"

	"fmt"
	"time"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type Output struct {
	status int
	msg    string
}

type BaseController struct {
	beego.Controller
	TplTheme  string //模板主题
	TplStatic string //模板静态文件
	IsLogin   int    //用户是否已登录
	Sys       models.Sys
	Out       Output
}

//初始化函数
func (this *BaseController) Prepare() {
	ctrl, _ := this.GetControllerAndAction()
	ctrl = strings.TrimSuffix(ctrl, "Controller")

	//设置默认模板
	this.TplTheme = "default"
	this.TplPrefix = "Home/" + this.TplTheme + "/" + ctrl + "/"
	this.Layout = "Home/" + this.TplTheme + "/layout.html"

	//防止跨站攻击
	//检测用户是否已经在cookie存在登录
	this.checkCookieLogin()

	//初始化
	this.Data["LoginUid"] = this.IsLogin
	//当前模板静态文件
	this.Data["TplStatic"] = "/static/Home/" + this.TplTheme

	version := helper.VERSION
	if helper.Debug { //debug模式下，每次更新js
		version = fmt.Sprintf("%v.%v", version, time.Now().Unix())
	}
	this.Sys, _ = models.NewSys().Get()
	this.Data["Version"] = version
	this.Data["Sys"] = this.Sys
	this.Data["Chanels"] = models.NewCategory().GetByPid(0, true)
	this.Data["Pages"], _, _ = models.NewPages().List(beego.AppConfig.DefaultInt("pageslimit", 6), 1)
	this.Data["AdminId"] = helper.Interface2Int(this.GetSession("AdminId"))
	this.Data["CopyrightDate"] = time.Now().Format("2006")

	this.Data["PreviewDomain"] = ""

	if cs, err := models.NewCloudStore(false); err == nil {
		this.Data["PreviewDomain"] = cs.GetPublicDomain()
	} else {
		helper.Logger.Error(err.Error())
	}

}

//是否已经登录，如果已登录，则返回用户的id
func (this *BaseController) CheckLogin() int {
	uid := this.GetSession("uid")
	if uid != nil {
		id, ok := uid.(int)
		if ok && id > 0 {
			return id
		}
	}
	return 0
}

//防止跨站攻击，在有表单的控制器放大中调用，不要直接在base控制器中调用，因为用户每访问一个页面都重新刷新cookie了
func (this *BaseController) Xsrf() {
	//使用的时候，直接在模板表单添加{{.xsrfdata}}
	this.Data["xsrfdata"] = template.HTML(this.XSRFFormHTML())
}

//检测用户登录的cookie是否存在
func (this *BaseController) checkCookieLogin() {
	secret := beego.AppConfig.DefaultString("CookieSecret", helper.DEFAULT_COOKIE_SECRET)
	timestamp, ok := this.GetSecureCookie(secret, "uid")
	if !ok {
		return
	}
	uid, ok := this.Ctx.GetSecureCookie(secret+timestamp, "token")
	if !ok || len(uid) == 0 {
		this.ResetCookie()
	}

	if this.IsLogin = helper.Interface2Int(uid); this.IsLogin > 0 {
		if info := models.NewUser().UserInfo(this.IsLogin); info.Status == false {
			//被封禁的账号，重置cookie
			this.ResetCookie()
		}
	}
}

//重置cookie
func (this *BaseController) ResetCookie() {
	this.Ctx.SetCookie("uid", "")
	this.Ctx.SetCookie("token", "")
}

//设置用户登录的cookie，其实uid是时间戳的加密，而token才是真正的uid
//@param            uid         interface{}         用户UID
func (this *BaseController) SetCookieLogin(uid interface{}) {
	secret := beego.AppConfig.DefaultString("CookieSecret", helper.DEFAULT_COOKIE_SECRET)
	timestamp := fmt.Sprintf("%v", time.Now().Unix())
	expire := 3600 * 24 * 365
	this.Ctx.SetSecureCookie(secret, "uid", timestamp, expire)
	this.Ctx.SetSecureCookie(secret+timestamp, "token", fmt.Sprintf("%v", uid), expire)
}

//校验文档是否已经存在
func (this *BaseController) DocExist() {
	if models.NewDocument().IsExistByMd5(this.GetString("md5")) > 0 {
		this.ResponseJson(true, "文档存在")
	}
	this.ResponseJson(false, "文档不存在")
}

//响应json
func (this *BaseController) ResponseJson(isSuccess bool, msg string, data ...interface{}) {
	status := 0
	if isSuccess {
		status = 1
	}
	ret := map[string]interface{}{"status": status, "msg": msg}
	if len(data) > 0 {
		ret["data"] = data[0]
	}
	this.Data["json"] = ret
	this.ServeJSON()
	this.StopRun()
}

//单页
func (this *BaseController) Pages() {
	alias := this.GetString(":page")
	page, err := models.NewPages().One(alias)
	if err != nil {
		helper.Logger.Error(err.Error())
		this.Abort("404")
	}
	if page.Id == 0 || page.Status == false {
		this.Abort("404")
	}
	this.Data["Seo"] = models.NewSeo().GetByPage("PC-Pages", page.Title, page.Keywords, page.Description, this.Sys.Site)
	page.Vcnt += 1
	orm.NewOrm().Update(&page, "Vcnt")
	cs, _ := models.NewCloudStore(false)
	page.Content = cs.ImageWithDomain(page.Content)

	this.Data["Page"] = page
	this.Data["Lists"], _, _ = models.NewPages().List(20, 1)
	this.Data["PageId"] = "wenku-content"
	this.TplName = "pages.html"
}
