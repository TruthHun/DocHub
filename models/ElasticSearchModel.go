package models

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/astaxie/beego/httplib"
)

//全文搜索
type ElasticSearch struct {
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
	DocType     int    //文档类型
}

//初始化搜索引擎，主要是检测能否ping通，以及索引、mapping等配置
func (this *ElasticSearch) InitSearch() {

}

//重建索引【全量】
func (this *ElasticSearch) RebuildAllIndex() {

}

//新增/更新索引
func (this *ElasticSearch) BuildIndex(es []ElasticSearch) {

}

//搜索内容
func (this *ElasticSearch) Search() {

}

//检验es服务能否连通
func (this *ElasticSearch) Ping(host string) error {
	req := httplib.Get(host).SetTimeout(1*time.Second, 1*time.Second)
	if strings.HasPrefix(strings.ToLower(host), "https") {
		req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
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

//生成索引
