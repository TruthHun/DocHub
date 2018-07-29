package HomeControllers

import (
	"html/template"
	"strings"

	"fmt"
	"time"

	"path/filepath"

	"io/ioutil"

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
	//this.Xsrf()//在有post表单的页面添加，避免每次都生成
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
	this.Sys, _ = models.ModelSys.Get()
	this.Data["Version"] = version
	this.Data["Sys"] = this.Sys
	this.Data["PreviewDomain"] = beego.AppConfig.String("oss::PreviewUrl")
	this.Data["Chanels"] = this.Chanels()
	//单页
	ModelPages := models.Pages{}
	this.Data["Pages"], _, _ = ModelPages.List(beego.AppConfig.DefaultInt("pageslimit", 6), 1)
}

//自定义的文档错误
func (this *BaseController) ErrorDiy(status, redirect, msg interface{}, timewait int) {
	this.TplPrefix = ""
	this.Data["status"] = status
	this.Data["redirect"] = redirect
	this.Data["msg"] = msg
	this.Data["timewait"] = timewait
	this.TplName = "Base/error_diy.html"
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
	secret := beego.AppConfig.String("cookieSecret")
	timestamp, b := this.GetSecureCookie(secret, "uid")
	if b {
		uid, b := this.Ctx.GetSecureCookie(secret+timestamp, "token")
		if b && len(uid) > 0 {
			this.IsLogin = helper.Interface2Int(uid)
		} else {
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
	secret := beego.AppConfig.String("cookieSecret")
	timestamp := fmt.Sprintf("%v", time.Now().Unix())
	expire := 3600 * 24 * 365
	this.Ctx.SetSecureCookie(secret, "uid", timestamp, expire)
	this.Ctx.SetSecureCookie(secret+timestamp, "token", fmt.Sprintf("%v", uid), expire)
}

//404
func (this *BaseController) Error404() {
	referer := this.Ctx.Request.Referer()
	this.Layout = ""
	this.Data["content"] = "Page Not Foud"
	this.Data["code"] = "404"
	this.Data["content_zh"] = "页面被外星人带走了"
	this.Data["Referer"] = referer
	if len(referer) > 0 {
		this.Data["IsReferer"] = true
	}
	this.TplName = "error.html"
}

//501
func (this *BaseController) Error501() {
	this.Layout = ""
	this.Data["code"] = "501"
	this.Data["content"] = "Server Error"
	this.Data["content_zh"] = "服务器被外星人戳炸了"
	this.TplName = "error.html"
}

//数据库错误
func (this *BaseController) ErrorDb() {
	this.Layout = ""
	this.Data["content"] = "Database is now down"
	this.Data["content_zh"] = "数据库别外星人抢走了"
	this.TplName = "error.html"
}

//获取频道
func (this *BaseController) Chanels() []orm.Params {
	key := "chanels"
	cache, err := helper.CacheGet("key")
	if fc, ok := cache.([]orm.Params); ok && err == nil && len(fc) > 0 {
		return fc
	}
	params, rows, _ := models.GetList("category", 1, 6, orm.NewCondition().And("Pid", 0), "Sort")
	if rows > 0 {
		helper.CacheSet(key, params, 10*time.Second)
	}
	return params
}

//校验文档是否已经存在
func (this *BaseController) DocExist() {
	if models.ModelDoc.IsExistByMd5(this.GetString("md5")) > 0 {
		this.ResponseJson(1, "文档存在")
	} else {
		this.ResponseJson(0, "文档不存在")
	}
}

//响应json
func (this *BaseController) ResponseJson(status int, msg string, data ...interface{}) {
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
	page, err := models.ModelPages.One(alias)
	if err != nil {
		helper.Logger.Error(err.Error())
		this.Abort("404")
	}
	if page.Id == 0 || page.Status == false {
		this.Abort("404")
	}
	this.Data["Seo"] = models.ModelSeo.GetByPage("PC-Pages", page.Title, page.Keywords, page.Description, this.Sys.Site)
	page.Vcnt += 1
	models.O.Update(&page, "Vcnt")
	page.Content = models.NewOss().HandleContent(page.Content, true)

	this.Data["Page"] = page
	this.Data["Lists"], _, _ = models.ModelPages.List(20, 1)
	this.Data["PageId"] = "wenku-content"
	this.TplName = "pages.html"
}

//静态文件
func (this *BaseController) StaticFile() {
	splat := this.GetString(":splat")
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(splat)))
	if ok, _ := helper.StaticExt[strings.ToLower(ext)]; ok && ext != ".conf" {
		if b, err := ioutil.ReadFile(splat); err == nil {
			this.Ctx.ResponseWriter.Write(b)
			return
		}
	}
	this.Error404()
}
