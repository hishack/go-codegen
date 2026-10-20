// Package errorutil 业务错误定义(示例工程的最小实现,与 shack 项目同构)。
// 这是手写代码,生成器不会修改本文件
package errorutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
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
	var be BusinessError
	if errors.As(err, &be) {
		return be
	}
	base := SystemError
	if IsClientError(err) {
		base = ParamsError
	}
	base.Message = err.Error()
	return base
}

// IsClientError 判断错误是否由客户端输入引起。
// 这类错误应记为 400 而不是 500,否则监控会把"用户传错参数"算成服务端故障
func IsClientError(err error) bool {
	if err == nil {
		return false
	}
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return true
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return true
	}
	var numErr *strconv.NumError
	if errors.As(err, &numErr) {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if errors.Is(err, multipart.ErrMessageTooLarge) {
		return true
	}
	return false
}

// 预置错误
var (
	Success     = BusinessError{Code: 0, Message: "success"}
	ParamsError = BusinessError{Code: 40000, Message: "参数错误"}
	SystemError = BusinessError{Code: 50000, Message: "系统繁忙,请稍后再试"}
)

// HTTPStatusOf 把错误映射为 HTTP 状态码,供**手写** handler 使用。
// 生成器产出的 handler 已固定为:参数/绑定类 400、业务处理失败 422,
// 不再调用本函数;这里保留给其它手写接口复用同一套约定。
// 避免所有错误都返回 200,让网关与监控能按状态码统计错误率
func HTTPStatusOf(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if IsClientError(err) {
		return http.StatusBadRequest
	}
	return http.StatusUnprocessableEntity
}
