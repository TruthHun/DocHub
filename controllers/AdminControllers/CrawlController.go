package AdminControllers

import (
	"strings"

	"github.com/TruthHun/DocHub/helper/conv"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego/orm"
)

type CrawlController struct {
	BaseController
}

//采集管理
func (this *CrawlController) Get() {
	Type := this.GetString("type", "gitbook") //默认是gitbook
	p, _ := this.GetInt("p", 1)               //页码
	listRows := 20                            //每页显示记录数
	//分类
	var cates []models.Category
	models.O.QueryTable(models.TableCategory).OrderBy("Sort", "Title").All(&cates)
	//标题关键字
	Title := this.GetString("title")
	Topic := this.GetString("topic")

	var (
		TotalRows int64 = 0 //默认总记录数
		data      []orm.Params
	)
	switch Type {
	case "gitbook":
		cond := orm.NewCondition().And("Id__gt", 0)
		if len(Title) > 0 {
			cond = cond.And("Title__contains", Title)
		}
		if len(Topic) > 0 {
			cond = cond.And("Topics__icontains", Topic)
		}
		data, _, _ = models.GetList("gitbook", p, listRows, cond, "PublishPdf", "PublishEpub", "PublishMobi", "-Stars")
		TotalRows = models.Count(models.TableGitbook, cond)
	}
	this.Data["IsCrawl"] = true
	this.Data["CatesJson"], _ = conv.InterfaceToJson(cates)
	this.Data["Cates"] = cates
	this.Data["Data"] = data
	this.Data["Type"] = Type
	this.Data["topic"] = Topic
	this.Data["title"] = Title
	this.Data["TotalRows"] = TotalRows
	this.TplName = "index.html"
}

//Gitbook发布
func (this *CrawlController) PublishGitbook() {
	if models.GlobalGitbookPublishing {
		this.ResponseJson(0, "存在正在发布的GitBook")
	}
	Chanel, _ := this.GetInt("Chanel")     //频道
	Parent, _ := this.GetInt("Parent")     //父类
	Children, _ := this.GetInt("Children") //子类
	Uid, _ := this.GetInt("Uid")           //发布人id
	Ids := this.GetString("Ids")           //文档id
	if Chanel*Parent*Children*Uid == 0 || len(Ids) == 0 {
		this.ResponseJson(0, "所有选项均为必填项")
	}
	//标记为正在发布
	models.GlobalGitbookPublishing = true
	//查询电子书
	IdsArr := strings.Split(Ids, ",")
	//执行发布操作
	go models.ModelGitbook.PublishBooks(Chanel, Parent, Children, Uid, IdsArr)
	this.ResponseJson(1, "GitBook发布操作提交成功，程序自会执行发布操作")
}
