package ioc

import (
	"context"
	"fmt"
	"os"
	"time"

	"arch3/internal/config"
	"arch3/pkg/logger"

	"github.com/hertz-contrib/obs-opentelemetry/provider"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
)

// TracingManager 管理 OpenTelemetry 相关资源
//
// 职责:
//   - 初始化和管理 OTEL Trace/Metrics Provider
//   - 初始化和管理 OTEL Log Provider（日志导出到 APMPlus）
//   - 资源的优雅关闭，确保数据完整发送
//
// 架构说明:
//
//	应用 → TracingManager
//	         ├── OtelProvider (Trace + Metrics)
//	         │     └── 通过 hertz-contrib/obs-opentelemetry 初始化
//	         └── LoggerProvider (Logs)
//	               └── 通过 otelzap bridge 集成 zap 日志
type TracingManager struct {
	provider    provider.OtelProvider  // Trace 和 Metrics 的 Provider
	logProvider *sdklog.LoggerProvider // Log 的 Provider
}

// NewTracingManager 创建并初始化 TracingManager
//
// 初始化顺序:
//  1. 创建 OtelProvider (Trace + Metrics)
//  2. 创建 LogProvider (Logs 导出)
//  3. 启用 otelzap bridge (让 zap 日志关联 trace)
//
// 返回值:
//   - nil, nil: tracing 被禁用，正常情况
//   - *TracingManager, nil: 初始化成功
//   - nil, error: 初始化失败
//
// 配置项:
//   - middleware.tracing.enabled: 是否启用
//   - middleware.tracing.endpoint: OTLP 端点
//   - middleware.tracing.service_name: 服务名称
//   - middleware.tracing.app_key: 火山引擎认证（可选）
//   - middleware.tracing.insecure: 是否使用非 TLS 连接
func NewTracingManager(cfg *config.Config) (*TracingManager, error) {
	if !cfg.Middleware.Tracing.Enabled {
		logger.Info("Tracing is disabled")
		return nil, nil
	}

	tm := &TracingManager{}

	// 构建 provider 选项
	providerOpts := []provider.Option{
		provider.WithServiceName(cfg.Middleware.Tracing.ServiceName),
		provider.WithExportEndpoint(cfg.Middleware.Tracing.Endpoint),
		provider.WithEnableMetrics(true), // 启用 Metrics 导出
		provider.WithEnableTracing(true), // 启用 Trace 导出
	}

	// TLS 配置: 开发环境通常使用 insecure
	if cfg.Middleware.Tracing.Insecure {
		providerOpts = append(providerOpts, provider.WithInsecure())
	}

	// 火山引擎 APMPlus 认证头
	if cfg.Middleware.Tracing.AppKey != "" {
		providerOpts = append(providerOpts, provider.WithHeaders(map[string]string{
			"x-byteapm-appkey": cfg.Middleware.Tracing.AppKey,
		}))
	}

	// 创建 OpenTelemetry Provider (Trace + Metrics)
	tm.provider = provider.NewOpenTelemetryProvider(providerOpts...)

	logger.Info("OpenTelemetry provider initialized",
		zap.String("endpoint", cfg.Middleware.Tracing.Endpoint),
		zap.String("service_name", cfg.Middleware.Tracing.ServiceName),
		zap.Bool("has_app_key", cfg.Middleware.Tracing.AppKey != ""),
	)

	// 初始化 OTEL Log Provider - 可选功能，失败不影响服务启动
	if err := tm.initLogProvider(cfg); err != nil {
		logger.Warn("Failed to initialize OTEL log provider, logs won't be sent to APMPlus",
			zap.Error(err),
		)
	}

	return tm, nil
}

// initLogProvider 初始化 OTEL Log Provider
//
// 功能:
//   - 创建 OTLP gRPC Log Exporter
//   - 配置 Resource 属性（服务名、主机名）
//   - 启用 otelzap bridge，让 zap 日志自动关联 trace
//
// 日志关联机制:
//
//	用户代码: logger.Ctx(ctx).Info("message")
//	         ↓
//	otelzap bridge: 从 ctx 提取 trace_id/span_id
//	         ↓
//	OTLP Exporter: 发送到 APMPlus
//
// 火山引擎 APMPlus 要求:
//   - Resource 必须包含 service.name 和 host.name
//   - Headers 必须包含 x-byteapm-appkey
func (tm *TracingManager) initLogProvider(cfg *config.Config) error {
	ctx := context.Background()

	// 构建 log exporter 选项（使用 gRPC 协议）
	exporterOpts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(cfg.Middleware.Tracing.Endpoint),
		otlploggrpc.WithTimeout(10 * time.Second),
	}

	if cfg.Middleware.Tracing.Insecure {
		exporterOpts = append(exporterOpts, otlploggrpc.WithInsecure())
	}

	// 火山引擎认证头
	if cfg.Middleware.Tracing.AppKey != "" {
		exporterOpts = append(exporterOpts, otlploggrpc.WithHeaders(map[string]string{
			"x-byteapm-appkey": cfg.Middleware.Tracing.AppKey,
		}))
	}

	// 创建 log exporter
	logExporter, err := otlploggrpc.New(ctx, exporterOpts...)
	if err != nil {
		return err
	}

	// 获取主机名用于 Resource 标识
	hostname, _ := os.Hostname()

	// 创建 log provider
	// Resource 配置是 APMPlus 识别日志来源的关键
	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.Middleware.Tracing.ServiceName),
			semconv.HostNameKey.String(hostname),
		)),
	)

	// 检查 logProvider 是否创建成功
	if logProvider == nil {
		// 关闭已创建的 exporter，避免资源泄漏
		if shutdownErr := logExporter.Shutdown(ctx); shutdownErr != nil {
			logger.Warn("Failed to shutdown log exporter", zap.Error(shutdownErr))
		}
		return fmt.Errorf("failed to create log provider")
	}

	tm.logProvider = logProvider

	// 设置全局 log provider，otelzap bridge 需要它来发送日志
	global.SetLoggerProvider(tm.logProvider)

	// 启用 otelzap bridge
	// 之后 logger.Ctx(ctx).Info() 的日志会同时:
	//   1. 输出到本地 (stdout/file)
	//   2. 发送到 APMPlus，并自动关联 trace_id
	if err := logger.EnableOTEL(cfg.Middleware.Tracing.ServiceName); err != nil {
		return err
	}

	logger.Info("OTEL log provider initialized with zap bridge")
	return nil
}

// Shutdown 关闭 OpenTelemetry 资源
//
// 关闭顺序:
//  1. LogProvider - 确保日志数据发送完成
//  2. OtelProvider - 确保 trace/metrics 数据发送完成
//
// 注意: 关闭时会等待缓冲区中的数据发送完毕，受 ctx 超时限制
func (tm *TracingManager) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down OpenTelemetry...")

	// 1. 先关闭 log provider，确保日志数据发送完成
	if tm.logProvider != nil {
		if err := tm.logProvider.Shutdown(ctx); err != nil {
			logger.Error("OTEL log provider shutdown error", zap.Error(err))
		}
	}

	// 2. 再关闭 tracing/metrics provider
	if tm.provider != nil {
		return tm.provider.Shutdown(ctx)
	}

	return nil
}

// GetProvider 获取 OpenTelemetry Provider
//
// 用于需要直接访问 Provider 的场景，如自定义 span 创建。
// 大多数情况下不需要直接使用此方法。
func (tm *TracingManager) GetProvider() provider.OtelProvider {
	return tm.provider
}
