package middleware

import (
	"doris"
	"errors"
)

// 异常捕获中间件
func Recovery() doris.HandlerFunc {
	return func(c *doris.Context) error {
		defer func() {
			// recover捕获panic异常
			if err := recover(); err != nil {
				// 修改响应码为500
				// 修改响应信息为捕获的异常信息
				c.Response.WriteHeader(500)
				doris.HTTPErrorMessages[500] = errors.New(err.(string))
			}
		}()
		c.Next()
		return nil
	}
}
