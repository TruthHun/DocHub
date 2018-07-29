package helper

import (
	"fmt"
	"strconv"
)

//获取配置
//@param            cate            配置分类
//@param            key             键
//@param			def				default，即默认值
//@return           val             值
func GetConfig(cate string, key string, def ...string) string {
	if val, ok := GlobalConfigMap.Load(fmt.Sprintf("%v.%v", cate, key)); ok {
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
func GetConfigBool(cate string, key string) (val bool) {
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
func GetConfigInt64(cate string, key string) (val int64) {
	val, _ = strconv.ParseInt(GetConfig(cate, key), 10, 64)
	return
}

//获取配置
//@param            cate            配置分类
//@param            key             键
//@return           val             值
func GetConfigFloat64(cate string, key string) (val float64) {
	val, _ = strconv.ParseFloat(GetConfig(cate, key), 64)
	return
}
