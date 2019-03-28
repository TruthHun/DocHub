package models

import (
	"fmt"
	"strings"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego/orm"
)

// ============================ //
// 废弃！
// 相关文档的功能，以后通过 ElasticSearch 来实现
// 废弃！
// ============================ //

// 关键字记录表，后期用这个来做相关资源功能
type Word struct {
	Id     int    `orm:"column(Id)"`
	Wd     string `orm:"column(Wd);size(20);unique"`       //关键字
	Cnt    int    `orm:"column(Cnt);default(0)"`           //统计
	Ids    string `orm:"column(Ids);default();type(text)"` //存在该关键字的文档id
	Status bool   `orm:"column(Status);default(true)"`     //bool值，默认该关键字合法，否则存在该关键字的都是不合法文档
}

func NewWord() *Word {
	return &Word{}
}

func GetTableWord() string {
	return getTable("word")
}

//添加关键字，多个关键字用英文逗号分隔
func (this *Word) AddWords(wds string, id interface{}) {
	var (
		wdMap   = make(map[string]string) //词汇map
		wdSlice []interface{}             //词汇切片
		wdData  []Word                    //词汇数据
		o       = orm.NewOrm()
	)
	slice := strings.Split(wds, ",")
	if len(slice) > 0 {
		for _, wd := range slice {
			wd = strings.TrimSpace(wd)
			cnt := strings.Count(wd, "") - 1
			if cnt > 1 && cnt <= 20 { //2-20个字符
				wdMap[wd] = wd
				wdSlice = append(wdSlice, wd)
			}
		}
		o.QueryTable(GetTableWord()).Filter("Wd__in", wdSlice...).All(&wdData)
		for _, w := range wdData { //存在数据，则更新
			w.Cnt = w.Cnt + 1
			w.Ids = fmt.Sprintf("%v,%v", w.Ids, id)
			if _, err := o.Update(&w, "Ids", "Cnt"); err == nil { //更新分词数据
				//删除map
				delete(wdMap, w.Wd)
			} else {
				helper.Logger.Error(err.Error())
			}
		}
		if len(wdMap) > 0 { //还存在数据，则新增
			var wdDataAdd []Word
			for _, w := range wdMap {
				wdDataAdd = append(wdDataAdd, Word{
					Wd:     w,
					Cnt:    1,
					Ids:    fmt.Sprintf("%v", id),
					Status: true,
				})
			}
			if len(wdDataAdd) > 0 {
				if _, err := o.InsertMulti(len(wdDataAdd), wdDataAdd); err != nil {
					helper.Logger.Error(err.Error())
				}
			}
		}
	}
}
