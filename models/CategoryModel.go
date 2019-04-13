package models

import (
	"errors"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego/orm"
)

//文档分类
type Category struct {
	Id     int    `orm:"column(Id)"`
	Pid    int    `orm:"default(0);column(Pid)"`           //父类ID【Pid为0时的id为频道Id(chanel_id)】
	Title  string `orm:"size(20);column(Title);default()"` //分类名称
	Cover  string `orm:"column(Cover);default()"`          //封面
	Cnt    int    `orm:"default(0);column(Cnt)"`           //当前分类下的文档数量统计
	Sort   int    `orm:"default(0);column(Sort)"`          //分类排序，值越小越靠前
	Alias  string `orm:"size(30);default();column(Alias)"` //英文别名
	Status bool   `orm:"default(true);column(Status)"`     //分类或频道状态，0表示关闭，1表示启用
}

func NewCategory() *Category {
	return &Category{}
}

func GetTableCategory() string {
	return getTable("category")
}

// 多字段唯一索引
func (this *Category) TableUnique() [][]string {
	return [][]string{
		[]string{"Pid", "Title"},
	}
}

//根据传递过来的id从分类表中查询标题
//@param                id              主键id
//@return               title           返回查询的标题名称
func (this *Category) GetTitleById(id interface{}) (title string) {
	orm.NewOrm().QueryTable(GetTableCategory()).Filter("Id", id).One(this, "Title")
	return this.Title
}

//根据id删除分类
//@param                id              需要删除的分类id
//@return               err             错误，nil表示删除成功
func (this *Category) Del(id ...interface{}) (err error) {
	var (
		cate     Category
		affected int64
	)
	qs := orm.NewOrm().QueryTable(GetTableCategory())
	if qs.Filter("Pid__in", id...).One(&cate); cate.Id > 0 {
		return errors.New("删除失败：当前分类存在子分类。")
	}
	if affected, err = qs.Filter("Id__in", id...).Filter("Cnt", 0).Delete(); affected > 0 {
		return err
	}
	return errors.New("删除失败：当前分类下存在文档")
}

//获取同级分类
//@param                id              当前同级分类的id
//@return               cates           分类列表数据
func (this *Category) GetSameLevelCategoryById(id interface{}) (cates []Category) {
	cate := Category{Id: helper.Interface2Int(id)}
	o := orm.NewOrm()
	o.Read(&cate)
	o.QueryTable(GetTableCategory()).Filter("Pid", cate.Pid).All(&cates)
	return
}

// 根据父级id获取分类
func (this *Category) GetByPid(pid int, status ...bool) (categories []Category) {
	q := orm.NewOrm().QueryTable(this).Filter("Pid", pid).OrderBy("Sort")
	if len(status) > 0 {
		q = q.Filter("Status", status[0])
	}
	q.All(&categories)
	return
}

// get all categories
func (this *Category) GetAll(status ...bool) (count int64, categories []Category) {
	q := orm.NewOrm().QueryTable(this)
	if len(status) > 0 {
		q = q.Filter("status", status[0])
	}
	count, _ = q.OrderBy("sort").All(&categories)
	return
}

func (this *Category) GetCategoriesById(id ...interface{}) (cates []Category, err error) {
	if len(id) == 0 {
		return
	}

	_, err = orm.NewOrm().QueryTable(this).Filter("Id__in", id...).All(&cates)
	if err != nil {
		helper.Logger.Error(err.Error())
	}

	return
}
