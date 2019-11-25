package tplfunc

import (
	"github.com/astaxie/beego"

	"DocHub/models"
)

func init() {
	beego.AddFuncMap("isOpenLdap", models.IsOpenLdap)
}
