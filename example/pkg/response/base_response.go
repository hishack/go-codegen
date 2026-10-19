// Package response 统一响应结构(示例工程的最小实现,与 shack 项目同构)。
// 这是手写代码,生成器不会修改本文件
package response

import (
	"errors"

	"codegen-example/pkg/errorutil"
)

// BaseResponse 统一响应
type BaseResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func NewSuccessResponse[T any](data T) *BaseResponse[T] {
	return &BaseResponse[T]{
		Code:    errorutil.Success.Code,
		Message: errorutil.Success.Message,
		Data:    data,
	}
}

func NewErrorResponse[T any](err error) *BaseResponse[T] {
	newError := errorutil.BusinessError{}
	if errors.As(err, &newError) {
		return &BaseResponse[T]{Code: newError.Code, Message: newError.Message}
	}
	newError = errorutil.ConvertError(err)
	return &BaseResponse[T]{Code: newError.Code, Message: newError.Message}
}



// PageResponse 分页响应
type PageResponse[T any] struct {
	Records   []T `json:"records"`
	PageNum   int `json:"pageNum"`
	PageSize  int `json:"pageSize"`
	TotalPage int `json:"totalPage"`
	TotalRow  int `json:"totalRow"`
}
