package xt

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinHandler Gin处理器
func GinHandler(handler MultiTenantHandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		mc := new(MultiTenantContext)
		mc.Context = ctx
		tid, ti, err := tenantIdResolver(ctx)
		if err != nil {
			ctx.JSON(http.StatusForbidden, response{http.StatusForbidden, nil, err.Error()})
			//ctx.Abort()
			return
		}
		mc.TenantId = tid
		mc.TenantInfo = ti
		clientMapLock.RLock()
		db, exist := clientMap[tid]
		clientMapLock.RUnlock()
		if exist {
			mc.DB = db
		} else {
			ctx.JSON(http.StatusForbidden, response{http.StatusForbidden, nil, "租户状态异常"})
			//ctx.Abort()
			return
		}
		mc.Reply = R(mc)
		handler(mc)
	}
}

// 获取租户ID
func getTenantId(ctx *gin.Context) (id string, info TenantInfo, err error) {
	var p tenantInfo
	if err = ctx.ShouldBindHeader(&p); err != nil {
		return
	}
	clientMapLock.RLock()
	info = clientInfoMap[p.TenantId]
	clientMapLock.RUnlock()
	return p.TenantId, info, nil
}

// GetAllTenantInfo 获取所有租户信息
func GetAllTenantInfo() (datas []TenantInfo) {
	clientMapLock.RLock()
	defer clientMapLock.RUnlock()
	for _, v := range clientInfoMap {
		datas = append(datas, v)
	}
	return
}
