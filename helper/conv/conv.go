package conv

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

//将interface数据转json
func InterfaceToJson(itf interface{}) (string, error) {
	b, err := json.Marshal(&itf)
	return string(b), err
}

//interface转整型
func InterfaceToInt(itf interface{}) (num int, err error) {
	return strconv.Atoi(fmt.Sprintf("%v", itf))
}

//首字母大写
func UpperFirst(str string) string {
	return strings.Replace(str, str[0:1], strings.ToUpper(str[0:1]), 1)
}

//把url路径中的path请求，变成key val，path参数值形式：/user/list/p/1，转成map[string]string
func Path2Map(path string) map[string]string {
	var data = make(map[string]string)
	path = strings.Trim(path, "/")
	slice := strings.Split(path, "/")
	cnt := len(slice)
	if cnt%2 == 1 {
		cnt = cnt - 1
	}
	if cnt > 0 {
		for i := 0; i < cnt; {
			data[slice[i]] = slice[(i + 1)]
			i = i + 2
		}
	}
	return data
}

//把url中的query请求参数变成key val，query参数形式：p=1&listRows=10&aaa=bbb，转成map[string]string
func Query2Map(query string) map[string]string {
	var data = make(map[string]string)
	query = strings.Trim(query, "=&")
	slice := strings.Split(query, "&")
	if len(slice) > 0 {
		for _, v := range slice {
			vv := strings.Split(v, "=")
			if len(vv) == 2 {
				data[vv[0]] = data[vv[1]]
			}
		}
	}
	return data
}
