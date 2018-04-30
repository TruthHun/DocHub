package models

import (
	"fmt"
	"time"

	"path/filepath"
	"strings"

	"os"

	"github.com/TruthHun/DocHub/helper"
)

//gitbook的表结构
//注意：gitbook的下载地址格式：https://www.gitbook.com/download/格式(epub,pdf,mobi)/book/gitbook的id(也就是Unid)
//访问用户的主页：https://www.gitbook.com/@用户名
type Gitbook struct {
	Id          int    `orm:"column(Id)"`                         //自增主键
	Unid        string `orm:"column(Unid);unique"`                //gitbook的id
	Title       string `orm:"column(Title)"`                      //标题
	Name        string `orm:"column(Name)"`                       //名称
	Description string `orm:"column(Description)"`                //摘要
	Topics      string `orm:"column(Topics)"`                     //话题，多个话题，用英文逗号分隔，这个相当于标签，文档分类时以这个为参考
	Cover       string `orm:"column(Cover)"`                      //封面，如果追加size=small，则获取到的就是小图
	Language    string `orm:"column(Language);size(10)"`          //所属语言，如en、zh、all等
	Md5Epub     string `orm:"column(Md5Epub);size(32)"`           //epub电子书的md5，注意，只有下载了的文档，才会有md5
	Md5Mobi     string `orm:"column(Md5Mobi);size(32)"`           //mobi电子书的md5，注意，只有下载了的文档，才会有md5
	Md5Pdf      string `orm:"column(Md5Pdf);size(32)"`            //pdf电子书的md5，注意，只有下载了的文档，才会有md5
	Stars       int    `orm:"column(Stars)"`                      //关注数量，采集时按照这个的倒叙排序进行采集和下载
	Status      bool   `orm:"column(Status)"`                     //是否已发布，这个是gitbook的状态
	DownPdf     bool   `orm:"column(DownPdf);default(false)"`     //是否下载了PDF
	DownEpub    bool   `orm:"column(DownEpub);default(false)"`    //是否下载了Epub
	DownMobi    bool   `orm:"column(DownMobi);default(false)"`    //是否下载了Mobi
	PublishPdf  bool   `orm:"column(PublishPdf);default(false)"`  //文档发布状态
	PublishEpub bool   `orm:"column(PublishEpub);defalt(false)"`  //文档发布状态
	PublishMobi bool   `orm:"column(PublishMobi);default(false)"` //文档发布状态
	ErrEpub     string `orm:"column(ErrEpub)"`                    //epub文档下载失败的错误提示
	ErrPdf      string `orm:"column(ErrPdf)"`                     //PDF文档下载失败的错误提示
	ErrMobi     string `orm:"column(ErrMobi)"`                    //ErrMobi文档下载失败的错误提示
}

//发布电子书
//@param            Chanel              频道
//@param            Parent              父类
//@param            Children            子类
//@param            Uid                 发布人
//@param            IdsArr              文档Id数组
func (this *Gitbook) PublishBooks(Chanel, Parent, Children, Uid int, IdsArr []string) {

	//遍历需要发布的书籍的ID
	for _, Id := range IdsArr {
		for {
			if GlobalGitbookNextAbled == true {
				GlobalGitbookNextAbled = false //不能继续发布下一本
				var book Gitbook
				O.QueryTable(TableGitbook).Filter("Id", Id).One(&book)
				if book.Id > 0 {
					this.execPublish(book, Chanel, Parent, Children, Uid)
					time.Sleep(1 * time.Second)
					GlobalGitbookNextAbled = true
					break
				}
			}
		}
		//暂停5分钟后再进行下一本书籍的发布，这里后面要修改成后台可控
		time.Sleep(300 * time.Second)
	}
	//重置为已完成本批次发布
	GlobalGitbookPublishing = false
}

//执行发布操作
func (this *Gitbook) execPublish(book Gitbook, Chanel, Parent, Children, Uid int) bool {
	SysData, _ := ModelSys.Get()
	//beego.Debug("发布书籍", book)
	//下载电子书，然后发布
	folder := fmt.Sprintf("uploads/%v/%v", time.Now().Format("2006/01/02"), Uid)
	fileType := map[string]string{"Pdf": "pdf", "Epub": "epub", "Mobi": "mobi"}
	for k, v := range fileType {
		if md5str, file, filename, err := helper.DownFile(fmt.Sprintf("https://www.gitbook.com/download/%v/book/%v", v, book.Unid), folder, ""); err != nil {
			//beego.Error(err)
			//将错误信息更新到数据库
			UpdateByField(TableGitbook, map[string]interface{}{fmt.Sprintf("Err%v", k): err.Error()}, "Id", book.Id)
		} else {
			var form FormUpload
			form.Ext = strings.TrimPrefix(filepath.Ext(file), ".")
			form.Md5 = md5str
			form.Title = book.Title
			form.Intro = book.Description
			form.Filename = filename
			form.Tags = helper.SegWord(book.Title)
			form.Chanel = Chanel
			form.Pid = Parent
			form.Cid = Children
			form.Price = 0
			//如果文档已经存在，则直接调用处理
			if ModelDoc.IsExistByMd5(form.Md5) > 0 {
				form.Exist = 1
				HandleExistDoc(Uid, form)
				//将信息更新到数据库
				UpdateByField(TableGitbook, map[string]interface{}{
					fmt.Sprintf("Md5%v", k):     md5str,
					fmt.Sprintf("Down%v", k):    1,
					fmt.Sprintf("Publish%v", k): 1,
				}, "Id", book.Id)
			} else {
				if info, err := os.Stat(file); err == nil {
					form.Size = int(info.Size())
					//beego.Debug(form)
					//将信息更新到数据库
					UpdateByField(TableGitbook, map[string]interface{}{
						fmt.Sprintf("Md5%v", k):     md5str,
						fmt.Sprintf("Down%v", k):    1,
						fmt.Sprintf("Publish%v", k): 1,
					}, "Id", book.Id)
					switch form.Ext {
					case "pdf": //处理pdf文档
						err = HandlePdf(Uid, file, form)
						if err != nil {
							helper.Logger.Error(err.Error())
						}
					case "umd", "epub", "chm", "txt", "mobi": //处理无法转码实现在线浏览的文档
						HandleUnOffice(Uid, file, form)
					default: //处理office文档
						HandleOffice(Uid, file, form)
					}
				} else {
					//将错误信息更新到数据库
					UpdateByField(TableGitbook, map[string]interface{}{fmt.Sprintf("Err%v", k): err.Error()}, "Id", book.Id)
				}
			}

			//积分增加
			price := SysData.Reward
			var log = CoinLog{
				Uid: Uid,
			}
			if form.Exist == 1 {
				price = 1 //已被分享过的文档，奖励1个金币
				log.Log = fmt.Sprintf("于%v成功分享了一篇已分享过的文档，获得 %v 个金币奖励", time.Now().Format("2006-01-02 15:04:05"), price)
			} else {
				log.Log = fmt.Sprintf("于%v成功分享了一篇未分享过的文档，获得 %v 个金币奖励", time.Now().Format("2006-01-02 15:04:05"), price)
			}
			log.Coin = price //金币变更
			if err := ModelCoinLog.LogRecord(log); err != nil {
				helper.Logger.Error(err.Error())
			}
			Regulate(TableUserInfo, "Coin", price, "Id=?", Uid)

		}
	}
	return true
}
