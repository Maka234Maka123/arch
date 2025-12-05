package middleware

import (
	"echo/internal/config"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/logger/accesslog"
)

// AccessLog 返回访问日志中间件
func AccessLog(cfg *config.AccessLogConfig) app.HandlerFunc {
	return accesslog.New(
		accesslog.WithFormat("[${time}] ${status} - ${latency} ${method} ${path} ${queryParams}"),
		accesslog.WithTimeFormat(cfg.TimeFormat),
	)
}
