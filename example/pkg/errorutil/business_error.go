// Package errorutil 业务错误定义(示例工程的最小实现,与 shack 项目同构)。
// 这是手写代码,生成器不会修改本文件
package errorutil

import (
	"errors"
	"fmt"
)

// BusinessError 业务错误
type BusinessError struct {
	Code    int
	Message string
}

func (e BusinessError) Error() string {
	return fmt.Sprintf("errorCode: %d, message: %s", e.Code, e.Message)
}

// WithMessage 携带自定义消息
func (e BusinessError) WithMessage(message string) BusinessError {
	e.Message = message
	return e
}

// ConvertError 将任意错误转换为业务错误
func ConvertError(err error) BusinessError {
	newErr := BusinessError{}
	if errors.As(err, &newErr) {
		return newErr
	}
	newErr = SystemError
	newErr.Message = err.Error()
	return newErr
}

// 预置错误
var (
	Success     = BusinessError{Code: 0, Message: "success"}
	ParamsError = BusinessError{Code: 40000, Message: "参数错误"}
	SystemError = BusinessError{Code: 50000, Message: "系统繁忙,请稍后再试"}
)
