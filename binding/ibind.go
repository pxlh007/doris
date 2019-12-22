/**
* 定义绑定用的接口
**/
package binding

import (
	"doris/validate"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

type (
	// 定义接口
	binding interface {
		Bind(r *http.Request, obj interface{}) error // 用于绑定
		Name() string                                // 获取绑定名
	}
	// 映射字段结构
	field struct {
		fieldName    string        // 参数名
		required     bool          // 是否必须
		regex        string        // 正则表达式
		validate     string        // 验证器序列
		hasDefault   bool          // 是否存在默认值
		defaultValue reflect.Value // 默认值
	}

	/**
	eg:
	type User struct {
		"Age" int8 `param:"age" default:18 lte:100 gte:10`
		"Username" string `param:"username" default:"louhao" regex:"\w+" validate:"(required,letter)|digit"`
	}
	*/
)

// 定义映射数组
var bitMap = map[reflect.Kind]int{
	reflect.Int:     32,
	reflect.Int16:   16,
	reflect.Int32:   32,
	reflect.Int64:   64,
	reflect.Int8:    8,
	reflect.Uint:    32,
	reflect.Uint16:  16,
	reflect.Uint32:  32,
	reflect.Uint64:  64,
	reflect.Uint8:   8,
	reflect.Float32: 32,
	reflect.Float64: 64,
}

// 定义错误提示
var (
	ErrStruct = errors.New("需要传入struct参数")
)

// 公共映射方法
func mapping(values url.Values, val reflect.Value, bType string) error {
	// 根据不同类型执行映射
	if bType == "query" || bType == "form" { // url/form 映射
		typ := val.Type() // 获取对象类型
		if val.Kind() == reflect.Ptr {
			if val.IsNil() { // 空指针
				return ErrStruct
			}

			// 取指针指向的元素
			val = val.Elem()

			// 递归映射
			return mapping(values, val, bType)

		} else if val.Kind() != reflect.Struct { // 非结构体
			return ErrStruct
		}

		// 循环结构体字段
		for i := 0; i < typ.NumField(); i++ {
			// 判断字段可导出
			if !val.Field(i).CanInterface() {
				continue
			}

			// 遍历所有的元素
			fieldT := typ.Field(i)
			param := fieldT.Tag.Get("param")
			if param == "" {
				return errors.New("参数不能为空")
			}

			// 获取url参数
			data := values.Get(param)

			// 获取验证规则
			vRule := fieldT.Tag.Get("validate")

			// 参数校验
			vd := new(validate.Validater)
			ok, err := vd.Validate(data, vRule)
			if !ok {
				if err != nil {
					return err
				}
				return errors.New("验证失败")
			}

			// 设置结构体
			err = setValue(data, val.Field(i))
			if err != nil {
				return err
			}
		}
	} else if bType == "json" {
		// json映射

	} else if bType == "file" {
		// file映射

	} else if bType == "protobuf" {
		// protobuf映射

	} else {
		// 其他类型
		return errors.New("不支持的类型")
	}
	return nil
}

// 实际执行映射的地方
func setValue(data string, val reflect.Value) (err error) {
	// 获取kind
	kind := val.Kind()

	// 根据不同的类调用设置函数
	switch kind {
	case reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int:
		var d int64
		d, err = strconv.ParseInt(data, 10, bitMap[kind])
		if err != nil {
			return
		}
		val.SetInt(d)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		var d uint64
		d, err = strconv.ParseUint(data, 10, bitMap[kind])
		if err != nil {
			return
		}
		val.SetUint(d)
	case reflect.Float64, reflect.Float32:
		var d float64
		d, err = strconv.ParseFloat(data, bitMap[kind])
		if err != nil {
			return
		}
		val.SetFloat(d)
	case reflect.String:
		if data == "" {
			return
		}
		val.SetString(data)
		return
	}

	return

}

// 优化反射性能
//func setValueWithPointer() {
//	acc := &Account{}
//	tp := reflect.TypeOf(acc).Elem()
//	field, _ := tp.FieldByName("Email")
//	fieldPtr := uintptr(unsafe.Pointer(acc)) + field.Offset
//	*((*string)(unsafe.Pointer(fieldPtr))) = "admin#otokaze.cn"
//	fmt.Println(acc) // stdout: &{admin#otokaze.cn   0}
//}
