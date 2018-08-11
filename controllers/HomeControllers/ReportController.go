package HomeControllers

import (
	"time"

	"fmt"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego/orm"
)

type ReportController struct {
	BaseController
}

//举报
func (this *ReportController) Get() {
	if this.IsLogin == 0 {
		this.ResponseJson(false, "您当前未登录，请先登录")
	} else {
		reason, _ := this.GetInt("Reason")
		did, _ := this.GetInt("Did")
		if reason > 0 && did > 0 {
			var report models.Report
			t := int(time.Now().Unix())
			report.Status = false
			report.Did = did
			report.TimeCreate = t
			report.TimeUpdate = t
			report.Uid = this.IsLogin
			report.Reason = reason
			rows, err := orm.NewOrm().Insert(&report)
			if err != nil {
				helper.Logger.Error("SQL执行失败：%v", err.Error())
			}
			if err == nil && rows > 0 {
				this.ResponseJson(true, fmt.Sprintf("恭喜您，举报成功。感谢您对 %v 的支持，我们将在24小时内对您举报的内容进行处理。", this.Sys.Site))
			} else {
				this.ResponseJson(false, "举报失败：您已举报过该文档")
			}
		} else {
			this.ResponseJson(false, "举报失败，请选择举报原因")
		}
	}

}
