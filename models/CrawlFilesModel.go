package models

//采集文档

type CrawlFiles struct {
	Id          int    `orm:"column(Id)"`                            //自增id
	Title       string `orm:"column(Title);default()"`               //文档名称
	Size        int    `orm:"column(Size);default(0)"`               //文档大小
	Md5         string `orm:"column(Md5);default();unique;size(32)"` //文档md5
	Description string `orm:"column(Description);default()"`         //文档摘要
	Filetype    string `orm:"column(Filetype);default();size(10)"`   //文档扩展名,pdf,ppt,rtf,xls,doc
	Savepath    string `orm:"column(Savepath);default()"`            //文档在服务器本地的存放路径
	Status      int    `orm:"column(Status);default(0)"`             //文档状态，0表示未下载，1表示已下载，2表示已发布，即文档数据发布到数据库，文档被转码并更新到OSS
	Url         string `orm:"column(Url);unique"`                    //文档的原下载链接，唯一
	Domain      string `orm:"column(Domain);index"`                  //域名
}
