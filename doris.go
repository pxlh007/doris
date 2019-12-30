package doris

import (
	//"bytes"
	//"crypto/tls"
	//"errors"
	"fmt"
	//"io"
	//"io/ioutil"
	//"net"
	"net/http"
	//"net/url"
	//"path"
	//"path/filepath"
	//"reflect"
	//"runtime"
	//"time"
	"strconv"
	"strings"
	"sync"

	"github.com/pxlh007/logger"
)

type (
	// doris结构
	Doris struct {
		RouteGroup                              // 组合继承组结构和方法
		maxParam         *int                   // 路由中的最大参数数
		trees            trees                  // Method路由树
		pool             sync.Pool              // 用于复用context上下文等对象
		HTTPErrorHandler HTTPErrorHandler       // http错误处理函数
		Config           map[string]interface{} // 全局用户配置器
		Debug            bool                   // 是否处于调试模式
		autoSlash        bool                   // 是否自动在路径的结尾添加'/'
		noRoute          HandlersChain          // 不存在路由处理链
		noMethod         HandlersChain          // 不存在方法处理链
		allowMethod      []string               // 允许的HTTP方法列表
		Logger           *logger.Logger         // 全局日志记录器
		ShowBanner       bool                   // 是否显示banner信息
		// beforeHandlers   HandlersChain       // 全局前向中间件调用链
		// afterHandlers    HandlersChain       // 全局后向中间件调用链
	}
	// 请求过程中出现的错误提示
	HTTPError struct {
		Code    int         `json:"-"`       // 错误编号
		Message interface{} `json:"message"` // 错误信息
	}
	// 定义http请求处理函数
	HandlerFunc func(*Context) error
	// 定义HandlerFunc数组
	HandlersChain []HandlerFunc
	// 集中式http错误处理器
	HTTPErrorHandler func(error, Context)
	// map[string]interface{}的简短定义
	D map[string]interface{}
)

// 定义方法列表
var httpMethods []string = []string{
	"GET",
	"POST",
	"PUT",
	"DELETE",
	"OPTIONS",
	"HEAD",
	"PATCH",
}

// 启动时打印框架版本和banner信息
const (
	// Version of Doris
	Version = "v1.0.0"
	website = "https://www.doris.com"
	// http://patorjk.com/software/taag/#p=display&f=Small%20Slant&t=Doris
	banner = `
       __           _     
  ____/ /___  _____(_)____
 / __  / __ \/ ___/ / ___/
/ /_/ / /_/ / /  / (__  ) 
\__,_/\____/_/  /_/____/ %s
High performance, High scalability Go web framework
website: %s
_________________________________Author: JonahLou__
                                    
`
)

// 实例化框架对象函数
func New() *Doris {
	doris := &Doris{
		maxParam:    new(int),
		Logger:      logger.NewLogger(),
		allowMethod: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS", "HEAD"},
	}
	// 注册默认404和405函数
	doris.NoMethod(defaultNoMethod)
	doris.NoRoute(defaultNoRoute)
	// 设置错误级别
	//doris.Logger.SetLevel(log.ERROR)
	doris.RouteGroup.doris = doris
	doris.pool.New = func() interface{} {
		return doris.allocateContext()
	}
	return doris
}

// 注册默认404函数
func defaultNoRoute(c *Context) error {
	return serveError(c, 404, "not found!")
}

// 注册默认405函数
func defaultNoMethod(c *Context) error {
	return serveError(c, 405, "method not allowed!")
}

// 分配一个新的上下文实例
func (doris *Doris) allocateContext() *Context {
	response := new(Response)
	return &Context{doris: doris, Response: response}
}

// Pre添加前中间件
func (doris *Doris) Pre(handlers ...HandlerFunc) IRoutes {
	// 追加处理器到beforeHandlers
	// 追加处理器到handlers首部
	return doris.RouteGroup.Pre(handlers...)
}

// Use添加后中间件
func (doris *Doris) Use(handlers ...HandlerFunc) IRoutes {
	debugPrintMessage("handlers", handlers, doris.Debug)
	debugPrintMessage("调试信息", "__debug__", doris.Debug)
	// 给错误处理器添加中间件
	doris.noRoute = append(doris.noRoute, handlers...)
	doris.noMethod = append(doris.noMethod, handlers...)
	// 追加处理器到handlers尾部
	return doris.RouteGroup.Use(handlers...)
}

// NoRoute用于注册没有路由时候的处理方法默认是404
func (doris *Doris) NoRoute(handlers ...HandlerFunc) {
	doris.noRoute = append(doris.noRoute, handlers...)
}

// NoRoute用于注册没有方法的处理默认405
func (doris *Doris) NoMethod(handlers ...HandlerFunc) {
	doris.noMethod = append(doris.noMethod, handlers...)
}

// 添加路由方法
func (doris *Doris) addRoute(method, path string, handlers HandlersChain) {
	// 初始断言
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(len(handlers) > 0, "there must be at least one handler")
	assert1(doris.validMethod(method), "method not support")
	// 注册路由
	if root := doris.trees.get(method); root != nil { // 树存在
		root.debug = doris.Debug // 设置调试参数
		root.addRoute(path, handlers)
	} else { // 构建树
		debugPrintMessage("创建树", "__print__", doris.Debug)
		debugPrintMessage("method", method, doris.Debug)
		node := new(node)
		node.fullPath = "/"
		node.label = '/'
		node.prefix = "/"
		node.debug = doris.Debug // 设置调试参数
		node.addRoute(path, handlers)
		if len(doris.trees) == 0 {
			doris.trees = make(map[string]*tree)
		}
		doris.trees[method] = &tree{
			root:   node,
			method: method,
			doris:  doris,
		}
	}
}

// 运行框架程序绑定端口
func (doris *Doris) Run(addr ...string) (err error) {
	address := ResolveAddress(addr)

	// 判断是否展示banner
	if doris.ShowBanner {
		// 显示banner信息
		fmt.Printf(banner, Version, website)
	}

	// 存在多监听的时候只取第一个
	pi := strings.Index(addr[0], ":")
	port := addr[0][pi+1:]

	// 打印引导信息
	fmt.Printf("⇨ http server started on \033[0;32m[::]:%s\033[0m \n\n", port)
	err = http.ListenAndServe(address, doris)

	return
}

// 实现ServerHTTP接口
func (doris *Doris) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := doris.pool.Get().(*Context)
	c.Response.reset(w)
	c.Request = req
	doris.handleHTTPRequest(c)
	doris.pool.Put(c)
}

// 实际处理http请求的地方
func (doris *Doris) handleHTTPRequest(c *Context) {
	httpMethod := c.Request.Method
	// 判断是否允许
	if !InSlice(httpMethod, doris.allowMethod) {
		c.handlers = doris.noMethod
		c.index = -1 // 默认设置为-1
		c.Next()     // 执行函数处理链
		return
	}
	rPath := c.Request.URL.Path
	debugPrintMessage("rPath", rPath, doris.Debug)
	// 查找method树
	if tree, ok := doris.trees[httpMethod]; ok {
		// 方法树存在
		nodev := tree.root.find(rPath)
		if nodev != nil && nodev.handlers != nil {
			c.handlers = nodev.handlers
			c.params = SliceToMap(nodev.params, nodev.pvalues)
			c.fullPath = nodev.fullPath
			c.index = -1 // 默认设置为-1
			c.Next()     // 执行函数处理链
			return
		}
	}
	// 方法树不存在
	c.handlers = doris.noRoute
	c.index = -1 // 默认设置为-1
	c.Next()     // 执行函数处理链
	return
}

// 断言函数
func assert1(guard bool, text string) {
	if !guard { // 弹出异常并统一捕获
		panic(text)
	}
}

// 校验Method的有效性
func (doris *Doris) validMethod(method string) bool {
	for _, m := range doris.allowMethod {
		if m == method {
			return true
		}
	}
	return false
}

// 处理错误
func serveError(c *Context, code int, defaultMessage string) error {
	c.Json(code, D{"code": code, "message": defaultMessage})
	return nil
}

// 循环字典树
func (doris *Doris) ScanTrees() {
	var childContainer []*node
	for _, method := range httpMethods {
		if tree, ok := doris.trees[method]; ok {
			root := tree.root
			level := 0 // 层级标识
			fmt.Println("=======" + method + "树路由开始======\n")
			childContainer = append(childContainer, root)
			for len(childContainer) > 0 {
				level++
				childContainer = debugPrintLevel(childContainer, level)
			}
			fmt.Println("\n=======" + method + "树路由结束======\n")
		}
	}
}

// 打印路由层信息
func debugPrintLevel(nodes []*node, level int) (childContainer []*node) {
	// 打印当前层获取下一层节点
	lstr := strconv.Itoa(level)
	if lstr == "" {
		lstr = "0"
	}
	fmt.Println("当前层级level : 第【" + lstr + "】层开始\n")
	if len(nodes) > 0 {
		for _, child := range nodes {
			if len(child.children) > 0 {
				childContainer = append(childContainer, child.children...)
			}
			fmt.Println("\n=========================当前节点开始============================\n")
			fmt.Printf("===节点类型：%v, 节点label：%v, 节点前缀：%v, 父节点：%v, 子节点：%v, 全路径：%v, 参数列表：%v, 节点处理链：%v===", child.nType, string(child.label), child.prefix, child.parent, child.children, child.fullPath, child.pList, child.handlers)
			fmt.Println("\n=========================当前节点结束============================\n")
		}
	}
	fmt.Println("\n当前层级level : 第【" + lstr + "】层结束\n")
	return
}
