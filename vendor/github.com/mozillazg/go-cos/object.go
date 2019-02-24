package cos

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// ObjectService ...
//
// Object 相关 API
type ObjectService service

// ObjectGetOptions ...
type ObjectGetOptions struct {
	ResponseContentType        string `url:"response-content-type,omitempty" header:"-"`
	ResponseContentLanguage    string `url:"response-content-language,omitempty" header:"-"`
	ResponseExpires            string `url:"response-expires,omitempty" header:"-"`
	ResponseCacheControl       string `url:"response-cache-control,omitempty" header:"-"`
	ResponseContentDisposition string `url:"response-content-disposition,omitempty" header:"-"`
	ResponseContentEncoding    string `url:"response-content-encoding,omitempty" header:"-"`
	Range                      string `url:"-" header:"Range,omitempty"`
	IfModifiedSince            string `url:"-" header:"If-Modified-Since,omitempty"`

	// 预签名授权 URL
	PresignedURL *url.URL `header:"-" url:"-" xml:"-"`
}

// Get Object 请求可以将一个文件（Object）下载至本地。
// 该操作需要对目标 Object 具有读权限或目标 Object 对所有人都开放了读权限（公有读）。
//
// https://www.qcloud.com/document/product/436/7753
func (s *ObjectService) Get(ctx context.Context, name string, opt *ObjectGetOptions) (*Response, error) {
	baseURL := s.client.BaseURL.BucketURL
	uri := "/" + encodeURIComponent(name)
	if opt != nil && opt.PresignedURL != nil {
		baseURL = opt.PresignedURL
		uri = ""
	}
	sendOpt := sendOptions{
		baseURL:          baseURL,
		uri:              uri,
		method:           http.MethodGet,
		optQuery:         opt,
		optHeader:        opt,
		disableCloseBody: true,
	}
	resp, err := s.client.send(ctx, &sendOpt)
	return resp, err
}

// ObjectPutHeaderOptions ...
type ObjectPutHeaderOptions struct {
	CacheControl       string `header:"Cache-Control,omitempty" url:"-"`
	ContentDisposition string `header:"Content-Disposition,omitempty" url:"-"`
	ContentEncoding    string `header:"Content-Encoding,omitempty" url:"-"`
	ContentType        string `header:"Content-Type,omitempty" url:"-"`
	ContentLength      int    `header:"Content-Length,omitempty" url:"-"`
	Expect             string `header:"Expect,omitempty" url:"-"`
	Expires            string `header:"Expires,omitempty" url:"-"`
	XCosContentSHA1    string `header:"x-cos-content-sha1,omitempty" url:"-"`
	// 自定义的 x-cos-meta-* header
	XCosMetaXXX      *http.Header `header:"x-cos-meta-*,omitempty" url:"-"`
	XCosStorageClass string       `header:"x-cos-storage-class,omitempty" url:"-"`
	// 可选值: Normal, Appendable
	//XCosObjectType string `header:"x-cos-object-type,omitempty" url:"-"`
}

// ObjectPutOptions ...
type ObjectPutOptions struct {
	*ACLHeaderOptions       `header:",omitempty" url:"-" xml:"-"`
	*ObjectPutHeaderOptions `header:",omitempty" url:"-" xml:"-"`

	// 预签名授权 URL
	PresignedURL *url.URL `header:"-" url:"-" xml:"-"`
}

// Put Object请求可以将一个文件（Oject）上传至指定Bucket。
//
// 当 r 不是 bytes.Buffer/bytes.Reader/strings.Reader 时，必须指定 opt.ObjectPutHeaderOptions.ContentLength
//
// https://www.qcloud.com/document/product/436/7749
func (s *ObjectService) Put(ctx context.Context, name string, r io.Reader, opt *ObjectPutOptions) (*Response, error) {
	baseURL := s.client.BaseURL.BucketURL
	uri := "/" + encodeURIComponent(name)
	if opt != nil && opt.PresignedURL != nil {
		baseURL = opt.PresignedURL
		uri = ""
	}
	sendOpt := sendOptions{
		baseURL:   baseURL,
		uri:       uri,
		method:    http.MethodPut,
		body:      r,
		optHeader: opt,
	}
	resp, err := s.client.send(ctx, &sendOpt)
	return resp, err
}

// ObjectCopyHeaderOptions ...
type ObjectCopyHeaderOptions struct {
	XCosMetadataDirective           string `header:"x-cos-metadata-directive,omitempty" url:"-" xml:"-"`
	XCosCopySourceIfModifiedSince   string `header:"x-cos-copy-source-If-Modified-Since,omitempty" url:"-" xml:"-"`
	XCosCopySourceIfUnmodifiedSince string `header:"x-cos-copy-source-If-Unmodified-Since,omitempty" url:"-" xml:"-"`
	XCosCopySourceIfMatch           string `header:"x-cos-copy-source-If-Match,omitempty" url:"-" xml:"-"`
	XCosCopySourceIfNoneMatch       string `header:"x-cos-copy-source-If-None-Match,omitempty" url:"-" xml:"-"`
	XCosStorageClass                string `header:"x-cos-storage-class,omitempty" url:"-" xml:"-"`
	// 自定义的 x-cos-meta-* header
	XCosMetaXXX    *http.Header `header:"x-cos-meta-*,omitempty" url:"-"`
	XCosCopySource string       `header:"x-cos-copy-source" url:"-" xml:"-"`
}

// ObjectCopyOptions ...
type ObjectCopyOptions struct {
	*ObjectCopyHeaderOptions `header:",omitempty" url:"-" xml:"-"`
	*ACLHeaderOptions        `header:",omitempty" url:"-" xml:"-"`
}

// ObjectCopyResult ...
type ObjectCopyResult struct {
	XMLName      xml.Name `xml:"CopyObjectResult"`
	ETag         string   `xml:"ETag,omitempty"`
	LastModified string   `xml:"LastModified,omitempty"`
}

// Copy ...
// Put Object Copy 请求实现将一个文件从源路径复制到目标路径。建议文件大小 1M 到 5G，
// 超过 5G 的文件请使用分块上传 Upload - Copy。在拷贝的过程中，文件元属性和 ACL 可以被修改。
//
// 用户可以通过该接口实现文件移动，文件重命名，修改文件属性和创建副本。
//
// 注意：在跨帐号复制的时候，需要先设置被复制文件的权限为公有读，或者对目标帐号赋权，同帐号则不需要。
//
// https://cloud.tencent.com/document/product/436/10881
func (s *ObjectService) Copy(ctx context.Context, name, sourceURL string, opt *ObjectCopyOptions) (*ObjectCopyResult, *Response, error) {
	var res ObjectCopyResult
	if opt == nil {
		opt = new(ObjectCopyOptions)
	}
	if opt.ObjectCopyHeaderOptions == nil {
		opt.ObjectCopyHeaderOptions = new(ObjectCopyHeaderOptions)
	}
	opt.XCosCopySource = sourceURL

	sendOpt := sendOptions{
		baseURL:   s.client.BaseURL.BucketURL,
		uri:       "/" + encodeURIComponent(name),
		method:    http.MethodPut,
		body:      nil,
		optHeader: opt,
		result:    &res,
	}
	resp, err := s.client.send(ctx, &sendOpt)
	return &res, resp, err
}

// Delete Object请求可以将一个文件（Object）删除。
//
// https://www.qcloud.com/document/product/436/7743
func (s *ObjectService) Delete(ctx context.Context, name string) (*Response, error) {
	sendOpt := sendOptions{
		baseURL: s.client.BaseURL.BucketURL,
		uri:     "/" + encodeURIComponent(name),
		method:  http.MethodDelete,
	}
	resp, err := s.client.send(ctx, &sendOpt)
	return resp, err
}

// ObjectHeadOptions ...
type ObjectHeadOptions struct {
	IfModifiedSince string `url:"-" header:"If-Modified-Since,omitempty"`
}

// Head Object请求可以取回对应Object的元数据，Head的权限与Get的权限一致
//
// https://www.qcloud.com/document/product/436/7745
func (s *ObjectService) Head(ctx context.Context, name string, opt *ObjectHeadOptions) (*Response, error) {
	sendOpt := sendOptions{
		baseURL:   s.client.BaseURL.BucketURL,
		uri:       "/" + encodeURIComponent(name),
		method:    http.MethodHead,
		optHeader: opt,
	}
	resp, err := s.client.send(ctx, &sendOpt)
	return resp, err
}

// ObjectOptionsOptions ...
type ObjectOptionsOptions struct {
	Origin                      string `url:"-" header:"Origin"`
	AccessControlRequestMethod  string `url:"-" header:"Access-Control-Request-Method"`
	AccessControlRequestHeaders string `url:"-" header:"Access-Control-Request-Headers,omitempty"`
}

// Options Object请求实现跨域访问的预请求。即发出一个 OPTIONS 请求给服务器以确认是否可以进行跨域操作。
//
// 当CORS配置不存在时，请求返回403 Forbidden。
//
// https://www.qcloud.com/document/product/436/8288
func (s *ObjectService) Options(ctx context.Context, name string, opt *ObjectOptionsOptions) (*Response, error) {
	sendOpt := sendOptions{
		baseURL:   s.client.BaseURL.BucketURL,
		uri:       "/" + encodeURIComponent(name),
		method:    http.MethodOptions,
		optHeader: opt,
	}
	resp, err := s.client.send(ctx, &sendOpt)
	return resp, err
}

// Append ...
//
// Append请求可以将一个文件（Object）以分块追加的方式上传至 Bucket 中。使用Append Upload的文件必须事前被设定为Appendable。
// 当Appendable的文件被执行Put Object的操作以后，文件被覆盖，属性改变为Normal。
//
// 文件属性可以在Head Object操作中被查询到，当您发起Head Object请求时，会返回自定义Header『x-cos-object-type』，该Header只有两个枚举值：Normal或者Appendable。
//
// 追加上传建议文件大小1M - 5G。如果position的值和当前Object的长度不致，COS会返回409错误。
// 如果Append一个Normal的Object，COS会返回409 ObjectNotAppendable。
//
// Appendable的文件不可以被复制，不参与版本管理，不参与生命周期管理，不可跨区域复制。
//
// 当 r 不是 bytes.Buffer/bytes.Reader/strings.Reader 时，必须指定 opt.ObjectPutHeaderOptions.ContentLength
//
// https://www.qcloud.com/document/product/436/7741
func (s *ObjectService) Append(ctx context.Context, name string, position int, r io.Reader, opt *ObjectPutOptions) (*Response, error) {
	u := fmt.Sprintf("/%s?append&position=%d", encodeURIComponent(name), position)
	sendOpt := sendOptions{
		baseURL:   s.client.BaseURL.BucketURL,
		uri:       u,
		method:    http.MethodPost,
		optHeader: opt,
		body:      r,
	}
	resp, err := s.client.send(ctx, &sendOpt)
	return resp, err
}

// ObjectDeleteMultiOptions ...
type ObjectDeleteMultiOptions struct {
	XMLName xml.Name `xml:"Delete" header:"-"`
	Quiet   bool     `xml:"Quiet" header:"-"`
	Objects []Object `xml:"Object" header:"-"`
	//XCosSha1 string `xml:"-" header:"x-cos-sha1"`
}

// ObjectDeleteMultiResult ...
type ObjectDeleteMultiResult struct {
	XMLName        xml.Name `xml:"DeleteResult"`
	DeletedObjects []Object `xml:"Deleted,omitempty"`
	Errors         []struct {
		Key     string
		Code    string
		Message string
	} `xml:"Error,omitempty"`
}

// DeleteMulti ...
//
// Delete Multiple Object请求实现批量删除文件，最大支持单次删除1000个文件。
// 对于返回结果，COS提供Verbose和Quiet两种结果模式。Verbose模式将返回每个Object的删除结果；
// Quiet模式只返回报错的Object信息。
//
// 此请求必须携带x-cos-sha1用来校验Body的完整性。
//
// https://www.qcloud.com/document/product/436/8289
func (s *ObjectService) DeleteMulti(ctx context.Context, opt *ObjectDeleteMultiOptions) (*ObjectDeleteMultiResult, *Response, error) {
	var res ObjectDeleteMultiResult
	sendOpt := sendOptions{
		baseURL: s.client.BaseURL.BucketURL,
		uri:     "/?delete",
		method:  http.MethodPost,
		body:    opt,
		result:  &res,
	}
	resp, err := s.client.send(ctx, &sendOpt)
	return &res, resp, err
}

type objectPresignedURLTestingOptions struct {
	authTime *AuthTime
}

// PresignedURL 生成预签名授权 URL，可用于无需知道 SecretID 和 SecretKey 就可以上传和下载文件 。
//
// httpMethod:
//   * 下载文件：http.MethodGet
//   * 上传文件: http.MethodPut
//
// https://cloud.tencent.com/document/product/436/14116
// https://cloud.tencent.com/document/product/436/14114
func (s *ObjectService) PresignedURL(ctx context.Context, httpMethod, name string, auth Auth, opt interface{}) (*url.URL, error) {
	sendOpt := sendOptions{
		baseURL:   s.client.BaseURL.BucketURL,
		uri:       "/" + encodeURIComponent(name),
		method:    httpMethod,
		optQuery:  opt,
		optHeader: opt,
	}
	req, err := s.client.newRequest(ctx, &sendOpt)
	if err != nil {
		return nil, err
	}

	var authTime *AuthTime
	if opt != nil {
		if opt, ok := opt.(*objectPresignedURLTestingOptions); ok {
			authTime = opt.authTime
		}
	}
	if authTime == nil {
		authTime = NewAuthTime(auth.Expire)
	}
	authorization := newAuthorization(auth, req, *authTime)
	sign := encodeURIComponent(authorization)

	if req.URL.RawQuery == "" {
		req.URL.RawQuery = fmt.Sprintf("sign=%s", sign)
	} else {
		req.URL.RawQuery = fmt.Sprintf("%s&sign=%s", req.URL.RawQuery, sign)
	}
	return req.URL, nil
}

// Object ...
type Object struct {
	Key          string `xml:",omitempty"`
	ETag         string `xml:",omitempty"`
	Size         int    `xml:",omitempty"`
	PartNumber   int    `xml:",omitempty"`
	LastModified string `xml:",omitempty"`
	StorageClass string `xml:",omitempty"`
	Owner        *Owner `xml:",omitempty"`
}
