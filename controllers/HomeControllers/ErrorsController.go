package HomeControllers

import "github.com/astaxie/beego"

type ErrorsController struct {
	beego.Controller
}

//404
func (this *ErrorsController) Error404() {
	referer := this.Ctx.Request.Referer()
	this.Layout = ""
	this.Data["content"] = "Page Not Foud"
	this.Data["code"] = "404"
	this.Data["content_zh"] = "页面不存在"
	this.Data["Referer"] = referer
	if len(referer) > 0 {
		this.Data["IsReferer"] = true
	}
	this.TplName = "error.html"
}

//501
func (this *ErrorsController) Error501() {
	this.Layout = ""
	this.Data["code"] = "501"
	this.Data["content"] = "Server Error"
	this.Data["content_zh"] = "服务内部错误"
	this.TplName = "error.html"
}

//数据库错误
func (this *ErrorsController) ErrorDb() {
	this.Layout = ""
	this.Data["content"] = "Database is now down"
	this.Data["content_zh"] = "数据库访问失败"
	this.TplName = "error.html"
}
