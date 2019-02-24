package qbox

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/qiniu/api.v7/conf"
	"github.com/qiniu/x/bytes.v7/seekable"
)

// Mac 七牛AK/SK的对象，AK/SK可以从 https://portal.qiniu.com/user/key 获取。
type Mac struct {
	AccessKey string
	SecretKey []byte
}

// NewMac 构建一个新的拥有AK/SK的对象
func NewMac(accessKey, secretKey string) (mac *Mac) {
	return &Mac{accessKey, []byte(secretKey)}
}

// Sign 对数据进行签名，一般用于私有空间下载用途
func (mac *Mac) Sign(data []byte) (token string) {
	h := hmac.New(sha1.New, mac.SecretKey)
	h.Write(data)

	sign := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s:%s", mac.AccessKey, sign)
}

// SignWithData 对数据进行签名，一般用于上传凭证的生成用途
func (mac *Mac) SignWithData(b []byte) (token string) {
	encodedData := base64.URLEncoding.EncodeToString(b)
	h := hmac.New(sha1.New, mac.SecretKey)
	h.Write([]byte(encodedData))
	digest := h.Sum(nil)
	sign := base64.URLEncoding.EncodeToString(digest)
	return fmt.Sprintf("%s:%s:%s", mac.AccessKey, sign, encodedData)
}

// SignRequest 对数据进行签名，一般用于管理凭证的生成
func (mac *Mac) SignRequest(req *http.Request) (token string, err error) {
	h := hmac.New(sha1.New, mac.SecretKey)

	u := req.URL
	data := u.Path
	if u.RawQuery != "" {
		data += "?" + u.RawQuery
	}
	io.WriteString(h, data+"\n")

	if incBody(req) {
		s2, err2 := seekable.New(req)
		if err2 != nil {
			return "", err2
		}
		h.Write(s2.Bytes())
	}

	sign := base64.URLEncoding.EncodeToString(h.Sum(nil))
	token = fmt.Sprintf("%s:%s", mac.AccessKey, sign)
	return
}

// SignRequestV2 对数据进行签名，一般用于高级管理凭证的生成
func (mac *Mac) SignRequestV2(req *http.Request) (token string, err error) {
	h := hmac.New(sha1.New, mac.SecretKey)

	u := req.URL

	//write method path?query
	io.WriteString(h, fmt.Sprintf("%s %s", req.Method, u.Path))
	if u.RawQuery != "" {
		io.WriteString(h, "?")
		io.WriteString(h, u.RawQuery)
	}

	//write host and post
	io.WriteString(h, "\nHost: ")
	io.WriteString(h, req.Host)

	//write content type
	contentType := req.Header.Get("Content-Type")
	if contentType != "" {
		io.WriteString(h, "\n")
		io.WriteString(h, fmt.Sprintf("Content-Type: %s", contentType))
	}

	io.WriteString(h, "\n\n")

	//write body
	if incBodyV2(req) {
		s2, err2 := seekable.New(req)
		if err2 != nil {
			return "", err2
		}
		h.Write(s2.Bytes())
	}

	sign := base64.URLEncoding.EncodeToString(h.Sum(nil))
	token = fmt.Sprintf("%s:%s", mac.AccessKey, sign)
	return
}

// 管理凭证生成时，是否同时对request body进行签名
func incBody(req *http.Request) bool {
	return req.Body != nil && req.Header.Get("Content-Type") == conf.CONTENT_TYPE_FORM
}

func incBodyV2(req *http.Request) bool {
	contentType := req.Header.Get("Content-Type")
	return req.Body != nil && (contentType == conf.CONTENT_TYPE_FORM || contentType == conf.CONTENT_TYPE_JSON)
}

// VerifyCallback 验证上传回调请求是否来自七牛
func (mac *Mac) VerifyCallback(req *http.Request) (bool, error) {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return false, nil
	}

	token, err := mac.SignRequest(req)
	if err != nil {
		return false, err
	}

	return auth == "QBox "+token, nil
}

// Sign 一般用于下载凭证的签名
func Sign(mac *Mac, data []byte) string {
	return mac.Sign(data)
}

// SignWithData 一般用于上传凭证的签名
func SignWithData(mac *Mac, data []byte) string {
	return mac.SignWithData(data)
}

// VerifyCallback 验证上传回调请求是否来自七牛
func VerifyCallback(mac *Mac, req *http.Request) (bool, error) {
	return mac.VerifyCallback(req)
}
