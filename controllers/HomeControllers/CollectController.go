package HomeControllers

import (
	"fmt"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego/orm"
)

type CollectController struct {
	BaseController
}

//收藏文档
func (this *CollectController) Get() {
	if this.IsLogin == 0 {
		this.ResponseJson(false, "您当前未登录，请先登录")
	}

	cid, _ := this.GetInt("Cid")
	did, _ := this.GetInt("Did")

	if cid == 0 || did == 0 {
		this.ResponseJson(false, "收藏失败：参数不正确")
	}

	collect := models.Collect{Did: did, Cid: cid}
	rows, err := orm.NewOrm().Insert(&collect)
	if err != nil {
		helper.Logger.Error("SQL执行失败：%v", err.Error())
	}

	if err != nil || rows == 0 {
		this.ResponseJson(false, "收藏失败：您已收藏过该文档")
	}

	//文档被收藏的数量+1
	models.Regulate(models.GetTableDocumentInfo(), "Ccnt", 1, fmt.Sprintf("`Id`=%v", did))

	//收藏夹的文档+1
	models.Regulate(models.GetTableCollectFolder(), "Cnt", 1, fmt.Sprintf("`Id`=%v", cid))

	this.ResponseJson(true, "恭喜您，文档收藏成功。")
}

//收藏夹列表
func (this *CollectController) FolderList() {
	uid, _ := this.GetInt("uid")
	if uid < 1 {
		uid = this.IsLogin
	}

	if uid == 0 {
		this.ResponseJson(false, "获取收藏夹失败：请先登录")
	}

	lists, rows, err := models.GetList(models.GetTableCollectFolder(), 1, 100, orm.NewCondition().And("Uid", uid), "-Id")
	if err != nil {
		helper.Logger.Error(err.Error())
	}
	if rows > 0 && err == nil {
		this.ResponseJson(true, "收藏夹获取获取成功", lists)
	}
	this.ResponseJson(false, "暂时没有收藏夹，请先在会员中心创建收藏夹")
}

//取消收藏文档
func (this *CollectController) CollectCancel() {
	if this.IsLogin == 0 {
		this.ResponseJson(false, "您当前未登录，请先登录")
	}

	cid, _ := this.GetInt("Cid")
	did, _ := this.GetInt("Did")
	if cid == 0 || did == 0 {
		this.ResponseJson(false, "收藏失败：参数不正确")
	}

	if err := models.NewCollect().Cancel(did, cid, this.IsLogin); err != nil {
		helper.Logger.Error(err.Error())
		this.ResponseJson(false, "移除收藏文档失败，可能您为收藏该文档")
	}

	//文档被收藏的数量-1
	models.Regulate(models.GetTableDocumentInfo(), "Ccnt", -1, "`Id`=?", did)

	//收藏夹的文档-1
	models.Regulate(models.GetTableCollectFolder(), "Cnt", -1, "`Id`=?", cid)

	this.ResponseJson(true, "恭喜您，删除收藏文档成功")
}
