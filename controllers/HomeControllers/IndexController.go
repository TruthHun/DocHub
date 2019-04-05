package HomeControllers

import (
	"fmt"

	"strings"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego/orm"
)

type IndexController struct {
	BaseController
}

func (this *IndexController) Get() {

	//获取横幅
	this.Data["Banners"], _, _ = models.GetList(models.GetTableBanner(), 1, 100, orm.NewCondition().And("status", 1), "Sort")

	//判断用户是否已登录，如果已登录，则返回用户信息
	if this.IsLogin > 0 {
		users, rows, err := models.NewUser().UserList(1, 1, "", "*", "i.`Id`=?", this.IsLogin)
		if err != nil {
			helper.Logger.Error(err.Error())
		}
		if rows > 0 {
			this.Data["User"] = users[0]
		} else {
			//如果用户不存在，则重置cookie
			this.IsLogin = 0
			this.ResetCookie()
		}
		this.Data["LoginUid"] = this.IsLogin
	} else {
		this.Xsrf()
	}

	modelCate := models.NewCategory()
	//首页分类显示
	_, this.Data["Cates"] = modelCate.GetAll(true)
	//获取最新的文档数据，这里News不是新闻的意思
	this.Data["Latest"], _, _ = models.NewDocument().SimpleList(fmt.Sprintf("d.`Id` in(%v)", strings.Trim(this.Sys.Trends, ",")), 5)
	this.Data["Seo"] = models.NewSeo().GetByPage("PC-Index", "文库首页", "文库首页", "文库首页", this.Sys.Site)
	this.Data["IsHome"] = true
	this.Data["PageId"] = "wenku-index"
	this.TplName = "index.html"
}
