package xt

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

// DatabaseClientInfo 数据库连接配置
type DatabaseClientInfo struct {
	TenantId uint   // 租户ID
	Host     string // 数据库地址
	Port     int    // 数据库端口
	User     string // 数据库用户名
	Password string // 数据库密码
	Db       string // 数据库名称
}

// GetDSN 返回 MySQL 连接字符串
func (c DatabaseClientInfo) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Db)
}

// =====================================================================================================================

// MultiTenantContext 多租户上下文
type MultiTenantContext struct {
	*gin.Context
	TenantId uint
	DB       *xorm.Engine
}

// MultiTenantHandlerFunc 处理函数
type MultiTenantHandlerFunc func(*MultiTenantContext)

// TenantDBProvider 租户数据库连接信息提供者
type TenantDBProvider func() []DatabaseClientInfo

// TenantIdResolver 租户Id解析器
type TenantIdResolver func(*gin.Context) (uint, error)

// =====================================================================================================================

// Header里面的租户信息
type tenantInfo struct {
	UserId   uint `header:"userId"`                      // 用户ID
	TenantId uint `header:"tenantId" binding:"required"` // 租户Id
}

// 返回数据包装
type response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"message"`
}
