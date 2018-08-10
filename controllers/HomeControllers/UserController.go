package HomeControllers

import (
	"fmt"

	"strings"

	"time"

	"os"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/helper/conv"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
)

type UserController struct {
	BaseController
}

func (this *UserController) Prepare() {
	this.BaseController.Prepare()
	this.Xsrf()
}

//会员中心
func (this *UserController) Get() {

	uid, _ := this.GetInt(":uid")
	path := this.GetString(":splat")
	params := conv.Path2Map(path)
	//排序
	sort := "new"
	if param, ok := params["sort"]; ok {
		sort = param
	}
	//页码
	p := 1
	if page, ok := params["p"]; ok {
		p = helper.Interface2Int(page)
		if p < 1 {
			p = 1
		}
	}

	switch sort {
	case "dcnt":
		sort = "dcnt"
	case "score":
		sort = "score"
	case "vcnt":
		sort = "vcnt"
	case "ccnt":
		sort = "ccnt"
	default:
		sort = "new"
	}
	//显示风格
	style := "list"
	if s, ok := params["style"]; ok {
		style = s
	}
	if style != "th" {
		style = "list"
	}
	//cid:collect folder id ,收藏夹id
	cid := 0
	if s, ok := params["cid"]; ok {
		cid = helper.Interface2Int(s)
	}
	if p < 1 {
		p = 1
	}
	if uid < 1 {
		uid = this.IsLogin
	}
	this.Data["Uid"] = uid
	if uid > 0 {
		listRows := 16
		user, rows, err := models.NewUser().GetById(uid)
		if err != nil {
			helper.Logger.Error(err.Error())
		}
		if rows == 0 {
			this.Redirect("/", 302)
			return
		}

		ModelUser := models.User{}
		if cid > 0 {
			sql := fmt.Sprintf("select Title,Cnt from %v where Id=? limit 1", models.GetTableCollectFolder())
			var params []orm.Params
			orm.NewOrm().Raw(sql, cid).Values(&params)
			if len(params) > 0 {
				this.Data["Folder"] = params[0]
				fields := "di.Id,di.`Uid`, di.`Cid`, di.`TimeCreate`, di.`Dcnt`, di.`Vcnt`, di.`Ccnt`, di.`Score`, di.`Status`, di.`ChanelId`, di.`Pid`,c.Title Category,u.Username,d.Title,ds.`Md5`, ds.`Ext`, ds.`ExtCate`, ds.`ExtNum`, ds.`Page`, ds.`Size`"
				sql_format := `
							select %v from %v di left join %v u on di.Uid=u.Id
							left join %v clt on clt.Did=di.Id
							left join %v d on d.Id=di.Id
							left join %v c on c.Id=di.cid
							left join %v ds on ds.Id=di.DsId
							where %v order by %v limit %v,%v
							`
				sql = fmt.Sprintf(sql_format,
					fields,
					models.GetTableDocumentInfo(),
					models.GetTableUser(),
					models.GetTableCollect(),
					models.GetTableDocument(),
					models.GetTableCategory(),
					models.GetTableDocumentStore(),
					fmt.Sprintf("clt.Cid=%v", cid),
					"clt.Id desc",
					(p-1)*listRows, listRows,
				)
				var data []orm.Params
				orm.NewOrm().Raw(sql).Values(&data)
				this.Data["Lists"] = data
				this.Data["Page"] = helper.Paginations(6, helper.Interface2Int(params[0]["Cnt"]), listRows, p, fmt.Sprintf("/user/%v/doc/cid/%v", user["Id"], cid), "sort", sort, "style", style)
			} else {
				this.Redirect(fmt.Sprintf("/user/%v/collect", uid), 302)
			}
		} else {
			this.Data["Lists"], _, _ = models.DocList(uid, 0, 0, 0, p, listRows, sort, 1)
			this.Data["Page"] = helper.Paginations(6, helper.Interface2Int(user["Document"]), listRows, p, fmt.Sprintf("/user/%v/doc", user["Id"]), "sort", sort, "style", style)
		}
		this.Data["Cid"] = cid
		this.Data["User"] = user
		this.Data["PageId"] = "wenku-user"
		this.Data["IsUser"] = true
		this.Data["Sort"] = sort
		this.Data["Style"] = style
		this.Data["P"] = p
		this.Data["Seo"] = models.NewSeo().GetByPage("PC-Ucenter-Doc", "文档列表-会员中心-"+user["Username"].(string), "会员中心,文档列表,"+user["Username"].(string), "文档列表-会员中心-"+user["Username"].(string), this.Sys.Site)
		this.Data["Ranks"], _, err = ModelUser.UserList(1, 8, "i.Document desc", "u.Id,u.Username,u.Avatar,u.Intro,i.Document", "i.Status=1")
		if err != nil {
			helper.Logger.Error(err.Error())
		}
		this.TplName = "index.html"
	} else {
		this.Redirect("/user/login", 302)
		return
	}
}

//金币记录
func (this *UserController) Coin() {

	uid, _ := this.GetInt(":uid")
	p, _ := this.GetInt("p", 1)
	if p < 1 {
		p = 1
	}
	if uid < 1 {
		uid = this.IsLogin
	}

	if uid > 0 {
		listRows := 16
		lists, _, _ := models.GetList(models.GetTableCoinLog(), p, listRows, orm.NewCondition().And("Uid", uid), "-Id")
		if p > 1 {
			this.ResponseJson(1, "数据获取成功", lists)
		} else {
			user, rows, err := models.NewUser().GetById(uid)
			if err != nil {
				helper.Logger.Error(err.Error())
			}
			if rows == 0 {
				this.Redirect("/", 302)
				return
			}
			ModelUser := models.User{}
			this.Data["Lists"] = lists
			this.Data["User"] = user
			this.Data["PageId"] = "wenku-user"
			this.Data["IsUser"] = true
			this.Data["Ranks"], _, err = ModelUser.UserList(1, 8, "i.Document desc", "u.Id,u.Username,u.Avatar,u.Intro,i.Document", "i.Status=1")
			if err != nil {
				helper.Logger.Error(err.Error())
			}
			this.Data["Seo"] = models.NewSeo().GetByPage("PC-Ucenter-Coin", "财富记录—会员中心-"+user["Username"].(string), "会员中心,财富记录,"+user["Username"].(string), "财富记录—会员中心-"+user["Username"].(string), this.Sys.Site)
			this.TplName = "coin.html"
		}
	} else {
		this.Redirect("/user/login", 302)
		return
	}
}

//收藏夹
func (this *UserController) Collect() {
	action := this.GetString("action")
	uid, _ := this.GetInt(":uid")
	p, _ := this.GetInt("p", 1)
	if p < 1 {
		p = 1
	}
	if uid < 1 {
		uid = this.IsLogin
	}
	this.Data["Uid"] = uid
	if uid > 0 {
		listRows := 100
		lists, _, _ := models.GetList(models.GetTableCollectFolder(), p, listRows, orm.NewCondition().And("Uid", uid), "-Id")
		if p > 1 {
			this.ResponseJson(1, "数据获取成功", lists)
		} else {
			user, rows, err := models.NewUser().GetById(uid)
			if err != nil {
				helper.Logger.Error(err.Error())
			}
			if rows == 0 {
				this.Redirect("/", 302)
				return
			}
			ModelUser := models.User{}
			this.Data["Lists"] = lists
			this.Data["User"] = user
			this.Data["PageId"] = "wenku-user"
			this.Data["IsUser"] = true
			this.Data["Ranks"], _, err = ModelUser.UserList(1, 8, "i.Document desc", "u.Id,u.Username,u.Avatar,u.Intro,i.Document", "i.Status=1")
			if err != nil {
				helper.Logger.Error(err.Error())
			}
			this.TplName = "collect.html"
			this.Data["Seo"] = models.NewSeo().GetByPage("PC-Ucenter-Folder", "收藏夹—会员中心-"+user["Username"].(string), "会员中心,收藏夹,"+user["Username"].(string), "收藏夹—会员中心-"+user["Username"].(string), this.Sys.Site)
			if action == "edit" {
				this.Data["Edit"] = true
			} else {
				this.Data["Edit"] = false
			}
		}
	} else {
		this.Redirect("/user/login", 302)
		return
	}
}

//用户登录
func (this *UserController) Login() {

	if this.IsLogin > 0 {
		this.Redirect("/user", 302)
		return
	}
	if this.Ctx.Request.Method == "POST" {
		type Post struct {
			Email, Password string
		}
		ret := map[string]interface{}{"status": 0, "msg": "登录失败，邮箱格式不正确"}
		var post Post
		this.ParseForm(&post)
		valid := validation.Validation{}
		res := valid.Email(post.Email, "Email")
		if res.Ok {
			ModelUser := models.NewUser()
			users, rows, err := ModelUser.UserList(1, 1, "", "", "u.`email`=? and u.`password`=?", post.Email, helper.MyMD5(post.Password))
			if rows > 0 && err == nil {

				user := users[0]
				this.IsLogin = helper.Interface2Int(user["Id"])
				if this.IsLogin > 0 {
					//查询用户有没有被封禁
					if info := ModelUser.UserInfo(this.IsLogin); info.Status == false { //被封禁了
						//登录失败，账号已被封禁
						ret["status"] = 0
						ret["msg"] = "登录失败，您的账号已被管理员禁用"
					} else {
						//登录成功
						ret["status"] = 1
						ret["msg"] = "登录成功"
						this.BaseController.SetCookieLogin(this.IsLogin)
					}
				}
			} else {
				if err != nil {
					helper.Logger.Error(err.Error())
				}
				//登录失败
				ret["msg"] = "登录失败，邮箱或密码不正确"
			}
		}
		this.Data["json"] = ret
		this.ServeJSON()
	}
	this.Data["Seo"] = models.NewSeo().GetByPage("PC-Login", "会员登录", "会员登录", "会员登录", this.Sys.Site)
	this.Data["IsUser"] = true
	this.Data["PageId"] = "wenku-reg"
	this.TplName = "login.html"
}

//用户退出登录
func (this *UserController) Logout() {
	this.ResetCookie()
	ret := map[string]interface{}{"status": 1, "msg": "退出登录成功"}
	if v, ok := this.Ctx.Request.Header["X-Requested-With"]; ok && v[0] == "XMLHttpRequest" {
		this.Data["json"] = ret
		this.ServeJSON()
		return
	}
	this.Redirect("/", 302)
	return
}

//会员注册
func (this *UserController) Reg() {
	if this.IsLogin > 0 {
		this.Redirect("/user", 302)
		return
	}
	this.Data["IsUser"] = true
	if this.Ctx.Request.Method == "POST" && this.Sys.IsCloseReg == false {
		ret := map[string]interface{}{"status": 0, "msg": "邮箱验证码不正确，请重新输入或重新获取"}
		//先验证邮箱验证码是否正确
		email := this.GetString("email")
		code := this.GetString("code")
		sess_email := fmt.Sprintf("%v", this.GetSession("RegMail"))
		sess_code := fmt.Sprintf("%v", this.GetSession("RegCode"))
		if sess_email == email && sess_code == code {
			ModelUser := models.User{}
			err, uid := ModelUser.Reg(
				email,
				this.GetString("username"),
				this.GetString("password"),
				this.GetString("repassword"),
				this.GetString("intro"),
			)
			if err == nil && uid > 0 {

				//站点用户数量增加
				models.Regulate(models.GetTableSys(), "CntUser", 1, "Id=1")
				this.IsLogin = uid
				this.SetCookieLogin(uid)
				ret["status"] = 1
				ret["msg"] = "会员注册成功"
			} else {
				ret["msg"] = err.Error()
			}
		}
		this.Data["json"] = ret
		this.ServeJSON()
	}
	this.Data["Seo"] = models.NewSeo().GetByPage("PC-Login", "会员注册", "会员注册", "会员注册", this.Sys.Site)
	this.Data["PageId"] = "wenku-reg"
	if this.Sys.IsCloseReg {
		this.TplName = "regclose.html"
	} else {
		this.TplName = "reg.html"
	}
}

//发送邮件
func (this *UserController) SendMail() {
	if len(this.Ctx.GetCookie(beego.AppConfig.String("SessionName"))) == 0 {
		this.Redirect("/", 302)
		return
	}
	ret := map[string]interface{}{"status": 0, "msg": "邮件发送类型不正确"}
	//发送邮件的类型：注册(reg)和找回密码(findpwd)
	t := this.GetString("type")
	mail := this.GetString("email")
	if t == "reg" || t == "findpwd" {
		valid := validation.Validation{}
		res := valid.Email(mail, "mail")
		helper.Logger.Debug(res.Error.Error())
		if res.Ok && res.Error == nil {
			//检测邮箱是否已被注册
			ModelUser := models.User{}
			u := ModelUser.GetUserField(orm.NewCondition().And("email", mail))
			//注册邮件
			if t == "reg" {
				if u.Id > 0 {
					ret["msg"] = "该邮箱已经被注册会员"
				} else {
					code := helper.RandStr(6, 0)
					err := models.SendMail(mail, fmt.Sprintf("%v会员注册验证码", this.Sys.Site), strings.Replace(this.Sys.TplEmailReg, "{code}", code, -1))
					if err == nil {
						ret["status"] = 1
						ret["msg"] = "邮件发送成功，请打开邮箱查看验证码"
						this.SetSession("RegMail", mail)
						this.SetSession("RegCode", code)
					} else {
						helper.Logger.Error("邮件发送失败：%v", err.Error())
						ret["msg"] = "邮件发送失败，请联系管理员检查邮箱配置是否正确"
					}
				}
			} else {
				//找回密码
				if u.Id > 0 {
					code := helper.RandStr(6, 0)
					err := models.SendMail(mail, fmt.Sprintf("%v找回密码验证码", this.Sys.Site), strings.Replace(this.Sys.TplEmailFindPwd, "{code}", code, -1))
					if err == nil {
						ret["status"] = 1
						ret["msg"] = "邮件发送成功，请打开邮箱查看验证码"
						this.SetSession("FindPwdMail", mail)
						this.SetSession("FindPwdCode", code)
					} else {
						helper.Logger.Error("邮件发送失败：%v", err.Error())
						ret["msg"] = "邮件发送失败，请联系管理员检查邮箱配置是否正确"
					}
				} else {
					ret["msg"] = "该邮箱不存在"
				}
			}

		} else {
			ret["msg"] = "邮箱格式不正确"
		}
	}
	this.Data["json"] = ret
	this.ServeJSON()
}

//会员签到，增加金币
func (this *UserController) Sign() {
	if this.IsLogin > 0 {
		var data = models.Sign{
			Uid:  this.IsLogin,
			Date: time.Now().Format("20060102"),
		}
		_, err := orm.NewOrm().Insert(&data)
		if err != nil {
			this.ResponseJson(0, "签到失败，您今天已签到")
		} else {
			if err := models.Regulate(models.GetTableUserInfo(), "Coin", this.Sys.Sign, fmt.Sprintf("Id=%v", this.IsLogin)); err == nil {
				log := models.CoinLog{
					Uid:  this.IsLogin,
					Coin: this.Sys.Sign,
					Log:  fmt.Sprintf("于%v签到成功，增加 %v 个金币", time.Now().Format("2006-01-02 15:04:05"), this.Sys.Sign),
				}
				models.NewCoinLog().LogRecord(log)
			}
			this.ResponseJson(1, fmt.Sprintf("恭喜您，今日签到成功，领取了 %v 个金币", this.Sys.Sign))
		}
	} else {
		this.ResponseJson(0, "签到失败，请先登录")
	}
}

//检测用户是否已登录
func (this *UserController) CheckLogin() {
	ret := map[string]interface{}{"status": 0, "msg": "您当前处于未登录状态，请先登录"}
	uid := this.BaseController.IsLogin
	if uid > 0 {
		ret["status"] = 1
		ret["msg"] = "已登录"
	}
	this.Data["json"] = ret
	this.ServeJSON()
	this.StopRun()
}

//创建收藏夹
func (this *UserController) CreateCollectFolder() {
	if this.IsLogin > 0 {
		cover := ""
		timestamp := int(time.Now().Unix())
		//文件在文档库中未存在，则接收文件并做处理
		f, fh, err := this.GetFile("Cover")
		if err == nil {
			defer f.Close()
			slice := strings.Split(fh.Filename, ".")
			ext := slice[len(slice)-1]
			dir := fmt.Sprintf("./uploads/%v/%v/", time.Now().Format("2006-01-02"), this.IsLogin)
			os.MkdirAll(dir, 0777)
			file := helper.MyMD5(fmt.Sprintf("%v-%v-%v", timestamp, this.IsLogin, fh.Filename)) + "." + ext
			err = this.SaveToFile("Cover", dir+file)
			if err == nil {
				//将图片移动到OSS
				err = models.NewOss().MoveToOss(dir+file, file, true, true)
				helper.Logger.Debug(dir + file)
				if err != nil {
					helper.Logger.Error(err.Error())
				}
				cover = file
			}
		} else {
			helper.Logger.Error(err.Error())
		}
		var folder = models.CollectFolder{
			Uid:         this.IsLogin,
			Title:       this.GetString("Title"),
			Description: this.GetString("Description"),
			TimeCreate:  int(time.Now().Unix()),
			Cnt:         0,
			Cover:       cover,
		}
		folder.Id, _ = this.GetInt("Id")
		if folder.Id > 0 {
			cols := []string{"Title", "Description"}
			if len(cover) > 0 {
				cols = append(cols, "Cover")
			}
			_, err = orm.NewOrm().Update(&folder, cols...)
		} else {
			if _, err = orm.NewOrm().Insert(&folder); err == nil {
				//收藏夹数量+1
				models.Regulate(models.GetTableUserInfo(), "Collect", 1, "Id=?", this.IsLogin)
			}
		}

		if err != nil {
			helper.Logger.Error(err.Error())
			this.ResponseJson(0, "操作失败，请重试")
		}

		if folder.Id == 0 {
			models.Regulate(models.GetTableUserInfo(), "Collect", 1, fmt.Sprintf("Id=%v", this.IsLogin))
			this.ResponseJson(1, "收藏夹创建成功")
		} else {
			this.ResponseJson(1, "收藏夹编辑成功")
		}

	} else {
		this.ResponseJson(0, "您当前未登录，请先登录")
	}
}

//找回密码
func (this *UserController) FindPwd() {
	if this.IsLogin > 0 {
		this.Redirect("/user", 302)
		return
	}
	if this.Ctx.Request.Method == "POST" {
		rules := map[string][]string{
			"username":   {"required", "mincount:2", "maxcount:16"},
			"email":      {"required", "email"},
			"code":       {"required", "len:6"},
			"password":   {"required", "mincount:6"},
			"repassword": {"required", "mincount:6"},
		}

		params, errs := helper.Valid(this.Ctx.Request.Form, rules)
		fmt.Println(this.Ctx.Request.Form, params, errs)
		if len(errs) > 0 {
			if _, ok := errs["username"]; ok {
				this.ResponseJson(0, "用户名限2-16个字符")
			}
			if _, ok := errs["email"]; ok {
				this.ResponseJson(0, "邮箱格式不正确")
			}
			if _, ok := errs["code"]; ok {
				this.ResponseJson(0, "请输入6位验证码")
			}
			if _, ok := errs["password"]; ok {
				this.ResponseJson(0, "密码长度，至少6个字符")
			}
			if _, ok := errs["repassword"]; ok {
				this.ResponseJson(0, "密码长度，至少6个字符")
			}
		}
		fmt.Println(this.Ctx.Request.Form)
		//校验验证码和邮箱是否匹配
		if fmt.Sprintf("%v", this.GetSession("FindPwdMail")) != params["email"].(string) || fmt.Sprintf("%v", this.GetSession("FindPwdCode")) != params["code"].(string) {
			this.ResponseJson(0, "验证码不正确，修改密码失败")
		}
		pwd := helper.MyMD5(params["password"].(string))
		repwd := helper.MyMD5(params["repassword"].(string))
		if pwd != repwd {
			this.ResponseJson(0, "确认密码和密码不一致")
		}
		ModelUser := models.User{}
		user := ModelUser.GetUserField(orm.NewCondition().And("Email", params["email"]))
		if user.Id > 0 && user.Username == params["username"].(string) {
			_, err := models.UpdateByIds("user", "Password", pwd, user.Id)
			if err != nil {
				helper.Logger.Error(err.Error())
				this.ResponseJson(0, "重置密码失败，请刷新页面重试")
			}
			this.DelSession("FindPwdMail")
			this.DelSession("FindPwdCode")
			this.ResponseJson(1, "重置密码成功，请重新登录")
		} else {
			this.ResponseJson(0, "重置密码失败，用户名与邮箱不匹配")
		}
	} else {
		this.Data["Seo"] = models.NewSeo().GetByPage("PC-Findpwd", "找回密码", "找回密码", "找回密码", this.Sys.Site)
		this.Data["IsUser"] = true
		this.Data["PageId"] = "wenku-reg"
		this.TplName = "findpwd.html"
	}
}

//删除文档
func (this *UserController) DocDel() {
	docid, _ := this.GetInt(":doc")
	if this.IsLogin > 0 {
		if docid > 0 {
			errs := models.NewDocumentRecycle().RemoveToRecycle(this.IsLogin, true, docid)
			if len(errs) > 0 {
				helper.Logger.Error("删除失败：%v", strings.Join(errs, "; "))
				this.ResponseJson(0, "删除失败，文档不存在")
			} else {
				this.ResponseJson(1, "删除成功")
			}

			//var doc = models.DocumentInfo{Id: docid}
			//err := orm.NewOrm().Read(&doc)
			//if err != nil {
			//	helper.Logger.Error(err.Error())
			//	this.ResponseJson(0, "删除失败，文档不存在")
			//} else {
			//	if doc.Uid == this.IsLogin {
			//		doc.Status = -1
			//		_, err := orm.NewOrm().Update(&doc)
			//		if err == nil {
			//			//总文档数量-1
			//			models.SetDecr("sys", "CntDoc", "Id=1")
			//			//用户的文档数量-1
			//			models.SetDecr("user_info", "Document", fmt.Sprintf("Id=%v", this.IsLogin))
			//			//分类下的文档统计-1
			//			models.SetDecr("category", "Cnt", fmt.Sprintf("Id in(%v,%v,%v)", doc.Cid, doc.ChanelId, doc.Pid))
			//			this.ResponseJson(1, "文档删除成功")
			//		} else {
			//			helper.Logger.Error(err.Error())
			//			this.ResponseJson(0, "删除失败，请重试")
			//		}
			//
			//	} else {
			//		this.ResponseJson(0, "删除失败，文档不存在")
			//	}
			//}
		} else {
			this.ResponseJson(0, "删除失败，文档不存在")
		}
	} else {
		this.ResponseJson(0, "请先登录")
	}

}

//文档编辑
func (this *UserController) DocEdit() {
	docid, _ := this.GetInt(":doc")
	if this.IsLogin > 0 {
		if docid > 0 {
			var info = models.DocumentInfo{Id: docid}
			err := orm.NewOrm().Read(&info)
			if err != nil {
				helper.Logger.Error(err.Error())
				this.Redirect("/user", 302)
			} else {
				if info.Uid == this.IsLogin {
					var doc = models.Document{Id: docid}
					//post
					if this.Ctx.Request.Method == "POST" {
						ruels := map[string][]string{
							"Title":  {"required", "unempty"},
							"Chanel": {"required", "gt:0", "int"},
							"Pid":    {"required", "gt:0", "int"},
							"Cid":    {"required", "gt:0", "int"},
							"Tags":   {"required"},
							"Intro":  {"required"},
							"Price":  {"required", "int"},
						}
						params, errs := helper.Valid(this.Ctx.Request.Form, ruels)
						if len(errs) > 0 {
							this.ResponseJson(0, "参数错误")
						}
						doc.Title = params["Title"].(string)
						doc.Keywords = params["Tags"].(string)
						doc.Description = params["Intro"].(string)
						info.Pid = params["Pid"].(int)
						info.Cid = params["Cid"].(int)
						info.ChanelId = params["Chanel"].(int)
						info.Price = params["Price"].(int)
						orm.NewOrm().Update(&doc, "Title", "Keywords", "Description")
						orm.NewOrm().Update(&info, "Pid", "Cid", "ChanelId", "Price")
						//原分类-1
						models.Regulate(models.GetTableCategory(), "Cnt", -1, fmt.Sprintf("Id in(%v,%v,%v)", info.ChanelId, info.Cid, info.Pid))
						//新分类+1
						models.Regulate(models.GetTableCategory(), "Cnt", 1, fmt.Sprintf("Id in(%v,%v,%v)", params["Chanel"], params["Cid"], params["Pid"]))
						this.ResponseJson(1, "文档编辑成功")
					} else {

						err := orm.NewOrm().Read(&doc)
						if err != nil {
							helper.Logger.Error(err.Error())
							this.Redirect("/user", 302)
						}
						cond := orm.NewCondition().And("status", 1)
						data, _, _ := models.GetList(models.GetTableCategory(), 1, 2000, cond, "sort")
						this.Data["User"], _, _ = models.NewUser().GetById(this.IsLogin)
						ModelUser := models.User{}
						this.Data["Ranks"], _, err = ModelUser.UserList(1, 8, "i.Document desc", "u.Id,u.Username,u.Avatar,u.Intro,i.Document", "i.Status=1")
						//cates := models.ToTree(data, "Pid", 0)
						this.Data["IsUser"] = true
						this.Data["Cates"], _ = conv.InterfaceToJson(data)
						this.Data["json"] = data
						this.Data["PageId"] = "wenku-user"
						this.Data["Info"] = info
						this.Data["Doc"] = doc
						this.TplName = "edit.html"
					}
				} else {
					this.Redirect("/user", 302)
				}
			}
		} else {
			this.Redirect("/user", 302)
		}
	} else {
		this.Redirect("/user", 302)
	}
}

//删除收藏(针对收藏夹)
func (this *UserController) CollectFolderDel() {
	cid, _ := this.GetInt(":cid")
	if cid > 0 && this.IsLogin > 0 {
		err := models.NewCollect().DelFolder(cid, this.IsLogin)
		if err != nil {
			helper.Logger.Error(err.Error())
			this.ResponseJson(0, err.Error())
		} else {
			this.ResponseJson(1, "收藏夹删除成功")
		}
	} else {
		this.ResponseJson(0, "删除失败，参数错误")
	}
}

//取消收藏(针对文档)
func (this *UserController) CollectCancel() {
	cid, _ := this.GetInt(":cid")
	did, _ := this.GetInt(":did")
	if err := models.NewCollect().Cancel(did, cid, this.IsLogin); err != nil {
		helper.Logger.Error(err.Error())
		this.ResponseJson(0, "移除收藏失败，可能您为收藏该文档")
	}
	this.ResponseJson(1, "移除收藏成功")
}

//更换头像
func (this *UserController) Avatar() {
	if this.IsLogin > 0 {
		dir := fmt.Sprintf("./uploads/%v/%v", time.Now().Format("2006-01-02"), this.IsLogin)
		os.MkdirAll(dir, 0777)
		f, fh, err := this.GetFile("Avatar")
		if err != nil {
			helper.Logger.Error("用户(%v)更新头像失败：%v", this.IsLogin, err.Error())
			this.ResponseJson(0, "头像文件上传失败")
		}
		defer f.Close()
		slice := strings.Split(fh.Filename, ".")
		ext := strings.ToLower(slice[len(slice)-1])
		fmt.Println(ext)
		if !(ext == "jpg" || ext == "jpeg" || ext == "png" || ext == "gif") {
			this.ResponseJson(0, "头像图片格式只支持jpg、jpeg、png和gif")
		}
		tmpfile := dir + "/" + helper.MyMD5(fmt.Sprintf("%v-%v-%v", fh.Filename, this.IsLogin, time.Now().Unix())) + "." + ext
		savefile := helper.MyMD5(tmpfile) + "." + ext
		err = this.SaveToFile("Avatar", tmpfile)
		if err != nil {
			helper.Logger.Error("用户(%v)头像保存失败：%v", this.IsLogin, err.Error())
			this.ResponseJson(0, "头像文件保存失败")
		}
		err = models.NewOss().MoveToOss(tmpfile, savefile, true, true)
		if err != nil {
			helper.Logger.Error(err.Error())
			this.ResponseJson(0, "头像文件保存失败")
		}
		//查询数据库用户数据
		var user = models.User{Id: this.IsLogin}
		orm.NewOrm().Read(&user)
		if len(user.Avatar) > 0 {
			//删除原头像图片
			go models.NewOss().DelFromOss(true, user.Avatar)
		}
		user.Avatar = savefile
		rows, err := orm.NewOrm().Update(&user, "Avatar")
		if rows > 0 && err == nil {
			this.ResponseJson(1, "头像更新成功")
		}
		if err != nil {
			helper.Logger.Error(err.Error())
		}
		this.ResponseJson(0, "头像更新失败")
	} else {
		this.ResponseJson(0, "请先登录")
	}
}

//编辑个人信息
func (this *UserController) Edit() {
	if this.IsLogin > 0 {
		changepwd := false
		cols := []string{"Intro"}
		rules := map[string][]string{
			"OldPassword": {"required"},
			"NewPassword": {"required"},
			"RePassword":  {"required"},
			"Intro":       {"required"},
		}
		params, errs := helper.Valid(this.Ctx.Request.Form, rules)
		if len(errs) > 0 {
			this.ResponseJson(0, "参数不正确")
		}
		var user = models.User{Id: this.IsLogin}
		orm.NewOrm().Read(&user)
		if len(params["OldPassword"].(string)) > 0 || len(params["NewPassword"].(string)) > 0 || len(params["RePassword"].(string)) > 0 {
			if len(params["NewPassword"].(string)) < 6 || len(params["RePassword"].(string)) < 6 {
				this.ResponseJson(0, "密码长度必须至少6个字符")
			}
			opwd := helper.MyMD5(params["OldPassword"].(string))
			npwd := helper.MyMD5(params["NewPassword"].(string))
			rpwd := helper.MyMD5(params["RePassword"].(string))
			if user.Password != opwd {
				this.ResponseJson(0, "原密码不正确")
			}
			if npwd != rpwd {
				this.ResponseJson(0, "确认密码和新密码必须一致")
			}
			if opwd == npwd {
				this.ResponseJson(0, "确认密码不能与原密码相同")
			}
			user.Password = rpwd
			cols = append(cols, "Password")
			changepwd = true
		}
		user.Intro = params["Intro"].(string)
		rows, err := orm.NewOrm().Update(&user, cols...)
		if err != nil {
			helper.Logger.Error(err.Error())
			this.ResponseJson(0, "设置失败，请刷新页面重试")
		}
		if rows > 0 {
			if changepwd {
				this.ResetCookie()
				this.ResponseJson(1, "设置成功，您设置了新密码，请重新登录")
			} else {
				this.ResponseJson(1, "设置成功")
			}
		} else {
			this.ResponseJson(1, "设置失败，可能您未对内容做更改")
		}
	} else {
		this.ResponseJson(0, "请先登录")
	}

}
