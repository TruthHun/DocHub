package HomeControllers

import (
	"strings"

	"lazybug.me/conv"
	"lazybug.me/util"
	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
)

type SearchController struct {
	BaseController
}

func (this *SearchController) Get() {

	var (
		p        int = 1  //默认页码
		listRows int = 10 //默认每页显示记录数
	)
	//path中的参数
	params := conv.Path2Map(this.GetString(":splat"))
	if _, ok := params["wd"]; !ok { //搜索关键字
		params["wd"] = this.GetString("wd")
	}

	//缺少搜索关键字，直接返回首页
	if len(params["wd"]) == 0 {
		this.Redirect("/", 302)
		return
	}

	if _, ok := params["type"]; !ok { //搜索类型
		params["type"] = this.GetString("type")
	}
	params["type"] = helper.Default(params["type"], "all") //默认全部搜索
	if _, ok := params["sort"]; !ok {                      //排序
		params["sort"] = this.GetString("sort")
	}
	params["sort"] = helper.Default(params["sort"], "default") //默认排序

	if _, ok := params["p"]; ok {
		p = helper.Interface2Int(params["p"])
	} else {
		p, _ = this.GetInt("p")
	}

	//页码处理
	p = util.NumberRange(p, 1, 100)
	res := models.Search(params["wd"], params["type"], params["sort"], p, listRows, 1)
	if res.Total > 0 && len(res.Ids) > 0 {
		data := models.ModelDoc.GetDocsByIds(res.Ids)
		if len(data) > 0 {
			for index, val := range data {
				if len(strings.TrimSpace(val["Description"].(string))) == 0 {
					if desc := models.ModelDocText.GetDescByMd5(val["Md5"], 120); len(desc) == 0 {
						data[index]["Description"] = val["Title"]
					} else {
						data[index]["Description"] = desc + "..."
					}
				}
			}
		}
		this.Data["Data"] = data
	}

	if p == 1 {
		wdSlice := strings.Split(this.Sys.DirtyWord, " ")
		for _, wd := range wdSlice {
			if !strings.Contains(params["wd"], wd) {
				models.ReplaceInto(models.TableSearchLog, map[string]interface{}{"Wd": params["wd"]})
				break
			}
		}
	}
	this.Data["Seo"] = models.ModelSeo.GetByPage("PC-Search", params["wd"], "文档搜索,"+params["wd"], "文档搜索,"+params["wd"], this.Sys.Site)
	this.Data["Page"] = helper.Paginations(6, int(res.Total), listRows, p, "/search/", "type", params["type"], "sort", params["sort"], "p", p, "wd", params["wd"])
	this.Data["Params"] = params
	this.Data["Result"] = res
	this.Data["ListRows"] = listRows
	this.Data["WordLen"] = len(res.Word) //分词的个数
	this.Data["SearchLog"] = models.ModelSearchLog.List(1, 10)
	this.Layout = ""
	this.Data["PageId"] = "wenku-search"
	this.TplName = "index.html"
}
