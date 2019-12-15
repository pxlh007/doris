package render

import (
	// "fmt"
	"bytes"
	"html/template"
	"net/http"

	"github.com/pxlh007/doris/internal/json"
)

// 返回普通的json格式
type Json struct {
	Data interface{} // 需要渲染的数据
}

// 返回jsonp格式
type Jsonp struct {
	Callback string      // 回调函数
	Data     interface{} // 需要渲染的数据
}

// 返回纯净版json格式
type PureJson struct {
	Data interface{} // 需要渲染的数据
}

// 返回asciiJson格式
type AsciiJson struct {
	Data interface{} // 需要渲染的数据
}

// 定义带缩进的json格式
type IndentedJson struct {
	Data interface{}
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

// 写类型接口
func (j Json) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

// 发送json体
func (j Json) WriteJson(w http.ResponseWriter) (err error) {
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&j.Data)
	return
}

// 写带缩进的json
func (j IndentedJson) Render(w http.ResponseWriter) error {
	jsonBytes, err := json.MarshalIndent(j.Data, "", "    ")
	if err != nil {
		return err
	}
	// 调用ResponseWriter的Write接口
	_, err = w.Write(jsonBytes)
	return err
}

// 写缩进版json的类型头
func (j IndentedJson) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

// 写jsonp版的json
func (j Jsonp) Render(w http.ResponseWriter) (err error) {
	ret, err := json.Marshal(j.Data)
	if err != nil {
		return err
	}
	if j.Callback == "" {
		_, err = w.Write(ret)
		return err
	}
	callback := template.JSEscapeString(j.Callback)
	_, err = w.Write([]byte(callback))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("("))
	if err != nil {
		return err
	}
	_, err = w.Write(ret)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(");"))
	if err != nil {
		return err
	}
	return nil
}

// 写jsonp版json的类型头
func (j Jsonp) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonpContentType)
}

// 写纯净版json
func (j PureJson) Render(w http.ResponseWriter) error {
	encoder := json.NewEncoder(w)
	// 设置不将html特殊字符转义为unicode编码
	encoder.SetEscapeHTML(false)
	return encoder.Encode(j.Data)
}

// 写纯净版json的类型头
func (j PureJson) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

// 写ASCII版Json
func (j AsciiJson) Render(w http.ResponseWriter) (err error) {
	ret, err := json.Marshal(j.Data)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	for _, r := range string(ret) {
		cvt := string(r)
		if r >= 128 {
			// 转化为4位十六进制无符号整型
			cvt = fmt.Sprintf("\\u%04x", int64(r))
		}
		buffer.WriteString(cvt)
	}

	_, err = w.Write(buffer.Bytes())
	return err
}

// 写ASCII版类型头
func (j AsciiJson) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonAsciiContentType)
}
