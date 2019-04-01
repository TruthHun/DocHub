package AdminControllers

import (
	"fmt"
	"os"
	"time"

	"github.com/TruthHun/DocHub/helper"
	"github.com/TruthHun/DocHub/models"
)

type KindEditorController struct {
	BaseController
}

//上传。这里是后台使用的，不限文件类型
func (this *KindEditorController) Upload() {
	f, fh, err := this.GetFile("imgFile")
	if err != nil {
		this.ResponseJson(false, err.Error())
	}
	defer f.Close()
	now := time.Now()
	dir := fmt.Sprintf("uploads/kindeditor/%v", now.Format("2006/01/02"))
	os.MkdirAll(dir, 0777)
	ext := helper.GetSuffix(fh.Filename, ".")
	filename := "article." + helper.MD5Crypt(fmt.Sprintf("%v-%v-%v", now, fh.Filename, this.AdminId)) + "." + ext
	//存储文件
	tmpFile := dir + "/" + filename
	err = this.SaveToFile("imgFile", tmpFile)
	if err != nil {
		this.Response(map[string]interface{}{"message": err.Error(), "error": 1})
	}
	defer os.RemoveAll(tmpFile)

	var cs *models.CloudStore
	if cs, err = models.NewCloudStore(false); err != nil {
		this.Response(map[string]interface{}{"message": err.Error(), "error": 1})
	}

	//将文件上传到OSS
	err = cs.Upload(tmpFile, filename)
	if err == nil {
		this.Response(map[string]interface{}{"url": cs.GetSignURL(filename), "error": 0})
	}
	this.Response(map[string]interface{}{"message": err.Error(), "error": 1})
}
