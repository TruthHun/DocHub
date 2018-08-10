package AdminControllers

import "github.com/TruthHun/DocHub/models"

type ReportController struct {
	BaseController
}

//查看举报处理情况
func (this *ReportController) Get() {
	this.Data["Data"], _, _ = models.NewReport().Lists(1, 1000)
	this.Data["IsReport"] = true
	this.TplName = "index.html"
}
