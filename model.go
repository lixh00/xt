package xt

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DatabaseClientInfo 数据库连接配置
type DatabaseClientInfo struct {
	TenantId string     `json:"tenantId"` // 租户ID
	Info     TenantInfo `json:"info"`     // 租户信息
	Host     string     `json:"host"`     // 数据库地址
	Port     int        `json:"port"`     // 数据库端口
	User     string     `json:"user"`     // 数据库用户名
	Password string     `json:"password"` // 数据库密码
	Db       string     `json:"db"`       // 数据库名称
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
	TenantId   string     // 租户Id
	TenantInfo TenantInfo // 租户信息
	DB         *gorm.DB   // 数据库连接
}

// MultiTenantHandlerFunc 处理函数
type MultiTenantHandlerFunc func(*MultiTenantContext)

// TenantDBProvider 租户数据库连接信息提供者
type TenantDBProvider func() []DatabaseClientInfo

// SyncModelsAfter 同步模型完成之后调用函数
type SyncModelsAfter func(db *gorm.DB, tenantId string) error

// TenantIdResolver 租户Id解析器
type TenantIdResolver func(*gin.Context) (string, TenantInfo, error)

// TenantInfo 租户信息
type TenantInfo struct {
	Id        string      `json:"id"`        // 租户ID
	Name      string      `json:"name"`      // 租户全名
	ShortName string      `json:"shortName"` // 租户简称
	Logo      string      `json:"logo"`      // 租户logo
	TypeCode  string      `json:"type"`      // 租户类型
	Extended  interface{} `json:"extended"`  // 自行维护的扩展信息
}

// =====================================================================================================================

// Header里面的租户信息
type tenantInfo struct {
	UserId   string `header:"userId"`                      // 用户ID
	TenantId string `header:"tenantId" binding:"required"` // 租户Id
}

// 返回数据包装
type response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"message"`
}
