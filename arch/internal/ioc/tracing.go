package ioc

import (
	"context"
	"echo/internal/config"
	"echo/pkg/logger"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TracerProvider 全局 tracer provider
var tracerProvider *sdktrace.TracerProvider

// InitTracing 初始化 OpenTelemetry 链路追踪
func InitTracing(cfg *config.Config) error {
	if !cfg.Middleware.Tracing.Enabled {
		logger.Info("Tracing is disabled")
		return nil
	}

	ctx := context.Background()

	// 创建 OTLP gRPC 连接
	var opts []grpc.DialOption
	if cfg.Middleware.Tracing.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.Middleware.Tracing.Endpoint, opts...)
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	// 创建 OTLP trace exporter
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// 创建 resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.Middleware.Tracing.ServiceName),
			semconv.DeploymentEnvironment(cfg.Server.Mode),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// 配置采样器
	var sampler sdktrace.Sampler
	if cfg.Middleware.Tracing.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.Middleware.Tracing.SampleRate <= 0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.Middleware.Tracing.SampleRate)
	}

	// 创建 TracerProvider
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tracerProvider)

	// 设置全局 Propagator（支持 W3C Trace Context）
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("Tracing initialized",
		zap.String("endpoint", cfg.Middleware.Tracing.Endpoint),
		zap.String("service_name", cfg.Middleware.Tracing.ServiceName),
		zap.Float64("sample_rate", cfg.Middleware.Tracing.SampleRate),
	)

	return nil
}

// ShutdownTracing 关闭 tracing
func ShutdownTracing(ctx context.Context) error {
	if tracerProvider == nil {
		return nil
	}

	logger.Info("Shutting down tracing...")
	return tracerProvider.Shutdown(ctx)
}

// GetTracerProvider 获取 TracerProvider
func GetTracerProvider() *sdktrace.TracerProvider {
	return tracerProvider
}
