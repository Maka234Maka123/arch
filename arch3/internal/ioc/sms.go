package ioc

import (
	"arch3/internal/config"
	"arch3/internal/integration/sms"
	"arch3/internal/integration/sms/volcengine"
	smsrepo "arch3/internal/repository/sms"
	userservice "arch3/internal/service/user"
	"arch3/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// InitSMSClient 初始化 SMS 客户端
func InitSMSClient(cfg *config.Config, rdb *redis.Client) (userservice.SMSClient, error) {
	smsCfg := &sms.Config{
		Provider:   cfg.SMS.Provider,
		AccessKey:  cfg.SMS.AccessKey,
		SecretKey:  cfg.SMS.SecretKey,
		SmsAccount: cfg.SMS.SmsAccount,
		SignName:   cfg.SMS.SignName,
		Templates: map[userservice.SMSType]string{
			userservice.SMSTypeLogin:    cfg.SMS.Templates.Login,
			userservice.SMSTypeRegister: cfg.SMS.Templates.Register,
			userservice.SMSTypeForget:   cfg.SMS.Templates.Forget,
		},
	}

	// 创建验证码存储 Repository
	codeRepo := smsrepo.NewCacheRepository(rdb)

	var client userservice.SMSClient
	var err error

	switch cfg.SMS.Provider {
	case "volcengine":
		client, err = volcengine.New(smsCfg, codeRepo)
	default:
		// 默认使用火山引擎
		client, err = volcengine.New(smsCfg, codeRepo)
	}

	if err != nil {
		return nil, err
	}

	logger.Info("SMS client initialized",
		zap.String("provider", cfg.SMS.Provider),
	)

	return client, nil
}
