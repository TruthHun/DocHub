package AdminControllers

import "github.com/TruthHun/DocHub/models"

type SeoController struct {
	BaseController
}

func (this *SeoController) Get() {
	this.Data["Data"], _, _ = models.GetList(models.TableSeo, 1, 50, nil, "-IsMobile")
	this.Data["IsSeo"] = true
	this.TplName = "index.html"
}
