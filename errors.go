// 定义错误编号，处理错误信息
package doris

import (
	"errors"
	"net/http"
)

// Errors
var HTTPErrorMessages = map[int]error{
	http.StatusOK:                    errors.New("Success"),
	http.StatusUnsupportedMediaType:  errors.New("Unsupported mediatype"),
	http.StatusNotFound:              errors.New("Not found"),
	http.StatusUnauthorized:          errors.New("Unauthorized"),
	http.StatusForbidden:             errors.New("Forbidden"),
	http.StatusMethodNotAllowed:      errors.New("Method not allowed"),
	http.StatusRequestEntityTooLarge: errors.New("Request entity too large"),
	http.StatusTooManyRequests:       errors.New("Too many requests"),
	http.StatusBadRequest:            errors.New("Bad request"),
	http.StatusBadGateway:            errors.New("Bad gateway"),
	http.StatusInternalServerError:   errors.New("Internal server error"),
	http.StatusRequestTimeout:        errors.New("Request timeout"),
	http.StatusServiceUnavailable:    errors.New("Service unavailable"),
}
