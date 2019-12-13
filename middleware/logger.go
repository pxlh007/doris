package middleware

import (
	"doris"
	// "errors"
	"fmt"
)

// 中间件函数的格式
func Logger() doris.HandlerFunc {
	return func(c *doris.Context) error {
		// 前向中间件可以只写c.Next()前面部分
		fmt.Println("这里是日志中间件")
		c.Next()
		// fmt.Println("这里是日志中间件结束")
		return nil
	}
}
