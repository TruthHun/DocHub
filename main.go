package main

import (
	"fmt"

	"github.com/TruthHun/DocHub/controllers/HomeControllers"
	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	_ "github.com/TruthHun/DocHub/routers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
)

func main() {
	go execTask()
	//定义错误和异常处理控制器
	beego.ErrorController(&HomeControllers.BaseController{})
	beego.Run()
}

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
	go helper.Segmenter.LoadDictionary("./dictionary/dictionary.txt")

	beego.AddFuncMap("TimestampFormat", helper.TimestampFormat)
	beego.AddFuncMap("Interface2Int", helper.Interface2Int)
	beego.AddFuncMap("Interface2String", helper.Interface2String)
	beego.AddFuncMap("Default", helper.Default)
	beego.AddFuncMap("FormatByte", helper.FormatByte)
	beego.AddFuncMap("CalcInt", helper.CalcInt)
	beego.AddFuncMap("StarVal", helper.StarVal)
	beego.AddFuncMap("Equal", helper.Equal)
	beego.AddFuncMap("SimpleList", models.ModelDoc.TplSimpleList)       //简易的文档列表
	beego.AddFuncMap("HandlePageNum", helper.HandlePageNum)             //处理文档页码为0的显示问题
	beego.AddFuncMap("DoesCollect", models.DoesCollect)                 //判断用户是否已收藏了该文档
	beego.AddFuncMap("DoesSign", models.ModelSign.DoesSign)             //用户今日是否已签到
	beego.AddFuncMap("Friends", models.ModelFriend.Friends)             //友情链接
	beego.AddFuncMap("DefPic", models.NewOss().DefaultPicture)          //获取默认图片
	beego.AddFuncMap("CategoryName", models.ModelCategory.GetTitleById) //根据分类id获取分类名称
	beego.AddFuncMap("IsIllegal", models.ModelDoc.IsIllegal)            //根据md5判断文档是否是非法文档
	beego.AddFuncMap("IsRemark", models.ModelDocRemark.IsRemark)        //根据文档是否存在备注
	beego.AddFuncMap("Xmd5", helper.Xmd5)                               //xmd5，MD5扩展加密
	beego.AddFuncMap("BuildURL", helper.BuildURL)                       //创建URL
	beego.AddFuncMap("HeightLight", helper.HeightLight)                 //高亮
	beego.AddFuncMap("ReportReason", models.ModelSys.GetReportReason)   //举报原因
	beego.AddFuncMap("GetDescByMd5", models.ModelDocText.GetDescByMd5)
	beego.AddFuncMap("GetDescByDsId", models.ModelDocText.GetDescByDsId)
	beego.AddFuncMap("GetDescByDid", models.ModelDocText.GetDescByDid)
}

//定时器，定时更新sitemap和ElasticSearch全量索引
func execTask() {
	//每天凌晨两点执行sitemap更新
	updateSitemap := toolbox.NewTask("updateSitemap", "0 0 2 * * *", func() error {
		models.ModelSeo.BuildSitemap()
		return nil
	})
	//每天凌晨3:00执行elasticsearch全文搜索全量更新
	updateIndex := toolbox.NewTask("updateIndex", "0 0 3 * * *", func() error {
		//TODO:只有开启了全文搜索，才执行索引更新
		return nil
	})

	toolbox.AddTask("updateSitemap", updateSitemap)
	toolbox.AddTask("updateIndex", updateIndex)
	toolbox.StartTask()
	//defer toolbox.StopTask()
}
