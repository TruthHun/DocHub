package AdminControllers

import (
	"strings"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
)

type SysController struct {
	BaseController
}

//系统配置管理
func (this *SysController) Get() {
	tab := models.ConfigCate(strings.ToLower(this.GetString("tab")))
	switch tab {
	case models.CONFIG_EMAIL, models.CONFIG_OSS, models.CONFIG_DEPEND, models.CONFIG_ELASTICSEARCH, models.CONFIG_LOGS:
	default:
		tab = "default"
	}
	if this.Ctx.Request.Method == "POST" {
		if tab == "default" {
			var sys models.Sys
			this.ParseForm(&sys)
			if i, err := models.O.Update(&sys); i > 0 && err == nil {
				models.ModelSys.UpdateGlobal() //更新全局变量
			} else {
				if err != nil {
					helper.Logger.Error(err.Error())
				}
				this.ResponseJson(0, "更新失败，可能您未对内容做更改")
			}
		} else {
			modelCfg := new(models.Config)
			for k, v := range this.Ctx.Request.Form {
				modelCfg.UpdateByKey(models.ConfigCate(tab), k, v[0])
			}
			//最后更新全局配置
			modelCfg.UpdateGlobal()
		}
		this.ResponseJson(1, "更新成功")
	} else {
		this.Data["Tab"] = tab
		this.Data["Title"] = "系统管理"
		this.Data["IsSys"] = true
		if tab == "default" {
			this.Data["Sys"], _ = models.ModelSys.Get()
		} else {
			this.Data["Configs"] = new(models.Config).All()
		}
		this.TplName = "index.html"
	}

}
