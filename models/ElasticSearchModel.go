package models

import (
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"net/http"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego/httplib"
)

//全文搜索客户端
type ElasticSearchClient struct {
	Host    string        //host
	Index   string        //索引
	On      bool          //是否启用全文搜索
	Timeout time.Duration //超时时间
}

//全文搜索
type ElasticSearchData struct {
	Id          int    //文档id
	Title       string //文档标题
	Keywords    string //文档关键字
	Description string //文档摘要
	Vcnt        int    //文档浏览次数	view count
	Ccnt        int    //文档收藏次数 collect count
	Dcnt        int    //文档下载次数，download count
	Score       int    //文档评分
	Size        int    //文档大小
	Page        int    //文档页数
	DocType     int    //文档类型，对应各格式的数字表示
	DsId        int    //DocumentStoreId，对应于md5
}

//创建全文搜索客户端
func NewElasticSearchClient() (client *ElasticSearchClient) {
	cateES := string(CONFIG_ELASTICSEARCH)
	//并未设置超时配置项
	timeout := helper.GetConfigInt64(cateES, "timeout")
	if timeout <= 0 { //默认超时时间为5秒
		timeout = 5
	}
	client = &ElasticSearchClient{
		Host:    helper.GetConfig(cateES, "host", "http://localhost:920/"),
		Index:   helper.GetConfig(cateES, "index", "dochub"),
		On:      helper.GetConfigBool(cateES, "on"),
		Timeout: time.Duration(timeout) * time.Second,
	}
	client.Host = strings.TrimRight(client.Host, "/") + "/"
	return
}

//初始化全文搜索客户端，包括检查索引是否存在，mapping设置等
func (this *ElasticSearchClient) Init() (err error) {
	//检测是否能ping同
	if err = this.ping(); err == nil {
		//TODO
		//检测索引是否存在

		//索引不存在，创建索引、创建mapping
	}
	return
}

//重建索引【全量】
func (this *ElasticSearchClient) RebuildAllIndex() {

}

//新增/更新索引
func (this *ElasticSearchClient) BuildIndex(es []ElasticSearchData) {

}

//搜索内容
func (this *ElasticSearchClient) Search() {

}

//检验es服务能否连通
func (this *ElasticSearchClient) ping() error {
	req := httplib.Get(this.Host).SetTimeout(this.Timeout, this.Timeout)
	if resp, err := req.Response(); err != nil {
		return err
	} else {
		if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			err = errors.New(resp.Status + "；" + string(body))
		}
	}
	return nil
}

//查询索引是否存在
//@param			err				nil表示索引存在，否则表示不存在
func (this *ElasticSearchClient) existIndex() (err error) {
	var resp *http.Response
	api := this.Host + this.Index
	if resp, err = httplib.Get(api).SetTimeout(this.Timeout, this.Timeout).Response(); err == nil {
		if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			b, _ := ioutil.ReadAll(resp.Body)
			err = errors.New(resp.Status + "：" + string(b))
		}
	}
	return
}

//创建索引
func (this *ElasticSearchClient) createIndex() (err error) {
	var resp *http.Response
	api := this.Host + this.Index
	if resp, err = httplib.Put(api).SetTimeout(this.Timeout, this.Timeout).Response(); err == nil {
		if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			b, _ := ioutil.ReadAll(resp.Body)
			err = errors.New(resp.Status + "：" + string(b))
		}
	}
	return
}

//生成索引
