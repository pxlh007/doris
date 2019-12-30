// time: 2019年10月27日
// author: jonah
// desp: utils包用于实现所有的公共函数，供框架其他模块使用
package doris

import (
	"fmt"
	"os"
	"path"
	"reflect"
)

// 连接路径公用方法
func JoinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'
	// debugPrintMessage("relativePath", relativePath, true)
	// debugPrintMessage("absolutePath", absolutePath, true)
	// debugPrintMessage("finalPath", finalPath, true)
	// debugPrintMessage("appendSlash", appendSlash, true)
	if appendSlash {
		return finalPath + "/"
	}
	return finalPath
}

// 解析网络地址
func ResolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			//debugPrint("Environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		//debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}

// 判断元素是否在切片中
func InSlice(item string, s []string) bool {
	sl := len(s)
	for i := 0; i < sl; i++ {
		if s[i] == item {
			return true
		}
	}
	return false
}

//将keys切片和values切片转为Map
func SliceToMap(keys []string, values []interface{}) map[string]interface{} {
	mObj := make(map[string]interface{})
	lk, lv := len(keys), len(values)
	if lk >= lv {
		keys = keys[:lv]
	} else {
		values = values[:lk]
	}
	for index, key := range keys {
		mObj[key] = values[index]
	}
	return mObj
}

// 获取字符串的末尾字符
func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

// debugPrint
func debugPrintMessage(name string, value interface{}, debug bool) {
	if !debug {
		return
	} else {
		if valueStr, ok := value.(string); ok && valueStr == "__print__" {
			fmt.Printf("调试信息：%s\n", name)
		} else {
			fmt.Printf("参数：%s; 其值是：%v\n", name, value)
		}
	}
}

// 判断interface类型的动态值是否为nil
func IsNil(i interface{}) bool {
	defer func() {
		recover()
	}()
	vi := reflect.ValueOf(i)
	return vi.IsNil()
}

// 实现字符串截取功能 Unicode编码的情况
// 普通的字符串可以直接使用切片截取
// 参考：https://cloud.tencent.com/developer/ask/50599
func SubString(str string, start int, end int) string {
	rs := []rune(str)
	// rs[开始索引:结束索引]
	return string(rs[start:end])
}
