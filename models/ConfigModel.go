package models

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"

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

type ConfigOss struct {
	AccessKey           string `dochub:"access-key"`
	SecretKey           string `dochub:"secret-key"`
	Endpoint            string `dochub:"endpoint"`
	PublicBucket        string `dochub:"public-bucket"`
	PublicBucketDomain  string `dochub:"public-bucket-domain"`
	PrivateBucket       string `dochub:"private-bucket"`
	PrivateBucketDomain string `dochub:"private-bucket-domain"`
	Expire              int64  `dochub:"expire"`
}

type ConfigMinio struct {
	AccessKey           string `dochub:"access-key"`
	SecretKey           string `dochub:"secret-key"`
	Endpoint            string `dochub:"endpoint"`
	PublicBucket        string `dochub:"public-bucket"`
	PublicBucketDomain  string `dochub:"public-bucket-domain"`
	PrivateBucket       string `dochub:"private-bucket"`
	PrivateBucketDomain string `dochub:"private-bucket-domain"`
	Expire              int64  `dochub:"expire"`
}

type ConfigCos struct {
	AccessKey           string `dochub:"access-key"`
	SecretKey           string `dochub:"secret-key"`
	Region              string `dochub:"region"`
	AppId               string `dochub:"app-id"`
	PublicBucket        string `dochub:"public-bucket"`
	PublicBucketDomain  string `dochub:"public-bucket-domain"`
	PrivateBucket       string `dochub:"private-bucket"`
	PrivateBucketDomain string `dochub:"private-bucket-domain"`
	Expire              int64  `dochub:"expire"`
}

type ConfigBos struct {
	AccessKey           string `dochub:"access-key"`
	SecretKey           string `dochub:"secret-key"`
	Endpoint            string `dochub:"endpoint"`
	PublicBucket        string `dochub:"public-bucket"`
	PublicBucketDomain  string `dochub:"public-bucket-domain"`
	PrivateBucket       string `dochub:"private-bucket"`
	PrivateBucketDomain string `dochub:"private-bucket-domain"`
	Expire              int64  `dochub:"expire"`
}

type ConfigObs struct {
	AccessKey           string `dochub:"access-key"`
	SecretKey           string `dochub:"secret-key"`
	Endpoint            string `dochub:"endpoint"`
	PublicBucket        string `dochub:"public-bucket"`
	PublicBucketDomain  string `dochub:"public-bucket-domain"`
	PrivateBucket       string `dochub:"private-bucket"`
	PrivateBucketDomain string `dochub:"private-bucket-domain"`
	Expire              int64  `dochub:"expire"`
}

type ConfigQiniu struct {
	AccessKey           string `dochub:"access-key"`
	SecretKey           string `dochub:"secret-key"`
	PublicBucket        string `dochub:"public-bucket"`
	PublicBucketDomain  string `dochub:"public-bucket-domain"`
	PrivateBucket       string `dochub:"private-bucket"`
	PrivateBucketDomain string `dochub:"private-bucket-domain"`
	Expire              int64  `dochub:"expire"`
}

type ConfigUpYun struct {
	Operator            string `dochub:"operator"`
	Password            string `dochub:"password"`
	Secret              string `dochub:"secret"`
	PublicBucket        string `dochub:"public-bucket"`
	PublicBucketDomain  string `dochub:"public-bucket-domain"`
	PrivateBucket       string `dochub:"private-bucket"`
	PrivateBucketDomain string `dochub:"private-bucket-domain"`
	Expire              int64  `dochub:"expire"`
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

// 获取云存储配置
func (this *Config) GetCloudStoreConfig(storeType helper.ConfigCate) (cfg interface{}) {
	switch storeType {
	case StoreCos:
		cfg = &ConfigCos{}
	case StoreBos:
		cfg = &ConfigBos{}
	case StoreOss:
		cfg = &ConfigOss{}
	case StoreMinio:
		cfg = &ConfigMinio{}
	case StoreUpyun:
		cfg = &ConfigUpYun{}
	case StoreQiniu:
		cfg = &ConfigQiniu{}
	case StoreObs:
		cfg = &ConfigObs{}
	}
	t := reflect.TypeOf(cfg)
	v := reflect.ValueOf(cfg)
	numFields := t.Elem().NumField()
	for i := 0; i < numFields; i++ {
		if v.Elem().Field(i).CanSet() {
			key := t.Elem().Field(i).Tag.Get("dochub")
			switch t.Elem().Field(i).Type.Kind() {
			case reflect.String:
				v.Elem().Field(i).Set(reflect.ValueOf(helper.GetConfig(storeType, key)))
			case reflect.Int64:
				v.Elem().Field(i).Set(reflect.ValueOf(helper.GetConfigInt64(storeType, key)))
			case reflect.Float64:
				v.Elem().Field(i).Set(reflect.ValueOf(helper.GetConfigFloat64(storeType, key)))
			case reflect.Bool:
				v.Elem().Field(i).Set(reflect.ValueOf(helper.GetConfigBool(storeType, key)))
			}
		}
	}
	return
}

func (this *Config) ParseForm(storeType helper.ConfigCate, form url.Values) (cfg interface{}, err error) {
	switch storeType {
	case StoreCos:
		cfg = &ConfigCos{}
	case StoreBos:
		cfg = &ConfigBos{}
	case StoreOss:
		cfg = &ConfigOss{}
	case StoreMinio:
		cfg = &ConfigMinio{}
	case StoreUpyun:
		cfg = &ConfigUpYun{}
	case StoreQiniu:
		cfg = &ConfigQiniu{}
	case StoreObs:
		cfg = &ConfigObs{}
	}
	t := reflect.TypeOf(cfg)
	v := reflect.ValueOf(cfg)
	numFields := t.Elem().NumField()
	for i := 0; i < numFields; i++ {
		if v.Elem().Field(i).CanSet() {
			key := t.Elem().Field(i).Tag.Get("dochub")
			val := form.Get(key)
			switch t.Elem().Field(i).Type.Kind() {
			case reflect.String:
				v.Elem().Field(i).Set(reflect.ValueOf(val))
			case reflect.Int64:
				intVal, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return nil, err
				}
				v.Elem().Field(i).Set(reflect.ValueOf(intVal))
			case reflect.Float64:
				floatVal, err := strconv.ParseFloat(val, 10)
				if err != nil {
					return nil, err
				}
				v.Elem().Field(i).Set(reflect.ValueOf(floatVal))
			case reflect.Bool:
				boolVal, err := strconv.ParseBool(val)
				if err != nil {
					return nil, err
				}
				v.Elem().Field(i).Set(reflect.ValueOf(boolVal))
			}
		}
	}
	return
}

// 注意：这里的 cfg 参数是值传递
func (this *Config) UpdateCloudStore(storeType helper.ConfigCate, cfg interface{}) (err error) {
	t := reflect.TypeOf(cfg)
	v := reflect.ValueOf(cfg)
	numFields := t.Elem().NumField()
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if err == nil {
			o.Commit()
		} else {
			o.Rollback()
		}
	}()
	for i := 0; i < numFields; i++ {
		key := t.Elem().Field(i).Tag.Get("dochub")
		params := orm.Params{"Value": v.Elem().Field(i).Interface()}
		_, err = o.QueryTable(this).Filter("Key", key).Filter("Category", storeType).Update(params)
		if err != nil {
			return
		}
	}
	this.UpdateGlobalConfig()
	return
}

//更新全局config配置
func (this *Config) UpdateGlobalConfig() {
	if cfgs := this.All(); len(cfgs) > 0 {
		for _, cfg := range cfgs {
			helper.ConfigMap.Store(fmt.Sprintf("%v.%v", cfg.Category, cfg.Key), cfg.Value)
		}
	} else {
		helper.Logger.Error("查询全局配置失败，config表中全局配置信息为空")
	}
}

func (this *Config) GetByCate(cate helper.ConfigCate) (configs []Config) {
	orm.NewOrm().QueryTable(GetTableConfig()).Filter("Category", cate).All(&configs)
	return
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
