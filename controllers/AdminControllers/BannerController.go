package AdminControllers

import (
	"fmt"

	"time"

	"os"

	"strings"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
	"github.com/astaxie/beego/orm"
)

//IT文库注册会员管理

type BannerController struct {
	BaseController
}

//横幅列表
func (this *BannerController) Get() {
	var err error
	if this.Data["Banners"], _, err = models.NewBanner().List(1, 100); err != nil && err != orm.ErrNoRows {
		helper.Logger.Error(err.Error())
	}
	this.Data["IsBanner"] = true
	this.TplName = "index.html"
}

//新增横幅
func (this *BannerController) Add() {
	f, h, err := this.GetFile("Picture")
	if err != nil {
		helper.Logger.Error(err.Error())
		this.ResponseJson(false, err.Error())
	}
	defer f.Close()

	dir := "uploads/" + time.Now().Format("2006/01/02")
	os.MkdirAll(dir, 0777)
	ext := helper.GetSuffix(h.Filename, ".")
	filePath := dir + "/" + helper.MD5Crypt(fmt.Sprintf("%v-%v", h.Filename, time.Now().Unix())) + "." + ext

	err = this.SaveToFile("Picture", filePath) // 保存位置
	if err != nil {
		helper.Logger.Error(err.Error())
		this.ResponseJson(false, err.Error())
	}
	defer os.RemoveAll(filePath)

	if err = helper.CropImage(filePath, helper.BannerWidth, helper.BannerHeight); err != nil {
		helper.Logger.Error("横幅裁剪失败：%v", err.Error())
		this.ResponseJson(false, err.Error())
	}

	var md5str string
	md5str, err = helper.FileMd5(filePath)
	if err != nil {
		helper.Logger.Error(err.Error())
		this.ResponseJson(false, err.Error())
	}

	save := md5str + "." + ext

	var cs *models.CloudStore
	if cs, err = models.NewCloudStore(false); err != nil {
		helper.Logger.Error(err.Error())
		this.ResponseJson(false, err.Error())
	}

	if err = cs.Upload(filePath, save); err != nil {
		helper.Logger.Error(err.Error())
		this.ResponseJson(false, err.Error())
	}

	var banner models.Banner
	this.ParseForm(&banner)
	banner.Picture = save
	banner.TimeCreate = int(time.Now().Unix())
	banner.Status = true
	_, err = orm.NewOrm().Insert(&banner)
	if err != nil {
		helper.Logger.Error(err.Error())
		this.ResponseJson(false, err.Error())
	}

	this.ResponseJson(true, "横幅添加成功")
}

//删除横幅
func (this *BannerController) Del() {
	var err error
	id := this.GetString("id")
	ids := strings.Split(id, ",")
	if len(ids) > 0 {
		//之所以这么做，是因为如果没有第一个参数，则参数编程了[]string，而不是[]interface{},有疑问可以自己验证试下
		if _, err = models.NewBanner().Del(ids[0], ids[1:]); err != nil {
			helper.Logger.Error(err.Error())
			this.ResponseJson(false, err.Error())
		}
	}
	this.ResponseJson(true, "删除成功")
}
