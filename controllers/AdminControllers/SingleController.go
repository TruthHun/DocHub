package AdminControllers

import (
	"time"

	"github.com/TruthHun/DocHub/models"
)

type SingleController struct {
	BaseController
}

//单页列表
func (this *SingleController) Get() {
	this.Data["IsSingle"] = true
	this.Data["Lists"], _, _ = models.ModelPages.List(1000)
	this.TplName = "index.html"
}

//单页编辑，只编辑文本内容
func (this *SingleController) Edit() {
	var page models.Pages
	this.Data["IsSingle"] = true
	alias := this.GetString(":alias")
	if this.Ctx.Request.Method == "POST" {
		this.ParseForm(&page)
		page.TimeCreate = int(time.Now().Unix())
		page.Content = models.ModelOss.HandleContent(page.Content, false)
		if rows, err := models.O.Update(&page); err == nil && rows > 0 {
			this.ResponseJson(1, "更新成功")
		} else if err != nil {
			this.ResponseJson(0, err.Error())
		} else {
			this.ResponseJson(0, "更新失败，可能您未对内容做更改")
		}
	} else {
		page, _ = models.ModelPages.One(alias)
		page.Content = models.ModelOss.HandleContent(page.Content, true)
		this.Data["Data"] = page
		this.TplName = "edit.html"
	}
}

//删除单页
func (this *SingleController) Del() {
	id, _ := this.GetInt("id")
	var page = models.Pages{Id: id}
	if err := models.O.Read(&page); err != nil {
		this.ResponseJson(0, err.Error())
	} else {
		if _, err = models.O.QueryTable(models.TablePages).Filter("Id", page.Id).Delete(); err != nil {
			this.ResponseJson(0, err.Error())
		} else {
			go models.ModelOss.DelByHtmlPics(page.Content)
			this.ResponseJson(1, "删除成功")
		}
	}
}
