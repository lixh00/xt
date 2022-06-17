package xt

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 返回数据包装
type responseData struct {
	Code   int    `json:"code"`
	Data   any    `json:"data"`
	Msg    string `json:"message"`
	ErrMsg string `json:"errMsg,omitempty"`
}

type resp struct {
	ctx *gin.Context
}

// 定义状态码
const (
	ERROR   = http.StatusInternalServerError
	SUCCESS = http.StatusOK
)

// R 工厂函数
func R(ctx *MultiTenantContext) *resp {
	x := ctx.Context
	x.Header("Tenant-Name", base64.StdEncoding.EncodeToString([]byte(ctx.TenantInfo.Name)))            // 租户名称
	x.Header("Tenant-Short-Name", base64.StdEncoding.EncodeToString([]byte(ctx.TenantInfo.ShortName))) // 租户简称
	x.Header("Tenant-Logo", base64.StdEncoding.EncodeToString([]byte(ctx.TenantInfo.Logo)))            // 租户logo
	x.Header("Tenant-Type-Code", ctx.TenantInfo.TypeCode)                                              // 租户类型代码

	return &resp{ctx: x}
}

// Result 手动组装返回结果
func (r resp) Result(code int, data any, msg, err string) {
	r.ctx.JSON(code, responseData{
		code,
		data,
		msg,
		err,
	})
}

// Ok 返回无数据的成功
func (r resp) Ok() {
	r.Result(SUCCESS, nil, "操作成功", "")
}

// OkWithMessage 返回自定义成功的消息
func (r resp) OkWithMessage(message string) {
	r.Result(SUCCESS, nil, message, "")
}

// OkWithData 自定义内容的成功返回
func (r resp) OkWithData(data any) {
	r.Result(SUCCESS, data, "操作成功", "")
}

// OkDetailed 自定义消息和内容的成功返回
func (r resp) OkDetailed(data any, message string) {
	r.Result(SUCCESS, data, message, "")
}

// Fail 返回默认失败
func (r resp) Fail() {
	r.Result(ERROR, nil, "操作失败", "")
}

// FailWithMessage 返回默认状态码自定义消息的失败
func (r resp) FailWithMessage(message string) {
	r.Result(ERROR, nil, message, "")
}

// FailWithError 返回自定义消息和内容的失败
func (r resp) FailWithError(msg string, err error) {
	r.Result(ERROR, nil, msg, err.Error())
}

// FailWithErrorAndCode 返回自定义消息和内容以及错误代码的失败
func (r resp) FailWithErrorAndCode(msg string, err error, code int) {
	r.Result(code, nil, msg, err.Error())
}

// FailWithMessageAndCode 返回自定义消息和状态码的失败
func (r resp) FailWithMessageAndCode(message string, code int) {
	r.Result(code, nil, message, "")
}
