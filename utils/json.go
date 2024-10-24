package utils

import (
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
)

// Get
// @description: 从json数据中获取指定路径的值
// @param data
// @param path
// @return jsoniter.Any
func Get(data any, path ...interface{}) jsoniter.Any {
	encode, _ := json.Marshal(data)
	return jsoniter.Get(encode, path...)
}
