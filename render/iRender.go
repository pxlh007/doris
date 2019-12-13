package render

import (
	"net/http"
)

type IRender interface {
	// 渲染接口
	Render(http.ResponseWriter) error
	// 写类型头接口
	WritecontentType(http.ResponseWriter)
}

// 声明接口的实现对象
var (
	_ IRender = Json{}
	// _ IRender = xml{}
	// _ IRender = ProtoBuf{}
	// _ IRender = Text{}
	// _ IRender = Yaml{}
)

// 更新当前请求的content_type头信息
func writeContentType(w http.ResponseWriter, value []string) {
	// 返回全部的header信息的map
	header := w.Header()
	// 已经设置直接跳过
	// 未设置就设置为指定的类型
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
