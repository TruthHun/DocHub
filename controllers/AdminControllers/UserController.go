package AdminControllers

import (
	"fmt"

	"github.com/TruthHun/DocHub/helper"

	"strings"

	"github.com/TruthHun/DocHub/helper/conv"
	"github.com/TruthHun/DocHub/models"
)

//IT文库注册会员管理

type UserController struct {
	BaseController
}

func (this *UserController) Prepare() {
	this.BaseController.Prepare()
	this.Data["IsUser"] = true
}

//用户列表
func (this *UserController) List() {
	var (
		condition []string
		listRows  = 10
		id        = 0
		p         = 1
		username  string
	)
	//path中的参数
	params := conv.Path2Map(this.GetString(":splat"))

	//页码处理
	if _, ok := params["p"]; ok {
		p = helper.Interface2Int(params["p"])
	} else {
		p, _ = this.GetInt("p")
	}
	p = helper.NumberRange(p, 1, 1000000)

	//搜索的用户id处理
	if _, ok := params["id"]; ok {
		id = helper.Interface2Int(params["id"])
	} else {
		id, _ = this.GetInt("id")
	}
	if id > 0 {
		condition = append(condition, fmt.Sprintf("i.Id=%v", id))
		this.Data["Id"] = id
	}

	//搜索的用户名处理
	if _, ok := params["username"]; ok {
		username = params["username"]
	} else {
		username = this.GetString("username")
	}
	if len(username) > 0 {
		condition = append(condition, fmt.Sprintf(`u.Username like "%v"`, "%"+username+"%"))
		this.Data["Username"] = username
	}

	data, totalRows, err := models.NewUser().UserList(p, listRows, "", "*", strings.Join(condition, " and "))
	if err != nil {
		this.Ctx.WriteString(err.Error())
		return
	}

	this.Data["Page"] = helper.Paginations(6, totalRows, listRows, p, "/admin/user/", "id", id, "username", username)
	this.Data["Users"] = data
	this.Data["ListRows"] = listRows
	this.Data["TotalRows"] = totalRows
	this.TplName = "list.html"
}
