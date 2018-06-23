package models

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TruthHun/DocHub/helper"

	"errors"

	"github.com/astaxie/beego/orm"
)

//文档表
type Document struct {
	Id          int    `orm:"Column(Id)"`
	Title       string `orm:"size(255);default();column(Title)"`       //文档名称【用户自定义的文档标题】
	Filename    string `orm:"size(255);default();column(Filename)"`    //文件名[文件的原文件名]
	Keywords    string `orm:"size(255);default();column(Keywords)"`    //文档标签、关键字
	Description string `orm:"size(255);default();column(Description)"` //文档摘要
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
	Dcnt        int  `orm:"default(0);column(Dcnt)"`           //下载次数
	Vcnt        int  `orm:"default(0);column(Vcnt)"`           //浏览次数
	Ccnt        int  `orm:"default(0);column(Ccnt)"`           //收藏次数
	Score       int  `orm:"default(30000);column(Score)"`      //默认30000，即表示3.0分。这是为了更准确统计评分的需要
	ScorePeople int  `orm:"default(0);column(ScorePeople)"`    //评分总人数
	Price       int  `orm:"default(0);column(Price)"`          //文档下载价格，0表示免费
	Status      int8 `orm:"default(0);column(Status)"`         //文档资源状态，1正常，0文档未转换成功，-1删除，同时把id录入文档回收站id，-2表示删除了文档文件，但是数据库记录还保留。同时后台也看不到该记录
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

//非法文档(侵权或不良信息文档)MD5记录表
type DocumentIllegal struct {
	Id  int    `orm:"column(Id)"`                            //文档id
	Md5 string `orm:"size(32);unique;default();column(Md5)"` //文档md5
}

//文档录入文档存档表
//@param            ds              DocumentStore           文档存储结构对象
//@return                           int64                   存储id
//@return                           error                   错误
func (this *Document) InsertDocStore(ds *DocumentStore) (int64, error) {
	return O.Insert(&ds)
}

//文档存入文档表
func (this *Document) InsertDoc(doc *Document) (int64, error) {
	return O.Insert(&doc)
}

//文档信息录入文档信息表
func (this *Document) InsertDocInfo(info *DocumentInfo) (int64, error) {
	return O.Insert(&info)
}

//根据md5判断文档是否是非法文档，如果是非法文档，则返回true
//@param                md5             md5
//@return               bool            如果文档存在于非法文档表中，则表示文档非法，否则合法
func (this *Document) IsIllegal(md5 string) bool {
	var ilg DocumentIllegal
	if O.QueryTable(TableDocIllegal).Filter("Md5", md5).One(&ilg); ilg.Id > 0 {
		return true
	}
	return false
}

//根据md5判断文档是否是非法文档，如果是非法文档，则返回true
//@param                id              文档id
//@return               bool            如果文档存在于非法文档表中，则表示文档非法，否则合法
func (this *Document) IsIllegalById(id interface{}) bool {
	var ilg DocumentIllegal
	if O.QueryTable(TableDocIllegal).Filter("Id", id).One(&ilg); ilg.Id > 0 {
		return true
	}
	return false
}

//根据文档id获取一篇文档的全部信息
//@param                id              文档id
//@return               params          文档信息
//@return               rows            记录数
//@return               err             错误
func (this *Document) GetById(id interface{}) (params orm.Params, rows int64, err error) {
	var data []orm.Params
	tables := []string{TableDocInfo + " info", TableDoc + " doc", TableDocStore + " ds", TableUser + " u"}
	fields := map[string][]string{
		"ds":   GetFields(ModelDocStore),
		"info": GetFields(ModelDocInfo),
		"u":    {"Username", "Id Uid"},
		"doc":  GetFields(ModelDoc),
	}
	on := []map[string]string{
		{"ds.Id": "info.DsId"},
		{"doc.Id": "info.Id"},
		{"u.Id": "info.Uid"},
	}
	if sql, err := LeftJoinSqlBuild(tables, on, fields, 1, 1, nil, nil, "info.Id=?"); err == nil {
		if rows, err = O.Raw(sql, id).Values(&data); len(data) > 0 {
			params = data[0]
		}
	}
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
	sql_format := `
	select %v from %v d left join %v di on di.Id=d.Id
	left join %v ds on ds.Id=di.DsId
	where %v group by d.Title order by di.%v desc limit %v
	`
	sql := fmt.Sprintf(sql_format, fields, TableDoc, TableDocInfo, TableDocStore, condition, order, limit)
	rows, err = O.Raw(sql).Values(&params)
	return params, rows, err
}

//根据md5判断文档是否存在
//@param                md5str              文档的md5
//@return               Id                  文档存储表的id
func (this *Document) IsExistByMd5(md5str string) (Id int) {
	var ds DocumentStore
	O.QueryTable(TableDocStore).Filter("Md5", md5str).One(&ds)
	return ds.Id
}

//处理已经存在了的文档
func HandleExistDoc(uid int, form FormUpload) error {
	var ds DocumentStore
	err := O.QueryTable(GetTable("document_store")).Filter("Md5", form.Md5).One(&ds)
	if err != nil {
		return err
	}
	doc := Document{
		Title:       form.Title,
		Filename:    form.Filename,
		Keywords:    form.Tags,
		Description: form.Intro,
	}
	docid, _ := O.Insert(&doc)
	docinfo := DocumentInfo{
		Uid:         uid,
		ChanelId:    form.Chanel,
		Cid:         form.Cid,
		Pid:         form.Pid,
		TimeCreate:  int(time.Now().Unix()),
		Dcnt:        0,
		Vcnt:        0,
		Ccnt:        0,
		Score:       30000,
		ScorePeople: 1,
		Status:      1,
	}
	docinfo.Id = int(docid)

	//记录关键字
	go ModelWord.AddWords(form.Tags, docid)

	docinfo.DsId = ds.Id
	i, err := O.Insert(&docinfo)
	if i > 0 && err == nil {
		DocCntInre(form.Chanel, form.Pid, form.Cid, uid)
	}
	return err
}

//处理上传的PDF文档
//@param            uid             上传文档的用户ID
//@param            tmpfile         临时存储的pdf文档
//@param            form            表单
func HandlePdf(uid int, tmpfile string, form FormUpload) error {
	sys, _ := ModelSys.Get()
	//TODO:系统后台可以配置，文档只提供多少页给读者阅读
	if sys.PreviewPage < 10 {
		sys.PreviewPage = 10
	}
	pageNum, err := helper.GetPdfPagesNum(tmpfile) //先用第三方包统计页码，如果不兼容，则在使用自己简单封装的函数获取pdf文档页码。但是好像也有些不兼容
	if err != nil || pageNum == 0 {
		if pageNum, err = helper.CountPdfPages(tmpfile); err != nil {
			helper.Logger.Error(err.Error())
		}
	}
	info, err := os.Stat(tmpfile)
	if err != nil {
		return err
	}

	ds := DocumentStore{
		Md5:         form.Md5,
		Ext:         form.Ext,
		Page:        pageNum,
		Size:        form.Size,
		ModTime:     int(info.ModTime().Unix()),
		PreviewPage: sys.PreviewPage,
		PreviewExt:  "svg",
	}
	ds.ExtCate, ds.ExtNum = helper.GetExtCate(ds.Ext)

	doc := Document{
		Title:       form.Title,
		Filename:    form.Filename,
		Keywords:    form.Tags,
		Description: form.Intro,
	}

	docinfo := DocumentInfo{
		Uid:         uid,
		ChanelId:    form.Chanel,
		Pid:         form.Pid,
		Cid:         form.Cid,
		TimeCreate:  int(time.Now().Unix()),
		Dcnt:        0,
		Vcnt:        0,
		Ccnt:        0,
		Score:       30000,
		ScorePeople: 0,
		Status:      1,
		Price:       form.Price,
	}
	//docinfo.Id,docinfo.DsId
	if _, id, err := O.ReadOrCreate(&ds, "Md5"); err == nil {

		docid, err := O.Insert(&doc)
		docinfo.DsId = int(id)
		if err == nil {
			//记录关键字
			go ModelWord.AddWords(form.Tags, docid)

			docinfo.Id = int(docid)
			_, err = O.Insert(&docinfo)
			if err == nil {
				//处理pdf文档，转成svg图片，再将svg图片压缩，上传到OSS预览库
				go func(tmpfile, md5str string, totalPage int) {
					//把原文档移动到存档库，先暂时不要删除本地的原PDF文件
					ModelOss.MoveToOss(tmpfile, md5str+".pdf", false, false)

					//将pdf文件转成svg
					if err := Pdf2Svg(tmpfile, totalPage, md5str); err != nil {
						helper.Logger.Error(err.Error())
					}

					//删除本地的PDF文件
					if err = os.Remove(tmpfile); err != nil {
						helper.Logger.Error(err.Error())
					}
				}(tmpfile, form.Md5, pageNum)

				//增加文档统计数量
				DocCntInre(form.Chanel, form.Pid, form.Cid, uid)
			}
			return err
		}
		return err
	}
	return err
}

//处理上传的office文档
//@param        uid         用户id
//@param        tmpfile     临时文件
//@param        form        上传表单
func HandleOffice(uid int, tmpfile string, form FormUpload) (err error) {
	//转化成PDF文档，转换成功之后，调用pdf文档处理
	if err := helper.OfficeToPdf(tmpfile); err == nil {
		pdf := strings.TrimSuffix(tmpfile, "."+form.Ext) + ".pdf"
		//处理pdf
		if err = HandlePdf(uid, pdf, form); err == nil {
			//如果转成pdf文档成功，则把原文档移动到OSS存储服务器
			ModelOss.MoveToOss(tmpfile, form.Md5+"."+form.Ext, false, true)
		} else {
			helper.Logger.Error("pdf文档（%v）处理错误：%v", pdf, err.Error())
		}
	} else {
		helper.Logger.Error("office文档（%v）处理失败：%v", tmpfile, err.Error())
	}
	return nil
}

//处理上传的非Office文档和非PDF文档
//@param        uid         用户id
//@param        tmpfile     临时文件
//@param        form        上传表单
func HandleUnOffice(uid int, tmpfile string, form FormUpload) error {
	if form.Ext != ".umd" { //calibre暂时无法转换umd文档
		if pdfFile, err := helper.UnofficeToPdf(tmpfile); err == nil {
			//如果转成pdf文档成功，则把原文档移动到OSS存储服务器
			defer ModelOss.MoveToOss(tmpfile, form.Md5+"."+form.Ext, false, true)
			return HandlePdf(uid, pdfFile, form)
		} else {
			helper.Logger.Error(err.Error())
		}
	}

	info, err := os.Stat(tmpfile)
	if err != nil {
		return err
	}
	ds := DocumentStore{
		Md5:     form.Md5,
		Ext:     form.Ext,
		Page:    0,
		Size:    form.Size,
		ModTime: int(info.ModTime().Unix()),
	}
	ds.ExtCate, ds.ExtNum = helper.GetExtCate(ds.Ext)
	doc := Document{
		Title:       form.Title,
		Filename:    form.Filename,
		Keywords:    form.Tags,
		Description: form.Intro,
	}

	docinfo := DocumentInfo{
		Uid:         uid,
		ChanelId:    form.Chanel,
		Pid:         form.Pid,
		Cid:         form.Cid,
		TimeCreate:  int(time.Now().Unix()),
		Dcnt:        0,
		Vcnt:        0,
		Ccnt:        0,
		Score:       30000,
		ScorePeople: 1,
		Status:      1,
		Price:       form.Price,
	}

	//docinfo.Id,docinfo.DsId
	if _, id, err := O.ReadOrCreate(&ds, "Md5"); err == nil {
		docid, err := O.Insert(&doc)
		docinfo.DsId = int(id)
		if err == nil {
			//记录关键字
			go ModelWord.AddWords(form.Tags, docid)
			docinfo.Id = int(docid)
			_, err = O.Insert(&docinfo)
			if err == nil {
				//把原文档移动到存档库
				go ModelOss.MoveToOss(tmpfile, form.Md5+"."+form.Ext, false, true)
				//增加文档统计数量
				DocCntInre(form.Chanel, form.Pid, form.Cid, uid)
			}
			return err
		}
		return err
	}
	return nil
}

//文档统计增加
//@param            chanel      文库频道
//@param            pid         文档一级分类
//@param            cid         文档二级分类
func DocCntInre(chanel, pid, cid, uid int) {
	//文档数量和分类+1
	Regulate(TableCategory, "Cnt", 1, "`Id` in(?,?,?)", chanel, pid, cid)
	//全站文档统计+1
	Regulate(TableSys, "CntDoc", 1, "Id=1")
	//用户文档数量+1
	Regulate(TableUserInfo, "Document", 1, "Id=?", uid)
}

//获取文档列表，其中status不传时，表示获取全部状态的文档，否则获取指定状态的文档，status:-1已删除，0转码中，1已转码
//排序order全部按倒叙排序，默认是按id倒叙排序，可选值：Id,Dcnt(下载),Vcnt(浏览),Ccnt(收藏)
func DocList(uid, chanelid, pid, cid, p, listRows int, order string, status ...int) ([]orm.Params, int64, error) {
	var (
		data     []orm.Params
		cond     = make(map[string]interface{})
		condQues []string
		args     []interface{}
		condStr  string
	)
	switch strings.ToLower(order) {
	case "dcnt":
		order = "di.Dcnt desc"
	case "vcnt":
		order = "di.Vcnt desc"
	case "ccnt":
		order = "di.Ccnt desc"
	case "score":
		order = "di.Score desc"
	default:
		order = "di.Id desc"
	}

	if uid > 0 {
		cond["di.Uid"] = uid
	}
	if chanelid > 0 {
		cond["di.ChanelId"] = chanelid
	}
	if pid > 0 {
		cond["di.Pid"] = pid
	}
	if cid > 0 {
		cond["di.Cid"] = cid
	}

	for k, v := range cond {
		condQues = append(condQues, fmt.Sprintf("%v=?", k))
		args = append(args, v)
	}
	if len(status) == 1 {
		condQues = append(condQues, fmt.Sprintf("di.Status in(%v)", status[0]))
	}
	if len(status) == 2 {
		condQues = append(condQues, fmt.Sprintf("di.Status in(%v,%v)", status[0], status[1]))
	}
	condStr = "true"
	if len(condQues) > 0 {
		condStr = strings.Join(condQues, " and ")
	}
	fields := "di.Id,di.`Uid`, di.`Cid`, di.`TimeCreate`, di.`Dcnt`, di.`Vcnt`, di.`Ccnt`, di.`Score`, di.`Status`, di.`ChanelId`, di.`Pid`,c.Title Category,u.Username,d.Title,ds.`Md5`, ds.Id DsId,ds.`Ext`, ds.`ExtCate`, ds.`ExtNum`, ds.`Page`, ds.`Size`"
	sql_format := `
		select %v from %v di left join %v u on di.Uid=u.Id
		left join %v d on d.Id=di.Id
		left join %v c on c.Id=di.cid
		left join %v ds on ds.Id=di.DsId
		where %v order by %v limit %v,%v
		`
	sql := fmt.Sprintf(sql_format,
		fields,
		GetTable("document_info"),
		GetTable("user"),
		GetTable("document"),
		GetTable("category"),
		GetTable("document_store"),
		condStr,
		order,
		(p-1)*listRows, listRows,
	)
	rows, err := O.Raw(sql, args...).Values(&data)
	return data, rows, err
}

//文档软删除，即把文档状态标记为-1，操作之后，需要把总文档数量、用户文档数量-1，同时把文档id移入回收站
func (this *Document) SoftDel(uid int, isAdmin bool, ids ...interface{}) (err error) {
	var (
		info []DocumentInfo
	)
	if len(ids) == 0 {
		return errors.New("文档id不能为空")
	}
	if O.QueryTable(TableDocInfo).Filter("Id__in", ids...).All(&info); len(info) > 0 {
		O.QueryTable(TableDocInfo).Filter("Id__in", ids...).Update(orm.Params{"Status": -1})
	}
	return nil
}

//注意：这里只能从回收站进行删除，因为这样就不需要再对统计数据进行增减计算操作了
//删除文档，这里是彻底删除（会先查询md5，然后根据md5查找到所有文档进行删除），被彻底删除的文档，会被标记为非法文档，不再允许用户上传该文档
func (this *Document) DocDeepDel(ids ...interface{}) (errs []string) {
	var (
		DocInfo     []DocumentInfo
		DocStore    []DocumentStore
		DocIllegal  []DocumentIllegal
		DsId, DocId []interface{}
	)

	//查询现有的文档
	if _, err := O.QueryTable(TableDocInfo).Filter("Id__in", ids...).All(&DocInfo); err != nil {
		helper.Logger.Error(err.Error())
		errs = append(errs, err.Error())
	}
	//获取DsId
	for _, info := range DocInfo {
		DsId = append(DsId, info.DsId)
	}
	//根据DsId查询所有需要删除的文档记录
	DocInfo, _, _ = this.GetDocInfoByDsId(DsId...)

	if len(DsId) > 0 {

		//根据DsId查询所有DocumentStore表中的数据
		if _, err := O.QueryTable(TableDocStore).Filter("Id__in", DsId...).All(&DocStore); err != nil {
			helper.Logger.Error(err.Error())
			errs = append(errs, err.Error())
		}

		//清空document_store表中的数据
		if _, err := O.QueryTable(TableDocStore).Filter("Id__in", DsId...).Delete(); err != nil {
			helper.Logger.Error(err.Error())
			errs = append(errs, err.Error())
		}
	}

	//记录非法的md5
	for _, store := range DocStore {
		var docIllegal = DocumentIllegal{Id: 0, Md5: store.Md5}
		O.Read(&docIllegal, "Md5")
		if docIllegal.Id == 0 {
			DocIllegal = append(DocIllegal, docIllegal)
		}
		go func() {
			//删除预览文档
			if err := ModelOss.DelFromOss(true, store.Md5+".pdf"); err != nil {
				helper.Logger.Error(err.Error())
			}
			//删除封面图片
			if err := ModelOss.DelFromOss(true, store.Md5+".jpg"); err != nil {
				helper.Logger.Error(err.Error())
			}
			//删除原文档
			if err := ModelOss.DelFromOss(false, store.Md5+"."+store.Ext); err != nil {
				helper.Logger.Error(err.Error())
			}
			if err := ModelOss.DelFromOss(false, store.Md5+".pdf"); err != nil {
				helper.Logger.Error(err.Error())
			}
		}()
	}
	//录入非法文档表
	if l := len(DocIllegal); l > 0 {
		if _, err := O.InsertMulti(l, &DocIllegal); err != nil {
			helper.Logger.Error(err.Error())
			errs = append(errs, err.Error())
		}
	}

	if len(DocId) > 0 {
		//清空回收站中的这些非法文档
		if _, err := O.QueryTable(TableDocRecycle).Filter("Id__in", DocId...).Delete(); err != nil {
			helper.Logger.Error(err.Error())
			errs = append(errs, err.Error())
		}
		//清空document表中的数据
		if _, err := O.QueryTable(TableDoc).Filter("Id__in", DocId...).Delete(); err != nil {
			helper.Logger.Error(err.Error())
			errs = append(errs, err.Error())
		}
		//清空document_info表中的数据
		if _, err := O.QueryTable(TableDocInfo).Filter("Id__in", DocId...).Delete(); err != nil {
			helper.Logger.Error(err.Error())
			errs = append(errs, err.Error())
		}

		//删除这些文档的所有评论记录
		if _, err := O.QueryTable(TableDocComment).Filter("Did__in", DocId...).Delete(); err != nil {
			helper.Logger.Error(err.Error())
			errs = append(errs, err.Error())
		}
		//处理收藏
		go ModelCollect.DelByDocId(DocId...)
	}

	return errs
}

//根据document_store表中的id查询document_info表中的数据
func (this *Document) GetDocInfoByDsId(DsId ...interface{}) (info []DocumentInfo, rows int64, err error) {
	if l := len(DsId); l > 0 {
		rows, err = O.QueryTable(GetTable("document_info")).Limit(l).Filter("DsId__in", DsId...).All(&info)
	}
	return
}

//根据document_store表中的id查询document_info表中的数据
func (this *Document) GetDocStoreByDsId(DsId ...interface{}) (store []DocumentStore, rows int64, err error) {
	if l := len(DsId); l > 0 {
		rows, err = O.QueryTable(TableDocStore).Limit(l).Filter("Id__in", DsId...).All(&store)
	}
	return
}

//根据document_store表中的id查询document_info表中的数据
func (this *Document) GetDocInfoById(Ids ...interface{}) (info []DocumentInfo, rows int64, err error) {
	if len(Ids) > 0 {
		rows, err = O.QueryTable(TableDocInfo).Limit(len(Ids)).Filter("Id__in", Ids...).All(&info)
	}
	return
}

//把文档标记为非法文档
//@param                ids             文档id
//@return               err             错误，nil表示成功
func (this *Document) SetIllegal(ids ...interface{}) (err error) {
	if l := int64(len(ids)); l > 0 {
		var (
			docinfo []DocumentInfo
			dsid    []interface{} //document_store表中的id
			did     []interface{} //文档id
			stores  []DocumentStore
		)
		l, err = O.QueryTable(TableDocInfo).Filter("Id__in", ids...).Limit(l).All(&docinfo)
		if l > 0 {
			for _, v := range docinfo {
				dsid = append(dsid, v.DsId)
			}
			if docinfo, l, err = this.GetDocInfoByDsId(dsid...); l > 0 {
				for _, v := range docinfo {
					did = append(did, v.Id)
				}
				//将文档移入回收站
				if errs := ModelDocRecycle.RemoveToRecycle(0, false, did...); len(errs) > 0 {
					helper.Logger.Error(strings.Join(errs, "; "))
				}
				//根据dsid查询文档md5，并把md5录入到非法文档表
				if stores, l, err = this.GetDocStoreByDsId(dsid...); err == nil && l > 0 {
					for _, store := range stores {
						var ilg = DocumentIllegal{
							Md5: store.Md5,
						}
						O.Insert(&ilg)
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
	tables := []string{TableDocInfo + " i", TableDoc + " d", TableDocStore + " ds"}
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
		O.Raw(sql).Values(&values)
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
