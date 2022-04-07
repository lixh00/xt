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
		tid, ti, err := tenantIdResolver(ctx)
		if err != nil {
			ctx.JSON(http.StatusForbidden, response{http.StatusForbidden, nil, err.Error()})
			ctx.Abort()
			return
		}
		mc.TenantId = tid
		mc.TenantInfo = ti
		if db, exist := clientMap[tid]; exist {
			mc.DB = db
		} else {
			ctx.JSON(http.StatusForbidden, response{http.StatusForbidden, nil, "租户状态异常"})
			ctx.Abort()
			return
		}
		handler(mc)
	}
}

// 获取租户ID
func getTenantId(ctx *gin.Context) (id uint, info TenantInfo, err error) {
	var p tenantInfo
	if err = ctx.ShouldBindHeader(&p); err != nil {
		return
	}
	return p.TenantId, clientInfoMap[p.TenantId], nil
}

// GetAllTenantInfo 获取所有租户信息
func GetAllTenantInfo() (datas []TenantInfo) {
	for _, v := range clientInfoMap {
		datas = append(datas, v)
	}
	return
}
