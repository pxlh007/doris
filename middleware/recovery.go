package middleware

import (
	"bytes"
	"doris"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"
)

// 全局变量定义
var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// 异常捕获中间件
func Recovery() doris.HandlerFunc {
	return func(c *doris.Context) error {
		defer func() {
			// recover捕获panic异常
			if err := recover(); err != nil {
				// 判断网络连接是否断开
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// 打印请求头信息和调用栈
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				headers := strings.Split(string(httpRequest), "\r\n")
				// 解析Authorization头
				for idx, header := range headers {
					current := strings.Split(header, ":")
					if current[0] == "Authorization" {
						headers[idx] = current[0] + ": *"
					}
				}

				// 组织日志信息
				if brokenPipe {
					// 网络断开
					// 修改响应信息为：请求头 + 捕获异常
					doris.HTTPErrorMessages[500] = errors.New(err.(string) + string(httpRequest))
					// 终止执行
					c.Abort()
				} else if c.Doris.Debug {
					// 调试模式
					// 获取stack信息[]byte
					stack := stack(3)
					// 修改响应信息为：捕获异常 + 函数调用链
					// 调用栈颜色配置：[\033[0;35m%s\033[0m]
					fmt.Printf("\n[\033[0;35m\n%s\n\n%s\033[0m]\n\n", strings.Join(headers, "\r\n"), string(stack))
					doris.HTTPErrorMessages[500] = errors.New(err.(string))
				} else {
					// 其他情况
					// 修改响应信息为：捕获异常
					doris.HTTPErrorMessages[500] = errors.New(err.(string))
				}

				// 修改响应码为500
				c.Response.WriteHeader(500)
			}
		}()
		c.Next()
		return nil
	}
}

// 从调用栈的第几段开始返回：借鉴gin框架
// 以一种相对有好的方式返回从指定的栈位开始的栈信息
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// 返回第n行被裁减之后的切片
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// 返回包含（持有）PC的函数名
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
