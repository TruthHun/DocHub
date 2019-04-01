package AdminControllers

import (
	"net/http"
	"time"

	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego/orm"
)

type SingleController struct {
	BaseController
}

//单页列表
func (this *SingleController) Get() {
	this.Data["IsSingle"] = true
	this.Data["Lists"], _, _ = models.NewPages().List(1000)
	this.TplName = "index.html"
}

//单页编辑，只编辑文本内容
func (this *SingleController) Edit() {
	var (
		page models.Pages
		cs   *models.CloudStore
		err  error
	)

	cs, err = models.NewCloudStore(false)
	if err != nil {
		this.CustomAbort(http.StatusInternalServerError, err.Error())
	}

	this.Data["IsSingle"] = true
	alias := this.GetString(":alias")

	if this.Ctx.Request.Method == "POST" {
		this.ParseForm(&page)
		page.TimeCreate = int(time.Now().Unix())
		page.Content = cs.ImageWithoutDomain(page.Content)
		_, err := orm.NewOrm().Update(&page)
		if err != nil {
			this.ResponseJson(false, err.Error())
		}
		this.ResponseJson(true, "更新成功")
	} else {
		page, _ = models.NewPages().One(alias)
		page.Content = cs.ImageWithDomain(page.Content)
		this.Data["Data"] = page
		this.TplName = "edit.html"
	}
}

//删除单页
func (this *SingleController) Del() {
	id, _ := this.GetInt("id")
	var page = models.Pages{Id: id}
	err := orm.NewOrm().Read(&page)
	if err != nil {
		this.ResponseJson(false, err.Error())
	}
	if _, err = orm.NewOrm().QueryTable(models.GetTablePages()).Filter("Id", page.Id).Delete(); err != nil {
		this.ResponseJson(false, err.Error())
	}

	cs, _ := models.NewCloudStore(false)
	go cs.DeleteImageFromHtml(page.Content)

	this.ResponseJson(true, "删除成功")
}
