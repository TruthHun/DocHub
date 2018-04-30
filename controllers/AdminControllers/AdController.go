package AdminControllers

type AdController struct {
	BaseController
}

func (this *AdController) Get() {
	this.Data["IsAd"] = true
	this.TplName = "index.html"
}
