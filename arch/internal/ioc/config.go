package ioc

import (
	"echo/internal/config"
)

// InitConfig 初始化配置
func InitConfig(path string) (*config.Config, error) {
	return config.Load(path)
}
