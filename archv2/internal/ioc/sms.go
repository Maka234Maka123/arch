package ioc

import (
	"archv2/internal/config"
	"archv2/internal/integration/sms"
	"archv2/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// InitSMSClient 初始化 SMS 客户端
func InitSMSClient(cfg *config.Config, rdb *redis.Client) (sms.Client, error) {
	smsCfg := &sms.Config{
		Provider:   cfg.SMS.Provider,
		AccessKey:  cfg.SMS.AccessKey,
		SecretKey:  cfg.SMS.SecretKey,
		SmsAccount: cfg.SMS.SmsAccount,
		SignName:   cfg.SMS.SignName,
		Templates: map[sms.Type]string{
			sms.TypeLogin:    cfg.SMS.Templates.Login,
			sms.TypeRegister: cfg.SMS.Templates.Register,
			sms.TypeForget:   cfg.SMS.Templates.Forget,
		},
	}

	client, err := sms.NewClient(smsCfg)
	if err != nil {
		return nil, err
	}

	// 注入 Redis（日志现在通过 context 自动获取 trace_id）
	if volcClient, ok := client.(*sms.VolcengineClient); ok {
		volcClient.WithRedis(rdb)
	}

	logger.Info("SMS client initialized",
		zap.String("provider", cfg.SMS.Provider),
	)

	return client, nil
}
