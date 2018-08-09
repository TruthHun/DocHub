package models

import (
	"fmt"

	"github.com/TruthHun/DocHub/helper"

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
			helper.GlobalConfigMap.Store(fmt.Sprintf("%v.%v", cfg.Category, cfg.Key), cfg.Value)
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
func (this *Config) UpdateByKey(cate ConfigCate, key, val string) (err error) {
	_, err = orm.NewOrm().QueryTable(GetTableConfig()).Filter("Category", cate).Filter("Key", key).Update(orm.Params{
		"Value": val,
	})
	return
}
