package helper

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"regexp"

	"github.com/astaxie/beego/validation"
)

var valid = validation.Validation{}

//测试数据验证，仅作为测试验证规则
func TestValid() {
	fmt.Println(len("地球"))
	fmt.Println(len("go"))
	fmt.Println(strings.Count("地球", "") - 1)
	fmt.Println("是否满足最小值要求", execValid("100", "min", "120"))
	fmt.Println("是否满足最大值要求", execValid("100", "max", "90"))
	fmt.Println("是否满足最大长度要求", execValid("123sdaswerew", "maxlen", "5"))
	fmt.Println("是否满足最小长度要求", execValid("21312dsafadf", "minlen", "50"))
	fmt.Println("是否满足指定长度", execValid("13fadfwerwr", "len", "50"))
	fmt.Println("邮箱格式验证", execValid("1231das#qq.com", "email"), execValid("1231das@qq.com", "email"))
	fmt.Println("座机", execValid("0771-6772237", "tel"))
	fmt.Println("座机", execValid("0771-677223711", "tel"))
	fmt.Println("手机号码", execValid("13687777777", "mobile"))
	fmt.Println("手机号码", execValid("53687717717", "mobile"))
	fmt.Println("enum", execValid("123", "enum", "1234", "457"))
	fmt.Println("range", execValid("100", "range", "1", "10"))
	fmt.Println("邮政编码", execValid("518000", "zipcode"))
	fmt.Println("邮政编码", execValid("5180001", "zipcode"))
	fmt.Println("IP", execValid("127.0.0.1", "ip"))
	fmt.Println("IP", execValid("1809.1.1.1", "ip"))
	fmt.Println("字母验证", execValid("12312asdasda", "alpha"))
	fmt.Println("字母验证", execValid("asADASDdasda", "alpha"))
	fmt.Println("是否是数字", execValid("123123", "numeric"))
	fmt.Println("是否是数字", execValid("12312.23", "numeric"))
	fmt.Println("数字字母", execValid("1231212asdfasdfAASDAS", "alphanumeric"))
	fmt.Println("数字字母", execValid("1231212asd..fasdfAASDAS", "alphanumeric"))
	fmt.Println("数字字母横线", execValid("1231212asd|\asfasdfAASDAS", "alphadash"))
	fmt.Println("数字字母横线", execValid("1231212-_dfAASDAS", "alphadash"))
	fmt.Println("正则规则验证", execValid("pe111?sach", "regexp", "p([a-z]+)ch"))
}

//参数验证
//规则中，required表示必传参数，有无值则是另外一回事
//验证规则有以下(大小写不敏感)：
//1、required    参数必传，有无值则是另外一回事，使用示例："required"
//1、unempty     值不能为空，使用示例："unempty"
//2、min         最小值验证，使用示例："min:1"
//3、max         最大值验证，使用示例："max:1000"
//4、minlen、mincount      最小长度验证，使用示例："minlen:2"         说明：count是字符个数校验，len是字节个数校验，如"地球"，如果是len，则是6个字节，如果是count，则是2个字符
//5、maxlen、maxcount      最大长度验证，使用示例："maxcount:255"       说明：count是字符个数校验，len是字节个数校验，如"地球"，如果是len，则是6个字节，如果是count，则是2个字符
//6、len、count            字符串长度验证，使用示例："len:40"         说明：count是字符个数校验，len是字节个数校验，如"地球"，如果是len，则是6个字节，如果是count，则是2个字符
//7、email       邮箱格式验证，使用示例："email"
//8、tel         座机号码格式验证，使用示例："tel"
//9、mobile      手机号码格式验证，使用示例："mobile"
//10、phone      电话号码格式验证，包括手机号码和座机号码，使用示例："phone"
//11、enum       参数值验证，参数值只能是其中指定的一个值，使用示例："enum:男:女:保密"
//12、range      参数区间，适用于数字变化范围，使用示例："range:1:100"
//13、int,int8,int32,int64,float32,float64,string           只要有这些数据类型的字段，则表示类型转换，使用示例："int64"
//14、zipcode    邮政编码验证，使用示例："zipcode"
//15、ip         IP地址格式验证，使用示例："ip"
//16、alpha      字符验证，字符必须是大小写字母，使用示例："alpha"
//17、numeric    数字验证，必须是0-9中的数字，使用示例："numeric"
//18、alphanumeric   大小写字母和数字验证，值必须是大小写字母和数字，使用示例："alphanumeric"
//19、alphadash      大小写字母、数字、横杠(-_)验证，使用示例："alphadash"
//20、gt,lt,gte,lte  大于、小于、大于等于或者小于等于指定数值：使用示例:"gt:0"    说明：如果是等于指定数值，使用enum，如"enum:1",即等于1
//21、regexp     正则校验，使用示例："regexp:[a-zA-z0-9\.\-\_]{10}"，第一个冒号后面的表示正则规则
//############################################
//@param            params          url.Values参数
//@param            rules           验证规则
//@return                           返回验证后通过的数据
//@return                           返回错误的map
func Valid(params url.Values, rules map[string][]string) (map[string]interface{}, map[string]string) {
	var (
		data   map[string]interface{}
		err    error
		errmsg map[string]string
		//可以执行类型转换的数据类型
		typeConvertStr = "int,int8,int32,int64,float32,float64,string"
	)
	data = make(map[string]interface{})
	errmsg = make(map[string]string)

	//为0的字符串和不为0的字符串做数字解析后为0，两种效果是不一样的
	for key, slice := range rules {
		var errs []string
		//如果没有验证规则，则值有则接收，无则不理会
		l := len(slice)
		if l == 0 {
			//如果这个key没有对应规则，key存在，则取值，不存在，则不取
			if v, ok := params[key]; ok {
				data[key] = v[0]
			}
		} else {
			//是否必传
			IsReuqired := false
			for _, v := range slice {
				v = strings.ToLower(v)
				if v == "required" {
					IsReuqired = true
					break
				}
			}

			//必填字段，也就是key必须传递过来
			if IsReuqired {
				if v, ok := params[key]; ok {
					data[key] = v[0]
				} else {
					errs = append(errs, fmt.Sprintf("%v参数必传", key))
				}
			}
			//非必须的情况下，有传递则直接验证
			if param, ok := params[key]; ok {
				v := param[0]
				//数据验证
				for _, item := range slice {
					//规则切分，如range:1:10
					itemSlice := strings.Split(item, ":")
					err = execValid(v, itemSlice[0], itemSlice[1:]...)
					if err == nil {
						data[key] = v
					} else {
						errs = append(errs, err.Error())
					}
				}

				//类型转换
				for _, item := range slice {
					//规则切分，如range:1:10
					itemSlice := strings.Split(item, ":")

					if strings.Contains(typeConvertStr, itemSlice[0]) {
						data[key], err = execConvert(v, itemSlice[0])
						if err != nil {
							errs = append(errs, err.Error())
						}
					}
				}
			}
		}
		if len(errs) > 0 {
			errmsg[key] = strings.Join(errs, ";")
		}
	}
	return data, errmsg
}

//数据验证
//@param            val         字符串
//@param            rule        验证规则，见Valid()注释
//@param            args        变参参数
//@return                       返回验证错误
func execValid(val, rule string, args ...string) error {
	rule = strings.ToLower(rule)
	switch rule {
	//验证邮箱格式
	case "email":
		res := valid.Email(val, "email")
		if !res.Ok || res.Error != nil {
			return errors.New("邮箱格式不正确")
		}
		//验证手机、座机格式
	case "phone":
		res := valid.Phone(val, "phone")
		if !res.Ok || res.Error != nil {
			return errors.New("联系电话格式不正确")
		}
		//验证手机号码
	case "mobile":
		res := valid.Phone(val, "mobile")
		if !res.Ok || res.Error != nil {
			return errors.New("手机号码格式不正确")
		}
		//验证座机号码
	case "tel":
		res := valid.Phone(val, "tel")
		if !res.Ok || res.Error != nil {
			return errors.New("座机号码格式不正确")
		}
		//字符长度校验
	case "len", "count":
		if len(args) > 0 {
			num, _ := strconv.ParseInt(args[0], 10, 64)
			if rule == "count" {
				if strings.Count(val, "")-1 != int(num) {
					return errors.New(fmt.Sprintf("字符个数限制在%v个字符", num))
				}
			} else {
				if len(val) != int(num) {
					return errors.New(fmt.Sprintf("字符个数限制在%v个字符", num))
				}
			}
		}
		//最大最小值校验
	case "min", "max":
		if len(args) > 0 {
			num, _ := strconv.ParseFloat(args[0], 64)
			parseVal, _ := strconv.ParseFloat(val, 64)
			if rule == "min" {
				if parseVal < num {
					return errors.New(fmt.Sprintf("最小值不能小于%v", num))
				}
			} else {
				if parseVal > num {
					return errors.New(fmt.Sprintf("最大值不能大于%v", num))
				}
			}
		}
	case "gt":
		if len(args) > 0 {
			num, _ := strconv.ParseFloat(args[0], 64)
			parseVal, _ := strconv.ParseFloat(val, 64)
			if parseVal <= num {
				return errors.New(fmt.Sprintf("值必须大于%v", num))
			}
		}
	case "lt":
		if len(args) > 0 {
			num, _ := strconv.ParseFloat(args[0], 64)
			parseVal, _ := strconv.ParseFloat(val, 64)
			if parseVal >= num {
				return errors.New(fmt.Sprintf("值必须小于%v", num))
			}
		}
	case "gte":
		if len(args) > 0 {
			num, _ := strconv.ParseFloat(args[0], 64)
			parseVal, _ := strconv.ParseFloat(val, 64)
			if parseVal < num {
				return errors.New(fmt.Sprintf("值必须大于或等于%v", num))
			}
		}
	case "lte":
		if len(args) > 0 {
			num, _ := strconv.ParseFloat(args[0], 64)
			parseVal, _ := strconv.ParseFloat(val, 64)
			if parseVal > num {
				return errors.New(fmt.Sprintf("值必须小于或等于%v", num))
			}
		}
	case "minlen", "maxlen":
		//最大最小长度校验
		if len(args) > 0 {
			l := len(val)
			num, _ := strconv.ParseInt(args[0], 10, 64)
			if rule == "minlen" {
				if l < int(num) {
					return errors.New(fmt.Sprintf("字节数不能少于%v个", num))
				}
			} else {
				if l > int(num) {
					return errors.New(fmt.Sprintf("字符数不能多于%v个", num))
				}
			}
		}
	case "unempty":
		if len(val) == 0 {
			return errors.New("值不能为空")
		}
		//最大最小长度校验
	case "mincount", "maxcount":
		if len(args) > 0 {
			l := strings.Count(val, "") - 1
			num, _ := strconv.ParseInt(args[0], 10, 64)
			if rule == "mincount" {
				if l < int(num) {
					return errors.New(fmt.Sprintf("字符个数不能少于%v个", num))
				}
			} else {
				if l > int(num) {
					return errors.New(fmt.Sprintf("字符个数不能多于%v个", num))
				}
			}
		}
		//变化范围
	case "range":
		if len(args) == 2 {
			min, _ := strconv.ParseFloat(args[0], 64)
			max, _ := strconv.ParseFloat(args[1], 64)
			parseVal, _ := strconv.ParseFloat(val, 64)
			if parseVal < min || parseVal > max {
				return errors.New(fmt.Sprintf("超出指定数值范围：%v-%v", args[0], args[1]))
			}
		}
	case "enum":
		if len(args) > 0 {
			for _, v := range args {
				if val == v {
					return nil
				}
			}
			return errors.New("参数不在规定的值内")
		}
		//邮政编码
	case "zipcode":
		res := valid.ZipCode(val, "zipcode")
		if !res.Ok || res.Error != nil {
			return errors.New("邮政编码格式不正确")
		}
		//IP
	case "ip":
		res := valid.IP(val, "ip")
		if !res.Ok || res.Error != nil {
			return errors.New("IP地址格式不正确")
		}
	case "alpha":
		res := valid.Alpha(val, "alpha")
		if !res.Ok || res.Error != nil {
			return errors.New("仅限大小写字母")
		}
	case "numeric":
		res := valid.Numeric(val, "numeric")
		if !res.Ok || res.Error != nil {
			return errors.New("仅限0-9的数字")
		}
	case "alphanumeric":
		res := valid.AlphaNumeric(val, "alphanumeric")
		if !res.Ok || res.Error != nil {
			return errors.New("仅限大小写字母和数字")
		}
	case "alphadash":
		res := valid.AlphaDash(val, "alphadash")
		if !res.Ok || res.Error != nil {
			return errors.New("仅限大小写字母、数字和-(横杠)、_(下划线)")
		}
	case "regexp":
		//正则规则验证
		if len(args) > 0 {
			//不排除正则规则里面有冒号(:)的存在，所以需要使用join
			b, err := regexp.MatchString(strings.Join(args, ":"), val)
			if err != nil {
				return err
			}
			if b == false {
				return errors.New("参数格式不正确")
			}
		}
		return errors.New("正则规则配置不正确")

	}
	return nil
}

//数据类型转换
//@param            val         需要转换的字符串
//@param            rule        需要转换的类型，如"int"
//@return                       返回转化后的数据类型
//@return                       返回错误
func execConvert(val, rule string) (interface{}, error) {
	rule = strings.ToLower(rule)
	switch rule {
	case "int", "int32", "int8", "int64":
		bit, _ := strconv.Atoi(strings.TrimPrefix(rule, "int"))
		if bit == 0 {
			bit = 32
		}
		num, err := strconv.ParseInt(val, 10, 64)
		switch bit {
		case 8:
			return int8(num), err
		case 32:
			return int(num), err
		default:
			return num, err
		}
		//转化成浮点型
	case "float", "float32", "float64":
		bit := 32
		if bit, _ := strconv.Atoi(strings.TrimPrefix(rule, "float")); bit == 0 {
			bit = 32
		}
		num, err := strconv.ParseFloat(val, 64)
		if bit == 32 {
			return float32(num), err
		}
		return num, err
	}
	return val, nil
}
