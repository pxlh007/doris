package doris

import (
	"net/http"
	"path"
	"regexp"
	"strings"
)

type (
	// 路由组结构用于配置路由，与基础路径和一个函数处理链相关联
	// 包含doris框架实例
	RouteGroup struct {
		Handlers HandlersChain // 组处理链
		basePath string        // 基础路径
		doris    *Doris        // 框架对象
		root     bool          // 是否为根节点
	}
	// 定义了所有路由的处理接口
	// 包含单个的路由和组路由等
	IRouter interface {
		// 接口组合
		IRoutes

		// 注册组的方法
		Group(string, ...HandlerFunc) *RouteGroup
	}
	// 定义了所有单路由处理接口
	IRoutes interface {
		// 注册中间件
		Use(...HandlerFunc) IRoutes
		Pre(...HandlerFunc) IRoutes

		// 注册各HTTP方法
		// 对应路由
		Handle(string, string, ...HandlerFunc) IRoutes
		Any(string, ...HandlerFunc) IRoutes
		CONNECT(string, ...HandlerFunc) IRoutes
		GET(string, ...HandlerFunc) IRoutes
		POST(string, ...HandlerFunc) IRoutes
		DELETE(string, ...HandlerFunc) IRoutes
		PATCH(string, ...HandlerFunc) IRoutes
		PUT(string, ...HandlerFunc) IRoutes
		OPTIONS(string, ...HandlerFunc) IRoutes
		HEAD(string, ...HandlerFunc) IRoutes

		// 注册静态文件
		// 对应路由
		StaticFile(string, string) IRoutes
		Static(string, string) IRoutes
		StaticFS(string, http.FileSystem) IRoutes
	}
)

// 声明一个类型但是不使用
// 相当于说明RouterGroup结构实现了IRouter接口类型
// 是IRouter接口的一个实现
// 但是这个变量由于被指定为_因此不会被使用
var _ IRouter = &RouteGroup{} // _代表丢弃不用

// 组的Pre方法同时也是doris
// 实例Pre方法属于组内的子路由
// 参数是各个中间件的列表
func (group *RouteGroup) Pre(middleware ...HandlerFunc) IRoutes {
	// 追加到追加到handlers前面
	group.Handlers = group.combineHandlers(middleware, true)
	return group.obj()
}

// 组的Use方法同时也是doris
// 实例Use方法属于组内的子路由
// 参数是各个中间件的列表
func (group *RouteGroup) Use(middleware ...HandlerFunc) IRoutes {
	group.Handlers = append(group.Handlers, middleware...)
	return group.obj()
}

// 组方法实现分组路由
// 同一个组的路由共用一组中间件函数
// 分组返回组的指针
func (group *RouteGroup) Group(relativePath string, handlers ...HandlerFunc) *RouteGroup {
	return &RouteGroup{
		Handlers: group.combineHandlers(handlers, false),
		basePath: group.calculateAbsolutePath(relativePath),
		doris:    group.doris,
	}
}

// 实际的处理路由组的函数
func (group *RouteGroup) handle(httpMethod, relativePath string, handlers ...HandlerFunc) IRoutes {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers, false)
	// debugPrintMessage("absolutePath", absolutePath, true)
	// debugPrintMessage("handlers", handlers, true)
	group.doris.addRoute(httpMethod, absolutePath, handlers)
	return group.obj()
}

// 函数本身用于给请求注册处理器和中间件
// 但是HTTP对应的各个方法都有相应的方法
// 此方法通常用于内部通信中（比如和代理服务的通信）
func (group *RouteGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) IRoutes {
	if matches, err := regexp.MatchString("^[A-Z]+$", httpMethod); !matches || err != nil {
		panic("http method " + httpMethod + " is not valid")
	}
	return group.handle(httpMethod, relativePath, handlers...)
}

// CONNECT方法
// 一般用于代理服务器转发
// 普通的网页应用不会使用
func (group *RouteGroup) CONNECT(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("CONNECT", relativePath, handlers...)
}

// GET方法
// 注册GET路由
func (group *RouteGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("GET", relativePath, handlers...)
}

// POST方法
// 注册POST路由
func (group *RouteGroup) POST(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("POST", relativePath, handlers...)
}

// PUT方法
// 注册PUT路由
func (group *RouteGroup) PUT(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("PUT", relativePath, handlers...)
}

// DELETE方法
// 注册GDELETE路由
func (group *RouteGroup) DELETE(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("DELETE", relativePath, handlers...)
}

// HEAD方法
// 注册HEAD路由
func (group *RouteGroup) HEAD(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("HEAD", relativePath, handlers...)
}

// OPTIONS方法
// 注册GET路由
func (group *RouteGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("OPTIONS", relativePath, handlers...)
}

// PATCH方法
// 注册PATCH路由
func (group *RouteGroup) PATCH(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("PATCH", relativePath, handlers...)
}

// TRACE方法
// 注册TRACE路由
func (group *RouteGroup) TRACE(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("TRACE", relativePath, handlers...)
}

// Any方法
// 注册HTTP的全部路由
func (group *RouteGroup) Any(relativePath string, handlers ...HandlerFunc) IRoutes {
	group.handle("CONNECT", relativePath, handlers...)
	group.handle("GET", relativePath, handlers...)
	group.handle("POST", relativePath, handlers...)
	group.handle("PUT", relativePath, handlers...)
	group.handle("DELETE", relativePath, handlers...)
	group.handle("HEAD", relativePath, handlers...)
	group.handle("OPTIONS", relativePath, handlers...)
	group.handle("PATCH", relativePath, handlers...)
	group.handle("TRACE", relativePath, handlers...)
	return group.obj()
}

// Match方法
// 注册部分HTTP方法的路由
// 用于定制支持的方法
func (group *RouteGroup) Match(methods []string, relativePath string, handlers ...HandlerFunc) IRoutes {
	for _, method := range methods {
		group.handle(method, relativePath, handlers...)
	}
	return group.obj()
}

// 处理单个文件的静态路由
// 调用方式tree.StaticFile("favicon.ico", "./resources/favicon.ico")
func (group *RouteGroup) StaticFile(relativePath, filepath string) IRoutes {
	// 包含:或者*视为非法字符
	// 直接抛出异常到外层
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static file")
	}
	// 定义闭包函数调用
	// context中的File方法
	handler := func(c *Context) error {
		c.File(filepath)
		return nil
	}
	// 注册GET和HEAD请求处理器
	group.GET(relativePath, handler)
	group.HEAD(relativePath, handler)
	return group.obj()
}

// Static方法用于从现有的文件系统中提供文件服务
// 在底层用的是http.FileServer因此当文件不存在时将使用http.NotFound
// 而不是Router路由的NotFound handler
// 使用案例：router.Static("/static", "/var/www")
func (group *RouteGroup) Static(relativePath, root string) IRoutes {
	return group.StaticFS(relativePath, Dir(root, false))
}

// StaticFS工作原理类似Static
// 只是使用的定制的http.FileSystem结构
func (group *RouteGroup) StaticFS(relativePath string, fs http.FileSystem) IRoutes {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	handler := group.createStaticHandler(relativePath, fs)
	urlPattern := path.Join(relativePath, "/*filepath")

	// 注册GET和HEAD方法的handler
	group.GET(urlPattern, handler)
	group.HEAD(urlPattern, handler)
	return group.obj()
}

// 创建静态文件处理函数
func (group *RouteGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := group.calculateAbsolutePath(relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(c *Context) error {
		if _, nolisting := fs.(*onlyfilesFS); nolisting {
			c.Response.WriteHeader(http.StatusNotFound)
		}

		file := c.Param("filepath")
		// 检查文件是否存在以及是否有权限访问
		if _, err := fs.Open(file.(string)); err != nil {
			c.Response.WriteHeader(http.StatusNotFound)
			// 将没有路由的函数链赋值给ctx的处理链
			c.handlers = group.doris.noRoute
			// 复位中间键索引值
			c.index = -1
			return err
		}

		// 调用文件服务的ServeHTTP方法
		fileServer.ServeHTTP(c.Response.Writer, c.Request)
		return nil
	}
}

// 合并处理器的方法
// 根据标志位分：前向合并和后向合并
func (group *RouteGroup) combineHandlers(handlers HandlersChain, isBefore bool) HandlersChain {
	finalSize := len(group.Handlers) + len(handlers)
	if finalSize >= int(abortIndex) {
		panic("too many handlers")
	}
	mergedHandlers := make(HandlersChain, finalSize)
	if isBefore {
		// 前向合并
		copy(mergedHandlers, handlers)
		copy(mergedHandlers[len(handlers):], group.Handlers)
	} else {
		// 后向合并
		copy(mergedHandlers, group.Handlers)
		copy(mergedHandlers[len(group.Handlers):], handlers)
	}
	return mergedHandlers
}

// 计算绝对路径的方法
func (group *RouteGroup) calculateAbsolutePath(relativePath string) string {
	// 调用公共的方法合并路径
	return JoinPaths(group.basePath, relativePath)
}

// 返回组的对象或者框架对象obj
func (group *RouteGroup) obj() IRoutes {
	if group.root {
		// 若为根节点返回框架实例
		return group.doris
	}
	// 否则返回group组实例
	return group
}
