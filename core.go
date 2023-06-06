package xt

import (
	"errors"
	"github.com/lixh00/xt/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
	"time"
)

var (
	clientMap                                map[string]*gorm.DB           // 存储所有的数据库连接
	clientMapLock                            sync.Mutex                    // 一把锁
	clientDbInfoMap                          map[string]DatabaseClientInfo // 存储所有的租户数据库连接信息
	clientInfoMap                            map[string]TenantInfo         // 租户信息
	syncModels                               []interface{}                 // 同步的模型
	syncModelsLock                           sync.Mutex                    // 一把锁
	autoSyncClient                           bool                          // 是否自动同步连接配置
	autoSyncClientTime                       int64                         // 自动同步连接配置的时间间隔
	syncModelsBefore                         SyncModelsBefore              // 同步模型前的回调
	syncModelsAfter                          SyncModelsAfter               // 同步模型后的回调
	syncModelsDisable                        bool                          // 是否禁用同步模型
	tenantDBProvider                         TenantDBProvider              // 租户数据库提供者
	tenantIdResolver                         TenantIdResolver              // 租户ID解析器
	logs                                     logger.Interface              // 日志输出
	disableForeignKeyConstraintWhenMigrating bool                          // 禁用自动创建外键约束
)

func init() {
	clientMap = make(map[string]*gorm.DB)
	clientInfoMap = make(map[string]TenantInfo)
	clientDbInfoMap = make(map[string]DatabaseClientInfo)
	syncModels = make([]interface{}, 0)
	autoSyncClientTime = 5 // 默认五分钟同步一次
	logs = logger.Default
}

// SetLogger 设置日志输出工具
func SetLogger(out logger.Interface) {
	logs = out
}

// SetSyncClientTime 自动同步连接配置的时间间隔(单位:分钟)
func SetSyncClientTime(minute int64) {
	autoSyncClientTime = minute
}

// DisableSyncModels 设置同步模型是否禁用
func DisableSyncModels(disable bool) {
	syncModelsDisable = disable
}

// Init 初始化
func Init(p TenantDBProvider, i TenantIdResolver, auto ...bool) error {
	if p == nil {
		return errors.New("db provider is nil")
	}
	tenantDBProvider = p
	if i == nil {
		tenantIdResolver = getTenantId
	} else {
		tenantIdResolver = i
	}
	clients := tenantDBProvider()
	for _, c := range clients {
		if err := Add(c); err != nil {
			return err
		}
	}
	// 启用了自动同步连接池
	if len(auto) > 0 {
		autoSyncClient = auto[0]
		go autoSyncClientHandle()
	}
	return nil
}

//	SetSyncModelsAfter
//	@description: 设置同步模型后的回调
//	@param handle
func SetSyncModelsAfter(handle SyncModelsAfter) {
	syncModelsAfter = handle
}

//	SetSyncModelsBefore
//	@description: 设置同步模型前的回调
//	@param handle
func SetSyncModelsBefore(handle SyncModelsBefore) {
	syncModelsBefore = handle
}

//	SetDisableForeignKeyConstraintWhenMigrating
//	@description: 设置是否禁用自动创建外键约束
//	@param flag
func SetDisableForeignKeyConstraintWhenMigrating(flag bool) {
	disableForeignKeyConstraintWhenMigrating = flag
}

// 自动同步连接配置
func autoSyncClientHandle() {
	for autoSyncClient {
		clients := tenantDBProvider()
		// 先筛选出已经不存在的租户
		var inIds, newIds []string
		for tenantId, _ := range clientInfoMap {
			inIds = append(inIds, tenantId)
		}
		for _, client := range clients {
			newIds = append(newIds, client.TenantId)
		}
		// 算差集，找出已经删除的租户
		needClearIds := utils.Difference(inIds, newIds)
		// 清理租户
		for _, tenantId := range needClearIds {
			clientMapLock.Lock()
			delete(clientMap, tenantId)
			clientMapLock.Unlock()
			delete(clientInfoMap, tenantId)
		}

		// 更新租户信息
		for _, c := range clients {
			// 循环已存在数据，匹配是否需要更新
			for k, old := range clientInfoMap {
				if c.TenantId == k {
					// 判断租户信息是否还一致，只要有一项不一致就给改掉 TODO 后面看看需不需要加锁
					newInfo := c.Info
					if newInfo.Name == old.Name && newInfo.ShortName == old.ShortName && newInfo.Logo == old.Logo && newInfo.TypeCode == old.TypeCode {
						break
					}
					clientInfoMap[k] = newInfo
					break
				}
			}
		}

		// 处理租户连接信息
		for _, c := range clients {
			// 新增租户连接
			if err := Add(c); err != nil {
				continue
			}
		}
		// 休眠五分钟再来
		time.Sleep(time.Minute * time.Duration(autoSyncClientTime))
	}
}

// Add 添加一个数据库连接
func Add(tdb DatabaseClientInfo) error {
	clientMapLock.Lock()
	defer clientMapLock.Unlock()
	// 如果已经存在且账号密码无变动，则跳过
	if data, exist := clientDbInfoMap[tdb.TenantId]; exist && data.User == tdb.User && data.Password == tdb.Password {
		return nil
	}

	// gorm配置
	conf := gorm.Config{
		Logger:                                   logs,
		DisableForeignKeyConstraintWhenMigrating: disableForeignKeyConstraintWhenMigrating,
	}

	// 创建数据库连接
	engine, err := gorm.Open(mysql.Open(tdb.GetDSN()), &conf)
	if err != nil {
		return err
	}

	// 同步模型
	if err = syncModel(engine, tdb.TenantId); err != nil {
		return err
	}

	// 操作成功了，保存连接信息
	clientMap[tdb.TenantId] = engine
	clientInfoMap[tdb.TenantId] = tdb.Info
	clientDbInfoMap[tdb.TenantId] = tdb

	return nil
}

// GetByTenantId 根据租户Id获取数据库连接对象
func GetByTenantId(tenantId string) (*gorm.DB, error) {
	clientMapLock.Lock()
	defer clientMapLock.Unlock()
	if client, exist := clientMap[tenantId]; exist {
		return client, nil
	}
	return nil, errors.New("not found")
}

// AddModel 添加一个需要同步的模型
func AddModel(m interface{}) error {
	// 加把锁
	syncModelsLock.Lock()
	defer syncModelsLock.Unlock()

	syncModels = append(syncModels, m)
	return nil
}

// AddModels 添加一堆需要同步的模型
func AddModels(m ...interface{}) error {
	if len(m) == 0 {
		return nil
	}

	// 加把锁
	syncModelsLock.Lock()
	defer syncModelsLock.Unlock()
	var err error
	for _, v := range m {
		err = AddModel(v)
	}

	return err
}

// 同步模型到数据库
func syncModel(e *gorm.DB, tenantId string) error {
	// 如果禁用了，跳过执行
	if syncModelsDisable {
		return nil
	}

	// 执行前回调
	if syncModelsBefore != nil {
		if err := syncModelsBefore(e, tenantId); err != nil {
			return err
		}
	}

	if e == nil || syncModels == nil {
		return errors.New("engine or model is nil")
	}
	syncModelsLock.Lock()
	defer syncModelsLock.Unlock()

	// 开启事务
	tx := e.Begin()
	// 同步表结构
	if err := tx.AutoMigrate(syncModels...); err != nil {
		tx.Rollback()
		return err
	}
	// 回调
	if syncModelsAfter != nil {
		if err := syncModelsAfter(tx, tenantId); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	tx.Commit()
	return nil
}
