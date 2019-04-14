package helper

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"html/template"
	"image"
	"math/rand"

	"crypto/md5"

	"time"

	"strconv"

	"os"

	"io"
	"strings"

	"crypto/sha1"

	"net/url"

	"os/exec"

	"io/ioutil"

	"regexp"

	"errors"

	"net/http"

	"path/filepath"

	"github.com/PuerkitoBio/goquery"
	"github.com/TruthHun/DocHub/helper/crawl"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/disintegration/imaging"
	"github.com/huichen/sego"
	"rsc.io/pdf"
)

func init() {
	//如果存在配置文件，则表示程序已经安装
	_, err := os.Stat("conf/app.conf")
	if err == nil {
		IsInstalled = true
	}
	setDefaultConfig()
	exts := strings.Split(beego.AppConfig.DefaultString("StaticExt", DEFAULT_STATIC_EXT), ",")
	for _, ext := range exts {
		StaticExt[strings.ToLower(strings.TrimSpace(ext))] = true
	}
	if _, err = os.Stat(RootPath); err != nil {
		err = os.MkdirAll(RootPath, os.ModePerm)
		if err != nil {
			Logger.Error(err.Error())
			panic(err.Error())
		}
	}
}

//比较两个内容的字符串类型是否相等
func Equal(itf1, itf2 interface{}) bool {
	return fmt.Sprintf("%v", itf1) == fmt.Sprintf("%v", itf2)
}

//语言国际化，目前默认为中文
func I18n(tag string, lang ...string) string {
	if len(lang) == 0 {
		lang[0] = "zh-CN"
	}
	return beego.AppConfig.DefaultString(fmt.Sprintf("%v::%v", lang[0], tag), fmt.Sprintf("{%v}", tag))
}

//MD5加密函数
//@str          string          需要进行加密的字符串
func MD5Crypt(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

//时间戳格式化
//@timestamp            int             时间戳
//@format               string          时间格式，如："2006-01-02 15:04:05"
func TimestampFormat(timestamp int, format ...string) string {
	formats := "2006-01-02 15:04:05"
	if len(format) > 0 {
		formats = format[0]
	}
	return time.Unix(int64(timestamp), 0).Format(formats)
}

//将字符串或者其他可转化数字的内容格式化成int数字类型返回
//@param        a            interface{}         需要转化成数字的内容
func Interface2Int(a interface{}) int {
	i, _ := strconv.Atoi(fmt.Sprintf("%v", a))
	return i
}

//生成url链接
//@param            prefix              前缀
//@param            params              参数
//@return           urlstr              返回组装后的url
func BuildURL(prefix string, params ...interface{}) (urlstr string) {
	var (
		l   int
		url []string
	)
	url = append(url, "/"+strings.Trim(prefix, "/"))
	l = len(params)
	if l != (l/2)*2 {
		l = (l / 2) * 2
	}
	if l > 0 {
		for i := 0; i < l; {
			k := fmt.Sprintf("%v", params[i])
			v := fmt.Sprintf("%v", params[i+1])
			if len(k) > 0 && v != "0" {
				url = append(url, fmt.Sprintf("%v/%v", k, v))
			}
			i += 2
		}
	}
	urlstr = strings.TrimRight(strings.Join(url, "/"), "/")
	//fmt.Println("BuildURL:", prefix, params, len(params), url)
	return
}

//图片文件信息
type Info struct {
	Width, Height int    //图片宽高
	Size          int64  //图片文件大小
	Md5           string //图片md5
	ModTime       int    //图片修改时间戳
	Ext           string //图片后缀
}

//判断文件路径判断文件是否是符合要求的图片格式，jpeg,jpg,gif,png,bmp,tif,tiff
//@path         string          文件路径
func IsImage(path string) bool {
	slice := strings.Split(path, ".")
	ext := strings.ToLower(strings.TrimSpace(slice[len(slice)-1]))
	exts := map[string]string{"jpeg": "jpeg", "jpg": "jpg", "gif": "gif", "png": "png", "bmp": "bmp", "tif": "tif", "tiff": "tiff"}
	_, ok := exts[ext]
	return ok
}

//获取文件的MD5
//@path         string          文件路径
func FileMd5(path string) (string, error) {
	var md5str string
	var err error
	var file *os.File
	file, err = os.Open(path)
	defer file.Close()
	if err != nil {
		return md5str, err
	}
	md5h := md5.New()
	_, err = io.Copy(md5h, file)
	if err == nil {
		md5str = fmt.Sprintf("%x", md5h.Sum(nil))
	}
	return md5str, err
}

//根据字符串后缀，比如根据文件路径获取后缀等
//@str          string          字符串或者文件路径
//@seg          string          分隔符
func GetSuffix(str, seg string) string {
	slice := strings.Split(str, seg)
	l := len(slice)
	if l > 1 {
		return slice[(l - 1)]
	}
	return ""
}

//将interface{}转为字符串，适合bool，数字等
//@v            interface{}         需要转化为字符串的值
func Interface2String(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

//整型计算
func CalcInt(num1 interface{}, oper string, num2 interface{}) int {
	val := 0
	n1 := Interface2Int(num1)
	n2 := Interface2Int(num2)

	switch oper {
	case "+":
		val = n1 + n2
	case "-":
		val = n1 - n2
	case "*":
		val = n1 * n2
	case "%":
		if n2 != 0 {
			val = n1 % n2
		}
	case "/":
		if n2 != 0 {
			val = n1 / n2
		}
	}
	return val
}

//根据切片或数组下表，移除元素
func DelSlice(slice []interface{}, index int) []interface{} {
	return append(slice[:index], slice[index+1:]...)
}

//sha1加密
//@param            str         string          需要加密的字符串
//@return            string      加密后的字符串
func MySha1(str string) string {
	return fmt.Sprintf("%x", sha1.New().Sum([]byte(str)))
}

//设置缓存
func CacheSet(key string, value interface{}, t time.Duration) error {
	bc, err := cache.NewCache("file", CACHE_CONF)
	if err == nil {
		return bc.Put(key, value, t)
	}
	return err
}

//获取缓存
func CacheGet(key string) (interface{}, error) {
	bc, err := cache.NewCache("file", CACHE_CONF)
	if err == nil && bc.IsExist(key) {
		return bc.Get(key), nil
	}
	return nil, err
}

//生成默认值
//@param            val         interface{}         值
//@param            defVal      interface{}         默认值
//@param            preventZero ...bool             是否保留0值
func Default(val, defVal interface{}, preventZero ...bool) string {
	v := fmt.Sprintf("%v", val)
	dv := fmt.Sprintf("%v", defVal)
	if v == "" {
		return dv
	}
	//保留0值
	if len(preventZero) > 0 {
		if v != "0" {
			return dv
		}
	}
	return v
}

//获取指定个数的随机字符串
//@param            size            int         指定的字符串个数
//@param            kind            int         0，纯数字；1，小写字母；2，大写字母；3，数字+大小写字母
//@return           string                      返回生成的随机字符串
func RandStr(size int, kind int) string {
	ikind, kinds, result := kind, [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}, make([]byte, size)
	isAll := kind > 2 || kind < 0
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		if isAll { // random ikind
			ikind = rand.Intn(3)
		}
		scope, base := kinds[ikind][0], kinds[ikind][1]
		result[i] = uint8(base + rand.Intn(scope))
	}
	return string(result)
}

//转换字节大小
func FormatByte(size int) string {
	fsize := float64(size)
	//字节单位
	units := [6]string{"B", "KB", "MB", "GB", "TB", "PB"}
	var i int
	for i = 0; fsize >= 1024 && i < 5; i++ {
		fsize /= 1024
	}

	num := fmt.Sprintf("%.2f", fsize)

	return string(num) + " " + units[i]
}

//获取文档扩展名分类以及扩展名分类对应的catenum数字
//@param            ext         扩展名
//@return           extCate     文档类型(字符串)
//@return           extNum      文档类型(整型)
func GetExtCate(ext string) (extCate string, extNum int) {
	ext = strings.ToLower(strings.TrimLeft(ext, "."))
	extNum = EXT_NUM_OTHER
	switch ext {
	case "doc", "docx", "rtf", "wps", "odt":
		extCate = EXT_CATE_WORD
		extNum = EXT_NUM_WORD
	case "ppt", "pptx", "pps", "ppsx", "dps", "odp", "pot":
		extCate = EXT_CATE_PPT
		extNum = EXT_NUM_PPT
	case "xls", "xlsx", "et", "ods":
		extCate = EXT_CATE_EXCEL
		extNum = EXT_NUM_EXCEL
	case "pdf":
		extCate = EXT_CATE_PDF
		extNum = EXT_NUM_PDF
	case "txt":
		extCate = EXT_CATE_TEXT
		extNum = EXT_NUM_TEXT
	case EXT_CATE_OTHER_UMD, EXT_CATE_OTHER_CHM, EXT_CATE_OTHER_EPUB, EXT_CATE_OTHER_MOBI: // cate other
		extCate = ext
	}
	return
}

//获取文档扩展名分类以及扩展名分类对应的catenum数字
//@param            ext         扩展名
//@return           extCate     文档类型(字符串)
//@return           extNum      文档类型(整型)
func GetExtCateByExtNum(num int) (extCate string) {
	switch num {
	case EXT_NUM_WORD:
		extCate = EXT_CATE_WORD
	case EXT_NUM_PPT:
		extCate = EXT_CATE_PPT
	case EXT_NUM_EXCEL:
		extCate = EXT_CATE_EXCEL
	case EXT_NUM_PDF:
		extCate = EXT_CATE_PDF
	case EXT_NUM_TEXT:
		extCate = EXT_CATE_TEXT
	case EXT_NUM_OTHER:
		extCate = EXT_CATE_OTHER
	}
	return
}

//PdfProcedure专用
type PdfRet struct {
	Cover   string //封面路径文件
	Preview string //
	Total   int
	Err     error
}

//获取pdf文件的页数
//@param            file            pdf文件
//@return           pages           pdf文件页码
//@return           err             错误
func getPdfPagesNum(file string) (pages int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
			return
		}
	}()
	if reader, err := pdf.Open(file); err == nil {
		pages = reader.NumPage()
	}
	return
}

//解析svg的原始宽高
//@param            file            svg文件
func ParseSvgWidthAndHeight(file string) (width, height int) {
	if bs, err := ioutil.ReadFile(file); err == nil {
		if doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bs))); err == nil {
			sel := doc.Find("svg")
			if w, ok := sel.Attr("width"); ok {
				if slice := strings.Split(strings.TrimRight(w, "pt"), "."); len(slice) > 0 {
					width = Interface2Int(slice[0])
				}
			}
			if h, ok := sel.Attr("height"); ok {
				if slice := strings.Split(strings.TrimRight(h, "pt"), "."); len(slice) > 0 {
					height = Interface2Int(slice[0])
				}
			}
		}
	} else {
		Logger.Error(err.Error())
	}
	return
}

//解析svg的原始宽高(TODO:水印效果不是很好，待优化)
//@param            file            svg文件
//@param			text			水印文字
//@param			x				x轴线位置
//@param			y				y轴位置
//@return			err				错误
func SvgTextWatermark(file, text string, x, y int) (err error) {
	if Debug {
		beego.Debug("svg添加水印", file, text, "位置", x, y)
	}
	if text != "" {
		var b []byte
		watermark := []string{}
		watermark = append(watermark, fmt.Sprintf(`<text x="%v" y="%v" style="fill:rgba(0,0,0,0.2)" transform="scale(2)">%v</text>`, x, y, text))
		if b, err = ioutil.ReadFile(file); err == nil {
			str := string(b)
			str = strings.Replace(str, "</svg>", strings.Join(watermark, "")+"</svg>", -1)
			err = ioutil.WriteFile(file, []byte(str), os.ModePerm)
		}
	}
	return
}

//扫描目录目录中的文件
func ScanDir(dir string) (files []string) {
	dir = strings.TrimSuffix(dir, "/")
	if infos, err := ioutil.ReadDir(dir); err == nil {
		for _, info := range infos {
			file := dir + "/" + info.Name()
			if info.IsDir() {
				files = append(files, ScanDir(file)...)
			} else {
				files = append(files, file)
			}
		}
	} else {
		Logger.Error(err.Error())
	}
	return
}

//统计PDF的页数
//@param            file                文件路径
//@return           pagenum             页码，当返回错误时，页码为0
//@return           err                 错误
func CountPDFPages(file string) (pageNum int, err error) {
	// 优先使用这种PDF分页统计的方式，如果PDF版本比支持，则再使用下一种PDF统计方式。所以这里报错了的话，不返回
	pageNum, err = getPdfPagesNum(file)
	if err == nil {
		return
	}

	var p []byte
	p, err = ioutil.ReadFile(file)
	if err != nil {
		return
	}

	content := string(p)
	arr := strings.Split(content, "/Pages")
	l := len(arr)
	if l > 0 {
		arr = strings.Split(arr[l-1], "endobj")
		if l = len(arr); l > 0 {
			return len(strings.Split(arr[0], "0 R")) - 1, nil
		}
		return 0, errors.New(fmt.Sprintf(`%v:"endobj"分割时失败`, file))
	}
	return 0, errors.New(fmt.Sprintf(`%v:"/Pages"分割时失败`, file))
}

//文档评分处理
//score：文档评分
//IsScore:是否返回处理后的评分，true：返回评分，如“4.5”，false返回css的星，如"45"
func StarVal(score int, IsScore bool) string {
	if IsScore {
		return fmt.Sprintf("%1.1f", float32(score)/10000.0)
	}
	return fmt.Sprintf("%d", score/1000)
}

//分页函数
//rollPage:展示分页的个数
//totalRows：总记录
//currentPage:每页显示记录数
//urlPrefix:url链接前缀
//urlParams:url键值对参数
func Paginations(rollPage, totalRows, listRows, currentPage int, urlPrefix string, urlParams ...interface{}) string {
	var (
		htmlPage, path string
		pages          []int
		params         []string
	)
	if listRows <= 0 {
		listRows = 10
	}
	//总页数
	totalPage := totalRows / listRows
	if totalRows%listRows > 0 {
		totalPage += 1
	}
	//只有1页的时候，不分页
	if totalPage < 2 {
		return ""
	}
	params_len := len(urlParams)
	if params_len > 0 {
		if params_len%2 > 0 {
			params_len = params_len - 1
		}
		for i := 0; i < params_len; {
			key := strings.TrimSpace(fmt.Sprintf("%v", urlParams[i]))
			val := strings.TrimSpace(fmt.Sprintf("%v", urlParams[i+1]))
			//键存在，同时值不为0也不为空
			if len(key) > 0 && len(val) > 0 && val != "0" {
				params = append(params, key, val)
			}
			i = i + 2
		}
	}

	path = strings.Trim(urlPrefix, "/")
	if len(params) > 0 {
		path = path + "/" + strings.Trim(strings.Join(params, "/"), "/")
	}
	//最后再处理一次“/”，是为了防止urlPrifix参数为空时，出现多余的“/”
	path = "/" + strings.Trim(path, "/") + "/p/"

	if currentPage > totalPage {
		currentPage = totalPage
	}
	if currentPage < 1 {
		currentPage = 1
	}
	index := 0
	rp := rollPage * 2
	for i := rp; i > 0; i-- {
		p := currentPage + rollPage - i
		if p > 0 && p <= totalPage {

			pages = append(pages, p)
		}
	}
	for k, v := range pages {
		if v == currentPage {
			index = k
		}
	}
	pages_len := len(pages)
	if currentPage > 1 {
		htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`1">1..</a></li><li><a class="num" href="`+path+`%d"><<</a></li>`, currentPage-1)
	}
	if pages_len <= rollPage {
		for _, v := range pages {
			if v == currentPage {
				htmlPage += fmt.Sprintf(`<li class="active"><a href="javascript:void(0);">%d</a></li>`, v)
			} else {
				htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`%d">%d</a></li>`, v, v)
			}
		}

	} else {
		index_min := index - rollPage/2
		index_max := index + rollPage/2
		page_slice := make([]int, 0)
		if index_min > 0 && index_max < pages_len { //切片索引未越界
			page_slice = pages[index_min:index_max]
		} else {
			if index_min < 0 {
				page_slice = pages[0:rollPage]
			} else if index_max > pages_len {
				page_slice = pages[(pages_len - rollPage):pages_len]
			} else {
				page_slice = pages[index_min:index_max]
			}

		}

		for _, v := range page_slice {
			if v == currentPage {
				htmlPage += fmt.Sprintf(`<li class="active"><a href="javascript:void(0);">%d</a></li>`, v)
			} else {
				htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`%d">%d</a></li>`, v, v)
			}
		}

	}
	if currentPage < totalPage {
		htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`%d">>></a></li><li><a class="num" href="`+path+`%d">..%d</a></li>`, currentPage+1, totalPage, totalPage)
	}
	return htmlPage
}

//将字符串切片转成interface切片
func StringSliceToInterfaceSlice(slice []string) (ret []interface{}) {
	for _, v := range slice {
		ret = append(ret, v)
	}
	return ret
}

//将字符串切片数组转成map
func StringSliceToMap(slice []string) (maps map[string]string) {
	maps = make(map[string]string)
	for _, v := range slice {
		maps[v] = v
	}
	return maps
}

//分词
//@param            str         需要分词的文字
func SegWord(str interface{}) (wds string) {
	//如果已经成功加载字典
	if Segmenter.Dictionary() != nil {
		wds = sego.SegmentsToString(Segmenter.Segment([]byte(fmt.Sprintf("%v", str))), true)
		var wdslice []string
		slice := strings.Split(wds, " ")
		for _, wd := range slice {
			w := strings.Split(wd, "/")[0]
			if (strings.Count(w, "") - 1) >= 2 {
				if Interface2Int(w) == 0 { //如果为0，则表示非数字
					wdslice = append(wdslice, w)
				}
			}
		}
		wds = strings.Join(wdslice, ",")
	}
	return
}

//url请求处理
func UrlEscape(str string) string {
	return strings.TrimSpace(url.QueryEscape(strings.Replace(str, "/", " ", -1)))
}

//高亮，主要用于搜索结果高亮
func HeightLight(title string, words []string) template.HTML {
	//注意：不能直接高亮，比如title里面有span标签作为关键字的时候就出现问题了
	for _, wd := range words {
		r, _ := regexp.Compile(fmt.Sprintf("(?i:%v)", wd))
		title = r.ReplaceAllString(title, fmt.Sprintf("<wenku.it>%v</wenku.it>", wd))
	}
	for _, wd := range words {
		r, _ := regexp.Compile(fmt.Sprintf("(?i:%v)", fmt.Sprintf("<wenku.it>%v</wenku.it>", wd)))
		title = r.ReplaceAllString(title, fmt.Sprintf(`<span class="text-danger">%v</span>`, wd))
	}
	return template.HTML(title)
}

//字符串截取
func SubStr(str interface{}, start, length int) string {
	v := fmt.Sprintf("%v", str)
	if start < 0 {
		start = 0
	}
	slice := strings.Split(v, "")
	l := len(slice)
	if l == 0 || start > l {
		return ""
	}
	if start+length+1 > l {
		return strings.Join(slice[start:], "")
	}
	return strings.Join(slice[start:length], "")
}

//文件下载，这里只是简单的文档下载，复杂点的，需要根据Content-Type去做处理
//@param            fileUrl             文件下载URL连接
//@param            savePath            文件存放文件夹
//@param            cookies             COOKIE
//@return           md5str              文档的md5
//@return           localFile           本地存放文件，err为nil时才需要处理文件，否则不需要处理
//@return           filename            文件下载时的文件名
//@return           err                 错误
func DownFile(fileUrl, savePath string, cookies string) (md5str, localFile, filename string, err error) {
	var (
		resp *http.Response //响应内容
		ext  string         //文件扩展名
	)
	//创建HTTP请求
	req := crawl.BuildRequest("get", fileUrl, "", cookies, "mac", true, false)
	resp, err = req.DoRequest()
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		err = errors.New(fmt.Sprintf("HTTP响应头错误：%v。文件下载地址：%v", resp.Status, fileUrl))
		return
	}
	if attachmentArr := strings.Split(resp.Header.Get("Content-Disposition"), "="); len(attachmentArr) == 2 {
		filename = strings.Trim(attachmentArr[1], `"'`)
		ext = filepath.Ext(filename)
	}
	if len(ext) == 0 {
		filepath.Ext(fileUrl)
	}
	ext = strings.ToLower(ext)
	os.MkdirAll(savePath, 0777)
	tmpFile := strings.TrimSuffix(savePath, "/") + "/" + MD5Crypt(filename) + ext
	if err = req.ToFile(tmpFile); err != nil {
		return
	}
	//计算文档md5
	if md5str, err = FileMd5(tmpFile); err != nil {
		return
	}
	//重命名文档
	localFile = strings.TrimSuffix(savePath, "/") + "/" + md5str + ext
	err = os.Rename(tmpFile, localFile) //重命名文件
	return
}

//
////给SVG加文字水印
////@param			svgfile			svg文件
////@param			text			水印文字
////@return			err				nil时表示添加水印成功
//func WatermarkByText(svgfile, text string) (err error) {
//	return
//}

//数字变化范围
//@param            val             int         需要处理的参数数字
//@param            min             int         最小值
//@param            max             int         最大值
//@return                           int         返回最大值与最小值之间的值
func NumberRange(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

//首字母大写
func UpperFirst(str string) string {
	if len(str) > 0 {
		strings.Replace(str, str[0:1], strings.ToUpper(str[0:1]), 1)
	}
	return str
}

//页数处理，处理页数为0或者页数为空的时候的显示
func HandlePageNum(PageNum interface{}) string {
	pn := strings.TrimSpace(fmt.Sprintf("%v", PageNum))
	if pn == "0" || pn == "" {
		return " -- "
	}
	return pn
}

//使用SVGO压缩svg文件
//@param            input           需要压缩的原文件
//@param            output          压缩后的文件路径
//@param            err             压缩错误
func CompressBySVGO(input, output string) (err error) {
	svgo := strings.TrimSpace(GetConfig("depend", "svgo", "svgo"))
	args := []string{"-i", input, "-o", output}
	if strings.HasPrefix(svgo, "sudo") {
		args = append([]string{strings.TrimPrefix(svgo, "sudo")}, args...)
		svgo = "sudo"
	}
	err = exec.Command(svgo, args...).Run()
	if err != nil {
		err = CompressSVG(input, output)
	}
	return
}

// 使用GZIP压缩文件
func CompressByGzip(file string) (err error) {
	var bs []byte
	bs, err = ioutil.ReadFile(file)
	if err != nil {
		return
	}
	var by bytes.Buffer
	w := gzip.NewWriter(&by)
	defer w.Close()
	w.Write(bs)
	w.Flush()
	err = ioutil.WriteFile(file, by.Bytes(), 0777)
	return
}

// 图片缩放居中裁剪
//图片缩放居中裁剪
//@param        file        图片文件
//@param        width       图片宽度
//@param        height      图片高度
//@return       err         错误
func CropImage(file string, width, height int) (err error) {
	var img image.Image
	img, err = imaging.Open(file)
	if err != nil {
		return
	}
	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ".jpeg", ".jpg", ".png", ".gif":
		img = imaging.Fill(img, width, height, imaging.Center, imaging.CatmullRom)
	default:
		err = errors.New("unsupported image format")
		return
	}
	return imaging.Save(img, file)
}

// 删除切片中的指定key，并返回
func DeleteSlice(slice []string, keys ...string) []string {
	if len(keys) == 0 {
		return slice
	}

	for _, key := range keys {
		var tmpSlice []string
		for _, item := range slice {
			if item != key {
				tmpSlice = append(tmpSlice, item)
			}
		}
		slice = tmpSlice
	}

	return slice
}

func ComputeFileMD5(file io.Reader) string {
	md5h := md5.New()
	io.Copy(md5h, file)
	return fmt.Sprintf("%x", md5h.Sum(nil))
}
