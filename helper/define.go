//定义一些常量和变量
package helper

import (
	"net/url"
	"sync"

	"github.com/astaxie/beego"
	"github.com/huichen/sego"
)

const (
	//DocHub Version
	VERSION = "v2.3"
	//Cache Config
	CACHE_CONF = `{"CachePath":"./cache/runtime","FileSuffix":".cache","DirectoryLevel":2,"EmbedExpiry":120}`

	DEFAULT_STATIC_EXT    = ".txt,.html,.ico,.jpeg,.png,.gif,.xml"
	DEFAULT_COOKIE_SECRET = "dochub"

	//	扩展名
	EXT_CATE_WORD       = "word"
	EXT_NUM_WORD        = 1
	EXT_CATE_PPT        = "ppt"
	EXT_NUM_PPT         = 2
	EXT_CATE_EXCEL      = "excel"
	EXT_NUM_EXCEL       = 3
	EXT_CATE_PDF        = "pdf"
	EXT_NUM_PDF         = 4
	EXT_CATE_TEXT       = "text"
	EXT_NUM_TEXT        = 5
	EXT_CATE_OTHER      = "other"
	EXT_NUM_OTHER       = 6
	EXT_CATE_OTHER_MOBI = "mobi"
	EXT_CATE_OTHER_EPUB = "epub"
	EXT_CATE_OTHER_CHM  = "chm"
	EXT_CATE_OTHER_UMD  = "umd"

	RootPath = "./virtualroot" // 虚拟根目录
)

type ConfigCate string

const (
	//word
	ExtDOC  = ".doc"
	ExtDOCX = ".docx"
	ExtRTF  = ".rtf"
	ExtWPS  = ".wps"
	ExtODT  = ".odt"

	// power point
	ExtPPT  = ".ppt"
	ExtPPTX = ".pptx"
	ExtPPS  = ".pps"
	ExtPPSX = ".ppsx"
	ExtDPS  = ".dps"
	ExtODP  = ".odp"
	ExtPOT  = ".pot"

	// excel
	ExtXLS  = ".xls"
	ExtXLSX = ".xlsx"
	ExtET   = ".et"
	ExtODS  = ".ods"

	// PDF
	ExtPDF = ".pdf"

	// text
	ExtTXT = ".txt"

	// other
	ExtEPUB = ".epub"
	ExtUMD  = ".umd"
	ExtMOBI = ".mobi"
	ExtCHM  = ".chm"
)

var (
	//develop mode
	Debug = beego.AppConfig.String("runmode") == "dev"

	//允许直接访问的文件扩展名
	StaticExt = make(map[string]bool)

	//分词器
	Segmenter sego.Segmenter

	//配置文件的全局map
	ConfigMap sync.Map

	//程序是否已经安装
	IsInstalled = false

	//允许上传的文档扩展名
	//AllowedUploadExt = ",doc,docx,rtf,wps,odt,ppt,pptx,pps,ppsx,dps,odp,pot,xls,xlsx,et,ods,txt,pdf,chm,epub,umd,mobi,"
	AllowedUploadDocsExt = map[string]bool{
		".doc": true, ".docx": true, ".rtf": true, ".wps": true, ".odt": true, //word
		".ppt": true, ".pptx": true, ".pps": true, ".ppsx": true, ".dps": true, ".odp": true, ".pot": true, // power point
		".xls": true, ".xlsx": true, ".et": true, ".ods": true, // excel
		".pdf":  true,                              //pdf
		".epub": true, ".mobi": true, ".txt": true, //other
		".umd": true, ".chm": true, //不能转化的电子书
	}

	// 图片尺寸
	CoverWidth   = beego.AppConfig.DefaultInt("cover_width", 140)
	CoverHeight  = beego.AppConfig.DefaultInt("cover_height", 200)
	BannerWidth  = beego.AppConfig.DefaultInt("banner_width", 825)
	BannerHeight = beego.AppConfig.DefaultInt("banner_height", 316)
	AvatarWidth  = beego.AppConfig.DefaultInt("avatar_width", 120)
	AvatarHeight = beego.AppConfig.DefaultInt("avatar_height", 120)
)

var (
	HeaderGzip = map[string]string{"Content-Encoding": "gzip"}
	HeaderSVG  = map[string]string{"Content-Type": "image/svg+xml"}
	HeaderPNG  = map[string]string{"Content-type": "image/png"}
	HeaderJPEG = map[string]string{"Content-type": "image/jpeg"}
)

func HeaderDisposition(name string) map[string]string {
	return map[string]string{"Content-Disposition": "attachment; filename=" + url.PathEscape(name)}
}
