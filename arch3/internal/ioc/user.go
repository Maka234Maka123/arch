package ioc

import (
	"arch3/internal/config"
	userhandler "arch3/internal/handler/user"
	userrepo "arch3/internal/repository/user"
	userservice "arch3/internal/service/user"
	"arch3/pkg/jwt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// InitUserHandler 初始化 User 模块的完整依赖链
//
// 依赖链: DAO → Repository → SMSClient → Service → Handler
func InitUserHandler(
	db *gorm.DB,
	rdb *redis.Client,
	jwtMgr *jwt.Manager,
	cfg *config.Config,
) (*userhandler.Handler, error) {
	// DAO 层
	userDAO := userrepo.NewDAO(db)

	// Repository 层
	userRepo := userrepo.NewRepository(userDAO)

	// SMS 客户端
	smsClient, err := InitSMSClient(cfg, rdb)
	if err != nil {
		return nil, err
	}

	// Service 层
	userSvc := userservice.NewService(smsClient, userRepo, jwtMgr)

	// Handler 层
	return userhandler.NewHandler(userSvc, jwtMgr), nil
}
