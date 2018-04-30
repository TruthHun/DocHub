package AdminControllers

import (
	"time"

	"github.com/TruthHun/DocHub/helper"

	"github.com/TruthHun/DocHub/models"
)

type FriendController struct {
	BaseController
}

//添加以及查看友链列表
func (this *FriendController) Get() {
	if this.Ctx.Request.Method == "POST" {
		var fr models.Friend
		this.ParseForm(&fr)
		fr.Status = true
		fr.TimeCreate = int(time.Now().Unix())
		if i, err := models.O.Insert(&fr); i > 0 && err == nil {
			this.ResponseJson(1, "友链添加成功")
		} else {
			if err != nil {
				helper.Logger.Error(err.Error())
			}
			this.ResponseJson(0, "友链添加失败，可能您要添加的友链已存在")
		}
	} else {
		this.Data["IsFriend"] = true
		this.Data["Friends"], _, _ = models.ModelFriend.GetListByStatus(-1)
		this.TplName = "index.html"
	}
}
