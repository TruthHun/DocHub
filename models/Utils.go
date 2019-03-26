package models

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego/orm"
)

// 文档处理
func DocumentProcess(uid int, form FormUpload) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if err != nil {
			o.Rollback()
		} else {
			o.Commit()
		}
	}()

	now := int(time.Now().Unix())
	isExist := false

	var doc = &Document{
		Title:       form.Title,
		Filename:    form.Filename,
		Keywords:    form.Tags,
		Description: form.Intro,
	}

	var info = &DocumentInfo{
		Uid:         uid,
		ChanelId:    form.Chanel,
		Cid:         form.Cid,
		Pid:         form.Pid,
		TimeCreate:  now,
		TimeUpdate:  now,
		Dcnt:        0,
		Vcnt:        0,
		Ccnt:        0,
		Score:       30000,
		ScorePeople: 1,
		Status:      DocStatusConverting,
		Price:       form.Price,
	}

	var store = &DocumentStore{}

	if form.Exist == 1 && len(form.Md5) == 32 {
		o.QueryTable(GetTableDocumentStore()).Filter("Md5", form.Md5).One(store)
		if store.Id > 0 {
			isExist = true
		}
	}

	if store.Id == 0 { // 计算文档md5，并入库

	}

	info.DsId = store.Id

	var (
		word = &Word{}
	)

	return
}

// 文档转换
func DocumentConvert(tmpFile string, fileMD5 string) {

}

//处理已经存在了的文档
func HandleExistDoc(uid int, form FormUpload) error {
	var ds DocumentStore
	err := orm.NewOrm().QueryTable(GetTableDocumentStore()).Filter("Md5", form.Md5).One(&ds)
	if err != nil {
		return err
	}
	doc := Document{
		Title:       form.Title,
		Filename:    form.Filename,
		Keywords:    form.Tags,
		Description: form.Intro,
	}
	docid, _ := orm.NewOrm().Insert(&doc)
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
	go NewWord().AddWords(form.Tags, docid)

	docinfo.DsId = ds.Id
	i, err := orm.NewOrm().Insert(&docinfo)
	if i > 0 && err == nil {
		SetDocCntIncre(form.Chanel, form.Pid, form.Cid, uid)
	}
	return err
}

//处理上传的PDF文档
//@param            uid             上传文档的用户ID
//@param            tmpfile         临时存储的pdf文档
//@param            form            表单
func HandlePdf(uid int, tmpfile string, form FormUpload) (err error) {
	var (
		pageNum     int
		previewPage = NewSys().GetByField("PreviewPage").PreviewPage
		fileInfo    os.FileInfo
		docId       int64
		docStoreId  int64
		o           = orm.NewOrm()
	)

	//启用事务
	o.Begin()
	defer func() {
		if err == nil {
			o.Commit()
			//生成索引
			go NewElasticSearchClient().BuildIndexById(int(docId))
		} else {
			o.Rollback()
		}
	}()

	//先用第三方包统计页码，如果不兼容，则在使用自己简单封装的函数获取pdf文档页码。但是好像也有些不兼容
	if pageNum, err = helper.GetPdfPagesNum(tmpfile); err != nil || pageNum == 0 {
		if pageNum, err = helper.CountPdfPages(tmpfile); err != nil {
			helper.Logger.Error(err.Error())
		}
	}

	if fileInfo, err = os.Stat(tmpfile); err != nil {
		helper.Logger.Error(err.Error())
		return err
	}

	//文档存档信息
	docStore := DocumentStore{
		Md5:         form.Md5,
		Ext:         form.Ext,
		Page:        pageNum,
		Size:        form.Size,
		ModTime:     int(fileInfo.ModTime().Unix()),
		PreviewPage: previewPage,
		PreviewExt:  "svg",
	}
	docStore.ExtCate, docStore.ExtNum = helper.GetExtCate(docStore.Ext)

	//文档标题等信息
	doc := Document{
		Title:       form.Title,
		Filename:    form.Filename,
		Keywords:    form.Tags,
		Description: form.Intro,
	}

	//文档基本信息
	now := int(time.Now().Unix())
	docInfo := DocumentInfo{
		Uid:         uid,
		ChanelId:    form.Chanel,
		Pid:         form.Pid,
		Cid:         form.Cid,
		TimeCreate:  now,
		TimeUpdate:  now,
		Dcnt:        0,
		Vcnt:        0,
		Ccnt:        0,
		Score:       30000,
		ScorePeople: 0,
		Status:      1,
		Price:       form.Price,
	}

	//创建存档信息
	if _, docStoreId, err = o.ReadOrCreate(&docStore, "Md5"); err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	//记录关键字
	NewWord().AddWords(form.Tags, docId)

	docId, err = o.Insert(&doc)
	docInfo.DsId = int(docStoreId)

	docInfo.Id = int(docId)
	if _, err = o.Insert(&docInfo); err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	//增加文档统计数量
	SetDocCntIncre(form.Chanel, form.Pid, form.Cid, uid)

	//处理pdf文档，转成svg图片，再将svg图片压缩，上传到OSS预览库
	go func(tmpfile, md5str string, totalPage int) {

		//把原文档移动到存档库(即私有库)，先暂时不要删除本地的原PDF文件
		NewOss().MoveToOss(tmpfile, md5str+".pdf", false, false)

		//转化供预览的总页数
		if previewPage > 0 {
			totalPage = previewPage
		}

		//将pdf文件转成svg
		if err := Pdf2Svg(tmpfile, totalPage, md5str); err != nil {
			helper.Logger.Error(err.Error())
		}

		//最后删除本地的PDF文件
		if err = os.Remove(tmpfile); err != nil {
			helper.Logger.Error(err.Error())
		}
	}(tmpfile, form.Md5, pageNum)

	return err
}

//处理上传的office文档
//@param        uid         用户id
//@param        tmpfile     临时文件
//@param        form        上传表单
func HandleOffice(uid int, tmpfile string, form FormUpload) (err error) {
	//转化成PDF文档，转换成功之后，调用pdf文档处理
	if err = helper.OfficeToPDF(tmpfile); err != nil {
		helper.Logger.Error("office文档（%v）处理失败：%v", tmpfile, err.Error())
		return err
	}

	pdf := strings.TrimSuffix(tmpfile, "."+form.Ext) + ".pdf"

	//处理pdf
	if err = HandlePdf(uid, pdf, form); err == nil {
		//如果转成pdf文档成功，则把原文档移动到OSS存储服务器
		NewOss().MoveToOss(tmpfile, form.Md5+"."+form.Ext, false, true)
	} else {
		helper.Logger.Error("pdf文档（%v）处理错误：%v", pdf, err.Error())
	}

	return
}

//处理上传的非Office文档和非PDF文档
//@param        uid         用户id
//@param        tmpfile     临时文件
//@param        form        上传表单
func HandleUnOffice(uid int, tmpfile string, form FormUpload) (err error) {
	var (
		fileInfo          os.FileInfo
		pdfFile           string
		o                 = orm.NewOrm()
		docId, docStoreId int64
	)
	if form.Ext != ".umd" { //calibre暂时无法转换umd文档
		//非umd文档，转PDF
		if pdfFile, err = helper.UnOfficeToPDF(tmpfile); err == nil {
			//如果转成pdf文档成功，则把原文档移动到OSS存储服务器
			defer NewOss().MoveToOss(tmpfile, form.Md5+"."+form.Ext, false, true)
			return HandlePdf(uid, pdfFile, form)
		} else {
			//转PDF失败，则作为无法预览的文档处理
			helper.Logger.Error(err.Error())
		}
	}

	//启动事务
	o.Begin()
	defer func() {
		if err == nil {
			o.Commit()
			//索引
			NewElasticSearchClient().BuildIndexById(int(docId))
		} else {
			o.Rollback()
		}
	}()

	//处理umd文档

	if fileInfo, err = os.Stat(tmpfile); err != nil {
		return err
	}

	docStore := DocumentStore{
		Md5:     form.Md5,
		Ext:     form.Ext,
		Page:    0,
		Size:    form.Size,
		ModTime: int(fileInfo.ModTime().Unix()),
	}
	docStore.ExtCate, docStore.ExtNum = helper.GetExtCate(docStore.Ext)

	doc := Document{
		Title:       form.Title,
		Filename:    form.Filename,
		Keywords:    form.Tags,
		Description: form.Intro,
	}

	now := int(time.Now().Unix())
	docInfo := DocumentInfo{
		Uid:         uid,
		ChanelId:    form.Chanel,
		Pid:         form.Pid,
		Cid:         form.Cid,
		TimeCreate:  now,
		TimeUpdate:  now,
		Dcnt:        0,
		Vcnt:        0,
		Ccnt:        0,
		Score:       30000,
		ScorePeople: 1,
		Status:      1,
		Price:       form.Price,
	}

	if _, docStoreId, err = o.ReadOrCreate(&docStore, "Md5"); err != nil {
		helper.Logger.Error(err.Error())
		return
	}
	if docId, err = o.Insert(&doc); err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	//记录关键字
	NewWord().AddWords(form.Tags, docId)

	docInfo.Id = int(docId)
	docInfo.DsId = int(docStoreId)
	if _, err = o.Insert(&docInfo); err == nil {
		//把原文档移动到存档库
		go NewOss().MoveToOss(tmpfile, form.Md5+"."+form.Ext, false, true)

		//增加文档统计数量
		SetDocCntIncre(form.Chanel, form.Pid, form.Cid, uid)
	}
	return
}

//文档统计增加
//@param            chanel      文库频道
//@param            pid         文档一级分类
//@param            cid         文档二级分类
func SetDocCntIncre(chanel, pid, cid, uid int) {
	//文档数量和分类+1
	Regulate(GetTableCategory(), "Cnt", 1, "`Id` in(?,?,?)", chanel, pid, cid)
	//全站文档统计+1
	Regulate(GetTableSys(), "CntDoc", 1, "Id=1")
	//用户文档数量+1
	Regulate(GetTableUserInfo(), "Document", 1, "Id=?", uid)
}

//获取文档列表，其中status不传时，表示获取全部状态的文档，否则获取指定状态的文档，status:-1已删除，0转码中，1已转码
//排序order全部按倒叙排序，默认是按id倒叙排序，可选值：Id,Dcnt(下载),Vcnt(浏览),Ccnt(收藏)
func GetDocList(uid, chanelid, pid, cid, p, listRows int, order string, status ...int) (data []orm.Params, rows int64, err error) {
	var (
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

	sqlFormat := `
		select %v from %v di left join %v u on di.Uid=u.Id
		left join %v d on d.Id=di.Id
		left join %v c on c.Id=di.cid
		left join %v ds on ds.Id=di.DsId
		where %v order by %v limit %v,%v
		`

	sql := fmt.Sprintf(sqlFormat,
		fields,
		GetTableDocumentInfo(),
		GetTableUser(),
		GetTableDocument(),
		GetTableCategory(),
		GetTableDocumentStore(),
		condStr,
		order,
		(p-1)*listRows, listRows,
	)

	rows, err = orm.NewOrm().Raw(sql, args...).Values(&data)

	return
}
