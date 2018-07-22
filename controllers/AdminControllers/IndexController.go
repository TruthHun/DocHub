package AdminControllers

type IndexController struct {
	BaseController
}

func (this *IndexController) Get() {
	this.Data["Ver"] = "v1.1"
	this.Data["IsIndex"] = true
	this.TplName = "index.html"
}
