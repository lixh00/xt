package xt

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// GinHandler Gin处理器
func GinHandler(handler MultiTenantHandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		mc := new(MultiTenantContext)
		mc.Context = ctx
		tid, err := getTenantId(ctx)
		if err != nil {
			ctx.JSON(http.StatusForbidden, response{http.StatusForbidden, nil, err.Error()})
			return
		}
		if db, exist := clientMap[tid]; exist {
			mc.DB = db
		} else {
			ctx.JSON(http.StatusForbidden, response{http.StatusForbidden, nil, "租户状态异常"})
			return
		}
		handler(mc)
	}
}

// 获取租户ID
func getTenantId(ctx *gin.Context) (uint, error) {
	var p tenantInfo
	if err := ctx.ShouldBindHeader(&p); err != nil {
		return 0, err
	}
	return p.TenantId, nil
}
