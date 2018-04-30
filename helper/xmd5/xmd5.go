package xmd5

import (
	"fmt"
	"strconv"
	"strings"
)

var dict = make(map[int]string, 64)

type Xmd5Obj struct {
	Dict map[int]string
}

//创建xmd5对象
func New() Xmd5Obj {
	setDict()
	return Xmd5Obj{Dict: dict}
}

func setDict() {
	for i := 0; i < 64; i++ {
		switch {
		case i <= 9:
			dict[i] = fmt.Sprintf("%v", i)
		case i >= 10 && i < 36:
			dict[i] = string(rune(i + 55))
		case i >= 36 && i < 62:
			dict[i] = string(rune(97 + i - 36))
		case i == 62:
			dict[i] = "."
		case i == 63:
			dict[i] = "_"
		}
	}
}

//xmd5，22位
func (this *Xmd5Obj) Xmd5(md5str string) string {
	bin := fmt.Sprintf("%v%v%v%v0000", hex2Bin(md5str[0:8]), hex2Bin(md5str[8:16]), hex2Bin(md5str[16:24]), hex2Bin(md5str[24:32]))
	var xmd5 []string
	cnt := len(bin) / 6
	for i := 0; i < cnt; i++ {
		str := bin[i*6 : i*6+6]
		val := interfaceToInt(string(str[0]))*32 + interfaceToInt(string(str[1]))*16 + interfaceToInt(string(str[2]))*8 + interfaceToInt(string(str[3]))*4 + interfaceToInt(string(str[4]))*2 + interfaceToInt(string(str[5]))*1
		xmd5 = append(xmd5, fmt.Sprintf("%v", dict[val]))
	}
	return strings.Join(xmd5, "")
}

//将16进制的字符串转成二进制
func hex2Bin(str string) string {
	v, _ := strconv.ParseInt(str, 16, 64)
	str = fmt.Sprintf("%b", v)
	l := len(str)
	if l < 32 {
		var slice = make([]string, l)
		for i := 0; i < 32-l; i++ {
			slice = append(slice, "0")
		}
		str = fmt.Sprintf("%v"+str, strings.Join(slice, ""))
	}
	return str
}

//interface{}转int，适用于一般的字符串、数字等
//@param    ift interface{}     需要转成int类型的数据
//@return   int                 转化后的结构
func interfaceToInt(itf interface{}) int {
	i, err := strconv.Atoi(fmt.Sprintf("%v", itf))
	if err != nil {
		return 0
	}
	return i
}
