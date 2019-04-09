package AdminControllers

import (
	"net/http"
	"strings"

	"io/ioutil"
	"time"

	"path/filepath"

	"os"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
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
	tab := helper.ConfigCate(strings.ToLower(this.GetString("tab")))
	var err error
	switch tab {
	case models.ConfigCateEmail, models.ConfigCateDepend, models.ConfigCateElasticSearch, models.ConfigCateLog:
	default:
		tab = "default"
	}
	if this.Ctx.Request.Method == http.MethodPost {
		defer func() {
			models.NewSys().UpdateGlobalConfig()
			models.NewConfig().UpdateGlobalConfig()
		}()
		if tab == "default" {
			var sys models.Sys
			this.ParseForm(&sys)
			_, err := orm.NewOrm().Update(&sys)
			if err != nil {
				helper.Logger.Error(err.Error())
				this.ResponseJson(false, err.Error())
			}
			this.ResponseJson(true, "更新成功")
		}

		var cfg interface{}
		cfg, err = models.NewConfig().ParseForm(tab, this.Ctx.Request.Form)
		if cfg != nil {
			switch tab {
			case models.ConfigCateEmail:
				if err != nil {
					this.ResponseJson(false, err.Error())
				}
				modelEmail := cfg.(*models.ConfigEmail)
				err = modelEmail.SendMail(modelEmail.TestUserEmail, "测试邮件", "这是一封测试邮件，用于检测是否能正常发送邮件")
			case models.ConfigCateElasticSearch:
				modelES := cfg.(*models.ElasticSearchClient)
				if modelES.On {
					modelES.Host = strings.TrimRight(modelES.Host, "/") + "/"
					modelES.Type = "fulltext"
					modelES.Timeout = 5 * time.Second
					err = modelES.Init()
				}
			}
		}
		if err != nil {
			this.ResponseJson(false, err.Error(), cfg)
		}

		o := orm.NewOrm()
		o.Begin()
		defer func() {
			if err != nil {
				o.Rollback()
			} else {
				o.Commit()
			}
		}()

		for k, v := range this.Ctx.Request.Form {
			if _, err = o.QueryTable(models.GetTableConfig()).Filter("Category", tab).Filter("Key", k).Update(orm.Params{"Value": v[0]}); err != nil {
				helper.Logger.Error(err.Error())
				this.ResponseJson(false, "ElasticSearch初始化失败："+err.Error())
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
		if tab == models.ConfigCateElasticSearch {
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
	to := helper.GetConfig(models.ConfigCateEmail, "test")
	if err := models.NewEmail().SendMail(to, "测试邮件", "这是一封测试邮件，用于检测是否能正常发送邮件"); err != nil {
		this.Response(map[string]interface{}{"status": 0, "msg": "邮件发送失败：" + err.Error()})
	}
	this.Response(map[string]interface{}{"status": 1, "msg": "邮件发送成功"})
}

// 云存储配置
func (this *SysController) CloudStore() {
	tab := this.GetString("tab", models.GlobalSys.StoreType)
	modelConfig := models.NewConfig()
	this.Data["Config"] = modelConfig.GetByCate(helper.ConfigCate(tab))
	this.Data["Tab"] = tab
	this.Data["IsCloudStore"] = true
	this.TplName = "cloud-store.html"
}

func (this *SysController) SetCloudStore() {
	storeType := helper.ConfigCate(this.GetString("tab", "cs-oss"))
	if storeType == "" {
		this.ResponseJson(false, "参数错误：存储类别不正确")
	}
	modelConfig := models.NewConfig()
	config, err := modelConfig.ParseForm(storeType, this.Ctx.Request.Form)
	if err != nil {
		this.ResponseJson(false, err.Error(), config)
	}

	csPublic, err := models.NewCloudStoreWithConfig(config, storeType, false)
	if err != nil {
		this.ResponseJson(false, err.Error(), config)
	}

	if err = csPublic.PingTest(); err != nil {
		this.ResponseJson(false, err.Error(), config)
	}

	csPrivate, err := models.NewCloudStoreWithConfig(config, storeType, true)
	if err != nil {
		this.ResponseJson(false, err.Error(), config)
	}

	if err = csPrivate.PingTest(); err != nil {
		this.ResponseJson(false, err.Error(), config)
	}

	err = modelConfig.UpdateCloudStore(storeType, config)
	if err != nil {
		this.ResponseJson(false, err.Error(), config)
	}
	this.ResponseJson(true, "更新成功", config)
}
