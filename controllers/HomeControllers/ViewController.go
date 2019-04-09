package HomeControllers

import (
	"fmt"

	"strings"

	"time"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type ViewController struct {
	BaseController
}

func (this *ViewController) Get() {
	id, _ := this.GetInt(":id")
	if id < 1 {
		this.Redirect("/", 302)
		return
	}

	doc, err := models.NewDocument().GetById(id)

	// 文档不存在、查询错误、被删除，报 404
	if err != nil || doc.Id <= 0 || doc.Status < models.DocStatusConverting {
		this.Abort("404")
	}

	var cates []models.Category
	cates, _ = models.NewCategory().GetCategoriesById(doc.Cid, doc.ChanelId, doc.Pid)
	breadcrumb := make(map[string]models.Category)
	for _, cate := range cates {
		switch cate.Id {
		case doc.ChanelId:
			breadcrumb["Chanel"] = cate
		case doc.Pid:
			breadcrumb["Parent"] = cate
		case doc.Cid:
			TimeStart := int(time.Now().Unix()) - this.Sys.TimeExpireHotspot
			this.Data["Hots"], _, _ = models.NewDocument().SimpleList(fmt.Sprintf("di.Cid=%v and di.TimeCreate>%v", doc.Cid, TimeStart), 10, "Dcnt")
			this.Data["Latest"], _, _ = models.NewDocument().SimpleList(fmt.Sprintf("di.Cid=%v", doc.Cid), 10, "Id")
			breadcrumb["Child"] = cate
		}
	}
	this.Data["Breadcrumb"] = breadcrumb

	models.Regulate(models.GetTableDocumentInfo(), "Vcnt", 1, "`Id`=?", id)

	pageShow := 5
	if doc.Page > pageShow {
		this.Data["PreviewPages"] = make([]string, pageShow)
	} else {
		this.Data["PreviewPages"] = make([]string, doc.Page)
	}
	this.Data["PageShow"] = pageShow

	this.Xsrf()
	if this.Data["Comments"], _, err = models.NewDocumentComment().GetCommentList(id, 1, 10); err != nil {
		helper.Logger.Error(err.Error())
	}

	content := models.NewDocText().GetDescByMd5(doc.Md5, 5000)
	seoTitle := fmt.Sprintf("%v - %v · %v · %v ", doc.Title, breadcrumb["Chanel"].Title, breadcrumb["Parent"].Title, breadcrumb["Child"].Title)
	seoKeywords := fmt.Sprintf("%v,%v,%v,", breadcrumb["Chanel"].Title, breadcrumb["Parent"].Title, breadcrumb["Child"].Title) + doc.Keywords
	seoDesc := beego.Substr(doc.Description+content, 0, 255)
	this.Data["Seo"] = models.NewSeo().GetByPage("PC-View", seoTitle, seoKeywords, seoDesc, this.Sys.Site)
	this.Data["Content"] = content
	this.Data["Reasons"] = models.NewSys().GetReportReasons()
	this.Data["IsViewer"] = true
	this.Data["PageId"] = "wenku-content"
	this.Data["Doc"] = doc

	doc.Ext = strings.TrimLeft(doc.Ext, ".")
	if doc.Page == 0 { //不能预览的文档
		this.Data["OnlyCover"] = true
		this.TplName = "disabled.html"
	} else {
		this.Data["ViewAll"] = doc.PreviewPage == 0 || doc.PreviewPage >= doc.Page
		this.TplName = "svg.html"
	}

}

//文档下载
func (this *ViewController) Download() {
	id, _ := this.GetInt(":id")
	if id <= 0 {
		this.ResponseJson(false, "文档id不正确")
	}

	if this.IsLogin == 0 {
		this.ResponseJson(false, "请先登录")
	}

	link, err := models.NewUser().CanDownloadFile(this.IsLogin, id)
	if err != nil {
		this.ResponseJson(false, err.Error())
	}
	this.ResponseJson(true, "下载链接获取成功", map[string]interface{}{"url": link})
}

//是否可以免费下载
func (this *ViewController) DownFree() {
	if this.IsLogin > 0 {
		did, _ := this.GetInt("id")
		if free := models.NewFreeDown().IsFreeDown(this.IsLogin, did); free {
			this.ResponseJson(true, fmt.Sprintf("您上次下载过当前文档，且仍在免费下载有效期(%v天)内，本次下载免费", this.Sys.FreeDay))
		}
	}
	this.ResponseJson(false, "不能免费下载，不在免费下载期限内")
}

//文档评论
func (this *ViewController) Comment() {
	id, _ := this.GetInt(":id")
	score, _ := this.GetInt("Score")
	answer := this.GetString("Answer")
	if answer != this.Sys.Answer {
		this.ResponseJson(false, "请输入正确的答案")
	}
	if id > 0 {
		if this.IsLogin > 0 {
			if score < 1 || score > 5 {
				this.ResponseJson(false, "请给文档评分")
			} else {
				comment := models.DocumentComment{
					Uid:        this.IsLogin,
					Did:        id,
					Content:    this.GetString("Comment"),
					TimeCreate: int(time.Now().Unix()),
					Status:     true,
					Score:      score * 10000,
				}
				cnt := strings.Count(comment.Content, "") - 1
				if cnt > 255 || cnt < 8 {
					this.ResponseJson(false, "评论内容限8-255个字符")
				} else {
					_, err := orm.NewOrm().Insert(&comment)
					if err != nil {
						this.ResponseJson(false, "发表评论失败：每人仅限给每个文档点评一次")
					} else {
						//文档评论人数增加
						sql := fmt.Sprintf("UPDATE `%v` SET `Score`=(`Score`*`ScorePeople`+%v)/(`ScorePeople`+1),`ScorePeople`=`ScorePeople`+1 WHERE Id=%v", models.GetTableDocumentInfo(), comment.Score, comment.Did)
						_, err := orm.NewOrm().Raw(sql).Exec()
						if err != nil {
							helper.Logger.Error(err.Error())
						}
						this.ResponseJson(true, "恭喜您，评论发表成功")
					}
				}
			}
		} else {
			this.ResponseJson(false, "评论失败，您当前处于未登录状态，请先登录")
		}
	} else {
		this.ResponseJson(false, "评论失败，参数不正确")
	}
}

//获取评论列表
func (this *ViewController) GetComment() {
	p, _ := this.GetInt("p", 1)
	did, _ := this.GetInt("did")
	if p > 0 && did > 0 {
		if rows, _, err := models.NewDocumentComment().GetCommentList(did, p, 10); err != nil {
			helper.Logger.Error(err.Error())
			this.ResponseJson(false, "评论列表获取失败")
		} else {
			this.ResponseJson(true, "评论列表获取成功", rows)
		}
	}
}
