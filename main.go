package main

import (
	"fmt"

	"time"

	"github.com/TruthHun/DocHub/controllers/HomeControllers"
	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	_ "github.com/TruthHun/DocHub/routers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

//初始化函数
func init() {
	fmt.Println("")
	fmt.Println("Powered By DocHub")
	fmt.Println("Author:进击的皇虫(TruthHun@QQ.COM)")
	fmt.Println("")
	//sitemap静态目录
	beego.SetStaticPath("/sitemap", "sitemap")

	//数据库初始化
	models.Init()
	//初始化日志
	helper.InitLogs()
	//初始化分词器
	go InitSego()
	beego.AddFuncMap("TimestampFormat", helper.TimestampFormat)
	beego.AddFuncMap("Interface2Int", helper.Interface2Int)
	beego.AddFuncMap("Interface2String", helper.Interface2String)
	beego.AddFuncMap("Default", helper.Default)
	beego.AddFuncMap("FormatByte", helper.FormatByte)
	beego.AddFuncMap("CalcInt", helper.CalcInt)
	beego.AddFuncMap("StarVal", helper.StarVal)
	beego.AddFuncMap("Equal", helper.Equal)
	beego.AddFuncMap("SimpleList", SimlpeList)                          //简易的文档列表
	beego.AddFuncMap("HandlePageNum", HandlePageNum)                    //处理文档页码为0的显示问题
	beego.AddFuncMap("DoesCollect", DoesCollect)                        //判断用户是否已收藏了该文档
	beego.AddFuncMap("DoesSign", DoesSign)                              //用户今日是否已签到
	beego.AddFuncMap("Friends", Friends)                                //友情链接
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

func main() {
	//定义错误和异常处理控制器
	beego.ErrorController(&HomeControllers.BaseController{})
	beego.Run()
}

//初始化分词词典
func InitSego() {
	fmt.Println("加载分词词典...")
	helper.Segmenter.LoadDictionary("./dictionary/dictionary.txt")
	fmt.Println("分词词典加载完成")
}

//文档简易列表
func SimlpeList(chinelid interface{}) []orm.Params {
	data, _, _ := models.ModelDoc.SimpleList(fmt.Sprintf("di.ChanelId=%v", helper.Interface2Int(chinelid)), 5)
	return data
}

//处理页数
func HandlePageNum(PageNum interface{}) string {
	pn := fmt.Sprintf("%v", PageNum)
	if pn == "0" {
		return " -- "
	}
	return pn
}

//是否已收藏文档
func DoesCollect(did, uid int) bool {
	if uid == 0 {
		return false
	}
	var params []orm.Params
	sql := fmt.Sprintf("select c.Id from %v cf left join %v c on c.cid=cf.id where c.Did=? and cf.Uid=? limit 1", models.GetTable("collect_folder"), models.GetTable("collect"))
	rows, err := models.O.Raw(sql, did, uid).Values(&params)
	if err != nil {
		helper.Logger.Error(err.Error())
	}
	if rows > 0 && err == nil {
		return true
	}
	return false
}

//检测用户今天是否已签到
func DoesSign(uid int) bool {
	if _, rows, err := models.GetList("sign", 1, 1, orm.NewCondition().And("Date", time.Now().Format("20060102")).And("Uid", uid)); err == nil && rows > 0 {
		return true
	}
	return false
}

//获取友链
func Friends() []orm.Params {
	rows, _, _ := models.GetList("friend", 1, 100, orm.NewCondition().And("Status", 1), "Sort")
	return rows
}
