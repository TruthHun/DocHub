package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/TruthHun/DocHub/helper"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

//文档回收站
type DocumentRecycle struct {
	Id   int  `orm:"column(Id)"`                  //对应的文档id
	Uid  int  `orm:"default(0);column(Uid)"`      //操作用户
	Date int  `orm:"default(0);column(Date)"`     //操作时间
	Self bool `orm:"default(false);column(Self)"` //是否是文档上传用户删除的，默认为false。如果是文档上传者删除的，设置为true
}

//将文档从回收站中恢复过来，文档的状态必须是-1才可以
//@param            ids             文档id
//@return           err             返回错误，nil表示恢复成功，否则恢复失败
func (this *DocumentRecycle) RecoverFromRecycle(ids ...interface{}) (err error) {
	if len(ids) > 0 {
		qs := O.QueryTable(TableDocInfo).Filter("Id__in", ids...).Filter("Status", -1)
		var docinfo []DocumentInfo
		qs.All(&docinfo)
		if affectedRows, err := qs.Update(orm.Params{"Status": 1}); affectedRows > 0 {
			//总文档数量增加
			Regulate(TableSys, "CntDoc", int(affectedRows), "Id=1")
			beego.Debug("查询到的文档", docinfo)
			if len(docinfo) > 0 {
				for _, v := range docinfo {
					//该用户的文档数量+1
					if err := Regulate(TableUserInfo, "Document", 1, "Id=?", v.Uid); err != nil {
						helper.Logger.Error(err.Error())
					}
					//该分类下的文档数量+1
					Regulate(TableCategory, "Cnt", 1, fmt.Sprintf("`Id` in(%v,%v,%v)", v.ChanelId, v.Cid, v.Pid))
				}
			}
			//从回收站中删除记录
			O.QueryTable(TableDocRecycle).Filter("Id__in", ids...).Delete()
			//从非法文档中将文档移除
			O.QueryTable(TableDocIllegal).Filter("Id__in", ids...).Delete()
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
	tables := []string{TableDocRecycle + " dr", TableDoc + " d", TableDocInfo + " di", TableUser + " u", TableDocStore + " ds"}
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
		rows, err = O.Raw(sql).Values(&params)
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
	var (
		docinfo []DocumentInfo
	)
	if len(ids) > 0 {
		O.QueryTable(TableDocInfo).Filter("Id__in", ids...).All(&docinfo)
		//总文档记录减少
		Regulate(TableSys, "CntDoc", -len(docinfo), "Id=1")
		for _, info := range docinfo {
			//文档分类统计数量减少
			if err := Regulate(TableCategory, "Cnt", -1, "Id in(?,?,?)", info.ChanelId, info.Pid, info.Cid); err != nil {
				helper.Logger.Error(err.Error())
			}
			//用户文档统计数量减少
			if err := Regulate(TableUserInfo, "Document", -1, "Id=?", info.Uid); err != nil {
				helper.Logger.Error(err.Error())
			}
		}
		//变更文档状态
		if _, err := UpdateByIds(TableDocInfo, "Status", -1, ids...); err != nil {
			helper.Logger.Error(err.Error())
			errs = append(errs, err.Error())
		}
		//移入回收站
		for _, id := range ids {
			var rc DocumentRecycle
			rc.Id = helper.Interface2Int(id)
			rc.Uid = helper.Interface2Int(uid)
			rc.Date = int(time.Now().Unix())
			rc.Self = self
			if _, err := O.Insert(&rc); err != nil {
				helper.Logger.Error(err.Error())
			}
		}
	} else {
		errs = append(errs, "参数错误:缺少文档id")
	}
	return errs
}
