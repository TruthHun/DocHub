package HomeControllers

import (
	"strings"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego"
)

type InstallController struct {
	beego.Controller
}

type installForm struct {
	Host     string `form:"host"`
	Port     int    `form:"port"`
	Database string `form:"database"`
	Prefix   string `form:"prefix"` //表前缀
	Username string `form:"username"`
	Password string `form:"password"`
	Charset  string `form:"charset"`
}

//安装程序
func (this *InstallController) Install() {
	if helper.IsInstalled { //如果程序已安装，在访问该路由是，跳转到首页
		this.Redirect("/", 302)
		return
	}
	if this.Ctx.Request.Method == "GET" {
		this.TplName = "Install/install.html"
		return
	}

	var form installForm
	var respData = map[string]interface{}{"status": 0}

	this.ParseForm(&form)

	if form.Charset == "" {
		respData["msg"] = "请选择您创建的数据库字符编码！！！请选择您创建的数据库字符编码！！！请选择您创建的数据库字符编码！！！"
	} else if form.Database == "" || form.Host == "" || form.Username == "" || form.Port <= 0 {
		respData["msg"] = "所有必填输入项均不能为空，请按要求进行填写"
	} else {
		if err := models.CheckDatabaseIsExist(form.Host, form.Port, form.Username, form.Password, form.Database); err != nil {
			respData["msg"] = "数据库连接失败：" + err.Error()
		} else {
			//生成app.conf配置项
			if form.Prefix = strings.TrimSpace(form.Prefix); form.Prefix == "" {
				form.Prefix = "hc_"
			}
			if err = helper.GenerateAppConf(form.Host, form.Port, form.Username, form.Password, form.Database, form.Prefix, form.Charset); err == nil {
				//重载app.conf
				if err = beego.LoadAppConfig("ini", "conf/app.conf"); err == nil {
					//初始化数据库
					models.Init()
					//将安装设置为true
					helper.IsInstalled = true
					respData["msg"] = "程序安装成功"
					respData["status"] = 1
				} else {
					respData["msg"] = "重载配置文件失败：" + err.Error()
				}
			} else {
				respData["msg"] = "生成配置文件失败：" + err.Error()
			}
		}
	}

	this.Data["json"] = respData
	this.ServeJSON()
}
