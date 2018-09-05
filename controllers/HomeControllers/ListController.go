package HomeControllers

import (
	"strings"

	"fmt"

	"time"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/helper/conv"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego/orm"
)

type ListController struct {
	BaseController
}

func (this *ListController) Get() {

	var (
		pid, cid    int // parent id && category id
		p, listRows = 1, this.Sys.ListRows
		totalRows   = 0
		seoStr      []string
	)

	if listRows <= 0 {
		listRows = 10
	}

	chanel := this.GetString(":chanel")
	params := conv.Path2Map(this.GetString(":splat"))
	chanels, rows, err := models.GetList(models.GetTableCategory(), 1, 1, orm.NewCondition().And("Alias", chanel))
	if err != nil {
		helper.Logger.Error("SQL语句执行错误：%v", err.Error())
	}

	if rows == 0 {
		this.Redirect("/", 302)
	}

	if v, ok := params["pid"]; ok {
		pid = helper.Interface2Int(v)
	}

	if v, ok := params["cid"]; ok {
		cid = helper.Interface2Int(v)
	}

	if v, ok := params["p"]; ok { //页码处理
		p = helper.NumberRange(helper.Interface2Int(v), 1, 100)
	}

	orderBy := []string{"Sort", "Title"} //分类排序
	totalRows = helper.Interface2Int(chanels[0]["Cnt"])
	seoStr = append(seoStr, chanels[0]["Title"].(string))
	if pid > 0 {
		totalRows = 0
		this.Data["Children"], _, _ = models.GetList(models.GetTableCategory(), 1, 50, orm.NewCondition().And("Pid", pid), orderBy...)
		if curParent, rows, err := models.GetList(models.GetTableCategory(), 1, 1, orm.NewCondition().And("Id", pid), orderBy...); err != nil {
			helper.Logger.Error(err.Error())
		} else if rows > 0 {
			this.Data["CurParent"] = curParent[0]
			totalRows = helper.Interface2Int(curParent[0]["Cnt"])
			seoStr = append(seoStr, curParent[0]["Title"].(string))
		}
	}

	if cid > 0 {
		totalRows = 0
		if curChildren, rows, err := models.GetList(models.GetTableCategory(), 1, 1, orm.NewCondition().And("Id", cid), orderBy...); err != nil {
			helper.Logger.Error(err.Error())
		} else if rows > 0 {
			this.Data["CurChildren"] = curChildren[0]
			totalRows = helper.Interface2Int(curChildren[0]["Cnt"])
			seoStr = append(seoStr, curChildren[0]["Title"].(string))
		}
	}

	//TODO 相关文档

	//热门文档，根据当前所属分类去获取
	TimeStart := int(time.Now().Unix()) - this.Sys.TimeExpireHotspot
	if cid > 0 {
		this.Data["Hots"], _, _ = models.NewDocument().SimpleList(fmt.Sprintf("di.Cid=%v and di.TimeCreate>%v", cid, TimeStart), 10, "vcnt")
	} else if pid > 0 {
		this.Data["Hots"], _, _ = models.NewDocument().SimpleList(fmt.Sprintf("di.Pid=%v and di.TimeCreate>%v", pid, TimeStart), 10, "vcnt")
	} else {
		this.Data["Hots"], _, _ = models.NewDocument().SimpleList(fmt.Sprintf("di.ChanelId=%v and di.TimeCreate>%v", chanels[0]["Id"], TimeStart), 10, "vcnt")
	}

	lists, rows, err := models.GetDocList(0, helper.Interface2Int(chanels[0]["Id"]), pid, cid, p, listRows, "Id", 0, 1)

	this.Data["PageId"] = "wenku-list"
	this.Data["Chanel"] = strings.ToLower(chanel)
	this.Data["CurChanel"] = chanels[0]
	this.Data["CurPid"] = pid
	this.Data["CurCid"] = cid
	this.Data["Lists"] = lists
	this.Data["Seo"] = models.NewSeo().GetByPage("PC-List", strings.Join(seoStr, "-"), strings.Join(seoStr, ","), strings.Join(seoStr, "-"), this.Sys.Site)
	this.Data["Page"] = helper.Paginations(6, totalRows, listRows, p, fmt.Sprintf("/list/%v", chanel), "pid", pid, "cid", cid)
	this.Data["Parents"], _, _ = models.GetList(models.GetTableCategory(), 1, 20, orm.NewCondition().And("Pid", chanels[0]["Id"]), orderBy...)
	this.Data["PageId"] = "wenku-list"
	this.TplName = "index.html"
}
