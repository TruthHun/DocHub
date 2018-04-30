package models

import (
	"fmt"

	"strconv"
)

//配置管理表
type Config struct {
	Id          int    `orm:"column(Id)"`                                //主键
	Title       string `orm:"column(Title);default()"`                   //名称
	Description string `orm:"column(Description);default()"`             //说明
	Key         string `orm:"column(Key);default();size(30)"`            //键
	Value       string `orm:"column(Value);default()"`                   //值
	Category    string `orm:"column(Category);default();index;size(30)"` //分类，如oss、email、redis等
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
func (this *Config) GetConfig(cate string, key string) (val string) {
	var ok bool
	if val, ok = GlobalConfigMap[fmt.Sprintf("%v.%v", cate, key)]; !ok {
		val = ""
	}
	return val
}

//获取配置
//@param            cate            配置分类
//@param            key             键
//@return           val             值
func (this *Config) GetConfigBool(cate string, key string) (val bool) {
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
func (this *Config) GetConfigInt64(cate string, key string) (val int64) {
	val, _ = strconv.ParseInt(this.GetConfig(cate, key), 10, 64)
	return
}

//获取配置
//@param            cate            配置分类
//@param            key             键
//@return           val             值
func (this *Config) GetConfigFloat64(cate string, key string) (val float64) {
	val, _ = strconv.ParseFloat(this.GetConfig(cate, key), 64)
	return
}

//更新全局config配置
func (this *Config) UpdateGlobal() {
	GlobalConfig = this.All()
	if len(GlobalConfig) > 0 {
		for _, cfg := range GlobalConfig {
			GlobalConfigMap[fmt.Sprintf("%v.%v", cfg.Category, cfg.Key)] = cfg.Value
		}
	}
}
