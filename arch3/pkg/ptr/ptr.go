// Package ptr 提供指针类型的通用转换函数
package ptr

// Of 将任意值转换为指针
// 示例: ptr.Of("hello") 返回 *string
func Of[T any](v T) *T {
	return &v
}

// Value 安全地获取指针的值，如果为 nil 返回零值
// 示例: ptr.Value(strPtr) 返回 string
func Value[T any](p *T) T {
	if p != nil {
		return *p
	}
	var zero T
	return zero
}

// ValueOr 安全地获取指针的值，如果为 nil 返回默认值
// 示例: ptr.ValueOr(strPtr, "default") 返回 string
func ValueOr[T any](p *T, defaultVal T) T {
	if p != nil {
		return *p
	}
	return defaultVal
}
