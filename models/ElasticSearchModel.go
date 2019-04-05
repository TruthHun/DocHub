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
	"github.com/TruthHun/gotil/util"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/orm"
)

//全文搜索客户端
type ElasticSearchClient struct {
	On      bool          `dochub:"on"`    //是否启用全文搜索
	Host    string        `dochub:"host"`  //host
	Index   string        `dochub:"index"` //索引
	Type    string        `dochub:"type"`  //type
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
	Price       int    `json:"Price"`       //文档售价
	TimeCreate  int    `json:"TimeCreate"`  //文档创建时间
}

//统计信息结构
type ElasticSearchCount struct {
	Shards struct {
		Failed     int `json:"failed"`
		Skipped    int `json:"skipped"`
		Successful int `json:"successful"`
		Total      int `json:"total"`
	} `json:"_shards"`
	Count int `json:"count"`
}

//搜索结果结构
type ElasticSearchResult struct {
	Shards struct {
		Failed     int `json:"failed"`
		Skipped    int `json:"skipped"`
		Successful int `json:"successful"`
		Total      int `json:"total"`
	} `json:"_shards"`
	Hits struct {
		Hits []struct {
			ID     string      `json:"_id"`
			Index  string      `json:"_index"`
			Score  interface{} `json:"_score"`
			Source struct {
				Ccnt        int    `json:"Ccnt"`
				Dcnt        int    `json:"Dcnt"`
				Description string `json:"Description"`
				DocType     int    `json:"DocType"`
				DsID        int    `json:"DsId"`
				ID          int    `json:"Id"`
				Keywords    string `json:"Keywords"`
				Page        int    `json:"Page"`
				Score       int    `json:"Score"`
				Size        int    `json:"Size"`
				Title       string `json:"Title"`
				Vcnt        int    `json:"Vcnt"`
				Price       int    `json:"Price"`
				TimeCreate  int    `json:"TimeCreate"`
			} `json:"_source"`
			Type      string `json:"_type"`
			Highlight struct {
				Title       []string `json:"Title"`
				Keywords    []string `json:"Keywords"`
				Description []string `json:"Description"`
			} `json:"highlight"`
			Sort []int `json:"sort"`
		} `json:"hits"`
		MaxScore interface{} `json:"max_score"`
		Total    int         `json:"total"`
	} `json:"hits"`
	TimedOut bool `json:"timed_out"`
	Took     int  `json:"took"`
}

//创建全文搜索客户端
func NewElasticSearchClient(configElasticSearch ...ElasticSearchClient) (client *ElasticSearchClient) {
	if len(configElasticSearch) > 0 {
		client = &configElasticSearch[0]
	}
	//并未设置超时配置项
	timeout := helper.GetConfigInt64(ConfigCateElasticSearch, "timeout")
	if timeout <= 0 { //默认超时时间为10秒
		timeout = 10
	}
	client = &ElasticSearchClient{
		Host:    helper.GetConfig(ConfigCateElasticSearch, "host", "http://localhost:9200/"),
		Index:   helper.GetConfig(ConfigCateElasticSearch, "index", "dochub"),
		Type:    "fulltext",
		On:      helper.GetConfigBool(ConfigCateElasticSearch, "on"),
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
			"search_analyzer": "ik_smart"
		},
		"Keywords": {
			"type": "text",
			"analyzer": "ik_max_word",
			"search_analyzer": "ik_smart"
		},
		"Description": {
			"type": "text",
			"analyzer": "ik_max_word",
			"search_analyzer": "ik_smart"
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

//搜索内容
//@param            wd          搜索关键字
//@param            sourceType  搜索的资源类型，可选择：doc、ppt、xls、pdf、txt、other、all
//@param            order       排序，可选值：new(最新)、down(下载)、page(页数)、score(评分)、size(大小)、collect(收藏)、view（浏览）、default(默认)
//@param            p           页码
//@param            listRows    每页显示记录数
func (this *ElasticSearchClient) Search(wd, sourceType, order string, p, listRows int) (result ElasticSearchResult, err error) {
	wd = strings.Replace(wd, "\"", " ", -1)
	wd = strings.Replace(wd, "\\", " ", -1)
	filter := ""
	//搜索资源类型，0表示全部类型（即不限资源类型）
	ExtNum := 0
	switch strings.ToLower(sourceType) {
	case "doc":
		ExtNum = helper.EXT_NUM_WORD
	case "ppt":
		ExtNum = helper.EXT_NUM_PPT
	case "xls":
		ExtNum = helper.EXT_NUM_EXCEL
	case "pdf":
		ExtNum = helper.EXT_NUM_PDF
	case "txt":
		ExtNum = helper.EXT_NUM_TEXT
	case "other":
		ExtNum = helper.EXT_NUM_OTHER
	}
	if ExtNum > 0 {
		filter = fmt.Sprintf(`,"filter":{"term":{"DocType":%v}}`, ExtNum)
	}

	//搜索结果排序
	sort := ""
	field := ""
	switch strings.ToLower(order) {
	case "new":
		field = "Id"
	case "down":
		field = "Dcnt"
	case "page":
		field = "Page"
	case "score":
		field = "Score"
	case "size":
		field = "Size"
	case "collect":
		field = "Ccnt"
	case "view":
		field = "Vcnt"
	}
	if field != "" {
		sort = fmt.Sprintf(`"sort":[{"%v":"desc"}],`, field)
	}

	//我elasticsearch不熟，只能这么用了...尴尬
	queryBody := `
{
  "query": {
    "bool":{
      "must":{
        "multi_match" : {
          "query":    "%v", 
          "fields": [ "Title", "Keywords","Description" ] 
        }
      }%v
    }
  },
  "from":%v,
  "size":%v,
  %v
  "highlight": {
        "fields" : {
          "Title" : {},
          "Description" : {}
        }
   }
}`
	queryBody = fmt.Sprintf(queryBody, wd, filter, (p-1)*listRows, listRows, sort)
	if helper.Debug {
		helper.Logger.Debug(queryBody)
	}
	api := this.Host + this.Index + "/" + this.Type + "/_search"
	if resp, errResp := this.post(api).Body(queryBody).Response(); errResp != nil {
		err = errResp
	} else {
		b, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(b, &result)
	}
	return
}

//重建索引【全量】
//采用批量重建索引的方式进行
//每次操作100条数据
func (this *ElasticSearchClient) RebuildAllIndex() {
	helper.ConfigMap.Store("indexing", true)
	defer helper.ConfigMap.Store("indexing", false)
	//假设有10个亿的文档...
	pageSize := 1000
	maxPage := int(1e7)
	for page := 1; page < maxPage; page++ {
		if infos, rows, err := NewDocument().GetDocInfoForElasticSearch(page, pageSize, 0); err != nil || rows == 0 {
			if err != nil && err != orm.ErrNoRows {
				helper.Logger.Error(err.Error())
			}
			page = maxPage
		} else {
			var ids []int
			for _, info := range infos {
				ids = append(ids, info.Id)
			}
			timeStart := time.Now().Unix()
			if data, err := NewDocument().GetDocForElasticSearch(ids...); err != nil {
				helper.Logger.Error("批量生成索引失败：" + err.Error())
			} else {
				if err := this.BuildIndexByBuck(data); err != nil {
					helper.Logger.Error("批量生成索引失败：" + err.Error())
				}
			}
			timeEnd := time.Now().Unix()
			//如果生成/更新索引耗时超过默认超时时间的一半，则休眠一小段时间，避免由于生成索引导致服务器负载过高，从而影响整站服务
			if spend := timeEnd - timeStart; spend > int64(this.Timeout)/2 {
				time.Sleep(time.Duration(spend) * time.Second)
			}
		}

	}
}

//根据id查询文档，并创建索引
func (this *ElasticSearchClient) BuildIndexById(id int) (err error) {
	if es, errES := NewDocument().GetDocForElasticSearch(id); errES != nil {
		err = errES
	} else {
		//基本只会有一项
		for _, item := range es {
			err = this.BuildIndex(item)
		}
	}
	return
}

//通过bulk，批量创建/更新索引
func (this *ElasticSearchClient) BuildIndexByBuck(data []ElasticSearchData) (err error) {
	var bodySlice []string
	if len(data) > 0 {
		for _, item := range data {
			action := fmt.Sprintf(`{"index":{"_index":"%v","_type":"%v","_id":%v}}`, this.Index, this.Type, item.Id)
			bodySlice = append(bodySlice, action)
			bodySlice = append(bodySlice, util.InterfaceToJson(item))
		}
		api := this.Host + "_bulk"
		body := strings.Join(bodySlice, "\n") + "\n"
		if helper.Debug {
			helper.Logger.Info("批量更新索引请求体")
			helper.Logger.Info(body)
		}
		if resp, errResp := this.post(api).Body(body).Response(); errResp != nil {
			err = errResp
		} else {
			if resp.StatusCode >= http.StatusMultipleChoices || resp.StatusCode < http.StatusOK {
				b, _ := ioutil.ReadAll(resp.Body)
				err = errors.New(resp.Status + "；" + string(b))
			}
		}
	}
	return
}

//创建索引
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

//查询索引量
//@return           count           统计数据
//@return           err             错误
func (this *ElasticSearchClient) Count() (count int, err error) {
	if !this.On {
		err = errors.New("未启用ElasticSearch")
		return
	}
	api := this.Host + this.Index + "/" + this.Type + "/_count"
	if resp, errResp := this.get(api).Response(); errResp != nil {
		err = errResp
	} else {
		b, _ := ioutil.ReadAll(resp.Body)
		body := string(b)
		if resp.StatusCode >= http.StatusMultipleChoices || resp.StatusCode < http.StatusOK {
			err = errors.New(resp.Status + "；" + body)
		} else {
			var cnt ElasticSearchCount
			if err = json.Unmarshal(b, &cnt); err == nil {
				count = cnt.Count
			}
		}
	}
	return
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
	resp, err := this.get(this.Host).Response()
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + "；" + string(body))
	}
	return err
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
