package HomeControllers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/astaxie/beego"

	"DocHub/helper"
)

type StaticController struct {
	beego.Controller
}

// 将除了static之外的静态资源导向到虚拟根目录
func (this *StaticController) Static() {
	splat := strings.TrimPrefix(this.GetString(":splat"), "../")
	if strings.HasPrefix(splat, ".well-known") {
		http.ServeFile(this.Ctx.ResponseWriter, this.Ctx.Request, splat)
		return
	}
	path := filepath.Join(helper.RootPath, splat)
	http.ServeFile(this.Ctx.ResponseWriter, this.Ctx.Request, path)
}
