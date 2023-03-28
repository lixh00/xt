package utils

//	Contain[T comparable]
//	@description: 判断切片中是否包含指定元素
//	@param slice
//	@param target
//	@return bool
func Contain[T comparable](slice []T, target T) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}

	return false
}

//	Difference[T comparable]
//	@description: 求两个切片的差集
//	@param slice
//	@param comparedSlice
//	@return []T
func Difference[T comparable](slice, comparedSlice []T) []T {
	result := []T{}

	for _, v := range slice {
		if !Contain(comparedSlice, v) {
			result = append(result, v)
		}
	}

	return result
}
