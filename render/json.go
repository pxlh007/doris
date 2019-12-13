package render

import (
	"net/http"
	// "github.com/json-iterator/go"
	"github.com/pxlh007/doris-internal/json"
)

// 返回普通的json格式
type Json struct {
	data interface{} // 需要渲染的数据
}

// 返回jsonp格式
type Jsonp struct {
	callback string      // 回调函数
	data     interface{} // 需要渲染的数据
}

// 返回纯净版json格式
type PureJson struct {
	data interface{} // 需要渲染的数据
}

// 返回asciiJson格式
type AsciiJson struct {
	data interface{} // 需要渲染的数据
}

// 定义带缩进的json格式
type IndentJson struct {
	data interface{}
}

// 定义各种content_type类型
var (
	jsonContentType      = []string{"application/json; charset=utf-8"}
	jsonpContentType     = []string{"application/javascript; charset=utf-8"}
	jsonAsciiContentType = []string{"application/json"}
)

// 实现渲染接口
func (j Json) Render(w http.ResponseWriter) error {
	if err := j.WriteJson(w); err != nil {
		panic(err)
	}
	return nil
}

// 实现写类型接口
func (j Json) WritecontentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

// 实际执行写json
func (j Json) WriteJson(w http.ResponseWriter) (err error) {
	j.WritecontentType(w)
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&j.data)
	return
}

// 写带缩进的json

// 写纯净版的json
