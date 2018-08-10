package crawl

import (
	"strings"

	"github.com/astaxie/beego/httplib"
)

//构造request请求
func BuildRequest(method, url, referrer, cookie, os string, iscn, isjson bool, headers ...map[string]string) *httplib.BeegoHTTPRequest {
	var req *httplib.BeegoHTTPRequest
	switch strings.ToLower(method) {
	case "get":
		req = httplib.Get(url)
	case "post":
		req = httplib.Post(url)
	case "put":
		req = httplib.Put(url)
	case "delete":
		req = httplib.Delete(url)
	case "head":
		req = httplib.Head(url)
	default:
		req = httplib.Get(url)
	}

	//设置referrer
	if len(referrer) > 0 {
		req.Header("Referrer", referrer)
	}
	//设置cookie
	if len(cookie) > 0 {
		req.Header("Cookie", cookie)
	}
	//设置host
	hostSlice := strings.Split(url, "://")
	if len(hostSlice) > 1 {
		host := strings.Split(hostSlice[1], "/")[0]
		req.SetHost(host)
	}
	//压缩
	req.Header("Accept-Encoding", "gzip, deflate, br")
	//中文
	if iscn {
		req.Header("Accept-Language", "zh-CN,zh;q=0.8,en;q=0.6")
	} else {
		req.Header("Accept-Language", "en-US,en;q=0.8,zh;q=0.6")
	}
	//是否是json采集
	if isjson {
		req.Header("Accept", "application/json")
		req.Header("X-Request", "JSON")
		req.Header("X-Requested-With", "XMLHttpRequest")
	} else {
		req.Header("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	}

	//系统设置
	switch strings.ToLower(os) {
	case "windows":
		req.Header("User-Agent", "Mozilla/5.0 (Windows NT 6.2; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/32.0.1667.0 Safari/537.36")
	case "linux":
		req.Header("User-Agent", "Mozilla/5.0 (X11; U; Linux i686) AppleWebKit/534.15 (KHTML, like Gecko) Ubuntu/10.10 Chromium/10.0.613.0 Chrome/10.0.613.0 Safari/534.15")
	case "mac":
		req.Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3107.4 Safari/537.36")
	case "android":
		req.Header("User-Agent", "MQQBrowser/26 Mozilla/5.0 (Linux; U; Android 2.3.7; MB200 Build/GRJ22; CyanogenMod-7) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1")
	case "ios":
		req.Header("User-Agent", "Mozilla/5.0(iPhone; CPU iPhone OS 9_3_3 like Mac OS X)AppleWebkit/601.1.46(KHTML,like Gecko)Mobile/13G3")
	default: //mac
		req.Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3107.4 Safari/537.36")
	}

	//设置headers
	if len(headers) > 0 {
		for _, header := range headers {
			for k, v := range header {
				req.Header(k, v)
			}
		}
	}
	return req
}

//普通采集函数封装
//不需要设置UA，采集函数会根据os设置ua
//如果在headers设置ua，将重写ua，如果需要重写ua，确实可以这么做
func Crawl(method, url, referrer, cookie, os string, iscn, isjson, isdebug bool, headers ...map[string]string) (string, error) {
	req := BuildRequest(method, url, referrer, cookie, os, iscn, isjson, headers...)
	if isdebug {
		req.Debug(isdebug)
	}
	//设置headers
	if len(headers) > 0 {
		for _, header := range headers {
			for k, v := range header {
				req.Header(k, v)
			}
		}
	}
	return req.String()
}
