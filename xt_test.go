package xt

import (
	"testing"
)

type User struct {
	Id   uint   `json:"id" gorm:"primary_key"`
	Name string `json:"name" form:"type:varchar(20);not null;comment:'姓名'"`
	Age  int    `json:"age" gorm:"type:tinyint(3);default:1;not null;comment:'年龄'"`
}

func (User) TableName() string {
	return "test_user"
}

type UserInfo struct {
	Id     uint   `json:"id" gorm:"primary_key"`
	Sex    string `json:"sex" form:"type:varchar(20);not null;comment:'姓名'"`
	Avatar int    `json:"avatar" gorm:"type:tinyint(3);default:1;not null;comment:'年龄'"`
}

func (UserInfo) TableName() string {
	return "test_user_info"
}

func getDBS() []DatabaseClientInfo {
	var dbs []DatabaseClientInfo
	dbs = append(dbs, DatabaseClientInfo{
		TenantId: 1,
		Info: TenantInfo{
			Name:      "李寻欢测试",
			ShortName: "测试",
			Logo:      "",
			TypeCode:  "school",
		},
		Host:     "10.11.0.10",
		Port:     3307,
		User:     "saas",
		Password: "saas123",
		Db:       "saas_hsxl",
	})
	return dbs
}

func TestSyncModels(t *testing.T) {
	_ = AddModel(User{})
	_ = AddModel(UserInfo{})

	//DisableSyncModels(true)
	SetSyncModelsAsync(true)

	err := Init(getDBS, nil, true)
	if err != nil {
		return
	}
	_, err = GetByTenantId(1)
	if err != nil {
		return
	}
	t.Log("成功")
}
