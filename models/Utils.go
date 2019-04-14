package models

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego/orm"
)

// 文档处理
func DocumentProcess(uid int, form FormUpload) (err error) {
	// 1. 计算文档 md5
	// 2. 判断文档是否合法
	// 3. 存入 document_store
	// 4. 存入 document
	// 5. 存入 document_info
	// 6. 文档分类数量增加 category 表
	// 7. 总文档数增加
	// 8. 用户积分和文档数量增加
	// 9. 积分记录

	sys, _ := NewSys().Get()
	score := 0 // 文档已被上传的话，获得的积分为0

	var doc = &Document{
		Title:       form.Title,
		Filename:    form.Filename,
		Keywords:    form.Tags,
		Description: form.Intro,
	}

	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if err != nil {
			o.Rollback()
			os.Remove(form.TmpFile)
		} else {
			o.Commit()
			if helper.Debug {
				beego.Debug("============")
				beego.Debug("form: %+v", form)
				beego.Debug("doc: %+v", doc)
				beego.Debug("============")
			}
			if score > 0 {
				go DocumentConvert(form.TmpFile, form.Md5, sys.PreviewPage)
			}
			go func() {
				if errIndex := NewElasticSearchClient().BuildIndexById(doc.Id); err != nil {
					helper.Logger.Error("重建索引失败：%v", errIndex.Error())
				}
			}()
		}
	}()

	var (
		errForbidden = errors.New("您上传的文档已被管理员禁止上传")
		errRetry     = errors.New("文档上传失败，请重新上传")
	)

	var file *os.File
	file, err = os.Open(form.TmpFile)
	if err == nil {
		defer file.Close()
		form.Md5 = helper.ComputeFileMD5(file)
	}

	if len(form.Md5) != 32 {
		return errRetry
	}

	if illegal := doc.IsIllegal(form.Md5); illegal {
		return errForbidden
	}

	now := int(time.Now().Unix())

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
		ScorePeople: 0,
		Status:      DocStatusConverting,
		Price:       form.Price,
	}

	var store = &DocumentStore{}

	o.QueryTable(GetTableDocumentStore()).Filter("Md5", form.Md5).One(store)
	if store.Id == 0 {
		// 文档未被上传过，设置用户获得的积分
		score = sys.Reward

		var fileInfo os.FileInfo
		fileInfo, err = os.Stat(form.TmpFile)
		if err != nil {
			return errRetry
		}
		store = &DocumentStore{
			Md5:         form.Md5,
			Ext:         strings.TrimLeft(form.Ext, "."),
			Size:        int(fileInfo.Size()),
			ModTime:     int(fileInfo.ModTime().Unix()),
			PreviewPage: sys.PreviewPage,
			PreviewExt:  "svg",
			Page:        0, //文档正在转换中，页码数设置为 0
		}
		if form.Ext == "" {
			form.Ext = filepath.Ext(form.TmpFile)
		}
		store.ExtCate, store.ExtNum = helper.GetExtCate(form.Ext)
		_, err = o.Insert(store)
		if err != nil {
			helper.Logger.Error(err.Error())
			return errRetry
		}
	} else {
		info.Status = DocStatusNormal
	}

	if _, err = o.Insert(doc); err != nil {
		helper.Logger.Error(err.Error())
		return errRetry
	}

	info.DsId = store.Id
	info.Id = doc.Id

	_, err = o.Insert(info)
	if err != nil {
		helper.Logger.Error(err.Error())
		return errRetry
	}

	// 分类统计数增加
	sqlCate := fmt.Sprintf("update `%v` set `Cnt`=`Cnt`+1 where `Id` in(?,?,?) limit 3", GetTableCategory())
	if _, err = o.Raw(sqlCate, form.Cid, form.Chanel, form.Pid).Exec(); err != nil {
		helper.Logger.Error(err.Error())
		return errRetry
	}
	// 总文档数增加
	sqlSys := fmt.Sprintf("update `%v` set `CntDoc`=`CntDoc`+1 where `Id`=1", GetTableSys())
	if _, err = o.Raw(sqlSys).Exec(); err != nil {
		helper.Logger.Error(err.Error())
		return errRetry
	}

	// 用户文档数量和积分数量增加
	sqlUser := fmt.Sprintf("update `%v` set `Document`=`Document`+1,`Coin`=`Coin`+? where `Id`=?", GetTableUserInfo())
	if _, err = o.Raw(sqlUser, score, uid).Exec(); err != nil {
		helper.Logger.Error(err.Error())
		return errRetry
	}

	coinLog := &CoinLog{
		Uid:        uid,
		Coin:       score,
		TimeCreate: now,
	}

	// 增加积分变更记录
	coinLog.Log = "分享了一篇已被分享过的文档《%v》，获得 %v 个积分"
	if score > 0 {
		coinLog.Log = "分享了一篇未被分享过的文档《%v》，获得 %v 个积分"
	}
	coinLog.Log = fmt.Sprintf(coinLog.Log, doc.Title, score)
	if _, err = o.Insert(coinLog); err != nil {
		helper.Logger.Error(err.Error())
		return errRetry
	}
	return
}

// 文档转换
// @param           tmpFile         临时存储的文件
// @param           fileMD5         文件md5
// @param           page            需要转换的页数，0表示全部转换
func DocumentConvert(tmpFile string, fileMD5 string, page ...int) (err error) {
	// 1. 先把原文档上传到云存储
	// 2. 把文档转成 PDF（如果原文档不是PDF的话）
	// 3. 提取 PDF 文档中的部分文本内容
	// 4. 把 PDF 转 svg
	// 5. 把 SVG 转 JPG 封面（封面需要裁剪）
	// 6. 判断云存储类型，以确定是否需要压缩 svg 成gzip（先加水印再压缩）
	// 7. 上传 svg、jpeg 等到云存储
	// 8. 更新 document_store 表的信息，如页数、SVG宽高
	// 9. 更新 document_info 中的文档转换状态为正常状态
	// 10. 更新 document_text 中的文档内容信息

	if _, err = os.Stat(tmpFile); err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	if helper.Debug {
		helper.Logger.Debug("文档转换中: %v ==> %v", tmpFile, fileMD5)
	}

	ext := filepath.Ext(tmpFile)
	extLower := strings.ToLower(ext)
	tmpDir := strings.TrimSuffix(tmpFile, ext)
	os.MkdirAll(tmpDir, os.ModePerm)
	pdfFile := tmpDir + helper.ExtPDF
	coverJPG := tmpDir + ".jpg" //default

	defer func() {
		os.Remove(tmpFile)
		os.Remove(pdfFile)
		os.Remove(coverJPG)
		os.RemoveAll(tmpDir)
	}()

	var (
		clientPrivate *CloudStore
		clientPublic  *CloudStore
	)

	if clientPrivate, err = NewCloudStore(true); err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	if clientPublic, err = NewCloudStore(false); err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	if err = clientPrivate.Upload(tmpFile, fileMD5+ext); err != nil {
		helper.Logger.Error(err.Error())
		return
	}

	if _, ok := helper.AllowedUploadDocsExt[extLower]; !ok {
		return errors.New("不允许上传的文档类型")
	}

	store := &DocumentStore{}
	text := &DocText{}

	o := orm.NewOrm()
	o.QueryTable(GetTableDocumentStore()).Filter("Md5", fileMD5).One(store)
	o.QueryTable(GetTableDocText()).Filter("Md5", fileMD5).One(text)
	text.Md5 = fileMD5

	// 只需要更新下转换状态
	if extLower == helper.ExtUMD || extLower == helper.ExtCHM {
		o.QueryTable(GetTableDocumentInfo()).Filter("DsId", store.Id).Update(orm.Params{
			"Status": DocStatusNormal,
		})
		return
	}

	if extLower != helper.ExtPDF {
		switch extLower {
		case helper.ExtMOBI, helper.ExtEPUB, helper.ExtTXT:
			if pdfFile, err = helper.ConvertByCalibre(tmpFile, helper.ExtPDF); err != nil {
				helper.Logger.Error("文件(%v)转PDF失败：%v", tmpFile, err.Error())
				return
			}
		default: // office
			if err = helper.OfficeToPDF(tmpFile); err != nil {
				helper.Logger.Error("文件(%v)转PDF失败：%v", tmpFile, err.Error())
				return
			}
		}
	} else {
		pdfFile = tmpFile
	}

	store.Page, err = helper.CountPDFPages(pdfFile)
	if err != nil {
		return
	}

	maxPreview := store.Page

	helper.Logger.Debug("store信息：", fmt.Sprintf("%+v", store))

	if store.PreviewPage > 0 && store.PreviewPage < store.Page {
		maxPreview = store.PreviewPage
	}

	if extLower == helper.ExtTXT { // 直接从txt文本中获取
		b, _ := ioutil.ReadFile(tmpFile)
		text.Content = string(b)
	} else {
		text.Content = helper.ExtractTextFromPDF(pdfFile, 1, 20)
	}

	text.Content = beego.Substr(text.Content, 0, 4500)

	ch := make(chan bool, 1)
	go func() {
		select {
		case <-ch:
			o.Begin()
			_, errOrm := o.QueryTable(GetTableDocumentInfo()).Filter("DsId", store.Id).Update(orm.Params{
				"Status": DocStatusNormal,
			})

			if errOrm == nil {
				_, errOrm = o.Update(store)
			}

			if text.Id > 0 {
				_, errOrm = o.Update(text)
			} else {
				_, errOrm = o.Insert(text)
			}

			if errOrm == nil {
				o.Commit()
			} else {
				helper.Logger.Error(errOrm.Error())
				o.Rollback()
			}
			close(ch)
		}

	}()

	// PDF 转 SVG
	for i := 0; i < maxPreview; i++ {
		pageNO := i + 1
		svg := filepath.Join(tmpDir, strconv.Itoa(pageNO)+".svg")
		if err = helper.ConvertPDF2SVG(pdfFile, svg, pageNO); err != nil {
			helper.Logger.Error(err.Error())
			continue
		}

		if i == 0 {
			store.Width, store.Height = helper.ParseSvgWidthAndHeight(svg)
			ch <- true

			if coverJPG, err = helper.ConvertToJPEG(svg); err == nil {
				err = helper.CropImage(coverJPG, helper.CoverWidth, helper.CoverHeight)
			}
			if err != nil {
				helper.Logger.Error(err.Error())
			}

			// 上传封面到云存储
			err = clientPublic.Upload(coverJPG, fileMD5+".jpg", helper.HeaderJPEG)
			if err != nil {
				helper.Logger.Error(err.Error())
			}
		}

		// 上传 svg 预览页到云存储
		var headers []map[string]string
		headers = append(headers, helper.HeaderSVG)
		save := fmt.Sprintf("%v/%v.svg", fileMD5, pageNO)
		if helper.Debug {
			beego.Debug("存储svg文件", svg, "==>", save)
		}

		err = helper.CompressBySVGO(svg, svg)
		if err != nil {
			helper.Logger.Error("SVGO压缩SVG失败：%v", err.Error())
		}
		if clientPublic.CanGZIP {
			if err = helper.CompressByGzip(svg); err != nil {
				helper.Logger.Error("GZIP压缩SVG失败：%v", err.Error())
			} else {
				headers = append(headers, helper.HeaderGzip)
			}
		}
		if err = clientPublic.Upload(svg, save, headers...); err != nil {
			helper.Logger.Error(err.Error())
		}

	}

	// 重置 err 为 nil
	if err != nil {
		err = nil
	}

	return
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

	if len(status) > 0 {
		var statusArr []string
		for _, item := range status {
			statusArr = append(statusArr, fmt.Sprint(item))
		}
		condQues = append(condQues, fmt.Sprintf("di.Status in(%v)", strings.Join(statusArr, ",")))
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
