//model公用操作
package models

import (
	"fmt"

	gomail "gopkg.in/gomail.v2"

	"github.com/TruthHun/DocHub/helper"

	"reflect"

	"strings"

	"errors"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"

	"os"

	"os/exec"
	"strconv"

	"time"

	"crypto/tls"

	_ "github.com/go-sql-driver/mysql"
)

//Ormer对象
var O orm.Ormer

//注意一下，varchar最多能存储65535个字符

//以下是数据表model对象
var (
	ModelAdmin         = new(Admin)           //管理员表
	ModelAd            = new(Ad)              //广告内容表
	ModelAdPosition    = new(AdPosition)      //广告位
	ModelCollectFolder = new(CollectFolder)   //收藏夹
	ModelBanner        = new(Banner)          //banner
	ModelCategory      = new(Category)        //分类
	ModelCoinLog       = new(CoinLog)         //金币记录
	ModelCollect       = new(Collect)         //收藏内容
	ModelDoc           = new(Document)        //文档
	ModelDocInfo       = new(DocumentInfo)    //文档数据信息
	ModelDocStore      = new(DocumentStore)   //文档存储
	ModelDocRecycle    = new(DocumentRecycle) //回收站
	ModelDocRemark     = new(DocumentRemark)  //文档备注
	ModelDocIllegal    = new(DocumentIllegal) //非法文档
	ModelDocComment    = new(DocumentComment) //文档评论
	ModelFriend        = new(Friend)          //友链
	ModelPages         = new(Pages)           //单页
	ModelRelate        = new(Relate)          //相关文档
	ModelReport        = new(Report)          //举报
	ModelSeo           = new(Seo)             //SEO
	ModelSign          = new(Sign)            //签到
	ModelSuggest       = new(Suggest)         //建议
	ModelSys           = new(Sys)             //系统
	ModelUser          = new(User)            //用户
	ModelUserInfo      = new(UserInfo)        //用户信息
	ModelWord          = new(Word)            //关键字
	ModelFreeDown      = new(FreeDown)        //免费下载，如果文档时收费下载，则用户下载第一次之后，在一定的时间范围内，再次下载则免费
	ModelSearchLog     = new(SearchLog)       //搜索日志
	ModelDocText       = new(DocText)         //文档文本内容
	ModelCrawlWords    = new(CrawlWords)      //采集关键字
	ModelCrawlFiles    = new(CrawlFiles)      //采集文件的信息
	ModelGitbook       = new(Gitbook)         //Gitbook文档信息存储表
	ModelConfig        = new(Config)          //配置
)

//以下是数据表
var (
	TableUser          = GetTable("user")
	TableUserInfo      = GetTable("user_info")
	TableCollect       = GetTable("collect")
	TableCollectFolder = GetTable("collect_folder")
	TableSeo           = GetTable("seo")
	TableSearchLog     = GetTable("search_log")
	TablePages         = GetTable("pages")
	TableReport        = GetTable("report")
	TableDocInfo       = GetTable("document_info")
	TableDoc           = GetTable("document")
	TableDocText       = GetTable("doc_text")
	TableDocStore      = GetTable("document_store")
	TableCategory      = GetTable("category")
	TableFreeDown      = GetTable("free_down")
	TableWord          = GetTable("word")
	TableDocIllegal    = GetTable("document_illegal")
	TableDocComment    = GetTable("document_comment")
	TableDocRecycle    = GetTable("document_recycle")
	TableDocRemark     = GetTable("document_remark")
	TableSys           = GetTable("sys")
	TableGitbook       = GetTable("gitbook")
	TableConfig        = GetTable("config")
)

//以下是数据库全局数据变量
var (
	GlobalSys               Sys          //全局系统设置
	GlobalGitbookPublishing bool = false //是否正在发布gitbook书籍，如果是，则不能再点击发布
	GlobalGitbookNextAbled  bool = true  //是否可以继续采集和发布下一本数据，如果true，则表示可以继续下载和发布下一本电子书，否则执行等待操作
)

//以下是表字段查询
var Fields = map[string]map[string]string{
	TableUser:     helper.StringSliceToMap(GetFields(ModelUser)),
	TableUserInfo: helper.StringSliceToMap(GetFields(ModelUserInfo)),
}

//初始化数据库注册
func Init() {
	//初始化数据库
	RegisterDB()
	runmode := beego.AppConfig.String("runmode")
	if runmode == "prod" {
		orm.Debug = false
		orm.RunSyncdb("default", false, false)
	} else {
		orm.Debug = true
		orm.RunSyncdb("default", false, true)
	}
	O = orm.NewOrm()

	//安装初始数据
	install()

	//全局变量赋值
	ModelConfig.UpdateGlobal() //配置文件全局变量更新
	ModelSys.UpdateGlobal()    //更新系统配置的全局变量

}

//注册数据库
func RegisterDB() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	models := []interface{}{
		ModelUser,
		ModelUserInfo,
		ModelAdmin,
		ModelCategory,
		ModelDoc,
		ModelDocInfo,
		ModelDocStore,
		ModelDocRecycle,
		ModelDocIllegal,
		ModelDocComment,
		ModelBanner,
		ModelRelate,
		ModelCollectFolder,
		ModelCollect,
		ModelAdPosition,
		ModelAd,
		ModelFriend,
		ModelSys,
		ModelWord,
		ModelSeo,
		ModelPages,
		ModelSign,
		ModelCoinLog,
		ModelReport,
		ModelSuggest,
		ModelDocRemark,
		ModelFreeDown,
		ModelSearchLog,
		ModelDocText,
		ModelCrawlWords,
		ModelCrawlFiles,
		ModelGitbook,
		ModelConfig,
	}
	orm.RegisterModelWithPrefix(beego.AppConfig.String("db::prefix"), models...)
	db_user := beego.AppConfig.String("db::user")
	db_password := beego.AppConfig.String("db::password")
	if envpass := os.Getenv("MYSQL_PASSWORD"); envpass != "" {
		db_password = envpass
	}
	db_database := beego.AppConfig.String("db::database")
	if envdatabase := os.Getenv("MYSQL_DATABASE"); envdatabase != "" {
		db_database = envdatabase
	}
	db_charset := beego.AppConfig.String("db::charset")
	db_host := beego.AppConfig.String("db::host")
	if envhost := os.Getenv("MYSQL_HOST"); envhost != "" {
		db_host = envhost
	}
	db_port := beego.AppConfig.String("db::port")
	if envport := os.Getenv("MYSQL_PORT"); envport != "" {
		db_port = envport
	}
	dblink := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&loc=%v", db_user, db_password, db_host, db_port, db_database, db_charset, "Asia%2FShanghai")
	//下面两个参数后面要放到app.conf提供用户配置使用
	// (可选)设置最大空闲连接
	maxIdle := beego.AppConfig.DefaultInt("db::maxIdle", 50)
	// (可选) 设置最大数据库连接 (go >= 1.2)
	maxConn := beego.AppConfig.DefaultInt("db::maxConn", 300)
	if err := orm.RegisterDataBase("default", "mysql", dblink, maxIdle, maxConn); err != nil {
		panic(err)
	}

}

//获取带表前缀的数据表
//@param            table               数据表
func GetTable(table string) string {
	prefix := beego.AppConfig.String("db::prefix")
	return prefix + strings.TrimPrefix(table, prefix)
}

//根据指定的表和id删除指定的记录，如果在删除记录的时候也删除记录中记录的文件，则不能调用该方法
//@param            table                   指定要删除记录的数据表
//@param            id                      要删除的记录的ID
//@return           affected                影响的记录数
//@return           err                     错误
func DelByIds(table string, id ...interface{}) (affected int64, err error) {
	return O.QueryTable(GetTable(table)).Filter("Id__in", id...).Delete()
}

//根据指定的表和id条件更新表字段，不支持批量更新
//@param            table                   需要更新的表
//@param            field                   需要更新的字段
//@param            value                   需要更新的字段的值
//@param            id                      id条件
//@return           affected                影响的记录数
//@return           err                     错误
func UpdateByIds(table string, field string, value interface{}, id ...interface{}) (affected int64, err error) {
	return O.QueryTable(GetTable(table)).Filter("Id__in", id...).Update(orm.Params{
		field: value,
	})
}

//根据指定的表和id条件更新表字段，不支持批量更新
//@param            table                   需要更新的表
//@param            data                    需要更新的字段
//@param            filter                  过滤条件，如"Id__in"
//@param            filterValue             过滤条件的值
//@return           affected                影响的记录数
//@return           err                     错误
func UpdateByField(table string, data map[string]interface{}, filter string, filterValue ...interface{}) (affected int64, err error) {
	return O.QueryTable(GetTable(table)).Filter(filter, filterValue...).Update(data)
}

//设置字段值减小
//@param                table           需要操作的数据表
//@param                field           需要对值进行增减的字段
//@param                step            增减的步长，正值为加，负值为减
//@param                condition       查询条件
//@param                conditionArgs   查询条件参数
//@return               err             返回错误
func Regulate(table, field string, step int, condition string, conditionArgs ...interface{}) (err error) {
	table = GetTable(table) //表处理
	mark := "+"             //符号
	if step < 0 {           //步长处理
		step = -step
		mark = "-"
	}
	sql := fmt.Sprintf("update %v set %v=%v%v? where %v", table, field, field, mark, condition)
	if len(conditionArgs) > 0 {
		_, err = O.Raw(sql, step, conditionArgs[0:]).Exec()
	} else {
		_, err = O.Raw(sql, step).Exec()
	}
	return err
}

//从单表中根据条件获取数据列表
//@param            table           需要查询的表
//@param            p               页码
//@param            listRows        每页显示记录数
//@param            condition       查询条件
//@param            orderby         排序
//@return           params          数据列表
//@return           rows            返回的记录数
//@return           err             错误
func GetList(table string, p, listRows int, condition *orm.Condition, orderby ...string) (params []orm.Params, rows int64, err error) {
	rows, err = O.QueryTable(GetTable(table)).SetCond(condition).Limit(listRows).Offset((p - 1) * listRows).OrderBy(orderby...).Values(&params)
	return params, rows, err
}

//获取指定Strut的字段
//@param            tableObj        Strut结构对象，引用传递
//@return           fields          返回字段数组
func GetFields(tableObj interface{}) (fields []string) {
	elem := reflect.ValueOf(tableObj).Elem()
	for i := 0; i < elem.NumField(); i++ {
		fields = append(fields, elem.Type().Field(i).Name)
	}
	return fields
}

//获取子节点
func GetChildrenNode(node string, value interface{}, params []orm.Params) (data []orm.Params) {
	strVal := fmt.Sprintf("%v", value)
	for _, v := range params {
		if strVal == fmt.Sprintf("%v", v[node]) {
			data = append(data, v)
		}
	}
	return data
}

//转成树形结构
func ToTree(params []orm.Params, Node string, value interface{}) []orm.Params {
	parents := GetChildrenNode(Node, value, params)
	if len(parents) > 0 {
		for _, v := range parents {
			children := GetChildrenNode(Node, v["Id"], params)
			if len(children) > 0 {
				for _, vv := range children {
					vv["Child"] = GetChildrenNode(Node, vv["Id"], params)
				}
			}
			v["Child"] = children
		}
	}
	return parents
}

//资源搜索
//@param            wd          搜索关键字
//@param            sourceType  搜索的资源类型，可选择：doc、ppt、xls、pdf、txt、other、all
//@param            order       排序，可选值：new(最新)、down(下载)、page(页数)、score(评分)、size(大小)、collect(收藏)、view（浏览）、default(默认)
//@param            p           页码
//@param            listRows    每页显示记录数
//@param            accuracy    是否精确搜索
func Search(wd, sourceType, order string, p, listRows, accuracy int) (res Result) {

	//========== like 查询  ==============
	//TODO:目前的查询没有排序、没有分类等，需要上elasticsearch

	//SELECT * from hc_document d left JOIN hc_document_info i on d.Id=i.Id LEFT JOIN hc_document_store s on i.DsId=s.Id where d.Title LIKE '%js%' GROUP BY s.Id  ORDER by i.Dcnt DESC
	//fields := "" //查询的字段
	start := time.Now().UnixNano()
	res.Word = []string{wd}
	res.Msg = "ok"
	res.Status = 1
	qs := O.QueryTable(TableDoc).Filter("Title__icontains", wd)
	if res.Total, _ = qs.Count(); res.Total > 0 {
		var (
			docs []Document
			ids  []string
		)
		qs.Limit(listRows).Offset((p-1)*listRows).All(&docs, "Id")
		for _, doc := range docs {
			ids = append(ids, strconv.Itoa(doc.Id))
		}
		res.Ids = strings.Join(ids, ",")
	}
	end := time.Now().UnixNano()
	res.Time = float64(end-start) / 1000000000
	return
}

//使用MySQL的like查询
//@param            wd          搜索关键字
//@param            sourceType  搜索的资源类型，可选择：doc、ppt、xls、pdf、txt、other、all
//@param            order       排序，可选值：new(最新)、down(下载)、page(页数)、score(评分)、size(大小)、collect(收藏)、view（浏览）、default(默认)
//@param            p           页码
//@param            listRows    每页显示记录数
func SearchByMysql(wd, sourceType, order string, p, listRows int) (data []orm.Params, total int64) {
	tables := []string{TableDocInfo + " i", TableDoc + " d", TableDocStore + " ds"}
	on := []map[string]string{
		{"i.Id": "d.Id"},
		{"i.DsId": "ds.Id"},
	}
	fields := map[string][]string{
		"i":  {"Score", "TimeCreate", "Id", "Dcnt", "Vcnt", "Price"},
		"d":  {"Title", "Description"},
		"ds": {"Page", "Size", "ExtCate", "Md5"},
	}
	//排序
	orderBy := []string{}
	switch strings.ToLower(order) {
	case "new":
		orderBy = []string{"i.Id desc"}
	case "down":
		orderBy = []string{"i.Dcnt desc"}
	case "page":
		orderBy = []string{"s.Page desc"}
	case "score":
		orderBy = []string{"i.Score desc"}
	case "size":
		orderBy = []string{"s.Size desc"}
	case "collect":
		orderBy = []string{"i.Ccnt desc"}
	case "view":
		orderBy = []string{"i.Vcnt desc"}
	}
	cond := " i.Status>=0 and d.Title like ? "
	//文档类型过滤条件
	ExtNum := 0 //这些也暂时写死了，后面再优化....
	switch strings.ToLower(sourceType) {
	case "doc":
		ExtNum = 1
	case "ppt":
		ExtNum = 2
	case "xls":
		ExtNum = 3
	case "pdf":
		ExtNum = 4
	case "txt":
		ExtNum = 5
	case "other":
		ExtNum = 6
	}
	if ExtNum > 0 {
		cond = cond + " and ds.ExtNum=" + strconv.Itoa(ExtNum)
	}

	//数量统计
	if sql, err := LeftJoinSqlBuild(tables, on, map[string][]string{"i": []string{"Count"}}, 1, 100000000, nil, []string{"i.DsId"}, cond); err == nil {
		sql = strings.Replace(sql, "i.Count", "count(d.Id) cnt", -1)
		var params []orm.Params
		O.Raw(sql, "%"+wd+"%").Values(&params)
		if len(params) > 0 {
			total, _ = strconv.ParseInt(params[0]["cnt"].(string), 10, 64)
		}
	} else {
		helper.Logger.Error(err.Error())
		helper.Logger.Debug(sql, wd)
	}
	if total == 0 {
		return
	}
	//数据查询
	if sql, err := LeftJoinSqlBuild(tables, on, fields, p, listRows, orderBy, []string{"i.DsId"}, cond); err == nil {
		helper.Logger.Debug(sql, wd)
		O.Raw(sql, "%"+wd+"%").Values(&data)
	} else {
		helper.Logger.Error(err.Error())
		helper.Logger.Debug(sql, wd)
	}

	return
}

//左联合查询创建SQL语句
//@param                tables                  需要作为联合查询的数据表。注意：数据表的第一个表是主表
//@param                on                      联合查询的on查询条件，必须必表(tables)少一个。比如user表和user_info表做联合查询，那么on查询条件只有一个，必tables的数组元素少一个
//@param                fields                  需要查询的字段
//@param                p                       页码
//@param                listRows                每页查询记录数
//@param                orderBy                 排序条件，可以穿空数组
//@param                groupBy                 按组查询
//@param                condition               查询条件
//@param                conditionArgs           查询条件参数
//@return               sql                     返回生成的SQL语句
//@return               err                     错误。如果返回的错误不为nil，则SQL语句为空字符串
//使用示例：
//tables := []string{"document", "document_info info", "document_store store"}
//fields := map[string][]string{
//"document": {"Id Did", "Title", "Filename"},
//"info":     {"Vcnt", "Dcnt"},
//"store":    {"Md5", "Page"},
//}
//on := []map[string]string{
//{"document.Id": "info.Id"},
//{"info.DsId": "store.Id"},
//}
//orderby := []string{"doc.Id desc", "store.Page desc"}
//sql, err := LeftJoinSqlBuild(tables, on, fields, 1, 100, orderby, nil, "")
//fmt.Println(sql, err)
func LeftJoinSqlBuild(tables []string, on []map[string]string, fields map[string][]string, p, listRows int, orderBy []string, groupBy []string, condition string) (sql string, err error) {
	if len(tables) < 2 || len(tables)-1 != len(on) {
		err = errors.New("参数不规范：联合查询的数据表数量必须在2个或2个以上，同时表数量比on条件多一个")
		return
	}
	var (
		FieldSlice   []string
		StrOrderBy   string
		StrGroupBy   string
		StrCondition string
		joinKV       string
		join         = []string{tables[0]}
		usedTables   = []string{}
	)
	for table, field := range fields {
		for _, f := range field {
			FieldSlice = append(FieldSlice, strings.Trim(fmt.Sprintf("%v.%v", table, f), "."))
		}
	}
	for index, table := range tables {
		slice := strings.Split(strings.TrimSpace(table), " ")
		if len(slice) == 1 {
			slice = append(slice, slice[0])
		}
		usedTables = append(usedTables, slice[1])
		if index > 0 {
			on, joinKV = joinOn(slice[1], usedTables, on)
			join = append(join, "left join "+table+" on "+joinKV)
		}
	}
	if len(orderBy) > 0 {
		StrOrderBy = " order by " + strings.Join(orderBy, ",")
	}
	if len(condition) > 0 {
		StrCondition = " where " + condition
	}
	if len(groupBy) > 0 {
		StrGroupBy = " group by " + strings.Join(groupBy, ",")
	}

	sql = fmt.Sprintf("select %v from %v %v %v %v limit %v offset %v", strings.Join(FieldSlice, ","), strings.Join(join, " "), StrCondition, StrGroupBy, StrOrderBy, listRows, (p-1)*listRows)
	return
}

//只供LeftJoinSqlBuild创建SQL语句使用
//@param                table               需要左联查询的表
//@param                usedTables          已使用的表
//@param                on                  联合查询条件
//@return               newon               新的联合查询条件(返回未被使用的联合查询条件)
//@return               ret                 返回组装联合查询条件
func joinOn(table string, usedTables []string, on []map[string]string) (newon []map[string]string, ret string) {
	table = table + "."
	lenon := len(on)
	for index, v := range on {
		for key, val := range v {
			if strings.HasPrefix(key, table) || strings.HasPrefix(val, table) {
				for _, used := range usedTables {
					if strings.HasPrefix(key, used) || strings.HasPrefix(val, used) {
						ret = key + "=" + val
						if index > 0 {
							newon = append(newon, on[0:index]...)
						}
						if index+1 <= lenon {
							newon = append(newon, on[(index+1):]...)
						}
						return
					}
				}
			}
		}
	}
	return
}

//将PDF文件转成svg，并把文件更新到oss上【注意：svg存放的文件夹是xmd5对MD5字符串加密后的文件夹】
//@param            file            pdf文件
//@param            totalPage       pdf文件页数
//@return           files           生成的pdf文件
//@return           err             错误
func Pdf2Svg(file string, totalPage int, md5str string) (err error) {
	var (
		width   int
		height  int
		content string
	)

	//文件夹
	folder := strings.TrimSuffix(strings.ToLower(file), ".pdf")
	folder = strings.TrimSuffix(folder, "/")
	//os.MkdirAll(folder, 0777)//注意：这里不要创建文件夹！！
	//如果文件夹folder已经存在了，则需要先删除
	os.MkdirAll(folder, os.ModePerm)
	defer os.RemoveAll(folder)

	pdf2svg := helper.GetConfig("depend", "pdf2svg", "pdf2svg")

	//compress := beego.AppConfig.DefaultBool("compressSvg", false) //是否压缩svg
	compress := true                            //强制为true
	content = helper.ExtractPdfText(file, 1, 5) //提取前5页的PDF文本内容
	watermarkText := ModelSys.GetByField("Watermark").Watermark
	//处理pdf转svg
	for i := 0; i < totalPage; i++ {
		num := i + 1
		svgfile := fmt.Sprintf("%v/%v.svg", folder, num)
		//Usage: pdf2svg <in file.pdf> <out file.svg> [<page no>]
		cmd := exec.Command(pdf2svg, file, svgfile, strconv.Itoa(num))
		if helper.Debug {
			beego.Debug("pdf转svg参数", cmd.Args)
		}
		if err := cmd.Run(); err != nil {
			helper.Logger.Error(err.Error())
		} else {
			if num == 1 {
				//封面处理
				if cover, err := helper.ConvertToJpeg(svgfile, false); err == nil {
					NewOss().MoveToOss(cover, md5str+".jpg", true, true)
				}
				//获取svg的宽高(pt)
				width, height = helper.ParseSvgWidthAndHeight(svgfile)
				if _, err := UpdateByField(TableDocStore, map[string]interface{}{"Width": width, "Height": height}, "Md5", md5str); err != nil {
					helper.Logger.Error(err.Error())
				}
			}
			//添加文字水印
			helper.SvgTextWatermark(svgfile, watermarkText, width/6, height/4)

			//压缩svg内容
			helper.CompressSvg(svgfile)
			NewOss().MoveToOss(svgfile, md5str+"/"+strconv.Itoa(num)+".svg", true, true, compress)
		}
	}

	//将内容更新到数据库
	if len(content) > 5000 {
		content = helper.SubStr(content, 0, 4800)
	}
	var docText = DocText{Md5: md5str, Content: content}
	if _, _, err := O.ReadOrCreate(&docText, "Md5"); err != nil {
		helper.Logger.Error(err.Error())
	}

	//扫尾工作，如果还存在文件，则继续将文件移到oss
	filenum := 1 //假设有一个svg或者jpg文件
	for {
		files := helper.ScanDir(folder)
		if filenum > 0 {
			//将l重置为0
			filenum = 0
			for _, file := range files {
				//svg结尾
				if strings.HasSuffix(file, ".svg") { //svg结尾的，都是文档页
					slice := strings.Split(file, "/")
					NewOss().MoveToOss(file, fmt.Sprintf("%v/%v", md5str, slice[len(slice)-1]), true, true, compress)
					filenum++ //
				} else if strings.HasSuffix(file, ".jpg") { //jpg结尾的，基本都是封面图片
					NewOss().MoveToOss(file, md5str+".jpg", true, true)
					filenum++
				}
			}
		} else {
			break
		}
	}
	//删除文件夹
	if filenum == 0 {
		go os.RemoveAll(folder)
	}
	return
}

//替换写入【注意：表中必须要有一个除了主键外的唯一键】
//@param            table           需要写入的table
//@param            params          需要写入的数据
//@return           err             返回错误
func ReplaceInto(table string, params map[string]interface{}) (err error) {
	var (
		fields []string
		values []interface{}
		sql    string
	)
	if len(params) > 0 {
		table = GetTable(table)
		for field, value := range params {
			fields = append(fields, field)
			values = append(values, value)
		}
		marks := make([]string, len(values)+1)
		sql = fmt.Sprintf("REPLACE INTO `%v`(`%v`) VALUES (%v)", table, strings.Join(fields, "`,`"), strings.Join(marks, "?"))
		_, err = O.Raw(sql, values...).Exec()
	} else {
		err = errors.New("需要写入的数据不能为空")
	}
	return
}

//对单表记录进行统计查询
//@param            table           需要查询或者统计的表
//@param            cond            查询条件
//@return           cnt             统计的记录数
func Count(table string, cond *orm.Condition) (cnt int64) {
	cnt, _ = O.QueryTable(GetTable(table)).SetCond(cond).Count()
	return
}

//发送邮件
//@param            to          string          收件人
//@param            subject     string          邮件主题
//@param            content     string          邮件内容
//@return           error                       发送错误
//func SendMail(to, subject, content string) error {
//	port := beego.AppConfig.DefaultInt("email::port", 80)
//	host := beego.AppConfig.String("email::host")
//	username := beego.AppConfig.String("email::username")
//	password := beego.AppConfig.String("email::password")
//	msg := &mail.Message{
//		mail.Header{
//			"From":         {username},
//			"To":           {to},
//			"Reply-To":     {beego.AppConfig.DefaultString("mail::replyto", username)},
//			"Subject":      {subject},
//			"Content-Type": {"text/html"},
//		},
//		strings.NewReader(content),
//	}
//	m := mailer.NewMailer(host, username, password, port)
//	err := m.Send(msg)
//	return err
//}

//发送邮件
//@param            to          string          收件人
//@param            subject     string          邮件主题
//@param            content     string          邮件内容
//@return           error                       发送错误
func SendMail(to, subject, content string) (err error) {
	port := int(helper.GetConfigInt64("email", "port"))
	host := helper.GetConfig("email", "host")
	username := helper.GetConfig("email", "username")
	password := helper.GetConfig("email", "password")
	replyto := helper.GetConfig("email", "replyto")
	m := gomail.NewMessage()
	m.SetHeader("From", username)
	m.SetHeader("To", to)
	if strings.TrimSpace(replyto) != "" {
		m.SetHeader("Reply-To", replyto)
	}
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	d := gomail.NewDialer(host, port, username, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send the email to Bob, Cora and Dan.
	err = d.DialAndSend(m)

	return
}
