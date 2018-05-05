package helper

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

//日志变量
var Logger = logs.NewLogger()

//日志初始化
func InitLogs() {
	//初始化日志各种配置
	LogsConf := `{"filename":"logs/dochub.log","level":7,"maxlines":5000,"maxsize":0,"daily":true,"maxdays":15}`
	Logger.SetLogger(logs.AdapterFile, LogsConf)
	if beego.AppConfig.String("runmode") == "dev" {
		Logger.SetLogger("console")
	} else {
		//是否异步输出日志
		Logger.Async(1e3)
	}
	Logger.EnableFuncCallDepth(true) //是否显示文件和行号
}
