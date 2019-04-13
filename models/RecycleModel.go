package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/TruthHun/DocHub/helper"

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
		docInfo []DocumentInfo
		dsId    []interface{} //document_store id
		o       = orm.NewOrm()
		md5Arr  []interface{}
	)

	if len(ids) == 0 {
		return
	}
	o.Begin()
	defer func() {
		if err != nil {
			o.Rollback()
		} else {
			o.Commit()
		}
	}()

	qs := o.QueryTable(GetTableDocumentInfo()).Filter("Id__in", ids...).Filter("Status", DocStatusDeleted)
	qs.All(&docInfo)
	_, err = qs.Update(orm.Params{"Status": DocStatusNormal})
	if err != nil {
		return
	}

	//总文档数量增加
	sqlSys := fmt.Sprintf("update %v set `CntDoc`=`CntDoc`+? where Id = 1", GetTableSys())
	_, err = o.Raw(sqlSys, len(docInfo)).Exec()
	if err != nil {
		return
	}
	reward := NewSys().GetByField("Reward").Reward
	client := NewElasticSearchClient()
	sqlUser := fmt.Sprintf("update %v set `Document`=`Document`+?,`Coin`=`Coin`+? where Id=?", GetTableUserInfo())
	sqlCate := fmt.Sprintf("update %v set `Cnt`=`Cnt`+? where `Id` in(?,?,?)", GetTableCategory())
	now := int(time.Now().Unix())
	doc := NewDocument()
	for _, v := range docInfo {
		dsId = append(dsId, v.DsId)

		_, err = o.Raw(sqlUser, 1, reward, v.Uid).Exec()
		if err != nil {
			return
		}

		// 积分变更
		log := &CoinLog{Uid: v.Uid, Coin: reward, TimeCreate: now}
		log.Log = fmt.Sprintf("系统恢复《%v》文档，获得 %v 个金币奖励", doc.GetDocument(v.Id, "Title").Title, reward)
		_, err = o.Insert(log)
		if err != nil {
			return
		}

		_, err = o.Raw(sqlCate, 1, v.ChanelId, v.Cid, v.Pid).Exec()
		if err != nil {
			return
		}

		client.BuildIndexById(v.Id) //新增索引
	}

	//从回收站中删除记录
	_, err = o.QueryTable(GetTableDocumentRecycle()).Filter("Id__in", ids...).Delete()

	if store, _, _ := NewDocument().GetDocStoreByDsId(dsId...); len(store) > 0 {
		for _, item := range store {
			md5Arr = append(md5Arr, item.Md5)
		}
	}
	_, err = o.QueryTable(GetTableDocumentIllegal()).Filter("Md5__in", md5Arr...).Delete()
	return
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
func (this *DocumentRecycle) RemoveToRecycle(uid interface{}, self bool, ids ...interface{}) (err error) {
	//软删除
	//1、将文档状态标记为-1
	//2、将文档id录入到回收站
	//3、用户文档数量减少
	//4、整站文档数量减少
	//5、分类下的文档减少
	//不需要删除用户的收藏记录
	//不需要删除文档的评分记录

	var docInfo []DocumentInfo
	sys, _ := NewSys().Get()

	if len(ids) == 0 {
		return
	}

	o := orm.NewOrm()

	o.QueryTable(GetTableDocumentInfo()).Filter("Id__in", ids...).Filter("Status__in", DocStatusNormal, DocStatusConverting).All(&docInfo)

	if len(docInfo) == 0 {
		return
	}

	o.Begin()
	defer func() {
		if err == nil {
			o.Commit()
		} else {
			o.Rollback()
		}
	}()

	sqlSys := fmt.Sprintf("update %v set `CntDoc`=`CntDoc`-? where Id=1", GetTableSys())
	_, err = o.Raw(sqlSys, len(docInfo)).Exec()
	if err != nil {
		return
	}

	sqlUser := fmt.Sprintf("update %v set `Document`=`Document`-? ,`Coin`=`Coin`-? where Id=?", GetTableUserInfo())
	sqlCate := fmt.Sprintf("update %v set `Cnt`=`Cnt`-? where `Id` in(?,?,?)", GetTableCategory())
	doc := NewDocument()
	now := int(time.Now().Unix())
	for _, info := range docInfo {
		_, err = o.Raw(sqlCate, 1, info.ChanelId, info.Pid, info.Cid).Exec()
		if err != nil {
			return
		}

		_, err = o.Raw(sqlUser, 1, sys.Reward, info.Uid).Exec() //用户个人的文档数量和金币减少
		if err != nil {
			return
		}

		log := CoinLog{
			Uid:        info.Uid,
			Coin:       -sys.Reward,
			TimeCreate: now,
		}

		log.Log = fmt.Sprintf("删除《%v》，扣除 %v 个金币", doc.GetDocument(info.Id, "Title").Title, sys.Reward)
		if !self {
			log.Log = "系统" + log.Log
		}
		_, err = o.Insert(&log)
		if err != nil {
			return
		}
	}

	//变更文档状态
	if _, err := UpdateByIds(GetTableDocumentInfo(), "Status", -1, ids...); err != nil {
		helper.Logger.Error(err.Error())
	}
	marks := strings.Join(make([]string, len(ids)+1), "?")
	sqlDocStatus := fmt.Sprintf("update %v set `Status`=-1 where Id in(%v)", GetTableDocumentInfo(), marks)
	_, err = o.Raw(sqlDocStatus, ids...).Exec()
	if err != nil {
		return
	}

	for _, id := range ids {
		var rc DocumentRecycle
		rc.Id = helper.Interface2Int(id)
		rc.Uid = helper.Interface2Int(uid)
		rc.Date = now
		rc.Self = self
		if _, err = o.Insert(&rc); err != nil { // 移入回收站
			return
		}

	}

	client := NewElasticSearchClient()
	for _, id := range ids { //删除索引
		client.DeleteIndex(helper.Interface2Int(id))
	}

	return
}

// 彻底删除文档，包括删除文档记录（被收藏的记录、用户的发布记录、扣除用户获得的积分(如果此刻文档的状态不是待删除)），删除文档文件
func (this *DocumentRecycle) DeepDel(ids ...interface{}) (err error) {
	// 文档id找到文档的dsId，再根据dsId查找到全部的文档id
	// 根据文档id，将文档全部移入回收站
	// 删除文档记录
	//
	var (
		dsId  []interface{}
		info  []DocumentInfo
		store []DocumentStore
		o     = orm.NewOrm()
	)

	info, _, err = NewDocument().GetDocInfoById(ids...)
	if err != nil && err != orm.ErrNoRows {
		return
	}

	if len(info) == 0 {
		return
	}

	for _, item := range info {
		dsId = append(dsId, item.DsId)
	}

	info, _, _ = NewDocument().GetDocInfoByDsId(dsId)
	if len(info) > 0 {
		ids = []interface{}{}
		for _, item := range info {
			ids = append(ids, item.Id)
		}
	}

	err = this.RemoveToRecycle(0, false, ids...)
	if err != nil {
		return
	}

	if err = this.DelRows(ids...); err != nil && err != orm.ErrNoRows {
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
			this.DelFile(item.Md5, item.Ext, item.PreviewExt, item.PreviewPage)
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

//根据md5，删除文档、封面、预览文件等
//@param                md5             文档md5
func (this *DocumentRecycle) DelFile(md5, oriExt, prevExt string, previewPagesCount int) (err error) {

	var (
		cover         = md5 + ".jpg" //封面文件
		pdfFile       = md5 + helper.ExtPDF
		oriFile       = md5 + "." + strings.TrimLeft(oriExt, ".")
		svgFormat     = md5 + "/%v." + strings.TrimLeft(prevExt, ".")
		clientPublic  *CloudStore
		clientPrivate *CloudStore
	)

	if previewPagesCount <= 0 {
		previewPagesCount = 1000 // default
	}

	clientPublic, err = NewCloudStore(false)
	if err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	clientPrivate, err = NewCloudStore(true)
	if err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	if err = clientPrivate.Delete(oriFile, pdfFile); err != nil {
		helper.Logger.Error(err.Error())
	}

	if err = clientPublic.Delete(cover); err != nil {
		helper.Logger.Error(err.Error())
	}

	var svgs []string
	for i := 0; i < previewPagesCount; i++ {
		svgs = append(svgs, fmt.Sprintf(svgFormat, i+1))
	}

	if err = clientPublic.Delete(svgs...); err != nil {
		helper.Logger.Error(err.Error())
	}

	return
}
