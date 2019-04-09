package main

import (
	"fmt"

	"github.com/TruthHun/DocHub/controllers/HomeControllers"
	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	_ "github.com/TruthHun/DocHub/routers"
	"github.com/astaxie/beego"
)

//初始化函数
func init() {

	fmt.Println("")
	fmt.Println("Powered By DocHub")
	fmt.Println("Version:", helper.VERSION)
	fmt.Println("")

	//sitemap静态目录
	beego.SetStaticPath("/sitemap", "sitemap")

	//初始化日志
	helper.InitLogs()

	//初始化分词器
	go func() {
		helper.Segmenter.LoadDictionary("./dictionary/dictionary.txt")
		beego.Info("==程序启动完毕==")
	}()

	beego.AddFuncMap("TimestampFormat", helper.TimestampFormat)
	beego.AddFuncMap("Interface2Int", helper.Interface2Int)
	beego.AddFuncMap("Interface2String", helper.Interface2String)
	beego.AddFuncMap("Default", helper.Default)
	beego.AddFuncMap("FormatByte", helper.FormatByte)
	beego.AddFuncMap("CalcInt", helper.CalcInt)
	beego.AddFuncMap("StarVal", helper.StarVal)
	beego.AddFuncMap("Equal", helper.Equal)
	beego.AddFuncMap("SimpleList", models.NewDocument().TplSimpleList)  //简易的文档列表
	beego.AddFuncMap("HandlePageNum", helper.HandlePageNum)             //处理文档页码为0的显示问题
	beego.AddFuncMap("DoesCollect", models.DoesCollect)                 //判断用户是否已收藏了该文档
	beego.AddFuncMap("DoesSign", models.NewSign().DoesSign)             //用户今日是否已签到
	beego.AddFuncMap("Friends", models.NewFriend().Friends)             //友情链接
	beego.AddFuncMap("CategoryName", models.NewCategory().GetTitleById) //根据分类id获取分类名称
	beego.AddFuncMap("IsIllegal", models.NewDocument().IsIllegal)       //根据md5判断文档是否是非法文档
	beego.AddFuncMap("IsRemark", models.NewDocumentRemark().IsRemark)   //根据文档是否存在备注
	beego.AddFuncMap("BuildURL", helper.BuildURL)                       //创建URL
	beego.AddFuncMap("HeightLight", helper.HeightLight)                 //高亮
	beego.AddFuncMap("ReportReason", models.NewSys().GetReportReason)   //举报原因
	beego.AddFuncMap("GetDescByMd5", models.NewDocText().GetDescByMd5)
	beego.AddFuncMap("GetDescByDsId", models.NewDocText().GetDescByDsId)
	beego.AddFuncMap("GetDescByDid", models.NewDocText().GetDescByDid)
	beego.AddFuncMap("DefaultImage", models.GetImageFromCloudStore) //获取默认图片
}

func main() {
	//定义错误和异常处理控制器
	beego.ErrorController(&HomeControllers.ErrorsController{})
	beego.Run()
}
