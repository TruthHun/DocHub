package AdminControllers

type IndexController struct {
	BaseController
}

func (this *IndexController) Get() {
	this.Data["IsIndex"] = true
	this.TplName = "index.html"
}
