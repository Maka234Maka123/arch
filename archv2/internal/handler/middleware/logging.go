package middleware

import (
	"context"
	"net/url"
	"strings"
	"time"

	"archv2/internal/config"
	"archv2/pkg/logger"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

// sensitiveParams 敏感参数名列表（小写），这些参数的值会被脱敏
var sensitiveParams = map[string]bool{
	"password":      true,
	"passwd":        true,
	"pwd":           true,
	"token":         true,
	"access_token":  true,
	"refresh_token": true,
	"apikey":        true,
	"api_key":       true,
	"secret":        true,
	"secret_key":    true,
	"authorization": true,
	"auth":          true,
	"credential":    true,
	"private_key":   true,
}

// AccessLog 返回访问日志中间件
// 每条访问日志自动包含 trace_id，便于日志关联和问题排查
//
// 配置项:
//   - SkipPaths: 不记录日志的路径列表，如 ["/health", "/metrics"]
//   - ErrorOnly: 只记录错误响应 (状态码 >= 400)，正常请求依赖 tracing
func AccessLog(cfg *config.AccessLogConfig) app.HandlerFunc {
	// 构建跳过路径的 map，提高查找效率
	skipPaths := make(map[string]bool, len(cfg.SkipPaths))
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Path())

		// 检查是否跳过该路径（支持前缀匹配）
		if shouldSkipPath(path, skipPaths, cfg.SkipPaths) {
			c.Next(ctx)
			return
		}

		start := time.Now()

		// 处理请求
		c.Next(ctx)

		statusCode := c.Response.StatusCode()

		// ErrorOnly 模式：只记录错误响应
		if cfg.ErrorOnly && statusCode < 400 {
			return
		}

		// 记录访问日志（自动包含 trace_id 和 span_id）
		latency := time.Since(start)
		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.Int64("latency_ms", latency.Milliseconds()),
			zap.String("method", string(c.Method())),
			zap.String("path", path),
			zap.String("query", sanitizeQuery(c.QueryArgs().String())),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", string(c.UserAgent())),
			zap.Int("body_size", len(c.Response.Body())),
		}

		// 确定日志消息：有错误时使用错误信息，否则使用 "access"
		msg := "access"
		if len(c.Errors) > 0 {
			if lastErr := c.Errors.Last(); lastErr != nil {
				msg = lastErr.Error()
			}
		}

		// 根据状态码选择日志级别
		if statusCode != 200 {
			logger.Ctx(ctx).Warn(msg, fields...)
		} else {
			logger.Ctx(ctx).Info(msg, fields...)
		}
	}
}

// sanitizeQuery 对查询字符串中的敏感参数进行脱敏
// 例如: "user=admin&password=123456" -> "user=admin&password=[REDACTED]"
func sanitizeQuery(queryStr string) string {
	if queryStr == "" {
		return ""
	}

	values, err := url.ParseQuery(queryStr)
	if err != nil {
		// 解析失败，返回 [PARSE_ERROR] 避免泄露原始内容
		return "[PARSE_ERROR]"
	}

	for key := range values {
		if sensitiveParams[strings.ToLower(key)] {
			values.Set(key, "[REDACTED]")
		}
	}

	return values.Encode()
}

// shouldSkipPath 检查路径是否应该跳过日志记录
// 支持精确匹配和前缀匹配（以 * 结尾）
func shouldSkipPath(path string, skipMap map[string]bool, skipPaths []string) bool {
	// 精确匹配
	if skipMap[path] {
		return true
	}

	// 前缀匹配（支持 /swagger/* 这样的配置）
	for _, skipPath := range skipPaths {
		if strings.HasSuffix(skipPath, "*") {
			prefix := strings.TrimSuffix(skipPath, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}

	return false
}
