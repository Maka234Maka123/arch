package ioc

import (
	"archv2/internal/config"
	userhandler "archv2/internal/handler/user"
	"archv2/internal/integration/sms"
	"archv2/internal/router"
	userservice "archv2/internal/service/user"

	"github.com/redis/go-redis/v9"
)

// ManualInitialize 手动初始化所有业务依赖
//
// 依赖注入链:
//
//	Redis → SMSClient → UserService → UserHandler → Router
//
// 返回值:
//   - Router: 路由管理器，用于注册所有业务路由
//   - redis.Client: Redis 客户端，App 需要在关闭时释放
//   - error: 初始化失败时返回错误
//
// 错误处理:
//   - Redis 失败: 返回错误，无法继续（基础设施依赖）
//   - SMS 失败: 返回已初始化的 Redis，允许部分功能可用
//
// 扩展指南:
//
//	添加新业务模块时，按照 Repository → Service → Handler 的顺序初始化，
//	然后将 Handler 注入到 Router 中。
//
// 参数:
//   - isShuttingDown: 检查服务是否正在关闭的函数，用于就绪探针
func ManualInitialize(cfg *config.Config, isShuttingDown router.ShutdownChecker) (*router.Router, *redis.Client, error) {
	// 1. 初始化 Redis - 基础设施，其他组件可能依赖
	rdb, err := InitRedis(cfg)
	if err != nil {
		return nil, nil, err
	}

	// 2. 初始化 SMS 客户端 - 第三方服务集成
	smsClient, err := InitSMSClient(cfg, rdb)
	if err != nil {
		return nil, rdb, err
	}

	// 3. 初始化 Service 层 - 业务逻辑
	userSvc := userservice.NewUserService(smsClient)

	// 4. 初始化 Handler 层 - HTTP 请求处理
	userHandler := userhandler.NewUserHandler(userSvc)

	// 5. 初始化 Router - 路由注册（传入 config 和关闭检查器）
	r := router.NewRouter(cfg, userHandler, isShuttingDown)

	return r, rdb, nil
}

// SMSClientProvider 提供 SMS 客户端
//
// 设计目的:
//   - 用于单元测试时的依赖注入
//   - 支持 mock SMS 客户端进行测试
//
// 使用示例:
//
//	func TestUserService(t *testing.T) {
//	    mockRedis := ... // mock redis client
//	    smsClient, err := ioc.SMSClientProvider(cfg, mockRedis)
//	    userSvc := userservice.NewUserService(smsClient)
//	    // 测试 userSvc ...
//	}
func SMSClientProvider(cfg *config.Config, rdb *redis.Client) (sms.Client, error) {
	return InitSMSClient(cfg, rdb)
}
