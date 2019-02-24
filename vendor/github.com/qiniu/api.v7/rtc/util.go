package rtc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/qiniu/api.v7/auth/qbox"
)

// resInfo is httpresponse infomation
type resInfo struct {
	Code int
	Err  error
}

func newResInfo() resInfo {
	info := resInfo{}
	return info
}

func getReqid(src *http.Header) string {
	for k, v := range *src {
		K := strings.Title(k)
		if strings.Contains(K, "Reqid") {
			return strings.Join(v, ", ")
		}
	}
	return ""
}

func buildURL(path string) string {
	if strings.Index(path, "/") != 0 {
		path = "/" + path
	}
	return "https://" + RtcHost + path
}

func postReq(httpClient *http.Client, mac *qbox.Mac, url string,
	reqParam interface{}, ret interface{}) *resInfo {
	info := newResInfo()
	var reqData []byte
	var err error

	switch v := reqParam.(type) {
	case *string:
		reqData = []byte(*v)
	case string:
		reqData = []byte(v)
	case *[]byte:
		reqData = *v
	case []byte:
		reqData = v
	default:
		reqData, err = json.Marshal(reqParam)
	}

	if err != nil {
		info.Err = err
		return &info
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqData))
	if err != nil {
		info.Err = err
		return &info
	}
	req.Header.Add("Content-Type", "application/json")
	return callReq(httpClient, req, mac, &info, ret)
}

func getReq(httpClient *http.Client, mac *qbox.Mac, url string, ret interface{}) *resInfo {
	info := newResInfo()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		info.Err = err
		return &info
	}
	return callReq(httpClient, req, mac, &info, ret)
}

func delReq(httpClient *http.Client, mac *qbox.Mac, url string, ret interface{}) *resInfo {
	info := newResInfo()
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		info.Err = err
		return &info
	}
	return callReq(httpClient, req, mac, &info, ret)
}

func callReq(httpClient *http.Client, req *http.Request, mac *qbox.Mac,
	info *resInfo, ret interface{}) (oinfo *resInfo) {
	oinfo = info
	accessToken, err := mac.SignRequestV2(req)
	if err != nil {
		info.Err = err
		return
	}
	req.Header.Add("Authorization", "Qiniu "+accessToken)
	client := httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		info.Err = err
		return
	}
	defer resp.Body.Close()
	info.Code = resp.StatusCode
	reqid := getReqid(&resp.Header)
	rebuildErr := func(msg string) error {
		return fmt.Errorf("Code: %v, Reqid: %v, %v", info.Code, reqid, msg)
	}

	if resp.ContentLength > 2*1024*1024 {
		err = rebuildErr(fmt.Sprintf("response is too long. Content-Length: %v", resp.ContentLength))
		info.Err = err
		return
	}
	resData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		info.Err = rebuildErr(err.Error())
		return
	}
	if info.Code != 200 {
		info.Err = rebuildErr(string(resData))
		return
	}
	if ret != nil {
		err = json.Unmarshal(resData, ret)
		if err != nil {
			info.Err = rebuildErr(fmt.Sprintf("err: %v, res: %v", err, resData))
		}
	}
	return
}
