package render

import (
	"fmt"
	"io"
	"net/http"
)

// 返回字符串格式
type String struct {
	Format string        // 字符串显示格式
	Data   []interface{} // 需要渲染的数据
}

// 定义各种content_type类型
var (
	plainContentType = []string{"text/plain; charset=utf-8"}
)

// 实现渲染接口
func (s String) Render(w http.ResponseWriter) error {
	if err := s.WriteString(w); err != nil {
		panic(err)
	}
	return nil
}

// 实现类型接口
func (s String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, plainContentType)
}

// 具体写功能
func (s String) WriteString(w http.ResponseWriter) (err error) {
	if len(s.Data) > 0 {
		// w.Write([]byte(s.Data))
		_, err = fmt.Fprintf(w, s.Format, s.Data...)
		return
	}
	// 数据为空
	_, err = io.WriteString(w, s.Format)
	return
}
