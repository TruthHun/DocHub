package models

import "errors"

//会员文档收藏的文件夹
type CollectFolder struct {
	Id          int    `orm:"column(Id)"`
	Cover       string `orm:"column(Cover);size(50);default()"`        //文档收藏夹(专辑封面)
	Title       string `orm:"column(Title);default(默认收藏夹)"`            //会员收藏文档的存放收藏夹
	Description string `orm:"column(Description);size(512);default()"` //会员创建的收藏夹的描述
	Uid         int    `orm:"column(Uid);index"`                       //归属于哪个会员的收藏夹
	TimeCreate  int    `orm:"column(TimeCreate)"`                      //收藏夹创建时间
	Cnt         int    `orm:"column(Cnt);default(0)"`                  //收藏夹默认的文档数量
}

// 文档收藏表多字段唯一索引
func (cf *CollectFolder) TableUnique() [][]string {
	return [][]string{
		[]string{"Title", "Uid"},
	}
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

//取消文档收藏
//@param                did             文档id
//@param                cid             CollectFolder表的id，即收藏夹id
//@param                uid             用户id
//@param                err             返回错误
func (this *Collect) Cancel(did, cid interface{}, uid int) (err error) {
	var affected int64
	if affected, err = O.QueryTable(TableCollect).Filter("Did", did).Filter("Cid", cid).Filter("Uid", uid).Delete(); err == nil && affected > 0 {
		Regulate(TableCollectFolder, "Cnt", -1, "Id=?", cid) //收藏夹收藏的文档数量-1
		Regulate(TableDocInfo, "Ccnt", -1, "Id=?", did)      //文档被收藏次数-1
	}
	return err
}

//删除收藏夹，当收藏夹里面收藏的文档不为空时，不允许删除
//@param                id              收藏夹id
//@param                uid             用户id
//@return               err             错误，如果错误为nil，则表示删除成功，否则删除失败
func (this *Collect) DelFolder(id, uid int) (err error) {
	//查询判断收藏夹是否是当前用户的收藏夹
	var cf = CollectFolder{Id: id}
	if err = O.Read(&cf); err == nil && cf.Uid == uid {
		if cf.Cnt > 0 {
			err = errors.New("收藏夹删除失败：您要删除的收藏夹不为空")
		} else {
			if _, err = O.Delete(&cf, "Id"); err == nil {
				if len(cf.Cover) > 0 {
					go ModelOss.DelFromOss(true, cf.Cover)
				}
				err = Regulate("user_info", "Collect", -1, "Uid=?", uid)
			}
		}
	}
	return
}

//删除指定的文档收藏，比如某文档是侵权或者非法，则凡是收藏了该文档的用户，该文档收藏都将被删除
//@param            dids            文档id
//@return           err             错误，nil表示删除成功
func (this *Collect) DelByDocId(dids ...interface{}) (err error) {
	var (
		clt []Collect //文档收藏记录
		ids []interface{}
	)
	if len(dids) > 0 {
		if _, err = O.QueryTable(TableCollect).Filter("Did__in", dids...).All(&clt); err == nil { //查询收藏
			for _, v := range clt {
				ids = append(ids, v.Id)
				Regulate(TableCollectFolder, "Cnt", -1, "Id=?", v.Cid) //收藏夹收藏的文档统计数量-1
				Regulate(TableDocInfo, "Ccnt", -1, "Id=?", v.Did)      //文档被收藏次数-1
			}
		}
		if len(ids) > 0 { //删除收藏
			_, err = O.QueryTable(TableCollect).Filter("Id__in", ids...).Delete()
		}
	}
	return err
}
