## 基于`xorm`和`Gin`封装的多租户处理框架

### 多租户方案
当下主流的多租户方案通常有三种：
1. 独立数据库
`每个租户单独的数据库，用户数据隔离级别最高，安全性更好，但是成本更高`
2. 共享数据库，隔离数据架构
`多个或所有租户共享同一个数据库，但是每个租户的scheme不同，用户数据隔离级别中等`
3. 共享数据库和数据架构
`所有租户共享同一个数据库和表，在数据表中通过TenantId字段区分不同租户，用户数据隔离级别最低，但是成本也更低`

本工具采用的是第一种(`独立数据库`)

### 实现原理
 1. 初始化
`传入一个连接配置，去读取数据库配置，并初始化数据库连接map单例`
 2. 根据租户Id返回不同的数据库连接对象和`gin.Context`

### 食用方式
0. 安装依赖
```shell
go get github.com/lixh00/xt
```
1. 先添加需要同步的Model，如果不需要可以不执行这个步骤
```go
if err := xt.AddModel(new(En));err != nil {
    panic(err)
}
```
2. 创建一个提供租户数据库配置信息的函数，此步骤为必须
```go
func GetTenantDbInfos() []xt.DatabaseClientInfo {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		"tenant", "tenant123", "192.168.1.37", 3307, "tenant")
	engine, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	var dbs []TDB
	if err = engine.Find(&dbs); err != nil {
		panic(err)
	}
	var d []xt.DatabaseClientInfo
	for _, ddd := range dbs {
		d = append(d, xt.DatabaseClientInfo{
			TenantId: ddd.TenantId,
			Host:     ddd.Host,
			Port:     ddd.Port,
			User:     ddd.Username,
			Password: ddd.Password,
			Db:       ddd.Db,
		})
	}
	logger.Say.Debugf("%+v", d)
	return d
}
```
3. 开始使用
```go
func main() {
	app := gin.Default()
	err := xt.Init(GetTenantDbInfos, nil)
	if err != nil {
		panic(err)
	}
	app.GET("/test", xt.GinHandler(func(ctx *xt.MultiTenantContext) {
		var us []En
		if err := ctx.DB.Find(&us); err != nil {
			logger.Say.Errorf("数据查询失败: %v", err)
			ctx.JSON(200, gin.H{
				"message": "系统错误",
			})
			return
		}
		ctx.JSON(200, gin.H{
			"message": us,
		})
	}))
	app.Run(":12345")
}
```