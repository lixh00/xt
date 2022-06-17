package xt

import "fmt"

// 二次封装的上下文处理

// MustParam 获取路径参数，如果不存在，就向前端返回错误
func (ctx *MultiTenantContext) MustParam(key string) (v string) {
	v = ctx.Param(key)
	if v == "" {
		R(ctx).FailWithError("参数错误", fmt.Errorf("参数%s不得为空", key))
		ctx.Abort()
		return
	}
	return
}
