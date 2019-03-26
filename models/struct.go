//公用的model数据结构
package models

//文档上传表单结构
type FormUpload struct {
	Title, Md5, Intro, Tags, Ext, Filename string
	Chanel, Pid, Cid, Exist, Size, Price   int
	TmpFile                                string // 临时文件
}

//默认的SEO结构
type DefSeo struct {
	Title, Keywords, Description, Sitename string
}

//这个不是数据表，这个是搜索结果的json数据解析结构
type Result struct {
	Status     int64    `json:"status"`
	TotalFound int64    `json:"total_found"`
	Total      int64    `json:"total"`
	Ids        string   `json:"ids"`
	Word       []string `json:"word"`
	Msg        string   `json:"msg"`
	Time       float64  `json:"time"`
}

//邮箱配置
type Email struct {
	Id       int    `orm:"column(Id)"`                 //主键
	Username string `orm:"column(Username);default()"` //邮箱用户名
	Email    string `orm:"column(Email);default()"`    //邮箱
	Host     string `orm:"column(Host);default()"`     //主机
	Port     int    `orm:"column(Port);default(25)"`   //端口
	Password string `orm:"column(Password);default()"` //密码
}
