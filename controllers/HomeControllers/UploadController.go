package HomeControllers

import (
	"fmt"

	"crypto/md5"
	"io"

	"strings"

	"time"

	"os"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"lazybug.me/conv"
	"lazybug.me/util"
	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
)

type UploadController struct {
	BaseController
}

//分词
func (this *UploadController) SegWord() {
	var wds string
	if this.IsLogin > 0 {
		wds = helper.SegWord(this.GetString("word"))
	}
	this.ResponseJson(1, "分词成功", wds)
}

//文档上传页面
func (this *UploadController) Get() {
	cond := orm.NewCondition().And("status", 1)
	data, _, _ := models.GetList("category", 1, 2000, cond, "Sort", "Title")
	//cates := models.ToTree(data, "Pid", 0)
	this.Xsrf()
	this.Data["Seo"] = models.ModelSeo.GetByPage("PC-Upload", "文档上传-文档分享", "文档上传,文档分享", "文档上传-文档分享", this.Sys.Site)
	this.Data["Cates"], _ = conv.InterfaceToJson(data)
	this.Data["json"] = data
	this.Data["IsUpload"] = true
	this.Data["PageId"] = "wenku-upload"
	this.Data["MaxSize"] = beego.AppConfig.DefaultInt("max_upload_size", 52428800)
	this.TplName = "index.html"
}

//文档执行操作
//处理流程：
//1、检测用户是否已登录，未登录不允许上传
//2、检测是否存在了该文档的md5，如果已存在，则根据md5查询存储在文档存档表中的数据；如果文档已经在文档存储表中存在，则该文档不需要再获取封面、大小、页码等数据
//3、检测文档格式是否符合要求。
//4、计算文档md5，然后根据md5再比对一次文档是否在存档表中存在
//5、文档未存在，则将文档数据录入文档存储表(document_store)
//6、执行文档转pdf，并获取文档页数、封面、摘要等
//7、获取文档大小
func (this *UploadController) Post() {
	var (
		ext     string //文档扩展名
		tmpfile string //存在服务器的临时文件
		dir     = fmt.Sprintf("./uploads/%v/%v", time.Now().Format("2006/01/02"), this.IsLogin)
		form    models.FormUpload
		err     error
	)

	//1、用户是否已登录
	if this.IsLogin == 0 {
		this.ResponseJson(0, "您当前未登录，请先登录")
	}

	this.ParseForm(&form)
	//检查必填字段是否已经填写完毕
	if len(form.Title) == 0 || form.Chanel*form.Pid*form.Cid == 0 {
		this.ResponseJson(0, "文档名称、频道、一级文档分类、二级文档分类均不能为空")
	}
	//允许上传的文档格式
	allowedExt := ",doc,docx,rtf,wps,odt,ppt,pptx,pps,ppsx,dps,odp,pot,xls,xlsx,et,ods,txt,pdf,chm,epub,umd,mobi,"

	//写死的范围，0-20
	form.Price = util.NumberRange(form.Price, 0, 20)

	//创建文档模型对象
	//ModelDoc := models.Document{}
	//文件在文档存档表中已存在，则不接收文档处理

	//非法文件，提示不允许上传，这里检测一次
	if models.ModelDoc.IsIllegal(form.Md5) {
		this.ResponseJson(0, "您上传的文档已被站点标记为不符合要求的文档，暂时不允许上传分享。")
	}

	if len(form.Md5) == 32 && form.Exist == 1 {
		err = models.HandleExistDoc(this.IsLogin, form)
	} else {
		//文件在文档库中未存在，则接收文件并做处理
		f, fh, err := this.GetFile("File")
		if err != nil {
			this.ResponseJson(0, err.Error())
		}
		defer f.Close()
		//判断文档格式是否被允许

		ext = strings.ToLower(helper.GetSuffix(fh.Filename, "."))
		if !strings.Contains(allowedExt, fmt.Sprintf(",%v,", ext)) {
			this.ResponseJson(0, "您上传的文档格式不正确，请上传正确格式的文档")
		}
		//获取文件MD5
		md5func := func(file io.Reader) string {
			md5h := md5.New()
			io.Copy(md5h, file)
			return fmt.Sprintf("%x", md5h.Sum(nil))
		}
		form.Md5 = md5func(f)
		form.Exist = 0 //哪怕存在了，这里也设置为0
		form.Ext = ext
		form.Filename = fh.Filename

		//非法文件，提示不允许上传。这里再检测一次，同时删除文档
		if models.ModelDoc.IsIllegal(form.Md5) {
			this.ResponseJson(0, "您上传的文档已被站点标记为不符合要求的文档，暂时不允许上传分享。")
		}

		//如果文档已经存在，则直接调用处理
		if models.ModelDoc.IsExistByMd5(form.Md5) > 0 {
			models.HandleExistDoc(this.IsLogin, form)
			this.ResponseJson(1, "文档上传成功")
		}

		os.MkdirAll(dir, 0777)
		tmpfile = dir + "/" + form.Md5 + "." + ext
		err = this.SaveToFile("File", tmpfile)
		if err != nil {
			this.ResponseJson(0, "文档存储失败，请重新上传")
		}
		if info, err := os.Stat(tmpfile); err == nil {
			form.Size = int(info.Size())
		}

		switch ext {
		case "pdf": //处理pdf文档
			err = models.HandlePdf(this.IsLogin, tmpfile, form)
			if err != nil {
				helper.Logger.Error(err.Error())
			}
		case "umd", "epub", "chm", "txt", "mobi": //处理无法转码实现在线浏览的文档
			go models.HandleUnOffice(this.IsLogin, tmpfile, form)
		default: //处理office文档
			go models.HandleOffice(this.IsLogin, tmpfile, form)
			//this.ResponseJson(1, "^.^ 恭喜您，成功上传了一篇文档。感谢您为知识的传承献上自己的一份力量。由于Office文档转码需要些时间，转码成功后方可在线预览，请您稍稍等待")
		}
	}
	if err == nil {
		price := this.Sys.Reward
		var log = models.CoinLog{
			Uid: this.IsLogin,
		}
		if form.Exist == 1 {
			price = 1 //已被分享过的文档，奖励1个金币
			log.Log = fmt.Sprintf("于%v成功分享了一篇已分享过的文档，获得 %v 个金币奖励", time.Now().Format("2006-01-02 15:04:05"), price)
		} else {
			log.Log = fmt.Sprintf("于%v成功分享了一篇未分享过的文档，获得 %v 个金币奖励", time.Now().Format("2006-01-02 15:04:05"), price)
		}
		log.Coin = price //金币变更
		if err := models.ModelCoinLog.LogRecord(log); err != nil {
			helper.Logger.Error(err.Error())
		}
		models.Regulate(models.TableUserInfo, "Coin", price, "Id=?", this.IsLogin)
		this.ResponseJson(1, "^.^ 恭喜您，成功上传了一篇文档。感谢您为知识的传承献上自己的一份力量。由于文档转码处理需要些时间，转码成功后方可在线预览，请您稍稍等待")
	} else {
		this.ResponseJson(0, "啊哦，文档上传失败...再重试一下吧。")
	}

}
