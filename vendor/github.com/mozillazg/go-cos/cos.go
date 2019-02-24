package cos

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"text/template"

	"strconv"

	"github.com/google/go-querystring/query"
	"github.com/mozillazg/go-httpheader"
)

const (
	// Version ...
	Version               = "0.9.0"
	userAgent             = "go-cos/" + Version
	contentTypeXML        = "application/xml"
	defaultServiceBaseURL = "https://service.cos.myqcloud.com"
)

var bucketURLTemplate = template.Must(
	template.New("bucketURLFormat").Parse(
		"{{.Scheme}}://{{.BucketName}}-{{.AppID}}.cos.{{.Region}}.myqcloud.com",
	),
)

// BaseURL 访问各 API 所需的基础 URL
type BaseURL struct {
	// 访问 bucket, object 相关 API 的基础 URL（不包含 path 部分）
	// 比如：https://test-1253846586.cos.ap-beijing.myqcloud.com
	// 详见 https://cloud.tencent.com/document/product/436/6224
	BucketURL *url.URL
	// 访问 service API 的基础 URL（不包含 path 部分）
	// 比如：https://service.cos.myqcloud.com
	ServiceURL *url.URL
}

// NewBaseURL 生成 BaseURL
func NewBaseURL(bucketURL string) (u *BaseURL, err error) {
	bu, err := url.Parse(bucketURL)
	if err != nil {
		return
	}
	su, _ := url.Parse(defaultServiceBaseURL)
	u = &BaseURL{
		BucketURL:  bu,
		ServiceURL: su,
	}
	return
}

// NewBucketURL 生成 BaseURL 所需的 BucketURL
//
//   bucketName: bucket 名称
//   AppID: 应用 ID
//   Region: 区域代码，详见 https://cloud.tencent.com/document/product/436/6224
//   secure: 是否使用 https
func NewBucketURL(bucketName, appID, region string, secure bool) *url.URL {
	scheme := "https"
	if !secure {
		scheme = "http"
	}

	w := bytes.NewBuffer(nil)
	bucketURLTemplate.Execute(w, struct {
		Scheme     string
		BucketName string
		AppID      string
		Region     string
	}{
		scheme, bucketName, appID, region,
	})

	u, _ := url.Parse(w.String())
	return u
}

// A Client manages communication with the COS API.
type Client struct {
	client *http.Client

	UserAgent string
	BaseURL   *BaseURL

	common service

	Service *ServiceService
	Bucket  *BucketService
	Object  *ObjectService
}

type service struct {
	client *Client
}

// NewClient returns a new COS API client.
func NewClient(uri *BaseURL, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	baseURL := &BaseURL{}
	if uri != nil {
		baseURL.BucketURL = uri.BucketURL
		baseURL.ServiceURL = uri.ServiceURL
	}
	if baseURL.ServiceURL == nil {
		baseURL.ServiceURL, _ = url.Parse(defaultServiceBaseURL)
	}

	c := &Client{
		client:    httpClient,
		UserAgent: userAgent,
		BaseURL:   baseURL,
	}
	c.common.client = c
	c.Service = (*ServiceService)(&c.common)
	c.Bucket = (*BucketService)(&c.common)
	c.Object = (*ObjectService)(&c.common)
	return c
}

func (c *Client) newRequest(ctx context.Context, opt *sendOptions) (req *http.Request, err error) {
	baseURL := opt.baseURL
	uri := opt.uri
	method := opt.method
	body := opt.body
	optQuery := opt.optQuery
	optHeader := opt.optHeader

	uri, err = addURLOptions(uri, optQuery)
	if err != nil {
		return
	}
	u, _ := url.Parse(uri)
	urlStr := baseURL.ResolveReference(u).String()

	var reader io.Reader
	contentType := ""
	contentMD5 := ""
	xsha1 := ""
	if body != nil {
		// 上传文件
		if r, ok := body.(io.Reader); ok {
			reader = r
		} else {
			b, err := xml.Marshal(body)
			if err != nil {
				return nil, err
			}
			contentType = contentTypeXML
			reader = bytes.NewReader(b)
			contentMD5 = base64.StdEncoding.EncodeToString(calMD5Digest(b))
			//xsha1 = base64.StdEncoding.EncodeToString(calSHA1Digest(b))
		}
	} else {
		contentType = contentTypeXML
	}

	req, err = http.NewRequest(method, urlStr, reader)
	if err != nil {
		return
	}

	req.Header, err = addHeaderOptions(req.Header, optHeader)
	if err != nil {
		return
	}
	if v := req.Header.Get("Content-Length"); req.ContentLength == 0 && v != "" && v != "0" {
		req.ContentLength, _ = strconv.ParseInt(v, 10, 64)
	}

	if contentMD5 != "" {
		req.Header["Content-MD5"] = []string{contentMD5}
	}
	if xsha1 != "" {
		req.Header.Set("x-cos-sha1", xsha1)
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	if req.Header.Get("Content-Type") == "" && contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return
}

func (c *Client) doAPI(ctx context.Context, req *http.Request, result interface{}, closeBody bool) (*Response, error) {
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}

	defer func() {
		if closeBody {
			// Close the body to let the Transport reuse the connection
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	response := newResponse(resp)

	err = checkResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return response, err
	}

	if result != nil {
		if w, ok := result.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = xml.NewDecoder(resp.Body).Decode(result)
			if err == io.EOF {
				err = nil // ignore EOF errors caused by empty response body
			}
		}
	}

	return response, err
}

type sendOptions struct {
	// 基础 URL
	baseURL *url.URL
	// URL 中除基础 URL 外的剩余部分
	uri string
	// 请求方法
	method string

	body interface{}
	// url 查询参数
	optQuery interface{}
	// http header 参数
	optHeader interface{}
	// 用 result 反序列化 resp.Body
	result interface{}
	// 是否禁用自动调用 resp.Body.Close()
	// 自动调用 Close() 是为了能够重用连接
	disableCloseBody bool
}

func (c *Client) send(ctx context.Context, opt *sendOptions) (resp *Response, err error) {
	req, err := c.newRequest(ctx, opt)
	if err != nil {
		return
	}

	resp, err = c.doAPI(ctx, req, opt.result, !opt.disableCloseBody)
	if err != nil {
		return
	}
	return
}

// addURLOptions adds the parameters in opt as URL query parameters to s. opt
// must be a struct whose fields may contain "url" tags.
func addURLOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	// 保留原有的参数，并且放在前面。因为 cos 的 url 路由是以第一个参数作为路由的
	// e.g. /?uploads
	q := u.RawQuery
	rq := qs.Encode()
	if q != "" {
		if rq != "" {
			u.RawQuery = fmt.Sprintf("%s&%s", q, qs.Encode())
		}
	} else {
		u.RawQuery = rq
	}
	return u.String(), nil
}

// addHeaderOptions adds the parameters in opt as Header fields to req. opt
// must be a struct whose fields may contain "header" tags.
func addHeaderOptions(header http.Header, opt interface{}) (http.Header, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return header, nil
	}

	h, err := httpheader.Header(opt)
	if err != nil {
		return nil, err
	}

	for key, values := range h {
		for _, value := range values {
			header.Add(key, value)
		}
	}
	return header, nil
}

// Owner ...
type Owner struct {
	UIN         string `xml:"uin,omitempty"`
	ID          string `xml:",omitempty"`
	DisplayName string `xml:",omitempty"`
}

// Initiator ...
type Initiator Owner

// Response API 响应
type Response struct {
	*http.Response
}

func newResponse(resp *http.Response) *Response {
	return &Response{
		Response: resp,
	}
}

// ACLHeaderOptions ...
type ACLHeaderOptions struct {
	XCosACL              string `header:"x-cos-acl,omitempty" url:"-" xml:"-"`
	XCosGrantRead        string `header:"x-cos-grant-read,omitempty" url:"-" xml:"-"`
	XCosGrantWrite       string `header:"x-cos-grant-write,omitempty" url:"-" xml:"-"`
	XCosGrantFullControl string `header:"x-cos-grant-full-control,omitempty" url:"-" xml:"-"`
}

// ACLGrantee ...
type ACLGrantee struct {
	Type        string `xml:"type,attr"`
	UIN         string `xml:"uin,omitempty"`
	ID          string `xml:",omitempty"`
	DisplayName string `xml:",omitempty"`
	SubAccount  string `xml:"Subaccount,omitempty"`
}

// ACLGrant ...
type ACLGrant struct {
	Grantee    *ACLGrantee
	Permission string
}

// ACLXml ...
type ACLXml struct {
	XMLName           xml.Name `xml:"AccessControlPolicy"`
	Owner             *Owner
	AccessControlList []ACLGrant `xml:"AccessControlList>Grant,omitempty"`
}
