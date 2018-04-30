package HomeControllers

import (
	"fmt"

	"strings"

	"time"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego/orm"
)

type IndexController struct {
	BaseController
}

func (this *IndexController) Get() {

	//获取横幅
	this.Data["Banners"], _, _ = models.GetList("banner", 1, 100, orm.NewCondition().And("status", 1), "Sort")

	//判断用户是否已登录，如果已登录，则返回用户信息
	if this.IsLogin > 0 {
		ModelUser := models.User{}
		users, rows, err := ModelUser.UserList(1, 1, "", "*", "i.`Id`=?", this.IsLogin)
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

	//首页分类显示
	this.Data["Cates"] = this.GetHomeCates(this.Sys.HomeCates)
	//获取最新的文档数据，这里News不是新闻的意思
	this.Data["Latest"], _, _ = models.ModelDoc.SimpleList(fmt.Sprintf("d.`Id` in(%v)", strings.Trim(this.Sys.Trends, ",")), 5)
	this.Data["Seo"] = models.ModelSeo.GetByPage("PC-Index", "文库首页", "文库首页", "文库首页", this.Sys.Site)
	this.Data["IsHome"] = true
	this.Data["PageId"] = "wenku-index"
	this.TplName = "index.html"
}

//获取首页分类缓存
//@param            catesId         string          频道ids
//@return           map[int]interface{}             分类数据
func (this *IndexController) GetHomeCates(catesId string) interface{} {
	key := "home_cates"
	cache, err := helper.CacheGet("home_cates")
	if fc, ok := cache.([]orm.Params); ok && len(fc) > 0 && err == nil {
		return fc
	}
	catesIdSlice := strings.Split(catesId, ",")
	chanels, _, err := models.GetList("category", 1, 23, orm.NewCondition().And("Id__in", catesIdSlice), "sort")
	for _, v := range catesIdSlice {
		for _, chanel := range chanels {
			if fmt.Sprintf("%v", v) == fmt.Sprintf("%v", chanel["Id"]) {
				chanel["child"], _, _ = models.GetList("category", 1, 8, orm.NewCondition().And("Pid", chanel["Id"]), "sort")
			}
		}
	}
	if len(chanels) > 0 {
		err = helper.CacheSet(key, chanels, 10*time.Second)
	}
	return chanels
}
