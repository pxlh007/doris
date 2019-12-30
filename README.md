## doris高性能、易扩展、又好用的go web开发框架

### 使用案例

```

package main

import (
	"doris"
	"doris/middleware"
	"fmt"
	"time"
)

const (
	Nanosecond  time.Duration = 1
	Microsecond               = 1000 * Nanosecond
	Millisecond               = 1000 * Microsecond
	Second                    = 1000 * Millisecond
	Minute                    = 60 * Second
	Hour                      = 60 * Minute
)

type Person struct {
	Name string `json:"username"`
	Sex  string `json:"sex"`
	Age  int    `json:"age"`
	Url  string `json:"url"`
}

func main() {

	// 实例化框架对象
	d := doris.New()
	d.ShowBanner = true
	d.Debug = false

	// 中间件部分测试
	// 全局中间件测试
	d.Use(middleware.Test())
	d.Use(middleware.Logger())

	// 测试路由冲突
	d.GET("/", func(c *doris.Context) error {
		c.IndentedJson(200, doris.D{"code": 200, "message": "success!"})
		return nil
	})

	// /:a/:e/:id
	d.GET("/:a/:e/:id", func(c *doris.Context) error {
		fmt.Println("/:a/:e/:id!")
		//c.Response.Writer.Write([]byte("SUCCESS!HELLO WORLD!"))
		return nil
	})

	// /:a/:id
	d.GET("/:a/:id", func(c *doris.Context) error {
		var p Person
		p.Name = "张胜男"
		p.Sex = "女"
		p.Age = 12
		p.Url = "http://www.geeksniu.com/<html></html>"
		c.PureJson(200, p)
		return nil
	})

	d.POST("/hello", func(c *doris.Context) error {
		// 测试参数绑定结构体
		type Louhao struct {
			AA       string `param:"aa"`
			bB       string `param:"bb"`
			sb       string `param:"sb"`
			Age      int    `param:"age"`
			nickname string `param:"nickname"`
			Address  string `param:"street"`
		}
		h := Louhao{}
		_ = c.Form(&h)
		_ = c.FormParam("age00")
		time.Sleep(50 * Millisecond)
		c.String(200, "hello world!")
		return nil
	})

	// /:a/:b/:c/:id
	d.GET("/:a/:b/:c/:id", func(c *doris.Context) error {
		fmt.Println("/:a/:b/:c/:id!")
		return nil
	})

	d.GET("/xml", func(c *doris.Context) error {
		var p Person
		p.Name = "张胜男"
		p.Sex = "女"
		p.Age = 12
		p.Url = "http://www.geeksniu.com/<html></html>"
		c.Xml(200, p)
		return nil
	})

	d.GET("/pong", func(c *doris.Context) error {
		fmt.Println("/pong!")
		return nil
	})

	// 这种形式生成的树会漏掉当前的路由
	// 期望的情况是从l分割开来
	// 下面三个互换位置再测
	d.POST("/hello/", func(c *doris.Context) error {
		fmt.Println("POST /hello/!")
		c.Response.Writer.Write([]byte("SUCCESS!HELLO!"))
		return nil
	})

	d.POST("/hello", func(c *doris.Context) error {
		fmt.Println("POST /hello!")
		c.Response.Writer.Write([]byte("SUCCESS!HELLO!"))
		return nil
	})

	d.POST("/hell", func(c *doris.Context) error {
		fmt.Println("POST /hell!")
		//c.Response.Writer.Write([]byte("SUCCESS!HELLO WORLD!"))
		return nil
	})

	// 静态路由和参数路由冲突测试
	// 静态路由优先的原则
	d.POST("/hello/name", func(c *doris.Context) error {
		fmt.Println("POST /hello/name!")
		c.Response.Writer.Write([]byte("SUCCESS!HELLO!"))
		return nil
	})

	// bug curl -XPOST localhost:8002/hello/name9
	// 404错误
	// 期望是/hello/:name

	// 注意如果注册两个相同的路由会出现报错
	d.POST("/hello/name/", func(c *doris.Context) error {
		c.Response.Writer.Write([]byte("SUCCESS!HELLO!"))
		return nil
	})
	d.POST("/hello/name/", func(c *doris.Context) error {
		c.Response.Writer.Write([]byte("SUCCESS!HELLO!"))
		return nil
	})

	// 参数路由
	d.POST("/hello/:name", func(c *doris.Context) error {
		c.Response.Writer.Write([]byte("SUCCESS! /hello/:name!"))
		return nil
	})

	// 参数路由
	d.POST("/hello/:age", func(c *doris.Context) error {
		c.Response.Writer.Write([]byte("SUCCESS!/hello/:age!"))
		return nil
	})

	// 下面两个互换位置再测
	// 参数路由
	d.POST("/hello/:age/:sex", func(c *doris.Context) error {
		c.Response.Writer.Write([]byte("SUCCESS!/hello/:age/:sex!"))
		return nil
	})

	// 参数路由
	d.POST("/hello/:age/sex/:sex", func(c *doris.Context) error {
		c.Response.Writer.Write([]byte("SUCCESS!/hello/:age/sex/:sex!"))
		return nil
	})

	// 全量路由
	d.POST("/hello/*/", func(c *doris.Context) error {
		c.Response.Writer.Write([]byte("SUCCESS!/hello/*/!"))
		return nil
	})

	// 定义组路由
	v1 := d.Group("/v1/", func(c *doris.Context) error {
		c.Response.Writer.Write([]byte("SUCCESS!/hello/*/!"))
		return nil
	})

	// 组中间件测试
	v1.Use(middleware.Test())
	v1.Pre(middleware.Logger())

	// localhost:8002/v1/hello
	// 这个群组测试失败，需要继续验证
	v1.POST("/hello", func(c *doris.Context) error {
		fmt.Println("GROUP /v1/hello")
		c.Response.Writer.Write([]byte("SUCCESS!/v1/hello"))
		return nil
	})

	// 这个群组测试失败，需要继续验证
	// 已经修复
	v1.POST("/hello/LOUHAO", func(c *doris.Context) error {
		fmt.Println("GROUP /v1/LOUHAO")
		c.Response.Writer.Write([]byte("SUCCESS!/v1/LOUHAO"))
		return nil
	})

	v1.POST("/hell", func(c *doris.Context) error {
		fmt.Println("GROUP /v1/hell")
		c.Response.Writer.Write([]byte("SUCCESS!/v1/hell"))
		return nil
	})

	v1.POST("/hello/:id/", func(c *doris.Context) error {
		fmt.Println("GROUP /v1/hello/:id/")
		c.Response.Writer.Write([]byte("SUCCESS!/v1/hello/:id/"))
		return nil
	})

	// debug
	// d.ScanTrees()           // 打印路由树
	d.Run("localhost:9527") // listen and serve on 0.0.0.0:8080
}

```