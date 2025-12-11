package middleware

import (
	"context"
	"net/http"

	"arch3/pkg/logger"
	"arch3/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"go.uber.org/zap"
)

// Recovery 返回 panic 恢复中间件
//
// 功能:
//   - 捕获 handler 中的 panic，防止服务崩溃
//   - 记录错误日志（包含 trace_id 和调用栈）
//   - 返回统一格式的错误响应（HTTP 200 + 业务码 9500）
//
// 这是中间件链的最外层，确保任何 panic 都能被捕获。
func Recovery() app.HandlerFunc {
	return recovery.Recovery(recovery.WithRecoveryHandler(recoveryHandler))
}

// recoveryHandler 自定义 panic 恢复处理器
//
// 使用 logger.Ctx(ctx) 记录错误，确保日志包含 trace_id，
// 便于在 APMPlus 中关联 trace 和错误日志。
//
// 记录的上下文信息:
//   - error: panic 的错误信息
//   - stack: 调用栈
//   - path/method: 请求路径和方法
//   - client_ip: 客户端 IP（用于安全分析）
//   - user_agent: 用户代理（用于问题复现）
//   - query: 查询参数（脱敏后，用于问题诊断）
func recoveryHandler(ctx context.Context, c *app.RequestContext, err interface{}, stack []byte) {
	method := string(c.Method())
	path := string(c.Path())

	// 使用统一的 logger 记录，确保 trace_id 关联
	// 记录完整的请求上下文，便于问题诊断
	logger.Ctx(ctx).Error("[Recovery] panic recovered",
		zap.Any("error", err),
		zap.ByteString("stack", stack),
		zap.String("path", path),
		zap.String("method", method),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", string(c.UserAgent())),
		zap.String("query", sanitizeQuery(c.QueryArgs().String())),
	)

	// 记录 panic 指标
	RecordPanic(ctx, method)

	c.JSON(http.StatusOK, &response.Result{
		Code:    response.CodeInternal,
		Message: "Internal Server Error",
	})
}
