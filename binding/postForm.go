// 绑定form表单参数
package binding

import (
	"net/http"
	"reflect"
)

// 定义结构体
type postFormBind struct{}

// 实现Name接口
func (p postFormBind) Name() string {
	return "form"
}

// 实现bind接口
func (p postFormBind) Bind(r http.Request, obj interface{}) error {
	// 绑定url查询参数
	val := r.ParseForm()
	err := mapping(val, obj, p.Name())
	return err
}
