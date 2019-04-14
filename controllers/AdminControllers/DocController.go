package AdminControllers

import (
	"fmt"

	"sort"

	"strings"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/helper/conv"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego/orm"
)

type DocController struct {
	BaseController
}

func (this *DocController) Prepare() {
	this.BaseController.Prepare()
	this.Data["IsDoc"] = true
	cond := orm.NewCondition().And("pid", 0)
	if data, _, err := models.GetList(models.GetTableCategory(), 1, 20, cond, "sort", "-id"); err != nil {
		helper.Logger.Error(err.Error())
	} else {
		this.Data["Chanels"] = data
	}
}

//文库频道管理
func (this *DocController) Get() {
	this.Data["Tab"] = "chanel"
	this.TplName = "chanel.html"
}

//文档分类管理
func (this *DocController) Category() {
	cond := orm.NewCondition().And("id__gt", 0)
	data, _, _ := models.GetList(models.GetTableCategory(), 1, 2000, cond, "pid", "sort", "-id")
	cates := models.ToTree(data, "Pid", 0)
	this.Data["Cates"] = cates
	this.Data["Tab"] = "cate"
	this.TplName = "cates.html"

}

//文档列表管理
func (this *DocController) List() {
	var (
		p, listRows             = 1, this.Sys.ListRows
		totalRows               = this.Sys.CntDoc
		pid, cid, chanelid, uid int
		slice                   []int
		level                   int //当前分类的level，频道级别，level为0，父类级别，level为1，children级别，level为2
		CurId                   int //当前分类的id
	)
	params := conv.Path2Map(this.GetString(":splat"))
	if v, ok := params["p"]; ok {
		//页码处理
		p = helper.NumberRange(helper.Interface2Int(v), 1, 100)
	}
	if v, ok := params["uid"]; ok {
		//页码处理
		uid = helper.Interface2Int(v)
	}
	if v, ok := params["chanelid"]; ok {
		chanelid = helper.Interface2Int(v)
		this.Data["ChanelId"] = chanelid
		slice = append(slice, chanelid)
	}
	if v, ok := params["pid"]; ok {
		pid = helper.Interface2Int(v)
		this.Data["Pid"] = pid
		slice = append(slice, pid)
		level = 1
	}
	if v, ok := params["cid"]; ok {
		cid = helper.Interface2Int(v)
		this.Data["Cid"] = cid
		slice = append(slice, cid)
		level = 2
	}
	if len(slice) > 0 {
		sort.Ints(slice)
		CurId = slice[len(slice)-1]
	}
	cates := models.NewCategory().GetSameLevelCategoryById(CurId)
	for _, cate := range cates {
		if cate.Id == CurId {
			totalRows = cate.Cnt
		}
	}
	this.Data["Cates"] = cates
	this.Data["CurId"] = CurId
	this.Data["Level"] = level
	lists, _, err := models.GetDocList(uid, chanelid, pid, cid, p, listRows, "Id", 0, 1)
	if err != nil {
		helper.Logger.Error("SQL语句执行错误：%v", err.Error())
	}
	this.Data["Lists"] = lists
	this.Data["Page"] = helper.Paginations(6, totalRows, listRows, p, "/admin/doc/list/", "uid", uid, "chanelid", chanelid, "pid", pid, "cid", cid)
	this.Data["Tab"] = "list"
	this.TplName = "list.html"
}

//文档回收站
func (this *DocController) Recycle() {
	p, _ := this.GetInt("p", 1)
	//页码处理
	p = helper.NumberRange(p, 1, 10000)
	listRows := this.Sys.ListRows
	this.Data["Lists"], _, _ = models.NewDocumentRecycle().RecycleList(p, listRows)
	this.Data["Tab"] = "recycle"
	this.TplName = "recycle.html"
}

//新增文库频道
func (this *DocController) AddChanel() {
	var cate models.Category
	this.ParseForm(&cate)
	if len(cate.Title) > 0 && len(cate.Alias) > 0 {
		cate.Status = true
		orm.NewOrm().Insert(&cate)
		this.ResponseJson(true, "频道新增成功")
	} else {
		this.ResponseJson(false, "名称和别名均不能为空")
	}
}

//根据频道获取下一级分类
func (this *DocController) GetCateByCid() {
	cid, _ := this.GetInt("Cid")
	if cid > 0 {
		if data, _, err := models.GetList(models.GetTableCategory(), 1, 100, orm.NewCondition().And("Pid", cid).And("Status", 1), "sort", "-id"); err != nil {
			this.ResponseJson(false, err.Error())
		} else {
			this.ResponseJson(true, "数据获取成功", data)
		}
	} else {
		this.ResponseJson(false, "频道ID参数不正确，必须大于0")
	}

}

//新增文档分类
func (this *DocController) AddCate() {
	var (
		cates   []models.Category
		cate    models.Category
		cid, _  = this.GetInt("Cid")        //频道id
		pid, _  = this.GetInt("Pid")        //父类id
		content = this.GetString("Content") //内容
	)
	slice := strings.Split(content, "\n")
	for _, v := range slice {
		if pid > cid {
			cate.Pid = pid
		} else {
			cate.Pid = cid
		}
		if v = strings.TrimSpace(v); len(v) > 0 {
			cate.Title = v
			cate.Status = true
			cates = append(cates, cate)
		}
	}
	if l := len(cates); l > 0 {
		if _, err := orm.NewOrm().InsertMulti(l, &cates); err != nil {
			this.ResponseJson(false, err.Error())
		} else {
			this.ResponseJson(true, "分类添加成功")
		}
	}
	this.ResponseJson(false, "添加失败，缺少分类")
}

//删除分类
func (this *DocController) DelCate() {
	id, _ := this.GetInt("id")
	if err := models.NewCategory().Del(id); err != nil {
		this.ResponseJson(false, err.Error())
	}
	this.ResponseJson(true, "删除成功")
}

//对文档进行操作，type类型的值包括remove（移入回收站），del(删除文档记录)，clear（清空通用户的内容)，deepdel（深度删除，在删除文档记录的同时删除文档文件），forbidden(禁止文档，把文档md5标记为禁止上传，只要文档的md5是这个，则该文档禁止被上传)
func (this *DocController) Action() {
	var errs []string
	ActionType := strings.ToLower(this.GetString("type"))
	ids := helper.StringSliceToInterfaceSlice(strings.Split(this.GetString("id"), ","))
	recycle := models.NewDocumentRecycle()
	switch ActionType {
	case "deepdel": //彻底删除文档：删除文档记录的同时也删除文档
		if err := recycle.DeepDel(ids...); err != nil {
			errs = append(errs, err.Error())
		}
	case "del-row": //只是删除该文档的文档记录
		if err := recycle.DelRows(ids...); err != nil {
			errs = append(errs, err.Error())
		}
	case "recover": //恢复文档，只有文档状态是-1时，才可以进行恢复【OK】
		if err := recycle.RecoverFromRecycle(ids...); err != nil {
			errs = append(errs, err.Error())
		}
	case "illegal": //将文档标记为非法文档【OK】
		if err := models.NewDocument().SetIllegal(ids...); err != nil {
			errs = append(errs, err.Error())
		}
	case "remove": //将文档移入回收站【OK】
		if err := recycle.RemoveToRecycle(this.AdminId, false, ids...); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		this.ResponseJson(false, fmt.Sprintf("操作失败：%v", strings.Join(errs, "; ")))
	}
	this.ResponseJson(true, "操作成功")
}

//获取文档备注模板
func (this *DocController) RemarkTpl() {
	if this.Ctx.Request.Method == "GET" {
		DsId, _ := this.GetInt("dsid")
		if DsId > 0 {
			remark := models.NewDocumentRemark().GetContentTplByDsId(DsId)
			this.ResponseJson(true, "获取成功", remark)
		} else {
			this.ResponseJson(false, "DsId不能为空")
		}
	} else {
		var rm models.DocumentRemark
		this.ParseForm(&rm)
		if err := models.NewDocumentRemark().Insert(rm); err != nil {
			this.ResponseJson(false, fmt.Sprintf("操作失败：%v", err.Error()))
		} else {
			this.ResponseJson(true, "操作成功")
		}
	}
}
