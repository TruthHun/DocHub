package models

import (
	"fmt"

	"github.com/TruthHun/DocHub/helper"

	"github.com/astaxie/beego/orm"
)

const (
	ConfigCateEmail         helper.ConfigCate = "email"         //email
	ConfigCateDepend        helper.ConfigCate = "depend"        //依赖
	ConfigCateElasticSearch helper.ConfigCate = "elasticsearch" //全文搜索
	ConfigCateLog           helper.ConfigCate = "logs"          //日志配置管理

	// 存储类型, cs 前缀表示 CloudStore
	StoreOss   helper.ConfigCate = "cs-oss"   //oss存储
	StoreMinio helper.ConfigCate = "cs-minio" //minio存储
	StoreCos   helper.ConfigCate = "cs-cos"   //腾讯云存储
	StoreObs   helper.ConfigCate = "cs-obs"   //华为云存储
	StoreBos   helper.ConfigCate = "cs-bos"   //百度云存储
	StoreQiniu helper.ConfigCate = "cs-qiniu" //七牛云储存
	StoreUpyun helper.ConfigCate = "cs-upyun" //又拍云存储
)
const (
	InputText     string = "string"   //对应input的text
	InputBool     string = "bool"     //对应input的radio，两个选项
	InputNumber   string = "number"   //对应input的number
	InputTextarea string = "textarea" //对应textarea
	IinputSelect  string = "select"   //对应textarea
)

//配置管理表
type Config struct {
	Id          int    `orm:"column(Id)"`                                //主键
	Title       string `orm:"column(Title);default()"`                   //名称
	InputType   string `orm:"column(InputType);default();size(10)"`      //类型：float、int、bool，string, textarea (空表示字符串类型)
	Description string `orm:"column(Description);default()"`             //说明
	Key         string `orm:"column(Key);default();size(30)"`            //键
	Value       string `orm:"column(Value);default()"`                   //值
	Category    string `orm:"column(Category);default();index;size(30)"` //分类，如oss、email、redis等
	Options     string `orm:"column(Options);default();size(4096)"`      //枚举值列举
}

func NewConfig() *Config {
	return &Config{}
}

func GetTableConfig() string {
	return getTable("config")
}

// 多字段唯一键
func (this *Config) TableUnique() [][]string {
	return [][]string{
		[]string{"Key", "Category"},
	}
}

//获取全部配置文件
//@return           configs         所有配置
func (this *Config) All() (configs []Config) {
	orm.NewOrm().QueryTable(GetTableConfig()).All(&configs)
	return
}

//更新全局config配置
func (this *Config) UpdateGlobal() {
	if cfgs := this.All(); len(cfgs) > 0 {
		for _, cfg := range cfgs {
			helper.ConfigMap.Store(fmt.Sprintf("%v.%v", cfg.Category, cfg.Key), cfg.Value)
		}
	} else {
		helper.Logger.Error("查询全局配置失败，config表中全局配置信息为空")
	}
}

//根据key更新配置
//@param            cate            配置分类
//@param            key             配置项
//@param            val             配置项的值
//@return           err             错误
func (this *Config) UpdateByKey(cate helper.ConfigCate, key, val string) (err error) {
	_, err = orm.NewOrm().QueryTable(GetTableConfig()).Filter("Category", cate).Filter("Key", key).Update(orm.Params{
		"Value": val,
	})
	return
}
