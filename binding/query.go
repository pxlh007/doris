// 绑定查询字符串格式参数
package binding

import (
	"net/http"
	"reflect"
)

// 定义结构体
type queryBind struct{}

// 实现Name接口
func (q queryBind) Name() string {
	return "query"
}

// 实现bind接口
func (q queryBind) Bind(r http.Request, obj interface{}) error {
	// 绑定url查询参数
	values := r.URL.Query()
	err := mapping(values, obj, q.Name())
	return err
}
