package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/TruthHun/DocHub/helper"

	"strconv"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/astaxie/beego/orm"
)

//文档回收站
type DocumentRecycle struct {
	Id   int  `orm:"column(Id)"`                  //对应的文档id
	Uid  int  `orm:"default(0);column(Uid)"`      //操作用户
	Date int  `orm:"default(0);column(Date)"`     //操作时间
	Self bool `orm:"default(false);column(Self)"` //是否是文档上传用户删除的，默认为false。如果是文档上传者删除的，设置为true
}

func NewDocumentRecycle() *DocumentRecycle {
	return &DocumentRecycle{}
}

func GetTableDocumentRecycle() string {
	return getTable("document_recycle")
}

//将文档从回收站中恢复过来，文档的状态必须是-1才可以
//@param            ids             文档id
//@return           err             返回错误，nil表示恢复成功，否则恢复失败
func (this *DocumentRecycle) RecoverFromRecycle(ids ...interface{}) (err error) {
	var (
		docInfo      []DocumentInfo
		dsId         []interface{} //document_store id
		o            = orm.NewOrm()
		md5Arr       []interface{}
		affectedRows int64
	)
	if len(ids) > 0 {
		qs := o.QueryTable(GetTableDocumentInfo()).Filter("Id__in", ids...).Filter("Status", -1)
		qs.All(&docInfo)
		if affectedRows, err = qs.Update(orm.Params{"Status": 1}); affectedRows > 0 {

			//总文档数量增加
			Regulate(GetTableSys(), "CntDoc", int(affectedRows), "Id=1")

			if len(docInfo) > 0 {
				client := NewElasticSearchClient()
				for _, v := range docInfo {
					dsId = append(dsId, v.DsId)

					//该用户的文档数量+1
					if err := Regulate(GetTableUserInfo(), "Document", 1, "Id=?", v.Uid); err != nil {
						helper.Logger.Error(err.Error())
					}

					//该分类下的文档数量+1
					Regulate(GetTableCategory(), "Cnt", 1, fmt.Sprintf("`Id` in(%v,%v,%v)", v.ChanelId, v.Cid, v.Pid))

					//新增索引
					client.BuildIndexById(v.Id)
				}
			}

			//从回收站中删除记录
			o.QueryTable(GetTableDocumentRecycle()).Filter("Id__in", ids...).Delete()

			if store, _, _ := NewDocument().GetDocStoreByDsId(dsId...); len(store) > 0 {
				for _, item := range store {
					md5Arr = append(md5Arr, item.Md5)
				}
			}
			//从非法文档中将文档记录移除（需要根据md5来进行移除）
			o.QueryTable(GetTableDocumentIllegal()).Filter("Md5__in", md5Arr...).Delete()
			return nil
		} else if err != nil {
			return err
		}
	}
	return errors.New("恢复的文档id不能为空")
}

//回收站文档列表
func (this *DocumentRecycle) RecycleList(p, listRows int) (params []orm.Params, rows int64, err error) {
	var sql string
	tables := []string{GetTableDocumentRecycle() + " dr", GetTableDocument() + " d", GetTableDocumentInfo() + " di", GetTableUser() + " u", GetTableDocumentStore() + " ds"}
	on := []map[string]string{
		{"dr.Id": "d.Id"},
		{"d.Id": "di.Id"},
		{"u.Id": "di.Uid"},
		{"di.DsId": "ds.Id"},
	}
	fields := map[string][]string{
		"dr": {"Date", "Self"},
		"d":  {"Title", "Filename", "Id"},
		"ds": {"Md5", "Ext", "ExtCate", "Page", "Size"},
		"u":  {"Username", "Id Uid"},
	}
	if sql, err = LeftJoinSqlBuild(tables, on, fields, p, listRows, []string{"dr.Date desc"}, nil, "dr.Id>0"); err == nil {
		rows, err = orm.NewOrm().Raw(sql).Values(&params)
	}
	return
}

//将文档移入回收站(软删除)
//@param            uid         操作人，即将文档移入回收站的人
//@param            self        是否是用户自己操作
//@param            ids         文档id，即需要删除的文档id
//@return           errs        错误
func (this *DocumentRecycle) RemoveToRecycle(uid interface{}, self bool, ids ...interface{}) (errs []string) {
	//软删除
	//1、将文档状态标记为-1
	//2、将文档id录入到回收站
	//3、用户文档数量减少
	//4、整站文档数量减少
	//5、分类下的文档减少
	//不需要删除用户的收藏记录
	//不需要删除文档的评分记录

	var docInfo []DocumentInfo
	if len(ids) > 0 {
		orm.NewOrm().QueryTable(GetTableDocumentInfo()).Filter("Id__in", ids...).All(&docInfo)
		//总文档记录减少
		Regulate(GetTableSys(), "CntDoc", -len(docInfo), "Id=1")
		for _, info := range docInfo {
			//文档分类统计数量减少
			if err := Regulate(GetTableCategory(), "Cnt", -1, "Id in(?,?,?)", info.ChanelId, info.Pid, info.Cid); err != nil {
				helper.Logger.Error(err.Error())
			}
			//用户文档统计数量减少
			if err := Regulate(GetTableUserInfo(), "Document", -1, "Id=?", info.Uid); err != nil {
				helper.Logger.Error(err.Error())
			}
		}
		//变更文档状态
		if _, err := UpdateByIds(GetTableDocumentInfo(), "Status", -1, ids...); err != nil {
			helper.Logger.Error(err.Error())
			errs = append(errs, err.Error())
		}

		client := NewElasticSearchClient()
		//移入回收站
		for _, id := range ids {
			var rc DocumentRecycle
			rc.Id = helper.Interface2Int(id)
			rc.Uid = helper.Interface2Int(uid)
			rc.Date = int(time.Now().Unix())
			rc.Self = self
			if _, err := orm.NewOrm().Insert(&rc); err != nil {
				helper.Logger.Error(err.Error())
			}
			//删除索引
			client.DeleteIndex(rc.Id)
		}
	} else {
		errs = append(errs, "参数错误:缺少文档id")
	}
	return errs
}

//彻底删除文档，包括删除文档记录（被收藏的记录、用户的发布记录、扣除用户获得的积分），删除文档文件
func (this *DocumentRecycle) DeepDel(ids ...interface{}) (err error) {
	//根据md5，找到OSS文件（封面文件、PDF文件、文件夹）
	var (
		dsId  []interface{}
		store []DocumentStore
		o     = orm.NewOrm()
	)

	if info, rows, errInfo := NewDocument().GetDocInfoById(ids...); rows > 0 {
		for _, item := range info {
			dsId = append(dsId, item.DsId)
		}
	} else if errInfo != orm.ErrNoRows && errInfo != nil {
		return errInfo
	} else if rows == 0 {
		return
	}

	if err = this.DelRows(ids...); err != orm.ErrNoRows && err != nil {
		return
	}

	if store, _, err = NewDocument().GetDocStoreByDsId(dsId...); err != orm.ErrNoRows && err != nil {
		return
	}

	if _, err = o.QueryTable(GetTableDocumentStore()).Filter("Id__in", dsId).Delete(); err != orm.ErrNoRows && err != nil {
		return
	}

	go func() {
		for _, item := range store {
			this.DelFile(item.Md5, item.Ext, item.PreviewExt)
		}
	}()

	return
}

//删除文档记录
func (this *DocumentRecycle) DelRows(ids ...interface{}) (err error) {
	//1、删除被收藏的收藏记录
	//2、删除文档的评论(评分)记录
	//3、删除document表的记录
	//4、删除document_info表的记录
	//【这个不删除】5、删除document_store表的记录
	//6、删除回收站中的记录
	var (
		o = orm.NewOrm()
	)

	defer func() {
		if err != nil {
			fmt.Println("Ids:", ids)
			helper.Logger.Error(err.Error())
		}
	}()

	if err = NewCollect().DelByDocId(ids...); err != orm.ErrNoRows && err != nil {
		return
	}

	//删除评论
	if err = NewDocumentComment().DelCommentByDocId(ids...); err != orm.ErrNoRows && err != nil {
		return
	}

	if _, err = o.QueryTable(GetTableDocumentInfo()).Filter("Id__in", ids).Delete(); err != orm.ErrNoRows && err != nil {
		return
	}

	if _, err = o.QueryTable(GetTableDocument()).Filter("Id__in", ids...).Delete(); err != orm.ErrNoRows && err != nil {
		return
	}

	if _, err = o.QueryTable(GetTableDocumentRecycle()).Filter("Id__in", ids...).Delete(); err != orm.ErrNoRows && err != nil {
		return
	}

	return
}

//根据md5，删除文档封面等
//@param                md5             文档md5
//@param                ext             源文档(为转成pdf之前的文档)的扩展名
//@param                prevExt         预览文件的扩展名，一般是svg
func (this *DocumentRecycle) DelFile(md5 string, ext string, prevExt string) {
	var (
		bucketPrivate *oss.Bucket
		bucketPublic  *oss.Bucket
		err           error

		cover     = md5 + ".jpg" //封面文件
		prevFiles []string

		folder       = md5
		originalFile = md5 + "." + strings.TrimLeft(ext, ".")
		pdfFile      = md5 + ".pdf"
	)

	fmt.Println("=====删除文件====", md5)

	if !strings.HasPrefix(prevExt, ".") {
		prevExt = "." + prevExt
	}

	if bucketPrivate, err = NewOss().NewBucket(false); err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	if bucketPublic, err = NewOss().NewBucket(true); err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	if err = bucketPublic.DeleteObject(cover); err != nil {
		helper.Logger.Error(err.Error())
	}

	if _, err = bucketPrivate.DeleteObjects([]string{originalFile, pdfFile}); err != nil {
		helper.Logger.Error(err.Error())
	}

	//OSS SDK没发现有可以直接删除文件夹的，所以这样去删除文件
	for i := 1; i <= 1000; i++ {
		prevFiles = append(prevFiles, folder+"/"+strconv.Itoa(i)+prevExt)
	}
	if _, err = bucketPublic.DeleteObjects(prevFiles); err != nil {
		helper.Logger.Error(err.Error())
	}

}
