// 绑定form表单参数
package binding

import (
	"net/http"
	"reflect"
)

// 定义结构体
type FormBind struct{}

// 实现Name接口
func (f FormBind) Name() string {
	return "form"
}

// 实现bind接口
func (f FormBind) Bind(r http.Request, obj interface{}) error {
	// 绑定form表单参数
	r.ParseForm()
	// 解析出url.Values
	form := r.Form
	val := reflect.ValueOf(obj)
	return mapping(form, val, f.Name())
}
