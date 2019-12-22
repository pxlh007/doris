// 绑定查询字符串格式参数
package binding

import (
	"net/http"
)

// 定义结构体
type QueryBind struct{}

// 实现Name接口
func (q QueryBind) Name() string {
	return "query"
}

// 实现bind接口
func (q QueryBind) Bind(r http.Request, obj interface{}) error {
	// 绑定url查询参数
	values := r.URL.Query()
	err := mapping(values, obj, q.Name())
	return err
}
