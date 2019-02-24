package storage

// Config 为文件上传，资源管理等配置
type Config struct {
	Zone          *Zone  //空间所在的机房
	UseHTTPS      bool   //是否使用https域名
	UseCdnDomains bool   //是否使用cdn加速域名
	CentralRsHost string //中心机房的RsHost，用于list bucket
	RsHost        string
	RsfHost       string
	UpHost        string
	ApiHost       string
	IoHost        string
}

func (c *Config) RsReqHost() string {
	if c.RsHost == "" {
		c.RsHost = DefaultRsHost
	}
	scheme := "http://"
	if c.UseHTTPS {
		scheme = "https://"
	}
	return scheme + c.RsHost
}
