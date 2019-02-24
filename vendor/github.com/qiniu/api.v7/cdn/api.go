package cdn

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/qiniu/api.v7/auth/qbox"
)

// Fusion CDN服务域名
var (
	FusionHost = "http://fusion.qiniuapi.com"
)

// CdnManager 提供了文件和目录刷新，文件预取，获取域名带宽和流量数据，获取域名日志列表等功能
type CdnManager struct {
	mac *qbox.Mac
}

// NewCdnManager 用来构建一个新的 CdnManager
func NewCdnManager(mac *qbox.Mac) *CdnManager {
	return &CdnManager{mac: mac}
}

// TrafficReq 为批量查询带宽/流量的API请求内容
//	StartDate 	开始日期，格式例如：2016-07-01
//	EndDate 	结束日期，格式例如：2016-07-03
//	Granularity	取值粒度，取值可选值：5min/hour/day
//	Domains 	域名列表，彼此用 ; 连接
type TrafficReq struct {
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Granularity string `json:"granularity"`
	Domains     string `json:"domains"`
}

// TrafficResp 为带宽/流量查询响应内容
type TrafficResp struct {
	Code  int                    `json:"code"`
	Error string                 `json:"error"`
	Time  []string               `json:"time,omitempty"`
	Data  map[string]TrafficData `json:"data,omitempty"`
}

// TrafficData 为带宽/流量数据
type TrafficData struct {
	DomainChina   []int `json:"china"`
	DomainOversea []int `json:"oversea"`
}

// GetBandwidthData 方法用来获取域名访问带宽数据
//	StartDate	string		必须	开始日期，例如：2016-07-01
//	EndDate		string		必须	结束日期，例如：2016-07-03
//	Granularity	string		必须	粒度，取值：5min ／ hour ／day
//	Domains		[]string	必须	域名列表
func (m *CdnManager) GetBandwidthData(startDate, endDate, granularity string,
	domainList []string) (bandwidthData TrafficResp, err error) {
	domains := strings.Join(domainList, ";")
	reqBody := TrafficReq{
		StartDate:   startDate,
		EndDate:     endDate,
		Granularity: granularity,
		Domains:     domains,
	}

	resData, reqErr := postRequest(m.mac, "/v2/tune/bandwidth", reqBody)
	if reqErr != nil {
		err = reqErr
		return
	}
	umErr := json.Unmarshal(resData, &bandwidthData)
	if umErr != nil {
		err = umErr
		return
	}
	return
}

// GetFluxData 方法用来获取域名访问流量数据
//	StartDate	string		必须	开始日期，例如：2016-07-01
//	EndDate		string		必须	结束日期，例如：2016-07-03
//	Granularity	string		必须	粒度，取值：5min ／ hour ／day
//	Domains		[]string	必须	域名列表
func (m *CdnManager) GetFluxData(startDate, endDate, granularity string,
	domainList []string) (fluxData TrafficResp, err error) {
	domains := strings.Join(domainList, ";")
	reqBody := TrafficReq{
		StartDate:   startDate,
		EndDate:     endDate,
		Granularity: granularity,
		Domains:     domains,
	}

	resData, reqErr := postRequest(m.mac, "/v2/tune/flux", reqBody)
	if reqErr != nil {
		err = reqErr
		return
	}

	umErr := json.Unmarshal(resData, &fluxData)
	if umErr != nil {
		err = umErr
		return
	}

	return
}

// RefreshReq 为缓存刷新请求内容
type RefreshReq struct {
	Urls []string `json:"urls"`
	Dirs []string `json:"dirs"`
}

// RefreshResp 缓存刷新响应内容
type RefreshResp struct {
	Code          int      `json:"code"`
	Error         string   `json:"error"`
	RequestID     string   `json:"requestId,omitempty"`
	InvalidUrls   []string `json:"invalidUrls,omitempty"`
	InvalidDirs   []string `json:"invalidDirs,omitempty"`
	URLQuotaDay   int      `json:"urlQuotaDay,omitempty"`
	URLSurplusDay int      `json:"urlSurplusDay,omitempty"`
	DirQuotaDay   int      `json:"dirQuotaDay,omitempty"`
	DirSurplusDay int      `json:"dirSurplusDay,omitempty"`
}

// RefreshUrlsAndDirs 方法用来刷新文件或目录
// urls	要刷新的单个url列表，单次方法调用总数不超过100条；单个url，即一个具体的url，
// 例如：http://bar.foo.com/index.html
// dirs	要刷新的目录url列表，单次方法调用总数不超过10条；目录dir，即表示一个目录级的url，
// 例如：http://bar.foo.com/dir/，
func (m *CdnManager) RefreshUrlsAndDirs(urls, dirs []string) (result RefreshResp, err error) {
	if len(urls) > 100 {
		err = errors.New("urls count exceeds the limit of 100")
		return
	}
	if len(dirs) > 10 {
		err = errors.New("dirs count exceeds the limit of 10")
		return
	}

	reqBody := RefreshReq{
		Urls: urls,
		Dirs: dirs,
	}

	resData, reqErr := postRequest(m.mac, "/v2/tune/refresh", reqBody)
	if reqErr != nil {
		err = reqErr
		return
	}
	umErr := json.Unmarshal(resData, &result)
	if umErr != nil {
		err = reqErr
		return
	}

	return
}

// RefreshUrls 刷新文件
func (m *CdnManager) RefreshUrls(urls []string) (result RefreshResp, err error) {
	return m.RefreshUrlsAndDirs(urls, nil)
}

// RefreshDirs 刷新目录
func (m *CdnManager) RefreshDirs(dirs []string) (result RefreshResp, err error) {
	return m.RefreshUrlsAndDirs(nil, dirs)
}

// PrefetchReq 文件预取请求内容
type PrefetchReq struct {
	Urls []string `json:"urls"`
}

// PrefetchResp 文件预取响应内容
type PrefetchResp struct {
	Code        int      `json:"code"`
	Error       string   `json:"error"`
	RequestID   string   `json:"requestId,omitempty"`
	InvalidUrls []string `json:"invalidUrls,omitempty"`
	QuotaDay    int      `json:"quotaDay,omitempty"`
	SurplusDay  int      `json:"surplusDay,omitempty"`
}

// PrefetchUrls 预取文件链接，每次最多不可以超过100条
func (m *CdnManager) PrefetchUrls(urls []string) (result PrefetchResp, err error) {
	if len(urls) > 100 {
		err = errors.New("urls count exceeds the limit of 100")
		return
	}

	reqBody := PrefetchReq{
		Urls: urls,
	}

	resData, reqErr := postRequest(m.mac, "/v2/tune/prefetch", reqBody)
	if reqErr != nil {
		err = reqErr
		return
	}

	umErr := json.Unmarshal(resData, &result)
	if umErr != nil {
		err = umErr
		return
	}

	return
}

// ListLogRequest 日志下载请求内容
type ListLogRequest struct {
	Day     string `json:"day"`
	Domains string `json:"domains"`
}

// ListLogResult 日志下载相应内容
type ListLogResult struct {
	Code  int                        `json:"code"`
	Error string                     `json:"error"`
	Data  map[string][]LogDomainInfo `json:"data"`
}

// LogDomainInfo 日志下载信息
type LogDomainInfo struct {
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	ModifiedTime int64  `json:"mtime"`
	URL          string `json:"url"`
}

// GetCdnLogList 获取CDN域名访问日志的下载链接
func (m *CdnManager) GetCdnLogList(day string, domains []string) (
	listLogResult ListLogResult, err error) {
	//new log query request
	logReq := ListLogRequest{
		Day:     day,
		Domains: strings.Join(domains, ";"),
	}

	resData, reqErr := postRequest(m.mac, "/v2/tune/log/list", logReq)
	if reqErr != nil {
		err = fmt.Errorf("get response error, %s", reqErr)
		return
	}

	if decodeErr := json.Unmarshal(resData, &listLogResult); decodeErr != nil {
		err = fmt.Errorf("get response error, %s", decodeErr)
		return
	}

	if listLogResult.Error != "" {
		err = fmt.Errorf("get log list error, %d %s", listLogResult.Code, listLogResult.Error)
		return
	}

	return
}

// RequestWithBody 带body对api发出请求并且返回response body
func postRequest(mac *qbox.Mac, path string, body interface{}) (resData []byte,
	err error) {
	urlStr := fmt.Sprintf("%s%s", FusionHost, path)
	reqData, _ := json.Marshal(body)
	req, reqErr := http.NewRequest("POST", urlStr, bytes.NewReader(reqData))
	if reqErr != nil {
		err = reqErr
		return
	}

	accessToken, signErr := mac.SignRequest(req)
	if signErr != nil {
		err = signErr
		return
	}

	req.Header.Add("Authorization", "QBox "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		err = respErr
		return
	}
	defer resp.Body.Close()

	resData, ioErr := ioutil.ReadAll(resp.Body)
	if ioErr != nil {
		err = ioErr
		return
	}

	return
}
