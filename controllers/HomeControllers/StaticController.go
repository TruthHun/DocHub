package HomeControllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/TruthHun/DocHub/models"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego"
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
	ext := strings.ToLower(filepath.Ext(splat))

	_, ok := helper.AllowedUploadDocsExt[ext]
	if !ok {
		if store, err := models.NewLocalStore(); err == nil {
			real := store.RealPath(splat)
			if _, err = os.Stat(real); err == nil {
				realHeader, err := ioutil.ReadFile(real + store.HeaderExt)
				if err == nil {
					var header map[string]string
					json.Unmarshal(realHeader, &header)
					for k, v := range header {
						this.Ctx.ResponseWriter.Header().Add(k, v)
					}
				}
				if ext == ".svg" {
					this.Ctx.ResponseWriter.Header().Add("Content-Type", "image/svg+xml")
				}
				b, err := ioutil.ReadFile(real)
				if err == nil {
					this.Ctx.ResponseWriter.Write(b)
					return
				}
			}
		}
	}
	http.ServeFile(this.Ctx.ResponseWriter, this.Ctx.Request, path)
}
