// 绑定查询字符串格式参数
package binding

import (
	// "fmt"
	"net/http"
	"reflect"
)

// 定义结构体
type QueryBind struct{}

// 实现Name接口
func (q QueryBind) Name() string {
	return "query"
}

// 实现bind接口
func (q QueryBind) Bind(r *http.Request, obj interface{}) error {
	// 绑定url查询参数
	values := r.URL.Query()
	val := reflect.ValueOf(obj)
	return mapping(values, val, q.Name())
}
