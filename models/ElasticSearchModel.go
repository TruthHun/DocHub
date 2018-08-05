package models

import (
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"net/http"

	"encoding/json"
	"strconv"

	"fmt"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/orm"
)

//全文搜索客户端
type ElasticSearchClient struct {
	Host    string        //host
	Index   string        //索引
	Type    string        //type
	On      bool          //是否启用全文搜索
	Timeout time.Duration //超时时间
}

//全文搜索
type ElasticSearchData struct {
	Id          int    `json:"Id"`          //文档id
	Title       string `json:"Title"`       //文档标题
	Keywords    string `json:"Keywords"`    //文档关键字
	Description string `json:"Description"` //文档摘要
	Vcnt        int    `json:"Vcnt"`        //文档浏览次数	view count
	Ccnt        int    `json:"Ccnt"`        //文档收藏次数 collect count
	Dcnt        int    `json:"Dcnt"`        //文档下载次数，download count
	Score       int    `json:"Score"`       //文档评分
	Size        int    `json:"Size"`        //文档大小
	Page        int    `json:"Page"`        //文档页数
	DocType     int    `json:"DocType"`     //文档类型，对应各格式的数字表示
	DsId        int    `json:"DsId"`        //DocumentStoreId，对应于md5
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
		Type:    "fulltext",
		On:      helper.GetConfigBool(cateES, "on"),
		Timeout: time.Duration(timeout) * time.Second,
	}
	client.Host = strings.TrimRight(client.Host, "/") + "/"
	return
}

//初始化全文搜索客户端，包括检查索引是否存在，mapping设置等
func (this *ElasticSearchClient) Init() (err error) {
	if !this.On { //未开启ElasticSearch，则不初始化
		return
	}
	//检测是否能ping同
	if err = this.ping(); err == nil {
		//检测索引是否存在；索引不存在，则创建索引；如果索引存在，则直接跳过初始化
		if err = this.existIndex(); err != nil {
			//创建索引成功
			if err = this.createIndex(); err == nil {
				//创建mapping
				js := `{
	"properties": {
		"Title": {
			"type": "text",
			"analyzer": "ik_max_word",
			"search_analyzer": "ik_max_word"
		},
		"Keywords": {
			"type": "text",
			"analyzer": "ik_max_word",
			"search_analyzer": "ik_max_word"
		},
		"Description": {
			"type": "text",
			"analyzer": "ik_max_word",
			"search_analyzer": "ik_max_word"
		},
		"Vcnt": {
			"type": "integer"
		},
		"Ccnt": {
			"type": "integer"
		},
		"Dcnt": {
			"type": "integer"
		},
		"Score": {
			"type": "integer"
		},
		"Size": {
			"type": "integer"
		},
		"Page": {
			"type": "integer"
		},
		"DocType": {
			"type": "integer"
		},
      	"DsId": {
			"type": "integer"
		}
	}
}`
				if helper.Debug {
					beego.Debug(" ==== ElasticSearch初始化mapping ==== ")
					beego.Info(js)
					beego.Debug(" ==== ElasticSearch初始化mapping ==== ")
				}
				api := this.Host + this.Index + "/" + this.Type + "/_mapping"
				req := this.post(api)
				if resp, errResp := req.Header("Content-Type", "application/json").Body(js).Response(); errResp != nil {
					err = errResp
				} else {
					if resp.StatusCode >= 300 || resp.StatusCode < 200 {
						err = errors.New(resp.Status)
					}
				}
			}
		}
	}
	return
}

//重建索引【全量】
func (this *ElasticSearchClient) RebuildAllIndex() {
	helper.GlobalConfigMap.Store("indexing", true)
	defer helper.GlobalConfigMap.Store("indexing", false)
	//执行全量索引更新
	pageSize := 100
	maxPage := 100000
	for page := 1; page < maxPage; page++ {
		if infos, rows, err := ModelDoc.GetDocInfoForElasticSearch(page, pageSize, 0); err != nil || rows == 0 {
			if err != nil && err != orm.ErrNoRows {
				helper.Logger.Error(err.Error())
			}
			page = maxPage
		} else {
			for _, info := range infos {
				//开始创建索引的时间
				tStart := time.Now().Unix()
				if err := this.BuildIndexById(info.Id); err != nil {
					helper.Logger.Error("创建索引失败：" + err.Error())
				}
				//创建索引完成时间
				tEnd := time.Now().Unix()
				//如果创建索引响应时长达到1秒钟，则休眠一定时长，以避免服务器负载过高，造成整站无法正常运行
				if t := tEnd - tStart; t >= 1 {
					time.Sleep(time.Duration(t) * time.Second)
				}
			}
		}
	}
}

//根据id查询文档，并创建索引
func (this *ElasticSearchClient) BuildIndexById(id int) (err error) {
	if es, errES := ModelDoc.GetDocForElasticSearch(id); errES != nil {
		err = errES
	} else {
		err = this.BuildIndex(es)
	}
	return
}

//新增/更新索引
func (this *ElasticSearchClient) BuildIndex(es ElasticSearchData) (err error) {
	var (
		js   []byte
		resp *http.Response
	)
	if !this.On {
		return
	}
	if helper.Debug {
		beego.Info("创建索引--------start--------")
		fmt.Printf("%+v\n", es)
		beego.Info("创建索引-------- end --------")
	}
	api := this.Host + this.Index + "/" + this.Type + "/" + strconv.Itoa(es.Id)
	if js, err = json.Marshal(es); err == nil {
		if resp, err = this.post(api).Body(js).Response(); err == nil {
			if resp.StatusCode >= 300 || resp.StatusCode < 200 {
				b, _ := ioutil.ReadAll(resp.Body)
				err = errors.New("生成索引失败：" + resp.Status + "；" + string(b))
			}
		}
	}
	return
}

//搜索内容
func (this *ElasticSearchClient) Search() {

}

//删除索引
//@param            id          索引id
//@return           err         错误
func (this *ElasticSearchClient) DeleteIndex(id int) (err error) {
	api := this.Host + this.Index + "/" + this.Type + "/" + strconv.Itoa(id)
	if resp, errResp := this.delete(api).Response(); errResp != nil {
		err = errResp
	} else {
		if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			b, _ := ioutil.ReadAll(resp.Body)
			err = errors.New("删除索引失败：" + resp.Status + "；" + string(b))
		}
	}
	return
}

//检验es服务能否连通
func (this *ElasticSearchClient) ping() error {
	if resp, err := this.get(this.Host).Response(); err != nil {
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
//@return			err				nil表示索引存在，否则表示不存在
func (this *ElasticSearchClient) existIndex() (err error) {
	var resp *http.Response
	api := this.Host + this.Index
	if resp, err = this.get(api).Response(); err == nil {
		if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			b, _ := ioutil.ReadAll(resp.Body)
			err = errors.New(resp.Status + "：" + string(b))
		}
	}
	return
}

//创建索引
//@return           err             创建索引
func (this *ElasticSearchClient) createIndex() (err error) {
	var resp *http.Response
	api := this.Host + this.Index
	if resp, err = this.put(api).Response(); err == nil {
		if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			b, _ := ioutil.ReadAll(resp.Body)
			err = errors.New(resp.Status + "：" + string(b))
		}
	}
	return
}

//put请求
func (this *ElasticSearchClient) put(api string) (req *httplib.BeegoHTTPRequest) {
	return httplib.Put(api).Header("Content-Type", "application/json").SetTimeout(this.Timeout, this.Timeout)
}

//post请求
func (this *ElasticSearchClient) post(api string) (req *httplib.BeegoHTTPRequest) {
	return httplib.Post(api).Header("Content-Type", "application/json").SetTimeout(this.Timeout, this.Timeout)
}

//delete请求
func (this *ElasticSearchClient) delete(api string) (req *httplib.BeegoHTTPRequest) {
	return httplib.Delete(api).Header("Content-Type", "application/json").SetTimeout(this.Timeout, this.Timeout)
}

//get请求
func (this *ElasticSearchClient) get(api string) (req *httplib.BeegoHTTPRequest) {
	return httplib.Get(api).Header("Content-Type", "application/json").SetTimeout(this.Timeout, this.Timeout)
}
