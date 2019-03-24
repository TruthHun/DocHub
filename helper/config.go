package helper

import (
	"fmt"
	"strconv"

	"time"

	"os"

	"io/ioutil"

	"github.com/astaxie/beego"
)

//获取配置
//@param            cate            配置分类
//@param            key             键
//@param			def				default，即默认值
//@return           val             值
func GetConfig(cate ConfigCate, key string, def ...string) string {
	if val, ok := ConfigMap.Load(fmt.Sprintf("%v.%v", cate, key)); ok {
		return val.(string)
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

//获取配置
//@param            cate            配置分类
//@param            key             键
//@return           val             值
func GetConfigBool(cate ConfigCate, key string) (val bool) {
	value := GetConfig(cate, key)
	if value == "true" || value == "1" {
		val = true
	}
	return
}

//获取配置
//@param            cate            配置分类
//@param            key             键
//@return           val             值
func GetConfigInt64(cate ConfigCate, key string) (val int64) {
	val, _ = strconv.ParseInt(GetConfig(cate, key), 10, 64)
	return
}

//获取配置
//@param            cate            配置分类
//@param            key             键
//@return           val             值
func GetConfigFloat64(cate ConfigCate, key string) (val float64) {
	val, _ = strconv.ParseFloat(GetConfig(cate, key), 64)
	return
}

//设置基本配置
func setDefaultConfig() {
	//在程序未安装之前才能设置
	if !IsInstalled {
		beego.BConfig.AppName = "DocHub"     //默认程序名称
		beego.BConfig.Listen.HTTPPort = 8090 //默认监听端口

		//程序安装的时候为开发模式，安装完成之后变为产品模式
		beego.BConfig.RunMode = "dev"   //默认运行模式
		beego.BConfig.EnableGzip = true //开启gzip压缩

		//程序安装的时候不启用，安装完成之后必须启用
		beego.BConfig.WebConfig.EnableXSRF = false                                                           //启用XSRF
		beego.BConfig.WebConfig.XSRFKey = MD5Crypt(fmt.Sprintf("%v", time.Now().UnixNano()) + RandStr(5, 3)) //生成随机key
		beego.BConfig.WebConfig.XSRFExpire = 3600                                                            //过期时间

		//SESSION基本配置
		beego.BConfig.WebConfig.Session.SessionOn = true
		beego.BConfig.WebConfig.Session.SessionName = "DocHub"
		beego.BConfig.WebConfig.Session.SessionProvider = "file"
		beego.BConfig.WebConfig.Session.SessionProviderConfig = "cache/session"

		//DIY配置
		ConfigMap.Store("CookieSecret", RandStr(16, 3))
		ConfigMap.Store("StaticExt", DEFAULT_STATIC_EXT)
	}
}

//生成app.conf配置文件
func GenerateAppConf(host string, port int, username, password, database, prefix string, charset ...string) (err error) {
	if !IsInstalled { //程序未安装状态才能生成app.conf文件
		os.Mkdir("conf", os.ModePerm)
		fileContent := `
# 项目名称
appname = DocHub

# 运行端口
httpport = 8090

# 运行模式：prod, dev【正式站点，务必设置为prod】
runmode = prod

# 开启GZip[建议开启]
EnableGzip=true

# 必须启用XSRF
enablexsrf = true
xsrfkey = %v
xsrfexpire = %v

# cookie的加密密钥
CookieSecret=%v

# 静态文件扩展名【注意：不要把.conf设置为扩展名，以避免泄露数据库账号密码】
StaticExt=%v

############ SESSION #############

# 必须启用session，否则无法登录
sessionon = true

# 使用文件的形式存储session
SessionProvider=file

#Session存放位置
SessionProviderConfig=cache/session

# session的名称
SessionName=dochub

############ SESSION #############



############ 数据库配置 start ############

#数据库配置
[db]

# 数据库host（之前分内网host和外网host，当初只是为了开发方便。但是后来发现，用户在使用的时候会感觉很不方便，所以移除了内网和外网的区分）
host=%v

#数据库端口
port=%v

# 数据库用户名
user=%v

# 数据库密码
password=%v

# 使用的数据库的名称
database=%v

# 表前缀
prefix=%v

# 数据库字符编码
charset=%v

#设置最大空闲连接
maxIdle= 50

#设置最大数据库连接
maxConn= 300


############ 数据库配置 end ############
`
		cs, _ := ConfigMap.Load("CookieSecret")
		se, _ := ConfigMap.Load("StaticExt")
		char := "utf8" //默认字符编码
		if len(charset) > 0 {
			char = charset[0]
		}
		//配置项配置
		fileContent = fmt.Sprintf(
			fileContent,
			beego.BConfig.WebConfig.XSRFKey,
			beego.BConfig.WebConfig.XSRFExpire,
			cs,
			se,
			host, port, username, password, database, prefix, char,
		)
		err = ioutil.WriteFile("conf/app.conf", []byte(fileContent), os.ModePerm)
	}
	return
}
