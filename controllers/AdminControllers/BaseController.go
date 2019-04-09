package AdminControllers

import (
	"strings"

	"github.com/TruthHun/DocHub/models"

	"time"

	"fmt"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego"
)

type BaseController struct {
	beego.Controller
	TplTheme  string //模板主题
	TplStatic string //模板静态文件
	AdminId   int    //管理员是否已登录，如果已登录，则管理员ID大于0
	Sys       models.Sys
}

//初始化函数
func (this *BaseController) Prepare() {
	var ok bool
	this.Sys, _ = models.NewSys().Get()
	//检测是否已登录，未登录则跳转到登录页
	AdminId := this.GetSession("AdminId")
	this.AdminId, ok = AdminId.(int)
	this.Data["Admin"], _ = models.NewAdmin().GetById(this.AdminId)
	if !ok || this.AdminId == 0 {
		this.Redirect("/admin/login", 302)
		return
	}

	version := helper.VERSION
	if helper.Debug {
		version = fmt.Sprintf("%v.%v", version, time.Now().Unix())
	}
	this.Data["Version"] = version
	//后台关闭XSRF功能
	this.EnableXSRF = false
	ctrl, _ := this.GetControllerAndAction()
	ctrl = strings.TrimSuffix(ctrl, "Controller")
	//设置默认模板
	this.TplTheme = "default"
	this.TplPrefix = "Admin/" + this.TplTheme + "/" + ctrl + "/"
	this.Layout = "Admin/" + this.TplTheme + "/layout.html"
	//当前模板静态文件
	this.Data["TplStatic"] = "/static/Admin/" + this.TplTheme
	//this.Data["PreviewDomain"] = beego.AppConfig.String("oss::PreviewUrl")
	if cs, err := models.NewCloudStore(false); err == nil {
		this.Data["PreviewDomain"] = cs.GetPublicDomain()
	} else {
		helper.Logger.Error(err.Error())
		this.Data["PreviewDomain"] = ""
	}
	this.Data["Sys"] = this.Sys
	this.Data["Title"] = "文库系统管理后台"
	this.Data["Lang"] = "zh-CN"
}

//自定义的文档错误
func (this *BaseController) ErrorDiy(status, redirect, msg interface{}, timewait int) {
	this.Data["status"] = status
	this.Data["redirect"] = redirect
	this.Data["msg"] = msg
	this.Data["timewait"] = timewait
	this.TplName = "error_diy.html"
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

//404
func (this *BaseController) Error404() {
	this.Layout = ""
	this.Data["content"] = "Page Not Foud"
	this.Data["code"] = "404"
	this.Data["content_zh"] = "页面被外星人带走了"
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
	this.Data["content_zh"] = "数据库被外星人抢走了"
	this.TplName = "error.html"
}

//更新内容
func (this *BaseController) Update() {
	id := strings.Split(this.GetString("id"), ",")
	i, err := models.UpdateByIds(this.GetString("table"), this.GetString("field"), this.GetString("value"), id)
	ret := map[string]interface{}{"status": 0, "msg": "更新失败，可能您未对内容作更改"}
	if i > 0 && err == nil {
		ret["status"] = 1
		ret["msg"] = "更新成功"
	}
	if err != nil {
		ret["msg"] = err.Error()
	}
	this.Data["json"] = ret
	this.ServeJSON()
}

//删除内容
func (this *BaseController) Del() {
	id := strings.Split(this.GetString("id"), ",")
	i, err := models.DelByIds(this.GetString("table"), id)
	ret := map[string]interface{}{"status": 0, "msg": "删除失败，可能您要删除的内容已经不存在"}
	if i > 0 && err == nil {
		ret["status"] = 1
		ret["msg"] = "删除成功"
	}
	if err != nil {
		ret["msg"] = err.Error()
	}
	this.Data["json"] = ret
	this.ServeJSON()
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

//响应json
func (this *BaseController) Response(data map[string]interface{}) {
	this.Data["json"] = data
	this.ServeJSON()
	this.StopRun()
}
