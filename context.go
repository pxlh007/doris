package doris

import (
	//"net"
	"math"
	"net/http"
	"render"
	"sync"
)

// context是doris框架中最重要的结构之一，主要功能：
// 1. 负责在各中间件中传递参数；
// 2. 负责整个执行流程的控制；
// 3. 负责请求参数的验证以及响应结构的渲染（比如json）
type Context struct {
	Response  Response               // 用于内部操作响应对象
	Request   *http.Request          // 请求对象
	handlers  HandlersChain          // 上下文方法链
	urlParams KeyValues              // 保存单个url的键值对参数
	index     int8                   // 执行的中间件索引
	fullPath  string                 // 全路径
	doris     *Doris                 // 框架对象
	params    map[string]interface{} // 保存同一个context下的参数（key/value）
	accepted  []string               // 保存被接受的内容协商类型
	lock      sync.RWMutex           // 上下文锁
	//errors   errorMsgs     // 保存同一个context下的所有中间件和主处理函数的错误信息
}

// 单个url参数包含key/value
type KeyValue struct {
	Key   string
	Value string
}

// 参数数组通常是route返回
type KeyValues []KeyValue

// 定义最大的中间件数目默认64
const abortIndex int8 = math.MaxInt8 / 2

// 定义Context的Next方法
func (c *Context) Next() {

	debugPrintMessage("c.handlers", c.handlers, true)

	// 通过Next启动处理链
	c.index++
	// 循环逐个执行注册的方法
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// 获取GET方法获取的参数
func (c *Context) Query(param string) {
	// 获取query参数

}

// 获取GET方法获取的参数带默认值
func (c *Context) DefaultQuery(param string) {
	// 获取query参数

}

// 获取POST方法的参数
func (c *Context) PostForm(param string) {

}

// 获取POST方法的参数带默认值
func (c *Context) DefaultPostForm(param string) {

}

// 处理静态文件方法
func (c *Context) File(filepath string) {
	//

}

// 根据参数名获取参数值
func (c *Context) Param(name string) interface{} {
	return c.params[name]
}

/******************/
/*** 响应渲染相关 ***/
/******************/

// 渲染函数
func (c *Context) render(code int, r render.IRender) {
	if !bodyAllowedCode(code) {
		r.render.WritecontentType(c.Response.Writer)
		c.Response.WriteHeaderNow()
		return nil
	}
	err := r.render.Render(c.Response.Writer)
	if err != nil {
		panic(err)
	}
}

// 输出json格式
func (c *Context) Json(code int, obj interface{}) {
	c.render(code, render.Json{data: boj})
}

// 输出字符串格式
func (c *Context) String() {

}

// 输出xml格式
func (c *Context) Xml() {

}

// 输出html格式
func (c *Context) Html() {

}
