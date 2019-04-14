package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"time"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego/orm"
)

//用户表
type User struct {
	Id       int    `orm:"column(Id)"`
	Email    string `orm:"size(50);unique;column(Email);default();"` //邮箱
	Password string `orm:"size(32);column(Password)"`                //密码
	Username string `orm:"size(16);unique;column(Username)"`         //用户名
	Intro    string `orm:"size(255);default();column(Intro)"`        //个性签名
	Avatar   string `orm:"size(50);default();column(Avatar)"`        //会员头像
}

func NewUser() *User {
	return &User{}
}

func GetTableUser() string {
	return getTable("user")
}

//用户信息表
type UserInfo struct {
	Id         int  `orm:"auto;pk;column(Id)"`                //主键，也就是User表的Id
	Coin       int  `orm:"default(10);index;column(Coin)"`    //金币积分
	Document   int  `orm:"default(0);index;column(Document)"` //文档数量
	Collect    int  `orm:"default(0);column(Collect)"`        //收藏专辑数量，每个收藏专辑下面有文档
	TimeCreate int  `orm:"column(TimeCreate);default(0)"`     //用户注册时间
	Status     bool `orm:"column(Status);default(true)"`      //用户信息状态，false(即0)表示被禁用
}

func NewUserInfo() *UserInfo {
	return &UserInfo{}
}
func GetTableUserInfo() string {
	return getTable("user_info")
}

//根据条件查询用户信息，比如用户登录、用户列表等的获取也可以使用这个函数
//@param            p           int         页码
//@param            listRows    int         每页显示记录数
//@param            orderby     string      排序，如"id desc"
//@param            fields      []string    需要查询的字段
//@param            cond        string      查询条件
//@param            args        ...interface{}  查询条件参数
func (u *User) UserList(p, listRows int, orderby, fields, cond string, args ...interface{}) (params []orm.Params, totalRows int, err error) {
	o := orm.NewOrm()
	if len(orderby) == 0 || orderby == "" {
		orderby = "i.Id desc"
	}
	if len(fields) == 0 {
		fields = "*"
	}
	if len(cond) > 0 {
		cond = "where " + cond
	}

	sqlCount := fmt.Sprintf("select count(i.Id) cnt from %v u left join %v i on u.Id=i.Id %v limit 1",
		GetTableUser(), GetTableUserInfo(), cond,
	)
	var one []orm.Params
	if rows, err := o.Raw(sqlCount, args...).Values(&one); err == nil && rows > 0 {
		totalRows = helper.Interface2Int(one[0]["cnt"])
	}

	sql := fmt.Sprintf("select %v from %v u left join %v i on u.Id=i.Id %v order by %v limit %v offset %v",
		fields, GetTableUser(), GetTableUserInfo(), cond, orderby, listRows, (p-1)*listRows,
	)
	_, err = o.Raw(sql, args...).Values(&params)
	return params, totalRows, err
}

//获取User表的字段
func (u *User) Fields() map[string]string {
	var fields map[string]string
	fields = make(map[string]string)
	v := reflect.ValueOf(u).Elem()
	k := v.Type()
	num := v.NumField()
	for i := 0; i < num; i++ {
		key := k.Field(i)
		fields[key.Name] = key.Name
	}
	return fields
}

//根据用户id获取用户info表的信息
//@param            uid         interface{}         用户UID
//@return           UserInfo                        用户信息
func (u *User) UserInfo(uid interface{}) UserInfo {
	var info UserInfo
	orm.NewOrm().QueryTable(GetTableUserInfo()).Filter("id", uid).One(&info)
	return info
}

//根据查询条件查询User表
//@param            cond            *orm.Condition          查询条件
//@return                           User                    返回查询到的User数据
func (u *User) GetUserField(cond *orm.Condition) (user User) {
	orm.NewOrm().QueryTable(GetTableUser()).SetCond(cond).One(&user)
	return user
}

//用户注册
//@param            email           string          邮箱
//@param            username        string          用户名
//@param            password        string          密码
//@param            repassword      string          确认密码
//@param            intro           string          签名
//@return                           error           错误
//@return                           int             注册成功时返回注册id
func (u *User) Reg(email, username, password, repassword, intro string) (error, int) {
	var (
		user User
		o    = orm.NewOrm()
		now  = time.Now().Unix()
	)

	l := strings.Count(username, "") - 1
	if l < 2 || l > 16 {
		return errors.New("用户名长度限制在2-16个字符"), 0
	}
	if o.QueryTable(GetTableUser()).Filter("Username", username).One(&user); user.Id > 0 {
		return errors.New("用户名已被注册，请更换新的用户名"), 0
	}
	if o.QueryTable(GetTableUser()).Filter("Email", email).One(&user); user.Id > 0 {
		return errors.New("邮箱已被注册，请更换新注册邮箱"), 0
	}
	pwd := helper.MD5Crypt(password)
	if pwd != helper.MD5Crypt(repassword) {
		return errors.New("密码和确认密码不一致"), 0
	}
	user = User{Email: email, Username: username, Password: pwd, Intro: intro}
	_, err := o.Insert(&user)
	if user.Id > 0 {
		//coin := beego.AppConfig.DefaultInt("coinreg", 10)
		coin := NewSys().GetByField("CoinReg").CoinReg
		var userinfo = UserInfo{Id: user.Id, Status: true, Coin: coin, TimeCreate: int(now)}
		_, err = o.Insert(&userinfo)
	}
	return err, user.Id
}

//获取除了用户密码之外的用户全部信息
//@param                id              用户id
//@return               params          用户信息
//@return               rows            记录数
//@return               err             错误
func (this *User) GetById(id interface{}) (params orm.Params, rows int64, err error) {
	var data []orm.Params
	tables := []string{GetTableUser() + " u", GetTableUserInfo() + " ui"}
	on := []map[string]string{
		{"u.Id": "ui.Id"},
	}
	fields := map[string][]string{
		"u":  GetFields(NewUser()),
		"ui": GetFields(NewUserInfo()),
	}
	if sql, err := LeftJoinSqlBuild(tables, on, fields, 1, 1, nil, nil, "u.Id=?"); err == nil {
		if rows, err = orm.NewOrm().Raw(sql, id).Values(&data); err == nil && len(data) > 0 {
			params = data[0]
		}
	}
	return
}

var (
	errFailedToDown = errors.New("文档下载失败")
	errParams       = errors.New("参数错误")
	errNotFound     = errors.New("文档不存在")
	errCannotDown   = errors.New("该文档不允许下载")
	errNotExistUser = errors.New("用户不存在")
	errLessCoin     = errors.New("金币不足")
)

// 判断用户是否可以下载指定文档
// err 为 nil 表示可以下载
func (this *User) CanDownloadFile(uid, docId int) (urlStr string, err error) {
	// 1. 判断用户和文档是否存在
	// 2. 判断用户是否可以免费下载，如果用户不可以免费下载，则再扣费之后，允许用户免费下载
	// 3. 文档被下载次数增加，用户(文档下载人和文档分享人)积分变更，并增加两个用户的积分记录
	// 4. 获取文档下载URL链接
	if uid <= 0 || docId <= 0 {
		err = errParams
		return
	}

	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if err == nil {
			o.Commit()
		} else {
			o.Rollback()
		}
	}()

	now := int(time.Now().Unix())

	u := &UserInfo{Id: uid}
	err = o.Read(u)
	if err != nil {
		return
	}

	var docInfo = DocumentInfo{Id: docId}
	if err = o.Read(&docInfo); err != nil {
		helper.Logger.Error(err.Error())
		err = errNotFound
		return
	}

	if docInfo.Price < 0 {
		err = errCannotDown
		return
	}

	doc := &Document{Id: docId}
	if err = o.Read(doc); err != nil {
		helper.Logger.Error(err.Error())
		err = errNotFound
		return
	}

	store := &DocumentStore{Id: docInfo.DsId}
	if err = o.Read(store); err != nil {
		helper.Logger.Error(err.Error())
		err = errNotFound
		return
	}

	price := docInfo.Price

	if u.Id == 0 {
		err = errNotExistUser
		return
	}

	isFree := NewFreeDown().IsFreeDown(uid, docId)
	if isFree {
		price = 0
	}

	if u.Coin < price {
		err = errLessCoin
		return
	}

	logDown := CoinLog{
		Uid:        uid,
		TimeCreate: now,
	}
	logShare := CoinLog{
		Uid:        docInfo.Uid,
		TimeCreate: now,
	}

	if price > 0 {
		// 文档下载人，扣除积分
		u.Coin = u.Coin - price
		if _, err = o.Update(u); err != nil {
			return
		}
		// 文档分享人，增加积分
		sqlShareUser := fmt.Sprintf("update `%v` set `Coin`=`Coin`+? where Id = ?", GetTableUserInfo())
		if _, err = o.Raw(sqlShareUser, price, docInfo.Uid).Exec(); err != nil {
			return
		}

		// 增加免费下载记录
		free := &FreeDown{Uid: uid, Did: docId, TimeCreate: now}
		if _, err = o.Insert(free); err != nil {
			return
		}

		logDown.Coin = -price
		logDown.Log = fmt.Sprintf("下载文档(%v)，花费 %v 金币", doc.Title, price)

		logShare.Coin = price
		logShare.Log = fmt.Sprintf("您分享的文档(%v)被其他用户下载，获得 %v 金币", doc.Title, price)
		if _, err = o.Insert(&logShare); err != nil {
			helper.Logger.Error(err.Error())
			err = errFailedToDown
			return
		}
	} else {
		logDown.Log = fmt.Sprintf("在免费期限内，下载同一篇文档《%v》免费", doc.Title)
	}

	if _, err = o.Insert(&logDown); err != nil {
		helper.Logger.Error(err.Error())
		err = errFailedToDown
		return
	}

	docInfo.Dcnt += 1
	if _, err = o.Update(&docInfo); err != nil {
		return
	}

	var cs *CloudStore
	cs, err = NewCloudStore(true)
	if err != nil {
		helper.Logger.Error(err.Error())
		err = errFailedToDown
		return
	}

	object := store.Md5 + "." + strings.TrimLeft(store.Ext, ".")
	urlStr = cs.GetSignURL(object)

	return
}
