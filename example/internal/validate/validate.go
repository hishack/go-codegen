// Package validate 提供各 logic 层共用的参数校验与错误构造。
//
// 单独成包的原因同 store:logic 层是生成器产物,手写代码不能放进去。
package validate

import (
	"fmt"
	"strings"

	"codegen-example/pkg/errorutil"
)

// wrap 在保留原始错误描述的前提下追加上下文。
// 直接用 BusinessError.WithMessage 会把描述整体覆盖掉,
// 例如把"用户不存在"顶成"id=999",丢掉关键信息。
func wrap(err errorutil.BusinessError, ctx string) errorutil.BusinessError {
	if ctx == "" {
		return err
	}
	return errorutil.BusinessError{Code: err.Code, Message: err.Message + ": " + ctx}
}

// Wrap 导出版本,供 logic 层调用
func Wrap(err errorutil.BusinessError, ctx string) errorutil.BusinessError {
	return wrap(err, ctx)
}

// 预置错误码
var (
	NotFound     = errorutil.BusinessError{Code: 40400, Message: "资源不存在"}
	Existed      = errorutil.BusinessError{Code: 40900, Message: "资源已存在"}
	Invalid      = errorutil.BusinessError{Code: 40001, Message: "参数不合法"}
	EmptyField   = errorutil.BusinessError{Code: 40002, Message: "必填字段为空"}
	OutOfRange   = errorutil.BusinessError{Code: 40003, Message: "数值超出允许范围"}
	TooLarge     = errorutil.BusinessError{Code: 41300, Message: "文件超过大小限制"}
	Unauthorized = errorutil.BusinessError{Code: 40100, Message: "未授权或凭证已失效"}
)

// Required 校验非空字符串,空白也算空
func Required(field, v string) error {
	if strings.TrimSpace(v) == "" {
		return wrap(EmptyField, field+" 不能为空")
	}
	return nil
}

// RequiredID 校验正数 id
func RequiredID(field string, id int64) error {
	if id <= 0 {
		return wrap(OutOfRange, field+" 必须为正数")
	}
	return nil
}

// RangeInt 校验整数区间,min/max 任一为 0 表示不检查该侧
func RangeInt(field string, v, min, max int) error {
	if v < min {
		return wrap(OutOfRange, fmt.Sprintf("%s 不能小于 %d(当前 %d)", field, min, v))
	}
	if max > 0 && v > max {
		return wrap(OutOfRange, fmt.Sprintf("%s 不能大于 %d(当前 %d)", field, max, v))
	}
	return nil
}

// RangeFloat 校验浮点区间
func RangeFloat(field string, v, min float64) error {
	if v < min {
		return wrap(OutOfRange, fmt.Sprintf("%s 不能小于 %g(当前 %g)", field, min, v))
	}
	return nil
}

// NonEmpty 校验切片非空
func NonEmpty(field string, n int) error {
	if n == 0 {
		return wrap(Invalid, field+" 至少需要一项")
	}
	return nil
}

// OneOf 校验取值在白名单内
func OneOf(field, v string, allowed ...string) error {
	for _, a := range allowed {
		if v == a {
			return nil
		}
	}
	return wrap(Invalid, fmt.Sprintf("%s 取值 %q 不合法,可选:%s", field, v, strings.Join(allowed, "/")))
}

// Page 归一化分页参数,返回页码与每页条数
func Page(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// Slice 对列表做内存分页
func Slice[T any](list []T, page, pageSize int) []T {
	page, pageSize = Page(page, pageSize)
	start := (page - 1) * pageSize
	if start >= len(list) {
		return []T{}
	}
	end := start + pageSize
	if end > len(list) {
		end = len(list)
	}
	return list[start:end]
}

// NotFoundErr 构造"资源不存在"错误
func NotFoundErr(what string, id int64) error {
	return wrap(NotFound, fmt.Sprintf("%s id=%d", what, id))
}
