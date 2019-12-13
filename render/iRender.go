package render

type Render interface {
	// 渲染接口
	Render()
	// 写类型头接口
	WritecontentType()
}

// 声明接口的实现对象
var (
	_ Render Json{}
	_ Render xml{}
	_ Render ProtoBuf{}
	_ Render Text{}
	_ Render Yaml{}
)
