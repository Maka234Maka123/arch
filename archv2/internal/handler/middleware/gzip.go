package middleware

import (
	"archv2/internal/config"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/gzip"
)

// Gzip 返回 Gzip 压缩中间件
func Gzip(cfg *config.GzipConfig) app.HandlerFunc {
	opts := []gzip.Option{
		gzip.WithExcludedExtensions(cfg.ExcludedExts),
	}
	return gzip.Gzip(cfg.Level, opts...)
}
