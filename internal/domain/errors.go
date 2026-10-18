package domain

import "errors"

// 跨层共享的哨兵错误。
//
// domain 用业务语言说明出了什么问题；adapter 决定如何在自己的传输层呈现
// （例如 HTTP 状态码）。调用方用 errors.Is 判断，adapter 可以为其附加上下文。
var (
	ErrNotFound   = errors.New("resource not found")
	ErrConflict   = errors.New("resource conflict")
	ErrValidation = errors.New("validation failed")
)
