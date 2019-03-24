package AdminControllers

import (
	"html/template"

	"time"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego/orm"
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
	this.AdminId = helper.Interface2Int(this.GetSession("AdminId"))
}

//登录后台
func (this *LoginController) Login() {
	this.EnableXSRF = true
	this.Data["Sys"], _ = models.NewSys().Get()
	if this.Ctx.Request.Method == "GET" {
		this.Xsrf()
		this.TplName = "index.html"
	} else {
		var (
			msg   string = "登录失败，用户名或密码不正确"
			admin models.Admin
		)
		this.ParseForm(&admin)
		if admin, err := models.NewAdmin().Login(admin.Username, admin.Password, admin.Code); err == nil && admin.Id > 0 {
			this.SetSession("AdminId", admin.Id)
			this.ResponseJson(true, "登录成功")
		} else {
			this.ResponseJson(false, msg)
		}
	}
}

//更新登录密码
func (this *LoginController) UpdatePwd() {
	if this.AdminId > 0 {
		PwdOld := this.GetString("password_old")
		PwdNew := this.GetString("password_new")
		PwdEnsure := this.GetString("password_ensure")
		if PwdOld == PwdNew || PwdNew != PwdEnsure {
			this.ResponseJson(false, "新密码不能与原密码相同，且确认密码必须与新密码一致")
		} else {
			var admin = models.Admin{Password: helper.MD5Crypt(PwdOld)}
			if orm.NewOrm().Read(&admin, "Password"); admin.Id > 0 {
				admin.Password = helper.MD5Crypt(PwdNew)
				if rows, err := orm.NewOrm().Update(&admin); rows > 0 {
					this.ResponseJson(true, "密码更新成功")
				} else {
					this.ResponseJson(false, "密码更新失败："+err.Error())
				}

			} else {
				this.ResponseJson(false, "原密码不正确")
			}
		}
	} else {
		this.Error404()
	}
}

//更新管理员信息
func (this *LoginController) UpdateAdmin() {
	if this.AdminId > 0 {
		code := this.GetString("code")
		if code == "" {
			this.ResponseJson(false, "登录验证码不能为空")
		}

		username := this.GetString("username")
		if username == "" {
			this.ResponseJson(false, "登录用户名不能为空")
		}

		email := this.GetString("email")

		admin := models.Admin{
			Id:       this.AdminId,
			Username: username,
			Code:     code,
			Email:    email,
		}

		if _, err := orm.NewOrm().Update(&admin, "Code", "Email", "Username"); err == nil {
			this.ResponseJson(true, "资料更新成功")
		} else {
			this.ResponseJson(false, "资料更新失败："+err.Error())
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
