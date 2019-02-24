package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Zone 为空间对应的机房属性，主要包括了上传，资源管理等操作的域名
type Zone struct {
	SrcUpHosts []string
	CdnUpHosts []string
	RsHost     string
	RsfHost    string
	ApiHost    string
	IovipHost  string
}

func (z *Zone) String() string {
	str := ""
	str += fmt.Sprintf("SrcUpHosts: %v\n", z.SrcUpHosts)
	str += fmt.Sprintf("CdnUpHosts: %v\n", z.CdnUpHosts)
	str += fmt.Sprintf("IovipHost: %s\n", z.IovipHost)
	str += fmt.Sprintf("RsHost: %s\n", z.RsHost)
	str += fmt.Sprintf("RsfHost: %s\n", z.RsfHost)
	str += fmt.Sprintf("ApiHost: %s\n", z.ApiHost)
	return str
}

func (z *Zone) GetRsfHost(useHttps bool) string {

	scheme := "http://"
	if useHttps {
		scheme = "https://"
	}

	return fmt.Sprintf("%s%s", scheme, z.RsfHost)
}

func (z *Zone) GetIoHost(useHttps bool) string {

	scheme := "http://"
	if useHttps {
		scheme = "https://"
	}

	return fmt.Sprintf("%s%s", scheme, z.IovipHost)
}

func (z *Zone) GetRsHost(useHttps bool) string {

	scheme := "http://"
	if useHttps {
		scheme = "https://"
	}

	return fmt.Sprintf("%s%s", scheme, z.RsHost)
}

func (z *Zone) GetApiHost(useHttps bool) string {

	scheme := "http://"
	if useHttps {
		scheme = "https://"
	}

	return fmt.Sprintf("%s%s", scheme, z.ApiHost)
}

// ZoneHuadong 表示华东机房
var ZoneHuadong = Zone{
	SrcUpHosts: []string{
		"up.qiniup.com",
		"up-nb.qiniup.com",
		"up-xs.qiniup.com",
	},
	CdnUpHosts: []string{
		"upload.qiniup.com",
		"upload-nb.qiniup.com",
		"upload-xs.qiniup.com",
	},
	RsHost:    "rs.qbox.me",
	RsfHost:   "rsf.qbox.me",
	ApiHost:   "api.qiniu.com",
	IovipHost: "iovip.qbox.me",
}

// ZoneHuabei 表示华北机房
var ZoneHuabei = Zone{
	SrcUpHosts: []string{
		"up-z1.qiniup.com",
	},
	CdnUpHosts: []string{
		"upload-z1.qiniup.com",
	},
	RsHost:    "rs-z1.qbox.me",
	RsfHost:   "rsf-z1.qbox.me",
	ApiHost:   "api-z1.qiniu.com",
	IovipHost: "iovip-z1.qbox.me",
}

// ZoneHuanan 表示华南机房
var ZoneHuanan = Zone{
	SrcUpHosts: []string{
		"up-z2.qiniup.com",
		"up-gz.qiniup.com",
		"up-fs.qiniup.com",
	},
	CdnUpHosts: []string{
		"upload-z2.qiniup.com",
		"upload-gz.qiniup.com",
		"upload-fs.qiniup.com",
	},
	RsHost:    "rs-z2.qbox.me",
	RsfHost:   "rsf-z2.qbox.me",
	ApiHost:   "api-z2.qiniu.com",
	IovipHost: "iovip-z2.qbox.me",
}

// ZoneBeimei 表示北美机房
var ZoneBeimei = Zone{
	SrcUpHosts: []string{
		"up-na0.qiniup.com",
	},
	CdnUpHosts: []string{
		"upload-na0.qiniup.com",
	},
	RsHost:    "rs-na0.qbox.me",
	RsfHost:   "rsf-na0.qbox.me",
	ApiHost:   "api-na0.qiniu.com",
	IovipHost: "iovip-na0.qbox.me",
}

// ZoneXinjiapo 表示新加坡机房
var ZoneXinjiapo = Zone{
	SrcUpHosts: []string{
		"up-as0.qiniup.com",
	},
	CdnUpHosts: []string{
		"upload-as0.qiniup.com",
	},
	RsHost:    "rs-as0.qbox.me",
	RsfHost:   "rsf-as0.qbox.me",
	ApiHost:   "api-as0.qiniu.com",
	IovipHost: "iovip-as0.qbox.me",
}

// for programmers
var Zone_z0 = ZoneHuadong
var Zone_z1 = ZoneHuabei
var Zone_z2 = ZoneHuanan
var Zone_na0 = ZoneBeimei
var Zone_as0 = ZoneXinjiapo

// UcHost 为查询空间相关域名的API服务地址
const UcHost = "https://uc.qbox.me"

// UcQueryRet 为查询请求的回复
type UcQueryRet struct {
	TTL int                            `json:"ttl"`
	Io  map[string]map[string][]string `json:"io"`
	Up  map[string]UcQueryUp           `json:"up"`
}

// UcQueryUp 为查询请求回复中的上传域名信息
type UcQueryUp struct {
	Main   []string `json:"main,omitempty"`
	Backup []string `json:"backup,omitempty"`
	Info   string   `json:"info,omitempty"`
}

var (
	zoneMutext sync.RWMutex
	zoneCache  = make(map[string]*Zone)
)

// GetZone 用来根据ak和bucket来获取空间相关的机房信息
func GetZone(ak, bucket string) (zone *Zone, err error) {
	zoneID := fmt.Sprintf("%s:%s", ak, bucket)
	//check from cache
	zoneMutext.RLock()
	if v, ok := zoneCache[zoneID]; ok {
		zone = v
	}
	zoneMutext.RUnlock()
	if zone != nil {
		return
	}

	//query from server
	reqURL := fmt.Sprintf("%s/v2/query?ak=%s&bucket=%s", UcHost, ak, bucket)
	var ret UcQueryRet
	ctx := context.TODO()
	qErr := DefaultClient.CallWithForm(ctx, &ret, "GET", reqURL, nil, nil)
	if qErr != nil {
		err = fmt.Errorf("query zone error, %s", qErr.Error())
		return
	}

	ioHost := ret.Io["src"]["main"][0]
	srcUpHosts := ret.Up["src"].Main
	if ret.Up["src"].Backup != nil {
		srcUpHosts = append(srcUpHosts, ret.Up["src"].Backup...)
	}
	cdnUpHosts := ret.Up["acc"].Main
	if ret.Up["acc"].Backup != nil {
		cdnUpHosts = append(cdnUpHosts, ret.Up["acc"].Backup...)
	}

	zone = &Zone{
		SrcUpHosts: srcUpHosts,
		CdnUpHosts: cdnUpHosts,
		IovipHost:  ioHost,
		RsHost:     DefaultRsHost,
		RsfHost:    DefaultRsfHost,
		ApiHost:    DefaultAPIHost,
	}

	//set specific hosts if possible
	setSpecificHosts(ioHost, zone)

	zoneMutext.Lock()
	zoneCache[zoneID] = zone
	zoneMutext.Unlock()
	return
}

func setSpecificHosts(ioHost string, zone *Zone) {
	if strings.Contains(ioHost, "-z1") {
		zone.RsHost = "rs-z1.qbox.me"
		zone.RsfHost = "rsf-z1.qbox.me"
		zone.ApiHost = "api-z1.qiniu.com"
	} else if strings.Contains(ioHost, "-z2") {
		zone.RsHost = "rs-z2.qbox.me"
		zone.RsfHost = "rsf-z2.qbox.me"
		zone.ApiHost = "api-z2.qiniu.com"
	} else if strings.Contains(ioHost, "-na0") {
		zone.RsHost = "rs-na0.qbox.me"
		zone.RsfHost = "rsf-na0.qbox.me"
		zone.ApiHost = "api-na0.qiniu.com"
	} else if strings.Contains(ioHost, "-as0") {
		zone.RsHost = "rs-as0.qbox.me"
		zone.RsfHost = "rsf-as0.qbox.me"
		zone.ApiHost = "api-as0.qiniu.com"
	}
}
