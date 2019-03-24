package models

import (
	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego/orm"
)

//管理员数据表
type Admin struct {
	Id       int    `orm:"column(Id)"`                       //自增主键
	Username string `orm:"size(16);unique;column(Username)"` //用户名，唯一
	Password string `orm:"size(32);column(Password)"`        //密码
	Email    string `orm:"size(50);default();column(Email)"` //邮箱
	Code     string `orm:"size(30);default();column(Code)"`  //您心目中的验证码
}

func NewAdmin() *Admin {
	return &Admin{}
}
func GetTableAdmin() string {
	return getTable("admin")
}

//管理员登录
//@param            username            用户名
//@param            password            经过md5加密后的密码
//@param            code                登录暗号
//@return           admin               管理员数据结构，如果登录成功，管理员id大于0
//@return           err                 SQL查询过程中出现的错误
func (this *Admin) Login(username, password, code string) (admin Admin, err error) {
	admin = Admin{Username: username, Password: helper.MD5Crypt(password), Code: code}
	err = orm.NewOrm().Read(&admin, "Username", "Password", "Code")
	return
}

//根据管理员ID获取管理员信息
//@param            id          管理员id
//@return           admin       管理员信息
//@return           err         错误信息
func (this *Admin) GetById(id int) (admin Admin, err error) {
	admin.Id = id
	err = orm.NewOrm().Read(&admin)
	return
}
