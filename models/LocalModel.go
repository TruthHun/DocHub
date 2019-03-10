package models

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/TruthHun/DocHub/helper"
)

type LocalStore struct {
	StoreDir  string // 存储路径
	HeaderExt string
}

// 新建本地存储
func NewLocalStore() (store *LocalStore, err error) {
	storeDir := "./store"
	if _, err = os.Stat(storeDir); err != nil {
		err = os.MkdirAll(storeDir, os.ModePerm)
		if err != nil {
			return
		}
	}
	store = &LocalStore{
		StoreDir:  storeDir,
		HeaderExt: ".header.json",
	}
	return
}

// 判断文件是否存在
// @param       object      文件是否存在
func (l *LocalStore) IsExist(object string) (err error) {
	_, err = os.Stat(l.RealPath(object))
	return
}

//文件移动到本地存储
//@param            local            本地文件
//@param            save             存储到OSS的文件
//@param            IsPreview        是否是预览，是预览的话，则上传到预览的OSS，否则上传到存储的OSS。存储的OSS，只作为文档的存储，以供下载，但不提供预览等访问，为私有
//@param            IsDel            文件上传后，是否删除本地文件
//@param            IsGzip           是否做gzip压缩，做gzip压缩的话，需要修改oss中对象的响应头，设置gzip响应
func (l *LocalStore) MoveToStore(local, save string, IsPreview, IsDel bool, IsGzip ...bool) (err error) {
	var bs []byte
	save = l.RealPath(save)

	isGzip := false //如果是开启了gzip，则需要设置文件对象的响应头
	if len(IsGzip) > 0 && IsGzip[0] == true {
		isGzip = true
	}

	if strings.ToLower(filepath.Ext(local)) == ".svg" && helper.GetConfigBool("depend", "svgo-on") {
		if err = helper.CompressBySVGO(local, local); err != nil {
			return
		}
	}

	if isGzip {
		if bs, err = ioutil.ReadFile(local); err != nil {
			helper.Logger.Error(err.Error())
			isGzip = false //设置为false
		} else {
			var by bytes.Buffer
			w := gzip.NewWriter(&by)
			defer w.Close()
			w.Write(bs)
			w.Flush()
			ioutil.WriteFile(local, by.Bytes(), 0777)
		}
		defer func() {
			if err == nil {
				headerFile := save + l.HeaderExt
				headerContent := `{"content-encoding": "gzip"}`
				ioutil.WriteFile(headerFile, []byte(headerContent), os.ModePerm)
			}
		}()
	}
	return os.Rename(local, save)
}

// 获取存储路径(dir)
// @param           object          文件对象。
func (l *LocalStore) getStoreDir(object string) (dir string) {
	dir = "./store/error"
	slice := strings.Split(object, "")
	if len(slice) > 5 {
		dir = filepath.Join(l.StoreDir, strings.Join(slice[:5], "/"))
	}
	return
}

// 获取文件的真实存储路径
func (l *LocalStore) RealPath(object string) (path string) {
	return filepath.Join(l.getStoreDir(object), object)
}
