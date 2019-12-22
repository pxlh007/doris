package doris

import (
	// "fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/pxlh007/doris/binding"
	"github.com/pxlh007/doris/render"
)

// context是doris框架中最重要的结构之一，主要功能：
// 1. 负责在各中间件中传递参数；
// 2. 负责整个执行流程的控制；
// 3. 负责请求参数的验证以及响应结构的渲染（比如json）
type Context struct {
	Response  *Response              // 用于内部操作响应对象
	Request   *http.Request          // 请求对象
	handlers  HandlersChain          // 上下文方法链
	urlParams KeyValues              // 保存单个url的键值对参数
	index     int8                   // 执行的中间件索引
	fullPath  string                 // 全路径
	doris     *Doris                 // 框架对象
	params    map[string]interface{} // 保存同一个context下的参数（key/value）
	accepted  []string               // 保存被接受的内容协商类型
	lock      sync.RWMutex           // 上下文锁
	// errors   errorMsgs     // 保存同一个context下的所有中间件和主处理函数的错误信息
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

/************************************/
/******** 中间件相关 ******************/
/************************************/
// 定义Context的Next方法
func (c *Context) Next() {
	// debug
	debugPrintMessage("c.handlers", c.handlers, c.doris.Debug)
	// 通过Next启动处理链
	c.index++
	// 循环逐个执行注册的方法
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

/************************************/
/******** 参数绑定/获取相关 ************/
/************************************/
// 获取GET方法获取的参数
func (c *Context) Query(obj interface{}) error {
	// 获取query参数
	b := binding.QueryBind{}
	return b.Bind(c.Request, obj)
}

// 获取POST方法的参数
func (c *Context) Form(param interface{}) error {
	// 获取query参数
	b := binding.FormBind{}
	return b.Bind(c.Request, param)
}

// 获取单个的查询参数
func (c *Context) QueryParam(param string) string {
	// 获取query参数
	var q url.Values = c.Request.URL.Query()
	return q.Get(param)
}

// 获取GET方法获取的参数带默认值
func (c *Context) DefaultQuery(param string, def string) string {
	// 获取query参数不存在则返回默认值
	var q url.Values = c.Request.URL.Query()
	if q.Get(param) == "" {
		return def
	}
	return q.Get(param)
}

// 获取单个的表单参数
func (c *Context) FormParam(param string) string {
	c.Request.ParseForm()
	var f url.Values = c.Request.Form
	return f.Get(param)
}

// 获取POST方法的参数带默认值
func (c *Context) DefaultFormParm(param string, def string) string {
	c.Request.ParseForm()
	var f url.Values = c.Request.Form
	if f.Get(param) == "" {
		return def
	}
	return f.Get(param)
}

// 处理静态文件方法
func (c *Context) File(filepath string) {
	//

}

// 根据参数名获取参数值
func (c *Context) Param(name string) interface{} {
	return c.params[name]
}

/************************************/
/******** 响应渲染相关 ****************/
/************************************/
// 渲染函数
func (c *Context) render(code int, r render.IRender) {
	r.WriteContentType(c.Response.Writer) // 设置contentType
	c.Status(code)                        // 设置status码
	if !bodyAllowedCode(code) {           // 非允许的code直接返回
		c.Response.WriteHeaderNow()
		return
	}
	err := r.Render(c.Response.Writer)
	if err != nil {
		panic(err)
	}
}

// 输出json格式
func (c *Context) Json(code int, obj interface{}) {
	c.render(code, render.Json{Data: obj})
}

// 输出pureJson格式
func (c *Context) PureJson(code int, obj interface{}) {
	c.render(code, render.PureJson{Data: obj})
}

// 输出IndentJson格式
func (c *Context) IndentedJson(code int, obj interface{}) {
	c.render(code, render.IndentedJson{Data: obj})
}

// 输出Jsonp格式
func (c *Context) Jsonp(code int, callback string, obj interface{}) {
	c.render(code, render.Jsonp{Callback: callback, Data: obj})
}

// 输出字符串格式
func (c *Context) String(code int, format string, values ...interface{}) {
	c.render(code, render.String{Format: format, Data: values})
}

// 输出xml格式
func (c *Context) Xml(code int, obj interface{}) {
	c.render(code, render.Xml{Data: obj})
}

// 输出html格式
func (c *Context) Html() {

}

// 检查传入的status是否是http包允许的
// 默认情况下304，204和100-199是不被允许加入到body中。
func bodyAllowedCode(code int) bool {
	switch {
	case code >= 100 && code <= 199:
		return false
	case code == http.StatusNoContent:
		return false
	case code == http.StatusNotModified:
		return false
	}
	return true
}

// 设置响应头状态码行
func (c *Context) Status(code int) {
	c.Response.Writer.WriteHeader(code)
}

/************************************/
/***** 请求/响应头和cookie设置相关 *****/
/************************************/
// 设置任意响应头信息
func (c *Context) SetResponseHeader(key, value string) {
	// 传递为空自动清除
	if value == "" {
		c.Response.Writer.Header().Del(key)
		return
	}
	c.Response.Writer.Header().Set(key, value)
}

// 设置任意请求头信息
func (c *Context) SetRequestHeader(key, value string) {
	// 传递为空自动清除
	if value == "" {
		c.Request.Header.Del(key)
		return
	}
	c.Request.Header.Add(key, value)
}

// 添加cookie信息
// 响应阶段设置set-cookie头信息
// 请求阶段自动携带Cookie头信息到服务端
func (c *Context) SetCookie(cookieParams map[string]interface{}) {
	// 根据字典传递的参数判断，不存在的设置默认值
	var (
		name     string
		value    string
		maxAge   int
		path     string
		domain   string
		secure   bool
		httpOnly bool
		isClient bool // 是否http作为客户端添加cookie
	)
	value, _ = cookieParams["value"].(string)
	name, _ = cookieParams["name"].(string)
	maxAge, _ = cookieParams["maxAge"].(int)
	path, _ = cookieParams["path"].(string)
	domain, _ = cookieParams["domain"].(string)
	secure, _ = cookieParams["secure"].(bool)
	httpOnly, _ = cookieParams["httpOnly"].(bool)
	isClient, _ = cookieParams["isClient"].(bool)
	if name == "" {
		return
	}
	if path == "" {
		path = "/"
	}
	// 设置cookie值
	var cookie *http.Cookie = &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	}
	if isClient {
		// 客户端设置
		c.Request.AddCookie(cookie)
	} else {
		// 服务端设置
		// http.SetCookie(c.Response.Writer, cookie)
		// 或者可以通过header方法实现
		// c.SetResponseHeader("Set-Cookie", cookie.String())
		// 需要将空格转换为urlencode的形式
		ck := strings.Replace(cookie.String(), " ", "%20", -1)
		c.SetResponseHeader("Set-Cookie", ck)
	}
}

// 获取请求中已设置的cookie信息
func (c *Context) Cookie(key string) (val string, err error) {
	var cookieVal *http.Cookie
	cookieVal, err = c.Request.Cookie(key)
	if err != nil {
		// 获取出错
		return "", err
	}
	// 获取成功
	return cookieVal.Value, nil
}

/************************************/
/******** 内容协商相关 ****************/
/************************************/
