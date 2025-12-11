package ioc

import (
	"time"

	"arch3/internal/config"
	"arch3/internal/handler/middleware"
	"arch3/internal/router"
	"arch3/pkg/jwt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Infrastructure 基础设施资源
// 用于统一管理需要在应用关闭时释放的资源
type Infrastructure struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// Container 依赖注入容器
// 包含应用运行所需的所有核心组件
type Container struct {
	Infra   *Infrastructure
	Tracing *TracingManager
	JWT     *jwt.Manager
	Server  *server.Hertz
	Router  *router.Router
}

// Close 关闭所有基础设施资源
func (i *Infrastructure) Close() error {
	var errs []error

	// 关闭数据库连接
	if i.DB != nil {
		if sqlDB, err := i.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// 关闭 Redis 连接
	if i.Redis != nil {
		if err := i.Redis.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// ManualInitialize 手动初始化所有依赖并构建容器
//
// 初始化顺序:
//  1. 基础设施层: DB, Redis
//  2. 可观测性层: Tracing, Metrics
//  3. 通用组件层: JWT
//  4. HTTP 层: Server, Middleware
//  5. 业务模块层: UserHandler
//  6. 路由层: Router
//
// 扩展指南:
//
//	添加新业务模块时，按照 Repository → Service → Handler 的顺序初始化，
//	然后将 Handler 注入到 Router 中。
func ManualInitialize(cfg *config.Config, isShuttingDown router.ShutdownChecker) (*Container, error) {
	// ========== 1. 基础设施层 ==========
	infra, err := initInfrastructure(cfg)
	if err != nil {
		return nil, err
	}

	// ========== 2. 可观测性层 ==========
	tracingMgr, err := NewTracingManager(cfg)
	if err != nil {
		infra.Close()
		return nil, err
	}

	if err := middleware.InitMetrics(); err != nil {
		infra.Close()
		return nil, err
	}

	// ========== 3. 通用组件层 ==========
	jwtMgr := initJWT(cfg, infra.Redis)

	// ========== 4. HTTP 层 ==========
	h, tracerCfg := initServer(cfg)
	registerMiddleware(h, cfg, tracerCfg, jwtMgr)

	// ========== 5. 业务模块层 ==========
	userHandler, err := InitUserHandler(infra.DB, infra.Redis, jwtMgr, cfg)
	if err != nil {
		infra.Close()
		return nil, err
	}

	// ========== 6. 路由层 ==========
	r := router.NewRouter(cfg, userHandler, isShuttingDown)
	r.Register(h)

	return &Container{
		Infra:   infra,
		Tracing: tracingMgr,
		JWT:     jwtMgr,
		Server:  h,
		Router:  r,
	}, nil
}

// initInfrastructure 初始化基础设施（DB、Redis 等）
func initInfrastructure(cfg *config.Config) (*Infrastructure, error) {
	infra := &Infrastructure{}

	db, err := InitDB(cfg)
	if err != nil {
		return nil, err
	}
	infra.DB = db

	rdb, err := InitRedis(cfg)
	if err != nil {
		infra.Close()
		return nil, err
	}
	infra.Redis = rdb

	return infra, nil
}

// initJWT 初始化 JWT 管理器
func initJWT(cfg *config.Config, rdb *redis.Client) *jwt.Manager {
	return jwt.NewManager(&jwt.Config{
		Secret:        cfg.JWT.Secret,
		AccessExpire:  time.Duration(cfg.JWT.AccessExpire) * time.Minute,
		RefreshExpire: time.Duration(cfg.JWT.RefreshExpire) * time.Minute,
		CookieSecure:  cfg.JWT.CookieSecure,
	}, rdb)
}
