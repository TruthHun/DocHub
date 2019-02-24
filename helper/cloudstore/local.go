package CloudStore

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// 本地存储
type Local struct {
	Secret    string
	Folder    string
	Domain    string
	headerExt string
}

// 新建本次存储的文件夹
func NewLocal(folder, domain, secret string) (local *Local, err error) {
	err = os.MkdirAll(folder, os.ModePerm)
	local = &Local{
		Folder:    folder,
		Domain:    domain,
		Secret:    secret,
		headerExt: ".header.json",
	}
	return
}

func (l *Local) PutObject(local, object string, header map[string]string) (err error) {

	if err = os.MkdirAll(filepath.Dir(object), os.ModePerm); err != nil {
		return
	}

	if err = os.Rename(local, object); err != nil {
		return
	}

	var b []byte
	if b, err = json.Marshal(header); err == nil {
		ioutil.WriteFile(object+l.headerExt, b, os.ModePerm)
	}

	return
}

func (l *Local) DeleteObjects(objects []string) (err error) {
	var errSlice []string
	for _, obj := range objects {
		if err = os.Remove(obj); err != nil {
			errSlice = append(errSlice, err.Error())
		}
	}

	if len(errSlice) > 0 {
		err = errors.New(strings.Join(errSlice, "; "))
	}

	return
}

func (l *Local) GetObjectURL(object string, expire int64) (urlStr string, err error) {
	// TODO: 获取链接

	return
}

func (l *Local) CheckObjectURLAccess(obj string, expire int64) {

}
