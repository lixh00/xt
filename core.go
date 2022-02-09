package xt

import (
	"errors"
	"github.com/go-xorm/xorm"
	"sync"
)

var (
	once             sync.Once
	clientMap        map[uint]*xorm.Engine // 存储所有的数据库连接
	clientMapLock    sync.Mutex            // 一把锁
	syncModels       []interface{}         // 同步的模型
	syncModelsLock   sync.Mutex            // 一把锁
	tenantIdResolver TenantIdResolver      // 租户ID解析器
)

func init() {
	clientMap = make(map[uint]*xorm.Engine)
	syncModels = make([]interface{}, 0)
}

// Init 初始化
func Init(p TenantDBProvider, i TenantIdResolver) error {
	if p == nil {
		return errors.New("db provider is nil")
	}
	if i == nil {
		tenantIdResolver = getTenantId
	} else {
		tenantIdResolver = i
	}
	clients := p()
	for _, c := range clients {
		if err := Add(c); err != nil {
			return err
		}
	}
	return nil
}

// Add 添加一个数据库连接
func Add(tdb DatabaseClientInfo) error {
	clientMapLock.Lock()
	defer clientMapLock.Unlock()
	// 如果已经存在，则不再添加
	if _, exist := clientMap[tdb.TenantId]; exist {
		return nil
	}
	// 创建数据库连接
	engine, err := xorm.NewEngine("mysql", tdb.GetDSN())
	if err != nil {
		return err
	}
	clientMap[tdb.TenantId] = engine
	return nil
}

// AddModel 添加一个需要同步的模型
func AddModel(m interface{}) error {
	// 加把锁
	syncModelsLock.Lock()
	defer syncModelsLock.Unlock()

	syncModels = append(syncModels, m)
	return nil
}

// 同步模型到数据库
func syncModel(e *xorm.Engine, m interface{}) error {
	if e == nil || m == nil {
		return errors.New("engine or model is nil")
	}
	if err := e.Sync2(m); err != nil {
		return err
	}
	return nil
}
