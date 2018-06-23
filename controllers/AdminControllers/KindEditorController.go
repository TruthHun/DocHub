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
	//imgFile
	//允许上传的文件的扩展名
	//AllowedExt := map[string][]string{
	//	"image": {"gif", "jpg", "jpeg", "png", "bmp"},
	//	"flash": {"swf", "flv"},
	//	"media": {"swf", "flv", "mp3", "wav", "wma", "wmv", "mid", "avi", "mpg", "asf", "rm", "rmvb"},
	//	"file":  {"doc", "docx", "xls", "xlsx", "ppt", "htm", "html", "txt", "zip", "rar", "gz", "bz2"},
	//}
	//文件在文档库中未存在，则接收文件并做处理
	f, fh, err := this.GetFile("imgFile")
	if err != nil {
		this.ResponseJson(0, err.Error())
	}
	defer f.Close()
	now := time.Now()
	dir := fmt.Sprintf("uploads/kindeditor/%v", now.Format("2006/01/02"))
	os.MkdirAll(dir, 0777)
	ext := helper.GetSuffix(fh.Filename, ".")
	ossfile := "article." + helper.MyMD5(fmt.Sprintf("%v-%v-%v", now, fh.Filename, this.AdminId)) + "." + ext
	//存储文件
	savefile := dir + "/" + ossfile
	err = this.SaveToFile("imgFile", savefile)
	if err != nil {
		this.Response(map[string]interface{}{"message": err.Error(), "error": 1})
	} else {
		//将文件上传到OSS
		err = models.ModelOss.MoveToOss(savefile, ossfile, true, true)
		if err == nil {
			this.Response(map[string]interface{}{"url": models.ModelOss.PreviewUrl + ossfile, "error": 0})
		} else {
			this.Response(map[string]interface{}{"message": err.Error(), "error": 1})
		}
	}
}
