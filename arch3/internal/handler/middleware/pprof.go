package middleware

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/pprof"
)

// RegisterPprof 注册 pprof 性能分析路由
func RegisterPprof(h *server.Hertz) {
	pprof.Register(h)
}
