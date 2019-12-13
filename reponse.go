// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package doris

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

// ResponseWriter接口定义
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	// Returns the HTTP response status code of the current request.
	Status() int

	// Returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// Writes the string into the response body.
	WriteString(string) (int, error)

	// Returns true if the response body was already written.
	Written() bool

	// Forces to write the http header (status code + headers).
	WriteHeaderNow()

	// get the http.Pusher for server push
	Pusher() http.Pusher
}

type Response struct {
	size   int
	status int
	Writer http.ResponseWriter
}

// Response结构实现了上述接口
var _ ResponseWriter = &Response{}

func (w *Response) reset(writer http.ResponseWriter) {
	w.Writer = writer
	w.size = noWritten
	w.status = defaultStatus
}

func (w *Response) WriteHeader(code int) {
	if code > 0 && w.status != code {
		if w.Written() {
			//debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
		}
		w.status = code
	}
}

func (w *Response) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
		w.Writer.WriteHeader(w.status)
	}
}

func (w *Response) Write(data []byte) (n int, err error) {
	w.WriteHeaderNow()
	n, err = w.Writer.Write(data)
	w.size += n
	return
}

func (w *Response) WriteString(s string) (n int, err error) {
	w.WriteHeaderNow()
	n, err = io.WriteString(w.Writer, s)
	w.size += n
	return
}

func (w *Response) Status() int {
	return w.status
}

func (w *Response) Size() int {
	return w.size
}

func (w *Response) Written() bool {
	return w.size != noWritten
}

// Hijack implements the http.Hijacker interface.
func (w *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.size < 0 {
		w.size = 0
	}
	return w.Writer.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotify interface.
func (w *Response) CloseNotify() <-chan bool {
	return w.Writer.(http.CloseNotifier).CloseNotify()
}

// Flush implements the http.Flush interface.
func (w *Response) Flush() {
	w.WriteHeaderNow()
	w.Writer.(http.Flusher).Flush()
}

func (w *Response) Pusher() (pusher http.Pusher) {
	if pusher, ok := w.Writer.(http.Pusher); ok {
		return pusher
	}
	return nil
}

func (w *Response) Header() http.Header {
	return nil
}
