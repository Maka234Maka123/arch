package config

// API 版本常量
// 统一管理 API 版本前缀，避免硬编码分散在各处
const (
	// APIVersion 当前 API 版本
	APIVersion = "v2"

	// APIPrefix API 路径前缀
	APIPrefix = "/api/" + APIVersion
)
