package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// Recovery 返回 panic 恢复中间件
func Recovery() app.HandlerFunc {
	return recovery.Recovery(recovery.WithRecoveryHandler(recoveryHandler))
}

// recoveryHandler 自定义 panic 恢复处理器
func recoveryHandler(ctx context.Context, c *app.RequestContext, err interface{}, stack []byte) {
	hlog.SystemLogger().CtxErrorf(ctx, "[Recovery] panic recovered: %v\n%s", err, stack)
	c.JSON(500, map[string]interface{}{
		"code":    500,
		"message": "Internal Server Error",
	})
}
