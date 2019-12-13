package middleware

import (
	"doris"
	// "errors"
	"fmt"
)

// 中间件函数的格式
func Test() doris.HandlerFunc {
	return func(c *doris.Context) error {
		// fmt.Println("这里是测试中间件开始")
		c.Next()
		// 后向中间件可以只写c.Next()后面部分
		fmt.Println("这里是测试中间件结束")
		return nil
	}
}
