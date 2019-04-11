package models

import (
	"errors"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego/orm"
)

//会员文档收藏的文件夹
type CollectFolder struct {
	Id          int    `orm:"column(Id)"`
	Cover       string `orm:"column(Cover);size(50);default()"`        //文档收藏夹(专辑封面)
	Title       string `orm:"column(Title);size(100);default(默认收藏夹)"`  //会员收藏文档的存放收藏夹
	Description string `orm:"column(Description);size(512);default()"` //会员创建的收藏夹的描述
	Uid         int    `orm:"column(Uid);index"`                       //归属于哪个会员的收藏夹
	TimeCreate  int    `orm:"column(TimeCreate)"`                      //收藏夹创建时间
	Cnt         int    `orm:"column(Cnt);default(0)"`                  //收藏夹默认的文档数量
}

//会员文档收藏表
type Collect struct {
	Id  int `orm:"column(Id)"`
	Cid int `orm:"column(Cid);index"` //文档收藏的自定义收藏的文件夹
	Did int `orm:"column(Did)"`       //文档id:document id
}

// 文档收藏表多字段唯一索引
func (clt *Collect) TableUnique() [][]string {
	return [][]string{
		[]string{"Did", "Cid"},
	}
}

// 文档收藏表多字段唯一索引
func (cf *CollectFolder) TableUnique() [][]string {
	return [][]string{
		[]string{"Title", "Uid"},
	}
}

func NewCollectFolder() *CollectFolder {
	return &CollectFolder{}
}

func NewCollect() *Collect {
	return &Collect{}
}

func GetTableCollectFolder() string {
	return getTable("collect_folder")
}

func GetTableCollect() string {
	return getTable("collect")
}

//取消文档收藏
//@param                did             文档id
//@param                cid             CollectFolder表的id，即收藏夹id
//@param                uid             用户id
//@param                err             返回错误
func (this *Collect) Cancel(did, cid interface{}, uid int) (err error) {
	var affected int64
	if affected, err = orm.NewOrm().QueryTable(GetTableCollect()).Filter("Did", did).Filter("Cid", cid).Delete(); err == nil && affected > 0 {
		Regulate(GetTableCollectFolder(), "Cnt", -1, "Id=?", cid) //收藏夹收藏的文档数量-1
		Regulate(GetTableDocumentInfo(), "Ccnt", -1, "Id=?", did) //文档被收藏次数-1
	}
	return err
}

//删除收藏夹，当收藏夹里面收藏的文档不为空时，不允许删除
//@param                id              收藏夹id
//@param                uid             用户id
//@return               err             错误，如果错误为nil，则表示删除成功，否则删除失败
func (this *Collect) DelFolder(id, uid int) (err error) {
	//查询判断收藏夹是否是当前用户的收藏夹
	var (
		cf = CollectFolder{Id: id}
		o  = orm.NewOrm()
	)
	err = o.Read(&cf)
	if err != nil {
		return
	}

	if cf.Uid != uid {
		return
	}

	if cf.Cnt > 0 {
		err = errors.New("收藏夹删除失败：您要删除的收藏夹不为空")
		return
	}

	o.Begin()
	defer func() {
		if err != nil {
			o.Rollback()
		}

		o.Commit()

		if len(cf.Cover) > 0 {
			if cs, errCS := NewCloudStore(false); errCS != nil {
				helper.Logger.Error(errCS.Error())
			} else {
				go cs.Delete(cf.Cover)
			}
		}

	}()

	if _, err = o.Delete(&cf, "Id"); err != nil {
		return
	}

	sql := "update `%v` set `Collect`=`Collect`-1 where `Collect`>0 AND Id = ?"
	_, err = o.Raw(sql, GetTableUserInfo(), uid).Exec()

	return
}

//删除指定的文档收藏，比如某文档是侵权或者非法，则凡是收藏了该文档的用户，该文档收藏都将被删除
//@param            dids            文档id
//@return           err             错误，nil表示删除成功
func (this *Collect) DelByDocId(dids ...interface{}) (err error) {
	var (
		clt []Collect //文档收藏记录
		ids []interface{}
		o   = orm.NewOrm()
	)
	if len(dids) > 0 {
		if _, err = o.QueryTable(GetTableCollect()).Filter("Did__in", dids...).All(&clt); err == nil { //查询收藏
			for _, v := range clt {
				ids = append(ids, v.Id)
				Regulate(GetTableCollectFolder(), "Cnt", -1, "Id=?", v.Cid) //收藏夹收藏的文档统计数量-1
				Regulate(GetTableDocumentInfo(), "Ccnt", -1, "Id=?", v.Did) //文档被收藏次数-1
			}
		}
		if len(ids) > 0 { //删除收藏
			_, err = o.QueryTable(GetTableCollect()).Filter("Id__in", ids...).Delete()
		}
	}
	return err
}
