package models

import (
	"fmt"
	"strings"

	"github.com/TruthHun/DocHub/helper"

	"errors"

	"strconv"

	"github.com/astaxie/beego/orm"
)

//文档资源状态，1正常，0文档未转换成功，-1删除，同时把id录入文档回收站id，-2表示删除了文档文件，但是数据库记录还保留。同时后台也看不到该记录
const (
	DocStatusFileDeleted int8 = -2
	DocStatusDeleted     int8 = -1
	DocStatusConverting  int8 = 0
	DocStatusNormal      int8 = 1
)

//文档表
type Document struct {
	Id          int    `orm:"Column(Id)"`
	Title       string `orm:"size(255);default();column(Title)"`       //文档名称【用户自定义的文档标题】
	Filename    string `orm:"size(255);default();column(Filename)"`    //文件名[文件的原文件名]
	Keywords    string `orm:"size(255);default();column(Keywords)"`    //文档标签、关键字
	Description string `orm:"size(255);default();column(Description)"` //文档摘要
}

func NewDocument() *Document {
	return &Document{}
}

func GetTableDocument() string {
	return getTable("document")
}

//文档信息表
type DocumentInfo struct {
	Id          int  `orm:"column(Id)"`
	DsId        int  `orm:"index;default(0);column(DsId)"`     //文档存档表Id,DocumentStore Id
	Uid         int  `orm:"index;default(0);column(Uid)"`      //文档上传用户的id
	ChanelId    int  `orm:"index;default(0);column(ChanelId)"` //文档所属频道
	Pid         int  `orm:"index;default(0);column(Pid)"`      //文档一级分类
	Cid         int  `orm:"index;default(0);column(Cid)"`      //频道下的最底层的分类id（二级分类），如幼儿教育下的幼儿读物等
	TimeCreate  int  `orm:"default(0);column(TimeCreate)"`     //文档上传时间
	TimeUpdate  int  `orm:"default(0);column(TimeUpdate)"`     //文档更新时间
	Dcnt        int  `orm:"default(0);column(Dcnt)"`           //下载次数
	Vcnt        int  `orm:"default(0);column(Vcnt)"`           //浏览次数
	Ccnt        int  `orm:"default(0);column(Ccnt)"`           //收藏次数
	Score       int  `orm:"default(30000);column(Score)"`      //默认30000，即表示3.0分。这是为了更准确统计评分的需要
	ScorePeople int  `orm:"default(0);column(ScorePeople)"`    //评分总人数
	Price       int  `orm:"default(0);column(Price)"`          //文档下载价格，0表示免费
	Status      int8 `orm:"default(0);column(Status)"`         //文档资源状态，1正常，0文档未转换成功，-1删除，同时把id录入文档回收站id，-2表示删除了文档文件，但是数据库记录还保留。同时后台也看不到该记录
}

func NewDocumentInfo() *DocumentInfo {
	return &DocumentInfo{}
}

func GetTableDocumentInfo() string {
	return getTable("document_info")
}

//文档存档表[供预览的文档存储在文档预览的OSS，完整文档存储在存储表]
type DocumentStore struct {
	Id          int    `orm:"column(Id)"`
	Md5         string `orm:"size(32);unique;column(Md5)"`             //文档md5
	Ext         string `orm:"size(10);default();column(Ext)"`          //文档扩展名，如pdf、xls等
	ExtCate     string `orm:"size(10);default();column(ExtCate)"`      //文档扩展名分类：word、ppt、text、pdf、xsl，code(这些分类配合图标一起使用，如word_24.png)
	ExtNum      int    `orm:"default(0);column(ExtNum)"`               //文档后缀的对应数字，主要是在coreseek搭建站内搜索时用到
	Page        int    `orm:"default(0);column(Page)"`                 //文档页数
	PreviewPage int    `orm:"default(50);column(PreviewPage)"`         //当前文档可预览页数
	Size        int    `orm:"default(0);column(Size)"`                 //文档大小
	ModTime     int    `orm:"default(0);column(ModTime)"`              //文档修改编辑时间
	PreviewExt  string `orm:"default(svg);column(PreviewExt);size(4)"` //文档预览的图片格式后缀，如jpg、png、svg等，默认svg
	Width       int    `orm:"default(0);column(Width)"`                //svg的原始宽度
	Height      int    `orm:"default(0);column(Height)"`               //svg的原始高度
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{}
}

func GetTableDocumentStore() string {
	return getTable("document_store")
}

// 完整的文档内容
type fullDocument struct {
	Id          int    `orm:"column(Id)"`
	Title       string `orm:"column(Title)"`              //文档名称【用户自定义的文档标题】
	Filename    string `orm:"column(Filename)"`           //文件名[文件的原文件名]
	Keywords    string `orm:"column(Keywords)"`           //文档标签、关键字
	Description string `orm:"column(Description)"`        //文档摘要
	Md5         string `orm:"column(Md5)"`                //文档md5
	Ext         string `orm:"column(Ext)"`                //文档扩展名，如pdf、xls等
	ExtCate     string `orm:"column(ExtCate)"`            //文档扩展名分类：word、ppt、text、pdf、xsl，code(这些分类配合图标一起使用，如word_24.png)
	ExtNum      int    `orm:"column(ExtNum)"`             //文档后缀的对应数字，主要是在coreseek搭建站内搜索时用到
	Page        int    `orm:"column(Page)"`               //文档页数
	PreviewPage int    `orm:"column(PreviewPage)"`        //当前文档可预览页数
	Size        int    `orm:"column(Size)"`               //文档大小
	ModTime     int    `orm:"column(ModTime)"`            //文档修改编辑时间
	PreviewExt  string `orm:"column(PreviewExt);size(4)"` //文档预览的图片格式后缀，如jpg、png、svg等，默认svg
	Width       int    `orm:"column(Width)"`              //svg的原始宽度
	Height      int    `orm:"column(Height)"`             //svg的原始高度
	DsId        int    `orm:"column(DsId)"`               //文档存档表Id,DocumentStore Id
	Uid         int    `orm:"column(Uid)"`                //文档上传用户的id
	Username    string `orm:"column(Username)"`           //文档上传用户的id
	ChanelId    int    `orm:"column(ChanelId)"`           //文档所属频道
	Pid         int    `orm:"column(Pid)"`                //文档一级分类
	Cid         int    `orm:"column(Cid)"`                //频道下的最底层的分类id（二级分类），如幼儿教育下的幼儿读物等
	TimeCreate  int    `orm:"column(TimeCreate)"`         //文档上传时间
	TimeUpdate  int    `orm:"column(TimeUpdate)"`         //文档更新时间
	Dcnt        int    `orm:"column(Dcnt)"`               //下载次数
	Vcnt        int    `orm:"column(Vcnt)"`               //浏览次数
	Ccnt        int    `orm:"column(Ccnt)"`               //收藏次数
	Score       int    `orm:"column(Score)"`              //默认30000，即表示3.0分。这是为了更准确统计评分的需要
	ScorePeople int    `orm:"column(ScorePeople)"`        //评分总人数
	Price       int    `orm:"column(Price)"`              //文档下载价格，0表示免费
	Status      int8   `orm:"column(Status)"`             //文档资源状态，1正常，0文档未转换成功，-1删除，同时把id录入文档回收站i
}

//非法文档(侵权或不良信息文档)MD5记录表
type DocumentIllegal struct {
	Id  int    `orm:"column(Id)"`                            //文档id
	Md5 string `orm:"size(32);unique;default();column(Md5)"` //文档md5
}

func NewDocumentIllegal() *DocumentIllegal {
	return &DocumentIllegal{}
}

func GetTableDocumentIllegal() string {
	return getTable("document_illegal")
}

//文档录入文档存档表
//@param            ds                   文档存储结构对象
//@return           id                   存储id
//@return           err                  错误
func (this *Document) InsertDocStore(ds *DocumentStore) (id int64, err error) {
	return orm.NewOrm().Insert(&ds)
}

//文档存入文档表
func (this *Document) InsertDoc(doc *Document) (int64, error) {
	return orm.NewOrm().Insert(&doc)
}

//文档信息录入文档信息表
func (this *Document) InsertDocInfo(info *DocumentInfo) (int64, error) {
	return orm.NewOrm().Insert(&info)
}

//根据md5判断文档是否是非法文档，如果是非法文档，则返回true
//@param                md5             md5
//@return               bool            如果文档存在于非法文档表中，则表示文档非法，否则合法
func (this *Document) IsIllegal(md5 string) bool {
	var ilg DocumentIllegal
	if orm.NewOrm().QueryTable(GetTableDocumentIllegal()).Filter("Md5", md5).One(&ilg); ilg.Id > 0 {
		return true
	}
	return false
}

//根据md5判断文档是否是非法文档，如果是非法文档，则返回true
//@param                id              文档id
//@return               bool            如果文档存在于非法文档表中，则表示文档非法，否则合法
func (this *Document) IsIllegalById(id interface{}) bool {
	var ilg DocumentIllegal
	if orm.NewOrm().QueryTable(GetTableDocumentIllegal()).Filter("Id", id).One(&ilg); ilg.Id > 0 {
		return true
	}
	return false
}

//根据文档id获取一篇文档的全部信息
//@param                id              文档id
//@return               params          文档信息
//@return               rows            记录数
//@return               err             错误
func (this *Document) GetById(id interface{}) (doc fullDocument, err error) {
	var sql string
	tables := []string{GetTableDocumentInfo() + " info", GetTableDocument() + " doc", GetTableDocumentStore() + " ds", GetTableUser() + " u"}
	fields := map[string][]string{
		"ds":   helper.DeleteSlice(GetFields(NewDocumentStore()), "Id"),
		"info": GetFields(NewDocumentInfo()),
		"u":    {"Username", "Id Uid"},
		"doc":  GetFields(NewDocument()),
	}
	on := []map[string]string{
		{"ds.Id": "info.DsId"},
		{"doc.Id": "info.Id"},
		{"u.Id": "info.Uid"},
	}
	helper.Logger.Debug("查询字段：%+v", fields)

	sql, err = LeftJoinSqlBuild(tables, on, fields, 1, 1, nil, nil, "info.Id=?")
	if err != nil {
		helper.Logger.Error(err.Error())
		err = errors.New("内部错误：数据查询失败")
		return
	}
	err = orm.NewOrm().Raw(sql, id).QueryRow(&doc)
	return
}

func (this *Document) GetDocument(id int, fields ...string) (doc Document) {
	orm.NewOrm().QueryTable(this).Filter("Id", id).One(&doc, fields...)
	return
}

//文档简易列表，用于首页或者其它查询简易字段，使用的时候，记得给条件加上表别名前缀。document_info别名前缀di，document_store别名前缀ds，document表名前缀d
//@param                condition               查询条件
//@param                limit                   查询记录限制
//@param                orderField              倒叙排序的字段。不需要表前缀。可用字段：Id，Ccnt，Dcnt，Vcnt，Score
//@return               params                  列表数据
//@return               rows                    记录数
//@return               err                     错误
func (this *Document) SimpleList(condition string, limit int, orderField ...string) (params []orm.Params, rows int64, err error) {
	condition = strings.Trim(condition, ",") + " and di.Status in(0,1)"
	order := "Id"
	if len(orderField) > 0 {
		odr := strings.ToLower(orderField[0])
		switch odr {
		case "ccnt", "dcnt", "vcnt", "score":
			order = helper.UpperFirst(odr)
		default:
			order = "Id"
		}
	}
	fields := "d.Title,d.Id,ds.Ext,ds.ExtCate"
	sqlFormat := `
	select %v from %v d left join %v di on di.Id=d.Id
	left join %v ds on ds.Id=di.DsId
	where %v group by d.Title order by di.%v desc limit %v
	`
	sql := fmt.Sprintf(sqlFormat, fields, GetTableDocument(), GetTableDocumentInfo(), GetTableDocumentStore(), condition, order, limit)
	if helper.Debug {
		helper.Logger.Debug("get simple list sql: %v", sql)
	}
	rows, err = orm.NewOrm().Raw(sql).Values(&params)
	return params, rows, err
}

//根据md5判断文档是否存在
//@param                md5str              文档的md5
//@return               Id                  文档存储表的id
func (this *Document) IsExistByMd5(md5str string) (Id int) {
	var ds DocumentStore
	orm.NewOrm().QueryTable(GetTableDocumentStore()).Filter("Md5", md5str).One(&ds)
	return ds.Id
}

//文档软删除，即把文档状态标记为-1，操作之后，需要把总文档数量、用户文档数量-1，同时把文档id移入回收站
func (this *Document) SoftDel(uid int, isAdmin bool, ids ...interface{}) (err error) {
	var (
		info []DocumentInfo
	)
	if len(ids) == 0 {
		return errors.New("文档id不能为空")
	}
	if orm.NewOrm().QueryTable(GetTableDocumentInfo()).Filter("Id__in", ids...).All(&info); len(info) > 0 {
		orm.NewOrm().QueryTable(GetTableDocumentInfo()).Filter("Id__in", ids...).Update(orm.Params{"Status": -1})
	}
	return nil
}

//根据document_store表中的id查询document_info表中的数据
func (this *Document) GetDocInfoByDsId(DsId ...interface{}) (info []DocumentInfo, rows int64, err error) {
	if l := len(DsId); l > 0 {
		rows, err = orm.NewOrm().QueryTable(GetTableDocumentInfo()).Filter("DsId__in", DsId...).All(&info)
	}
	return
}

//根据document_store表中的id查询document_info表中的数据
func (this *Document) GetDocStoreByDsId(DsId ...interface{}) (store []DocumentStore, rows int64, err error) {
	if l := len(DsId); l > 0 {
		rows, err = orm.NewOrm().QueryTable(GetTableDocumentStore()).Limit(l).Filter("Id__in", DsId...).All(&store)
	}
	return
}

//根据document_store表中的id查询document_info表中的数据
func (this *Document) GetOneDocStoreByDsId(DsId interface{}, fields ...string) (store DocumentStore, rows int64, err error) {
	err = orm.NewOrm().QueryTable(GetTableDocumentStore()).Filter("Id__in", DsId).One(&store, fields...)
	return
}

//根据document_store表中的id查询document_info表中的数据
func (this *Document) GetDocInfoById(Ids ...interface{}) (info []DocumentInfo, rows int64, err error) {
	if len(Ids) > 0 {
		rows, err = orm.NewOrm().QueryTable(GetTableDocumentInfo()).Limit(len(Ids)).Filter("Id__in", Ids...).All(&info)
	}
	return
}

//把文档标记为非法文档
//@param                ids             文档id
//@return               err             错误，nil表示成功
func (this *Document) SetIllegal(ids ...interface{}) (err error) {
	if length := int64(len(ids)); length > 0 {
		var (
			docInfo []DocumentInfo
			dsId    []interface{} //document_store表中的id
			did     []interface{} //文档id
			stores  []DocumentStore
			o       = orm.NewOrm()
		)

		length, err = o.QueryTable(GetTableDocumentInfo()).Filter("Id__in", ids...).Limit(length).All(&docInfo)
		if length > 0 {

			for _, v := range docInfo {
				dsId = append(dsId, v.DsId)
			}

			if docInfo, length, err = this.GetDocInfoByDsId(dsId...); length > 0 {
				for _, v := range docInfo {
					did = append(did, v.Id)
				}

				//将文档移入回收站，主要是避免之前文档没有被删除的情况
				if err = NewDocumentRecycle().RemoveToRecycle(0, false, did...); err != nil {
					helper.Logger.Error(err.Error())
				}

				//根据dsid查询文档md5，并把md5录入到非法文档表
				if stores, length, err = this.GetDocStoreByDsId(dsId...); err == nil && length > 0 {
					for _, store := range stores {
						o.Insert(&DocumentIllegal{Md5: store.Md5})
					}
				}
			}
		}
	}
	return
}

//根据文档id获取文档，并根据ids参数的id顺序返回搜索结果【主要用于搜索】
//@param                ids             文档id
//@param                num             记录数量
func (this *Document) GetDocsByIds(ids interface{}, num ...int) (data []orm.Params) {
	var values []orm.Params
	tables := []string{GetTableDocumentInfo() + " i", GetTableDocument() + " d", GetTableDocumentStore() + " ds"}
	on := []map[string]string{
		{"i.Id": "d.Id"},
		{"i.DsId": "ds.Id"},
	}
	fields := map[string][]string{
		"i":  {"Score", "TimeCreate", "Id", "Dcnt", "Vcnt", "Price"},
		"d":  {"Title", "Description"},
		"ds": {"Page", "Size", "ExtCate", "Md5"},
	}
	listRows := 10
	if len(num) > 0 {
		listRows = num[0]
	}
	if sql, err := LeftJoinSqlBuild(tables, on, fields, 1, listRows, nil, nil, fmt.Sprintf("i.Status>=0 and i.Id in(%v)", ids)); err == nil {
		orm.NewOrm().Raw(sql).Values(&values)
	} else {
		helper.Logger.Error(err.Error())
	}
	IdSlice := strings.Split(fmt.Sprintf("%v", ids), ",")
	for _, id := range IdSlice {
		for _, v := range values {
			if id == fmt.Sprintf("%v", v["Id"]) {
				data = append(data, v)
			}
		}
	}
	return
}

//文档简易列表
func (this *Document) TplSimpleList(chinelid interface{}) []orm.Params {
	data, _, _ := this.SimpleList(fmt.Sprintf("di.ChanelId=%v", helper.Interface2Int(chinelid)), 5)
	return data
}

//根据id查询搜索数据结构
//@param            id         根据id查询搜索文档
func (this *Document) GetDocForElasticSearch(id ...int) (es []ElasticSearchData, err error) {
	var (
		sql    string
		params []orm.Params
		num    int64
	)
	tables := []string{GetTableDocumentInfo() + " i", GetTableDocument() + " d", GetTableDocumentStore() + " ds"}
	on := []map[string]string{
		{"i.Id": "d.Id"},
		{"i.DsId": "ds.Id"},
	}
	fields := map[string][]string{
		"i":  {"Score", "Id", "Dcnt", "Vcnt", "Ccnt", "Price", "TimeCreate"},
		"d":  {"Title", "Description", "Keywords"},
		"ds": {"Page", "Size", "ExtNum DocType", "Id DsId"},
	}
	listRows := len(id)
	if listRows == 0 {
		err = errors.New("请至少传递一个文档Id")
		return
	}
	var idSlice []string
	for _, v := range id {
		idSlice = append(idSlice, strconv.Itoa(v))
	}
	if sql, err = LeftJoinSqlBuild(tables, on, fields, 1, listRows, nil, nil, fmt.Sprintf("i.Status>=0 and i.Id in(%v)", strings.Join(idSlice, ","))); err == nil {
		if num, err = orm.NewOrm().Raw(sql).Values(&params); num > 0 {
			for _, param := range params {
				es = append(es, ElasticSearchData{
					Id:          helper.Interface2Int(param["Id"]),
					Title:       param["Title"].(string),
					Keywords:    param["Keywords"].(string),
					Description: param["Description"].(string),
					Vcnt:        helper.Interface2Int(param["Vcnt"]),
					Ccnt:        helper.Interface2Int(param["Ccnt"]),
					Dcnt:        helper.Interface2Int(param["Dcnt"]),
					Score:       helper.Interface2Int(param["Score"]),
					Size:        helper.Interface2Int(param["Size"]),
					Page:        helper.Interface2Int(param["Page"]),
					DocType:     helper.Interface2Int(param["DocType"]),
					DsId:        helper.Interface2Int(param["DsId"]),
					Price:       helper.Interface2Int(param["Price"]),
					TimeCreate:  helper.Interface2Int(param["TimeCreate"]),
				})
			}
		}
	}
	return
}

//查询需要索引的稳定id
//@param            page            页面
//@param            pageSize        每页记录数
//@param            startTime       开始时间
//@param            fields          查询字段
//@return           infos           文档信息
//@return           rows            查询到的文档数量
//@return           err             查询错误
func (this *Document) GetDocInfoForElasticSearch(page, pageSize int, startTime int, fields ...string) (infos []DocumentInfo, rows int64, err error) {
	if len(fields) == 0 {
		fields = append(fields, "Id")
	}
	rows, err = orm.NewOrm().QueryTable(GetTableDocumentInfo()).Filter("Status__gte", 0).Filter("TimeUpdate__gte", startTime).Limit(pageSize).Offset((page-1)*pageSize).All(&infos, fields...)
	return
}
