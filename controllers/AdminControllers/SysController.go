package AdminControllers

import (
	"net/http"
	"strings"

	"fmt"

	"io/ioutil"
	"time"

	"path/filepath"

	"os"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/astaxie/beego/orm"
)

type SysController struct {
	BaseController
}

type logFile struct {
	Path    string
	ModTime time.Time
	Size    int
}

//系统配置管理
func (this *SysController) Get() {
	tab := models.ConfigCate(strings.ToLower(this.GetString("tab")))
	switch tab {
	case models.CONFIG_EMAIL, models.CONFIG_DEPEND, models.CONFIG_ELASTICSEARCH, models.CONFIG_LOGS:
	case models.StoreOss, models.StoreLocal, models.StoreBos, models.StoreCos, models.StoreQiniu:
	default:
		tab = "default"
	}
	if this.Ctx.Request.Method == http.MethodPost {
		if tab == "default" {
			var sys models.Sys
			this.ParseForm(&sys)
			if i, err := orm.NewOrm().Update(&sys); i > 0 && err == nil {
				models.NewSys().UpdateGlobal() //更新全局变量
			} else {
				if err != nil {
					helper.Logger.Error(err.Error())
				}
				this.ResponseJson(false, "更新失败，可能您未对内容做更改")
			}
		} else {
			modelCfg := models.NewConfig()
			for k, v := range this.Ctx.Request.Form {
				modelCfg.UpdateByKey(models.ConfigCate(tab), k, v[0])
			}
			//最后更新全局配置
			modelCfg.UpdateGlobal()
			if tab == models.CONFIG_ELASTICSEARCH {
				if err := models.NewElasticSearchClient().Init(); err != nil {
					this.ResponseJson(false, "ElasticSearch初始化失败："+err.Error())
				}
			}

		}
		this.ResponseJson(true, "更新成功")
	}

	this.Data["Tab"] = tab
	this.Data["Title"] = "系统管理"
	this.Data["IsSys"] = true
	this.Data["Store"] = this.GetString("store", string(models.StoreOss))
	if tab == "default" {
		this.Data["Sys"], _ = models.NewSys().Get()
	} else {
		this.Data["Configs"] = models.NewConfig().All()
		if tab == models.CONFIG_ELASTICSEARCH {
			count, errES := models.NewElasticSearchClient().Count()
			this.Data["Count"] = count
			if errES != nil {
				this.Data["ErrES"] = errES.Error()
			}
		} else if tab == "logs" {
			var logs []logFile
			if files, _ := ioutil.ReadDir("logs"); len(files) > 0 {
				for _, file := range files {
					if !file.IsDir() {
						logs = append(logs, logFile{
							Path:    "logs/" + file.Name(),
							ModTime: file.ModTime(),
							Size:    int(file.Size()),
						})
					}
				}
			}
			this.Data["Logs"] = logs
		}
	}
	this.TplName = "index.html"
}

//下载或者删除日志文件
func (this *SysController) HandleLogs() {
	file := this.GetString("file")
	action := this.GetString("action")
	if action == "del" { //删除
		if ext := filepath.Ext(file); ext == ".log" {
			if file == "logs/dochub.log" {
				this.Response(map[string]interface{}{"status": 0, "msg": "日志文件删除失败：logs/dochub.log日志文件禁止删除，否则程序无法写入日志"})
			}
			if err := os.Remove(file); err != nil {
				this.Response(map[string]interface{}{"status": 0, "msg": "日志文件删除失败：" + err.Error()})
			}
		}
		this.Response(map[string]interface{}{"status": 1, "msg": "删除成功"})
	} else { //下载
		if b, err := ioutil.ReadFile(file); err != nil {
			helper.Logger.Error(err.Error())
			this.Abort("404")
		} else {
			this.Ctx.ResponseWriter.Header().Add("Content-disposition", "attachment; filename="+filepath.Base(file))
			this.Ctx.ResponseWriter.Write(b)
		}
	}
}

//重建全量索引
func (this *SysController) RebuildAllIndex() {
	resp := map[string]interface{}{
		"status": 0, "msg": "全量索引重建失败",
	}
	if client := models.NewElasticSearchClient(); client.On {
		//再次初始化，避免elasticsearch未初始化
		if err := client.Init(); err != nil {
			resp["msg"] = "全量索引重建失败：" + err.Error()
		} else {
			//索引是否正在进行，如果正在进行，则不再执行全量索引
			exist := false
			if indexing, ok := helper.ConfigMap.Load("indexing"); ok {
				if b, ok := indexing.(bool); b && ok { //索引正在重建
					exist = true
				} else {
					exist = false
				}
			} else { //不存在正在重建的全量索引操作
				exist = false
			}
			if exist == false {
				go client.RebuildAllIndex()
				resp["msg"] = "全量索引重建提交成功，正在后端执行，请耐心等待."
				resp["status"] = 1
			} else {
				resp["msg"] = "全量索引重建失败：存在正在重建的全量索引"
			}
		}
	} else {
		resp["msg"] = "全量索引重建失败，您未启用ElasticSearch"
	}
	this.Response(resp)
}

//测试邮箱是否能发件成功
func (this *SysController) TestForSendingEmail() {
	//邮件接收人
	to := helper.GetConfig(string(models.CONFIG_EMAIL), "test")
	if err := models.SendMail(to, "测试邮件", "这是一封测试邮件，用于检测是否能正常发送邮件"); err != nil {
		this.Response(map[string]interface{}{"status": 0, "msg": "邮件发送失败：" + err.Error()})
	}
	this.Response(map[string]interface{}{"status": 1, "msg": "邮件发送成功"})
}

//测试OSS是否连通成功
func (this *SysController) TestOSS() {
	var (
		testFile        = "dochub-test.txt"
		content         = strings.NewReader("this is test content")
		public, private *oss.Bucket
		err             error
	)

	if public, err = models.NewOss().NewBucket(true); err == nil {
		err = public.PutObject(testFile, content)
	}
	if err != nil {
		this.Response(map[string]interface{}{"status": 0, "msg": fmt.Sprintf("Bucket(%v)连通失败：%v", public.BucketName, err.Error())})
	}
	public.DeleteObject(testFile)

	if private, err = models.NewOss().NewBucket(false); err == nil {
		err = private.PutObject(testFile, content)
	}
	if err != nil {
		this.Response(map[string]interface{}{"status": 0, "msg": fmt.Sprintf("Bucket(%v)连通失败：%v", private.BucketName, err.Error())})
	}

	private.DeleteObject(testFile)

	this.Response(map[string]interface{}{"status": 1, "msg": "OSS连通成功"})
}
