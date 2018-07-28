package models

import (
	"fmt"

	"github.com/TruthHun/DocHub/helper"

	"strconv"

	"github.com/astaxie/beego/orm"
)

type (
	ConfigCate string
)

const (
	CONFIG_EMAIL         ConfigCate = "email"         //email
	CONFIG_OSS           ConfigCate = "oss"           //oss
	CONFIG_DEPEND        ConfigCate = "depend"        //依赖
	CONFIG_ELASTICSEARCH ConfigCate = "elasticsearch" //全文搜索
	CONFIG_LOGS          ConfigCate = "logs"          //日志配置管理
)
const (
	INPUT_STRING string = "string"   //对应input的text
	INPUT_BOOL   string = "bool"     //对应input的radio，两个选项
	INPUT_NUMBER string = "number"   //对应input的number
	INPUT_TEXT   string = "textarea" //对应textarea
)

//配置管理表
type Config struct {
	Id          int    `orm:"column(Id)"`                                //主键
	Title       string `orm:"column(Title);default()"`                   //名称
	InputType   string `orm:"column(BoolType);default();size(10)"`       //类型：float、int、bool，string(空表示字符串类型)
	Description string `orm:"column(Description);default()"`             //说明
	Key         string `orm:"column(Key);default();size(30)"`            //键
	Value       string `orm:"column(Value);default()"`                   //值
	Category    string `orm:"column(Category);default();index;size(30)"` //分类，如oss、email、redis等
}

//邮箱配置
type configEmail struct {
}

//环境依赖
type configDepend struct {
	Pdf2svg   string //PDF转svg工具
	Soffice   string //libreoffice/openoffice将office文档转PDF文档的工具
	Calibre   string //calibre，将mobi等转PDF
	Pdftotext string //PDF文本提取工具
	Imagick   string //imagick设置，用于转换封面
}

//oss配置
type configOss struct {
}

//全文搜索配置
type configElasticSearch struct {
	On   bool   //是否开启全文搜索
	Host string //全文搜索地址
}

//日志配置
type configLogs struct {
	MaxDays  int //最大保留多长时间
	MaxLines int //一个日志文件，最大多少行日志
}

// 多字段唯一键
func (this *Config) TableUnique() [][]string {
	return [][]string{
		[]string{"Key", "Category"},
	}
}

//获取全部配置文件
func (this *Config) All() (configs []Config) {
	O.QueryTable(TableConfig).All(&configs)
	return
}

//获取配置
//@param            cate            配置分类
//@param            key             键
//@return           val             值
func (this *Config) GetConfig(cate ConfigCate, key string) string {
	if val, ok := helper.GlobalConfigMap.Load(fmt.Sprintf("%v.%v", cate, key)); ok {
		return val.(string)
	}
	return ""
}

//获取配置
//@param            cate            配置分类
//@param            key             键
//@return           val             值
func (this *Config) GetConfigBool(cate ConfigCate, key string) (val bool) {
	value := this.GetConfig(cate, key)
	if value == "true" || value == "1" {
		val = true
	}
	return
}

//获取配置
//@param            cate            配置分类
//@param            key             键
//@return           val             值
func (this *Config) GetConfigInt64(cate ConfigCate, key string) (val int64) {
	val, _ = strconv.ParseInt(this.GetConfig(cate, key), 10, 64)
	return
}

//获取配置
//@param            cate            配置分类
//@param            key             键
//@return           val             值
func (this *Config) GetConfigFloat64(cate ConfigCate, key string) (val float64) {
	val, _ = strconv.ParseFloat(this.GetConfig(cate, key), 64)
	return
}

//更新全局config配置
func (this *Config) UpdateGlobal() {
	if cfgs := this.All(); len(cfgs) > 0 {
		for _, cfg := range cfgs {
			helper.GlobalConfigMap.Store(fmt.Sprintf("%v.%v", cfg.Category, cfg.Key), cfg.Value)
		}
	}
}

//根据key更新配置
func (this *Config) UpdateByKey(cate ConfigCate, key, val string) (err error) {
	_, err = O.QueryTable(TableConfig).Filter("Category", cate).Filter("Key", key).Update(orm.Params{
		"Value": val,
	})
	return
}
