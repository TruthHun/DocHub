package AdminControllers

import (
	"html/template"

	"time"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
)

type LoginController struct {
	BaseController
}

//重置prepare方法，移除模板继承
func (this *LoginController) Prepare() {
	this.EnableXSRF = false
	//设置默认模板
	TplTheme := "default"
	this.TplPrefix = "Admin/" + TplTheme + "/Login/"
	this.Layout = ""
	//当前模板静态文件
	this.Data["TplStatic"] = "/static/Admin/" + TplTheme
}

//登录后台
func (this *LoginController) Login() {
	this.EnableXSRF = true
	this.Data["Sys"], _ = models.ModelSys.Get()
	if this.Ctx.Request.Method == "GET" {
		this.Xsrf()
		this.TplName = "index.html"
	} else {
		var (
			msg   string = "登录失败，用户名或密码不正确"
			admin models.Admin
		)
		this.ParseForm(&admin)
		if admin, err := models.ModelAdmin.Login(admin.Username, admin.Password, admin.Code); err == nil && admin.Id > 0 {
			this.SetSession("AdminId", admin.Id)
			this.ResponseJson(1, "登录成功")
		} else {
			this.ResponseJson(0, msg)
		}
	}
}

//更新登录密码
func (this *LoginController) UpdatePwd() {
	if helper.Interface2Int(this.GetSession("AdminId")) > 0 {
		PwdOld := this.GetString("password_old")
		PwdNew := this.GetString("password_new")
		PwdEnsure := this.GetString("password_ensure")
		if PwdOld == PwdNew || PwdNew != PwdEnsure {
			this.ResponseJson(0, "新密码不能与原密码相同，且确认密码必须与新密码一致")
		} else {
			var admin = models.Admin{Password: helper.MyMD5(PwdOld)}
			if models.O.Read(&admin, "Password"); admin.Id > 0 {
				admin.Password = helper.MyMD5(PwdNew)
				if rows, err := models.O.Update(&admin); rows > 0 {
					this.ResponseJson(1, "密码更新成功")
				} else {
					this.ResponseJson(0, "密码更新失败："+err.Error())
				}

			} else {
				this.ResponseJson(0, "原密码不正确")
			}
		}
	} else {
		this.Error404()
	}
}

//退出登录
func (this *LoginController) Logout() {
	this.DelSession("AdminId")
	this.Redirect("/admin/login?t="+time.Now().String(), 302)
}

//防止跨站攻击，在有表单的控制器中调用
func (this *LoginController) Xsrf() {
	//使用的时候，直接在模板表单添加{{.xsrfdata}}
	this.Data["xsrfdata"] = template.HTML(this.XSRFFormHTML())
}
