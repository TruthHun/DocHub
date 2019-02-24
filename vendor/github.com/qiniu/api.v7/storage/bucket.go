package storage

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/conf"
	"net/http"
)

// 资源管理相关的默认域名
const (
	DefaultRsHost  = "rs.qiniu.com"
	DefaultRsfHost = "rsf.qiniu.com"
	DefaultAPIHost = "api.qiniu.com"
	DefaultPubHost = "pu.qbox.me:10200"
)

// FileInfo 文件基本信息
type FileInfo struct {
	Hash     string `json:"hash"`
	Fsize    int64  `json:"fsize"`
	PutTime  int64  `json:"putTime"`
	MimeType string `json:"mimeType"`
	Type     int    `json:"type"`
}

func (f *FileInfo) String() string {
	str := ""
	str += fmt.Sprintf("Hash:     %s\n", f.Hash)
	str += fmt.Sprintf("Fsize:    %d\n", f.Fsize)
	str += fmt.Sprintf("PutTime:  %d\n", f.PutTime)
	str += fmt.Sprintf("MimeType: %s\n", f.MimeType)
	str += fmt.Sprintf("Type:     %d\n", f.Type)
	return str
}

// FetchRet 资源抓取的返回值
type FetchRet struct {
	Hash     string `json:"hash"`
	Fsize    int64  `json:"fsize"`
	MimeType string `json:"mimeType"`
	Key      string `json:"key"`
}

type listFilesRet2 struct {
	Marker string   `json:"marker"`
	Item   ListItem `json:"item"`
	Dir    string   `json:"dir"`
}

func (r *FetchRet) String() string {
	str := ""
	str += fmt.Sprintf("Key:      %s\n", r.Key)
	str += fmt.Sprintf("Hash:     %s\n", r.Hash)
	str += fmt.Sprintf("Fsize:    %d\n", r.Fsize)
	str += fmt.Sprintf("MimeType: %s\n", r.MimeType)
	return str
}

// ListItem 为文件列举的返回值
type ListItem struct {
	Key      string `json:"key"`
	Hash     string `json:"hash"`
	Fsize    int64  `json:"fsize"`
	PutTime  int64  `json:"putTime"`
	MimeType string `json:"mimeType"`
	Type     int    `json:"type"`
	EndUser  string `json:"endUser"`
}

// 接口可能返回空的记录
func (l *ListItem) IsEmpty() (empty bool) {
	return l.Key == "" && l.Hash == "" && l.Fsize == 0 && l.PutTime == 0
}

func (l *ListItem) String() string {
	str := ""
	str += fmt.Sprintf("Hash:     %s\n", l.Hash)
	str += fmt.Sprintf("Fsize:    %d\n", l.Fsize)
	str += fmt.Sprintf("PutTime:  %d\n", l.PutTime)
	str += fmt.Sprintf("MimeType: %s\n", l.MimeType)
	str += fmt.Sprintf("Type:     %d\n", l.Type)
	str += fmt.Sprintf("EndUser:  %s\n", l.EndUser)
	return str
}

// BatchOpRet 为批量执行操作的返回值
// 批量操作支持 stat，copy，delete，move，chgm，chtype，deleteAfterDays几个操作
// 其中 stat 为获取文件的基本信息，如果文件存在则返回基本信息，如果文件不存在返回 error 。
// 其他的操作，如果成功，则返回 code，不成功会同时返回 error 信息，可以根据 error 信息来判断问题所在。
type BatchOpRet struct {
	Code int `json:"code,omitempty"`
	Data struct {
		Hash     string `json:"hash"`
		Fsize    int64  `json:"fsize"`
		PutTime  int64  `json:"putTime"`
		MimeType string `json:"mimeType"`
		Type     int    `json:"type"`
		Error    string `json:"error"`
	} `json:"data,omitempty"`
}

// BucketManager 提供了对资源进行管理的操作
type BucketManager struct {
	Client *Client
	Mac    *qbox.Mac
	Cfg    *Config
}

// NewBucketManager 用来构建一个新的资源管理对象
func NewBucketManager(mac *qbox.Mac, cfg *Config) *BucketManager {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.CentralRsHost == "" {
		cfg.CentralRsHost = DefaultRsHost
	}

	return &BucketManager{
		Client: &DefaultClient,
		Mac:    mac,
		Cfg:    cfg,
	}
}

// NewBucketManagerEx 用来构建一个新的资源管理对象
func NewBucketManagerEx(mac *qbox.Mac, cfg *Config, client *Client) *BucketManager {
	if cfg == nil {
		cfg = &Config{}
	}

	if client == nil {
		client = &DefaultClient
	}
	if cfg.CentralRsHost == "" {
		cfg.CentralRsHost = DefaultRsHost
	}

	return &BucketManager{
		Client: client,
		Mac:    mac,
		Cfg:    cfg,
	}
}

// Buckets 用来获取空间列表，如果指定了 shared 参数为 true，那么一同列表被授权访问的空间
func (m *BucketManager) Buckets(shared bool) (buckets []string, err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	var reqHost string

	reqHost = m.Cfg.RsReqHost()
	reqURL := fmt.Sprintf("%s/buckets?shared=%v", reqHost, shared)
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, &buckets, "POST", reqURL, headers)
	return
}

// Stat 用来获取一个文件的基本信息
func (m *BucketManager) Stat(bucket, key string) (info FileInfo, err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqHost, reqErr := m.RsReqHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	reqURL := fmt.Sprintf("%s%s", reqHost, URIStat(bucket, key))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, &info, "POST", reqURL, headers)
	return
}

// Delete 用来删除空间中的一个文件
func (m *BucketManager) Delete(bucket, key string) (err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqHost, reqErr := m.RsReqHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	reqURL := fmt.Sprintf("%s%s", reqHost, URIDelete(bucket, key))
	err = m.Client.Call(ctx, nil, "POST", reqURL, headers)
	return
}

// Copy 用来创建已有空间中的文件的一个新的副本
func (m *BucketManager) Copy(srcBucket, srcKey, destBucket, destKey string, force bool) (err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqHost, reqErr := m.RsReqHost(srcBucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	reqURL := fmt.Sprintf("%s%s", reqHost, URICopy(srcBucket, srcKey, destBucket, destKey, force))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, nil, "POST", reqURL, headers)
	return
}

// Move 用来将空间中的一个文件移动到新的空间或者重命名
func (m *BucketManager) Move(srcBucket, srcKey, destBucket, destKey string, force bool) (err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqHost, reqErr := m.RsReqHost(srcBucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	reqURL := fmt.Sprintf("%s%s", reqHost, URIMove(srcBucket, srcKey, destBucket, destKey, force))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, nil, "POST", reqURL, headers)
	return
}

// ChangeMime 用来更新文件的MimeType
func (m *BucketManager) ChangeMime(bucket, key, newMime string) (err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqHost, reqErr := m.RsReqHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}
	reqURL := fmt.Sprintf("%s%s", reqHost, URIChangeMime(bucket, key, newMime))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, nil, "POST", reqURL, headers)
	return
}

// ChangeType 用来更新文件的存储类型，0表示普通存储，1表示低频存储
func (m *BucketManager) ChangeType(bucket, key string, fileType int) (err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqHost, reqErr := m.RsReqHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}
	reqURL := fmt.Sprintf("%s%s", reqHost, URIChangeType(bucket, key, fileType))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, nil, "POST", reqURL, headers)
	return
}

// DeleteAfterDays 用来更新文件生命周期，如果 days 设置为0，则表示取消文件的定期删除功能，永久存储
func (m *BucketManager) DeleteAfterDays(bucket, key string, days int) (err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqHost, reqErr := m.RsReqHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	reqURL := fmt.Sprintf("%s%s", reqHost, URIDeleteAfterDays(bucket, key, days))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, nil, "POST", reqURL, headers)
	return
}

// Batch 接口提供了资源管理的批量操作，支持 stat，copy，move，delete，chgm，chtype，deleteAfterDays几个接口
func (m *BucketManager) Batch(operations []string) (batchOpRet []BatchOpRet, err error) {
	if len(operations) > 1000 {
		err = errors.New("batch operation count exceeds the limit of 1000")
		return
	}
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	scheme := "http://"
	if m.Cfg.UseHTTPS {
		scheme = "https://"
	}
	reqURL := fmt.Sprintf("%s%s/batch", scheme, m.Cfg.CentralRsHost)
	params := map[string][]string{
		"op": operations,
	}
	err = m.Client.CallWithForm(ctx, &batchOpRet, "POST", reqURL, nil, params)
	return
}

// Fetch 根据提供的远程资源链接来抓取一个文件到空间并已指定文件名保存
func (m *BucketManager) Fetch(resURL, bucket, key string) (fetchRet FetchRet, err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)

	reqHost, rErr := m.IoReqHost(bucket)
	if rErr != nil {
		err = rErr
		return
	}
	reqURL := fmt.Sprintf("%s%s", reqHost, uriFetch(resURL, bucket, key))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, &fetchRet, "POST", reqURL, headers)
	return
}

func (m *BucketManager) RsReqHost(bucket string) (reqHost string, err error) {
	var reqErr error

	if m.Cfg.RsHost == "" {
		reqHost, reqErr = m.RsHost(bucket)
		if reqErr != nil {
			err = reqErr
			return
		}
	} else {
		reqHost = m.Cfg.RsHost
	}
	if !strings.HasPrefix(reqHost, "http") {
		reqHost = "http://" + reqHost
	}
	return
}

func (m *BucketManager) ApiReqHost(bucket string) (reqHost string, err error) {
	var reqErr error

	if m.Cfg.ApiHost == "" {
		reqHost, reqErr = m.ApiHost(bucket)
		if reqErr != nil {
			err = reqErr
			return
		}
	} else {
		reqHost = m.Cfg.ApiHost
	}
	if !strings.HasPrefix(reqHost, "http") {
		reqHost = "http://" + reqHost
	}
	return
}

func (m *BucketManager) RsfReqHost(bucket string) (reqHost string, err error) {
	var reqErr error

	if m.Cfg.RsfHost == "" {
		reqHost, reqErr = m.RsfHost(bucket)
		if reqErr != nil {
			err = reqErr
			return
		}
	} else {
		reqHost = m.Cfg.RsfHost
	}
	if !strings.HasPrefix(reqHost, "http") {
		reqHost = "http://" + reqHost
	}
	return
}

func (m *BucketManager) IoReqHost(bucket string) (reqHost string, err error) {
	var reqErr error

	if m.Cfg.IoHost == "" {
		reqHost, reqErr = m.IovipHost(bucket)
		if reqErr != nil {
			err = reqErr
			return
		}
	} else {
		reqHost = m.Cfg.IoHost
	}
	if !strings.HasPrefix(reqHost, "http") {
		reqHost = "http://" + reqHost
	}
	return
}

// FetchWithoutKey 根据提供的远程资源链接来抓取一个文件到空间并以文件的内容hash作为文件名
func (m *BucketManager) FetchWithoutKey(resURL, bucket string) (fetchRet FetchRet, err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)

	reqHost, rErr := m.IoReqHost(bucket)
	if rErr != nil {
		err = rErr
		return
	}
	reqURL := fmt.Sprintf("%s%s", reqHost, uriFetchWithoutKey(resURL, bucket))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, &fetchRet, "POST", reqURL, headers)
	return
}

// Prefetch 用来同步镜像空间的资源和镜像源资源内容
func (m *BucketManager) Prefetch(bucket, key string) (err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqHost, reqErr := m.IoReqHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}
	reqURL := fmt.Sprintf("%s%s", reqHost, uriPrefetch(bucket, key))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, nil, "POST", reqURL, headers)
	return
}

// SetImage 用来设置空间镜像源
func (m *BucketManager) SetImage(siteURL, bucket string) (err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqURL := fmt.Sprintf("http://%s%s", DefaultPubHost, uriSetImage(siteURL, bucket))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, nil, "POST", reqURL, headers)
	return
}

// SetImageWithHost 用来设置空间镜像源，额外添加回源Host头部
func (m *BucketManager) SetImageWithHost(siteURL, bucket, host string) (err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqURL := fmt.Sprintf("http://%s%s", DefaultPubHost,
		uriSetImageWithHost(siteURL, bucket, host))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, nil, "POST", reqURL, headers)
	return
}

// UnsetImage 用来取消空间镜像源设置
func (m *BucketManager) UnsetImage(bucket string) (err error) {
	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqURL := fmt.Sprintf("http://%s%s", DefaultPubHost, uriUnsetImage(bucket))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, nil, "POST", reqURL, headers)
	return err
}

type listFilesRet struct {
	Marker         string     `json:"marker"`
	Items          []ListItem `json:"items"`
	CommonPrefixes []string   `json:"commonPrefixes"`
}

// ListFiles 用来获取空间文件列表，可以根据需要指定文件的前缀 prefix，文件的目录 delimiter，循环列举的时候下次
// 列举的位置 marker，以及每次返回的文件的最大数量limit，其中limit最大为1000。
func (m *BucketManager) ListFiles(bucket, prefix, delimiter, marker string,
	limit int) (entries []ListItem, commonPrefixes []string, nextMarker string, hasNext bool, err error) {
	if limit <= 0 || limit > 1000 {
		err = errors.New("invalid list limit, only allow [1, 1000]")
		return
	}

	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqHost, reqErr := m.RsfReqHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	ret := listFilesRet{}
	reqURL := fmt.Sprintf("%s%s", reqHost, uriListFiles(bucket, prefix, delimiter, marker, limit))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	err = m.Client.Call(ctx, &ret, "POST", reqURL, headers)
	if err != nil {
		return
	}

	commonPrefixes = ret.CommonPrefixes
	nextMarker = ret.Marker
	entries = ret.Items
	if ret.Marker != "" {
		hasNext = true
	}

	return
}

// ListBucket 用来获取空间文件列表，可以根据需要指定文件的前缀 prefix，文件的目录 delimiter，流式返回每条数据。
func (m *BucketManager) ListBucket(bucket, prefix, delimiter, marker string) (retCh chan listFilesRet2, err error) {

	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	reqHost, reqErr := m.RsfReqHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	// limit 0 ==> 列举所有文件
	reqURL := fmt.Sprintf("%s%s", reqHost, uriListFiles2(bucket, prefix, delimiter, marker))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	retCh, err = m.Client.CallChan(ctx, "POST", reqURL, headers)
	return
}

// ListBucketCancel 用来获取空间文件列表，可以根据需要指定文件的前缀 prefix，文件的目录 delimiter，流式返回每条数据。
// 接受的context可以用来取消列举操作
func (m *BucketManager) ListBucketContext(ctx context.Context, bucket, prefix, delimiter, marker string) (retCh chan listFilesRet2, err error) {

	vctx := context.WithValue(ctx, "mac", m.Mac)
	reqHost, reqErr := m.RsfReqHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	// limit 0 ==> 列举所有文件
	reqURL := fmt.Sprintf("%s%s", reqHost, uriListFiles2(bucket, prefix, delimiter, marker))
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_FORM)
	retCh, err = m.Client.CallChan(vctx, "POST", reqURL, headers)
	return
}

type AsyncFetchParam struct {
	Url              string `json:"url"`
	Host             string `json:"host,omitempty"`
	Bucket           string `json:"bucket"`
	Key              string `json:"key,omitempty"`
	Md5              string `json:"md5,omitempty"`
	Etag             string `json:"etag,omitempty"`
	CallbackURL      string `json:"callbackurl,omitempty"`
	CallbackBody     string `json:"callbackbody,omitempty"`
	CallbackBodyType string `json:"callbackbodytype,omitempty"`
	FileType         int    `json:"file_type,omitempty"`
}

type AsyncFetchRet struct {
	Id   string `json:"id"`
	Wait int    `json:"wait"`
}

func (m *BucketManager) AsyncFetch(param AsyncFetchParam) (ret AsyncFetchRet, err error) {

	reqUrl, err := m.ApiReqHost(param.Bucket)
	if err != nil {
		return
	}

	reqUrl += "/sisyphus/fetch"

	ctx := context.WithValue(context.TODO(), "mac", m.Mac)
	headers := http.Header{}
	headers.Add("Content-Type", conf.CONTENT_TYPE_JSON)
	err = m.Client.CallWithJson(ctx, &ret, "POST", reqUrl, headers, param)
	return
}

func (m *BucketManager) RsHost(bucket string) (rsHost string, err error) {
	zone, err := m.Zone(bucket)
	if err != nil {
		return
	}

	rsHost = zone.GetRsHost(m.Cfg.UseHTTPS)
	return
}

func (m *BucketManager) RsfHost(bucket string) (rsfHost string, err error) {
	zone, err := m.Zone(bucket)
	if err != nil {
		return
	}

	rsfHost = zone.GetRsfHost(m.Cfg.UseHTTPS)
	return
}

func (m *BucketManager) IovipHost(bucket string) (iovipHost string, err error) {
	zone, err := m.Zone(bucket)
	if err != nil {
		return
	}

	iovipHost = zone.GetIoHost(m.Cfg.UseHTTPS)
	return
}

func (m *BucketManager) ApiHost(bucket string) (apiHost string, err error) {
	zone, err := m.Zone(bucket)
	if err != nil {
		return
	}

	apiHost = zone.GetApiHost(m.Cfg.UseHTTPS)
	return
}

func (m *BucketManager) Zone(bucket string) (z *Zone, err error) {

	if m.Cfg.Zone != nil {
		z = m.Cfg.Zone
		return
	}

	z, err = GetZone(m.Mac.AccessKey, bucket)
	return
}

// 构建op的方法，导出的方法支持在Batch操作中使用

// URIStat 构建 stat 接口的请求命令
func URIStat(bucket, key string) string {
	return fmt.Sprintf("/stat/%s", EncodedEntry(bucket, key))
}

// URIDelete 构建 delete 接口的请求命令
func URIDelete(bucket, key string) string {
	return fmt.Sprintf("/delete/%s", EncodedEntry(bucket, key))
}

// URICopy 构建 copy 接口的请求命令
func URICopy(srcBucket, srcKey, destBucket, destKey string, force bool) string {
	return fmt.Sprintf("/copy/%s/%s/force/%v", EncodedEntry(srcBucket, srcKey),
		EncodedEntry(destBucket, destKey), force)
}

// URIMove 构建 move 接口的请求命令
func URIMove(srcBucket, srcKey, destBucket, destKey string, force bool) string {
	return fmt.Sprintf("/move/%s/%s/force/%v", EncodedEntry(srcBucket, srcKey),
		EncodedEntry(destBucket, destKey), force)
}

// URIDeleteAfterDays 构建 deleteAfterDays 接口的请求命令
func URIDeleteAfterDays(bucket, key string, days int) string {
	return fmt.Sprintf("/deleteAfterDays/%s/%d", EncodedEntry(bucket, key), days)
}

// URIChangeMime 构建 chgm 接口的请求命令
func URIChangeMime(bucket, key, newMime string) string {
	return fmt.Sprintf("/chgm/%s/mime/%s", EncodedEntry(bucket, key),
		base64.URLEncoding.EncodeToString([]byte(newMime)))
}

// URIChangeType 构建 chtype 接口的请求命令
func URIChangeType(bucket, key string, fileType int) string {
	return fmt.Sprintf("/chtype/%s/type/%d", EncodedEntry(bucket, key), fileType)
}

// 构建op的方法，非导出的方法无法用在Batch操作中
func uriFetch(resURL, bucket, key string) string {
	return fmt.Sprintf("/fetch/%s/to/%s",
		base64.URLEncoding.EncodeToString([]byte(resURL)), EncodedEntry(bucket, key))
}

func uriFetchWithoutKey(resURL, bucket string) string {
	return fmt.Sprintf("/fetch/%s/to/%s",
		base64.URLEncoding.EncodeToString([]byte(resURL)), EncodedEntryWithoutKey(bucket))
}

func uriPrefetch(bucket, key string) string {
	return fmt.Sprintf("/prefetch/%s", EncodedEntry(bucket, key))
}

func uriSetImage(siteURL, bucket string) string {
	return fmt.Sprintf("/image/%s/from/%s", bucket,
		base64.URLEncoding.EncodeToString([]byte(siteURL)))
}

func uriSetImageWithHost(siteURL, bucket, host string) string {
	return fmt.Sprintf("/image/%s/from/%s/host/%s", bucket,
		base64.URLEncoding.EncodeToString([]byte(siteURL)),
		base64.URLEncoding.EncodeToString([]byte(host)))
}

func uriUnsetImage(bucket string) string {
	return fmt.Sprintf("/unimage/%s", bucket)
}

func uriListFiles(bucket, prefix, delimiter, marker string, limit int) string {
	query := make(url.Values)
	query.Add("bucket", bucket)
	if prefix != "" {
		query.Add("prefix", prefix)
	}
	if delimiter != "" {
		query.Add("delimiter", delimiter)
	}
	if marker != "" {
		query.Add("marker", marker)
	}
	if limit > 0 {
		query.Add("limit", strconv.FormatInt(int64(limit), 10))
	}
	return fmt.Sprintf("/list?%s", query.Encode())
}

func uriListFiles2(bucket, prefix, delimiter, marker string) string {
	query := make(url.Values)
	query.Add("bucket", bucket)
	if prefix != "" {
		query.Add("prefix", prefix)
	}
	if delimiter != "" {
		query.Add("delimiter", delimiter)
	}
	if marker != "" {
		query.Add("marker", marker)
	}
	return fmt.Sprintf("/v2/list?%s", query.Encode())
}

// EncodedEntry 生成URL Safe Base64编码的 Entry
func EncodedEntry(bucket, key string) string {
	entry := fmt.Sprintf("%s:%s", bucket, key)
	return base64.URLEncoding.EncodeToString([]byte(entry))
}

// EncodedEntryWithoutKey 生成 key 为null的情况下 URL Safe Base64编码的Entry
func EncodedEntryWithoutKey(bucket string) string {
	return base64.URLEncoding.EncodeToString([]byte(bucket))
}

// MakePublicURL 用来生成公开空间资源下载链接
func MakePublicURL(domain, key string) (finalUrl string) {
	domain = strings.TrimRight(domain, "/")
	srcUrl := fmt.Sprintf("%s/%s", domain, key)
	srcUri, _ := url.Parse(srcUrl)
	finalUrl = srcUri.String()
	return
}

// MakePrivateURL 用来生成私有空间资源下载链接
func MakePrivateURL(mac *qbox.Mac, domain, key string, deadline int64) (privateURL string) {
	publicURL := MakePublicURL(domain, key)
	urlToSign := publicURL
	if strings.Contains(publicURL, "?") {
		urlToSign = fmt.Sprintf("%s&e=%d", urlToSign, deadline)
	} else {
		urlToSign = fmt.Sprintf("%s?e=%d", urlToSign, deadline)
	}
	token := mac.Sign([]byte(urlToSign))
	privateURL = fmt.Sprintf("%s&token=%s", urlToSign, token)
	return
}
