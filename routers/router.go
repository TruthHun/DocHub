package routers

import (
	"github.com/TruthHun/DocHub/controllers/AdminControllers"

	"github.com/TruthHun/DocHub/controllers/HomeControllers"
	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {

	beego.InsertFilter("/*", beego.BeforeRouter, func(ctx *context.Context) {
		if !helper.IsInstalled && ctx.Request.URL.Path != "/install" { //程序未安装，且请求路径不是install，则跳转到install
			ctx.Redirect(302, "/install")
		}
	})

	front()
	back()
}

//前台路由
func front() {
	beego.Router("/", &HomeControllers.IndexController{})
	beego.Router("/install", &HomeControllers.InstallController{}, "get,post:Install")
	beego.Router("/list/:chanel", &HomeControllers.ListController{})
	beego.Router("/list/:chanel/*", &HomeControllers.ListController{})
	beego.Router("/user", &HomeControllers.UserController{})
	beego.Router("/user/sign", &HomeControllers.UserController{}, "get:Sign")
	beego.Router("/report", &HomeControllers.ReportController{})
	beego.Router("/collect", &HomeControllers.CollectController{})
	beego.Router("/collect/folder", &HomeControllers.CollectController{}, "get:FolderList")
	beego.Router("/user/:uid", &HomeControllers.UserController{})
	beego.Router("/user/:uid/doc", &HomeControllers.UserController{})
	beego.Router("/user/:uid/collect", &HomeControllers.UserController{}, "get:Collect")
	beego.Router("/user/:uid/coin", &HomeControllers.UserController{}, "get:Coin")
	beego.Router("/user/:uid/del/doc/:doc", &HomeControllers.UserController{}, "get:DocDel")
	beego.Router("/user/:uid/edit/doc/:doc", &HomeControllers.UserController{}, "get,post:DocEdit")
	beego.Router("/user/:uid/del/collect/:cid", &HomeControllers.UserController{}, "get:CollectFolderDel")
	beego.Router("/user/:uid/cancel/collect/:cid/:did", &HomeControllers.UserController{}, "get:CollectCancel") //取消收藏文档
	beego.Router("/user/:uid/doc/*", &HomeControllers.UserController{})
	beego.Router("/user/avatar", &HomeControllers.UserController{}, "post:Avatar")
	beego.Router("/user/edit", &HomeControllers.UserController{}, "post:Edit")
	beego.Router("/user/folder/add", &HomeControllers.UserController{}, "post:CreateCollectFolder")
	beego.Router("/user/login", &HomeControllers.UserController{}, "get,post:Login")
	beego.Router("/user/islogin", &HomeControllers.UserController{}, "get:CheckLogin")
	beego.Router("/user/logout", &HomeControllers.UserController{}, "get:Logout")
	beego.Router("/user/findpwd", &HomeControllers.UserController{}, "get,post:FindPwd")
	beego.Router("/user/reg", &HomeControllers.UserController{}, "get,post:Reg")
	beego.Router("/user/sendmail", &HomeControllers.UserController{}, "get:SendMail")
	beego.Router("/upload", &HomeControllers.UploadController{}, "get:Get")
	beego.Router("/upload", &HomeControllers.UploadController{}, "post:Post")
	beego.Router("/segwd", &HomeControllers.UploadController{}, "get:SegWord")
	beego.Router("/search/*", &HomeControllers.SearchController{})
	beego.Router("/view/:id", &HomeControllers.ViewController{})
	beego.Router("/comment/:id", &HomeControllers.ViewController{}, "post:Comment")
	beego.Router("/comment/list", &HomeControllers.ViewController{}, "get:GetComment")
	beego.Router("/down/:id", &HomeControllers.ViewController{}, "get:Download")
	beego.Router("/downfree", &HomeControllers.ViewController{}, "get:DownFree")
	beego.Router("/doc/check", &HomeControllers.BaseController{}, "get:DocExist")
	beego.Router("/pages/:page", &HomeControllers.BaseController{}, "get:Pages")
	beego.Router("/*", &HomeControllers.StaticController{}, "get:Static")
}

//后台路由
func back() {
	beego.Router("/admin", &AdminControllers.IndexController{})
	beego.Router("/admin/login", &AdminControllers.LoginController{}, "get,post:Login")
	beego.Router("/admin/updatePwd", &AdminControllers.LoginController{}, "post:UpdatePwd")
	beego.Router("/admin/update-admin", &AdminControllers.LoginController{}, "post:UpdateAdmin")
	beego.Router("/admin/logout", &AdminControllers.LoginController{}, "get:Logout")
	beego.Router("/admin/user", &AdminControllers.UserController{}, "get:List")
	beego.Router("/admin/user/*", &AdminControllers.UserController{}, "get:List")
	beego.Router("/admin/doc", &AdminControllers.DocController{})
	beego.Router("/admin/doc/cate", &AdminControllers.DocController{}, "get:Category")
	beego.Router("/admin/doc/addcate", &AdminControllers.DocController{}, "post:AddCate")
	beego.Router("/admin/doc/addchanel", &AdminControllers.DocController{}, "post:AddChanel")
	beego.Router("/admin/doc/delcate", &AdminControllers.DocController{}, "get:DelCate")
	beego.Router("/admin/doc/getCateByCid", &AdminControllers.DocController{}, "get:GetCateByCid")
	beego.Router("/admin/doc/action", &AdminControllers.DocController{}, "get:Action")
	beego.Router("/admin/doc/list", &AdminControllers.DocController{}, "get:List")
	beego.Router("/admin/doc/recycle", &AdminControllers.DocController{}, "get:Recycle")
	beego.Router("/admin/doc/remark", &AdminControllers.DocController{}, "get,post:RemarkTpl")
	beego.Router("/admin/doc/list/*", &AdminControllers.DocController{}, "get:List")
	beego.Router("/admin/sys", &AdminControllers.SysController{}, "get,post:Get")
	beego.Router("/admin/cloud-store", &AdminControllers.SysController{}, "get:CloudStore")
	beego.Router("/admin/cloud-store", &AdminControllers.SysController{}, "post:SetCloudStore")
	beego.Router("/admin/sys/handle-logs", &AdminControllers.SysController{}, "get:HandleLogs") //下载或者删除日志文件
	beego.Router("/admin/seo", &AdminControllers.SeoController{})
	beego.Router("/admin/seo/sitemap", &AdminControllers.SeoController{}, "get:UpdateSitemap") //更新站点地图
	beego.Router("/admin/ad", &AdminControllers.AdController{})
	beego.Router("/admin/friend", &AdminControllers.FriendController{}, "get,post:Get")
	beego.Router("/admin/update", &AdminControllers.BaseController{}, "get,post:Update")
	beego.Router("/admin/del", &AdminControllers.BaseController{}, "get,post:Del")
	beego.Router("/admin/single", &AdminControllers.SingleController{})
	beego.Router("/admin/single/:alias", &AdminControllers.SingleController{}, "get,post:Edit")
	//beego.Router("/admin/singledel/:alias", &AdminControllers.SingleController{}, "get:Del")
	beego.Router("/admin/kindeditor/upload", &AdminControllers.KindEditorController{}, "post:Upload")
	beego.Router("/admin/score", &AdminControllers.ScoreController{})
	beego.Router("/admin/banner", &AdminControllers.BannerController{})
	beego.Router("/admin/banner/add", &AdminControllers.BannerController{}, "post:Add")
	beego.Router("/admin/banner/del", &AdminControllers.BannerController{}, "get,post:Del")
	beego.Router("/admin/report", &AdminControllers.ReportController{})
	beego.Router("/admin/elasticsearch/rebuild", &AdminControllers.SysController{}, "get:RebuildAllIndex") //重建全量索引
	beego.Router("/admin/test/send-email", &AdminControllers.SysController{}, "get:TestForSendingEmail")
	//beego.Router("/admin/test/ping-oss", &AdminControllers.SysController{}, "get:TestOSS")
}
