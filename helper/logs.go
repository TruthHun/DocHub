package helper

import (
	"os"

	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

//日志变量
var Logger = logs.NewLogger()

//日志初始化
func InitLogs() {
	//创建日志目录
	if _, err := os.Stat("logs"); err != nil {
		os.Mkdir("logs", os.ModePerm)
	}
	var level = 7
	if Debug {
		level = 4
	}
	maxLines := GetConfigInt64("logs", "max_lines")
	if maxLines <= 0 {
		maxLines = 10000
	}
	maxDays := GetConfigInt64("logs", "max_days")
	if maxDays <= 0 {
		maxDays = 7
	}
	//初始化日志各种配置
	LogsConf := fmt.Sprintf(`{"filename":"logs/dochub.log","level":%v,"maxlines":%v,"maxsize":0,"daily":true,"maxdays":%v}`, level, maxLines, maxDays)
	Logger.SetLogger(logs.AdapterFile, LogsConf)
	if Debug {
		Logger.SetLogger("console")
		beego.Info("日志配置信息：" + LogsConf)
	} else {
		//是否异步输出日志
		Logger.Async(1e3)
	}
	Logger.EnableFuncCallDepth(true) //是否显示文件和行号
}
