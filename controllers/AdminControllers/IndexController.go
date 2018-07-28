package AdminControllers

import "github.com/astaxie/beego"

type IndexController struct {
	BaseController
}

func (this *IndexController) Get() {
	this.Data["BeegoVersion"] = beego.VERSION
	this.Data["IsIndex"] = true
	this.TplName = "index.html"
}
