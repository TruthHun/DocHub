package AdminControllers

import "github.com/TruthHun/DocHub/models"

type SeoController struct {
	BaseController
}

func (this *SeoController) Get() {
	this.Data["Data"], _, _ = models.GetList(models.GetTableSeo(), 1, 50, nil, "-IsMobile")
	this.Data["IsSeo"] = true
	this.TplName = "index.html"
}

func (this *SeoController) UpdateSitemap() {
	go models.NewSeo().BuildSitemap()
	this.ResponseJson(true, "Sitemap更新已提交后台执行，请耐心等待")
}
