package sms

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"archv2/pkg/tracer"

	"github.com/redis/go-redis/v9"
	volcsms "github.com/volcengine/volc-sdk-golang/service/sms"
)

var (
	minuteKeyFormat  = "sms:minute:%s"
	dayKeyFormat     = "sms:day:%s"
	logicalKeyFormat = "sms:logical:%s:%s"
)

// VolcengineClient 火山引擎短信客户端
type VolcengineClient struct {
	config *Config
	rdb    *redis.Client
}

// NewVolcengineClient 创建火山引擎短信客户端
func NewVolcengineClient(cfg *Config) (*VolcengineClient, error) {
	// 配置火山引擎SDK
	volcsms.DefaultInstance.Client.SetAccessKey(cfg.AccessKey)
	volcsms.DefaultInstance.Client.SetSecretKey(cfg.SecretKey)

	return &VolcengineClient{
		config: cfg,
	}, nil
}

// WithRedis 设置 Redis 客户端
func (c *VolcengineClient) WithRedis(rdb *redis.Client) *VolcengineClient {
	c.rdb = rdb
	return c
}

// Send 发送短信验证码
func (c *VolcengineClient) Send(ctx context.Context, smsType Type, phone string) error {
	ctx, span := tracer.Start(ctx, "sms.volcengine.Send")
	defer span.End()

	maskedPhone := maskPhone(phone)
	span.SetAttributes(
		tracer.String(tracer.AttrSMSProvider, "volcengine"),
		tracer.String(tracer.AttrSMSType, string(smsType)),
		tracer.String(tracer.AttrPhoneMasked, maskedPhone),
	)

	// 检查发送限制
	if err := c.checkSendLimit(ctx, smsType, phone); err != nil {
		tracer.RecordError(span, err)
		return err
	}

	// 生成验证码
	code := c.generateCode()

	// 存储验证码到 Redis
	if err := c.storeCode(ctx, smsType, phone, code); err != nil {
		tracer.RecordError(span, err)
		return err
	}

	// 获取模板ID
	templateID, ok := c.config.Templates[smsType]
	if !ok {
		err := fmt.Errorf("unsupported sms type: %s", smsType)
		tracer.RecordError(span, err)
		return err
	}

	// 发送短信
	if err := c.sendSMS(ctx, phone, templateID, code); err != nil {
		tracer.RecordError(span, err)
		return fmt.Errorf("发送短信失败")
	}

	// 记录发送次数（失败不影响主流程）
	if err := c.recordSend(ctx, smsType, phone); err != nil {
		tracer.AddEvent(span, "record_send_failed", tracer.String("error", err.Error()))
	}

	return nil
}

// Verify 验证短信验证码
func (c *VolcengineClient) Verify(ctx context.Context, smsType Type, phone, code string) error {
	ctx, span := tracer.Start(ctx, "sms.volcengine.Verify")
	defer span.End()

	maskedPhone := maskPhone(phone)
	span.SetAttributes(
		tracer.String(tracer.AttrSMSProvider, "volcengine"),
		tracer.String(tracer.AttrSMSType, string(smsType)),
		tracer.String(tracer.AttrPhoneMasked, maskedPhone),
	)

	// 测试验证码 (生产环境应移除)
	if code == "777777" {
		tracer.AddEvent(span, "test_code_used")
		return nil
	}

	key := c.getRedisKey(logicalKeyFormat, phone, smsType)
	storedCode, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err := fmt.Errorf("验证码已过期或不存在")
			tracer.RecordError(span, err)
			return err
		}
		tracer.RecordError(span, err)
		return fmt.Errorf("获取验证码失败")
	}

	if storedCode != code {
		err := fmt.Errorf("验证码不正确")
		tracer.RecordError(span, err)
		return err
	}

	// 验证成功后删除验证码
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		tracer.AddEvent(span, "delete_code_failed", tracer.String("error", err.Error()))
	}

	return nil
}

// checkSendLimit 检查发送限制
func (c *VolcengineClient) checkSendLimit(ctx context.Context, smsType Type, phone string) error {
	ctx, span := tracer.Start(ctx, "sms.checkSendLimit")
	defer span.End()

	if c.rdb == nil {
		return nil
	}

	minuteKey := c.getRedisKey(minuteKeyFormat, phone, smsType)
	dayKey := c.getRedisKey(dayKeyFormat, phone, smsType)

	// 检查分钟限制
	minuteCount, err := c.rdb.Get(ctx, minuteKey).Int()
	if err != nil && err != redis.Nil {
		tracer.RecordError(span, err)
		return err
	}
	if minuteCount >= 1 {
		err := fmt.Errorf("一分钟内最多发送1次")
		tracer.RecordError(span, err)
		return err
	}

	// 检查日限制
	dayCount, err := c.rdb.Get(ctx, dayKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		tracer.RecordError(span, err)
		return err
	}
	if dayCount >= 6 {
		err := fmt.Errorf("一天最多发送6次")
		tracer.RecordError(span, err)
		return err
	}

	span.SetAttributes(
		tracer.Int("sms.minute_count", minuteCount),
		tracer.Int("sms.day_count", dayCount),
	)

	return nil
}

// storeCode 存储验证码
func (c *VolcengineClient) storeCode(ctx context.Context, smsType Type, phone, code string) error {
	ctx, span := tracer.Start(ctx, "sms.storeCode")
	defer span.End()

	if c.rdb == nil {
		return nil
	}

	key := c.getRedisKey(logicalKeyFormat, phone, smsType)
	if err := c.rdb.Set(ctx, key, code, 5*time.Minute).Err(); err != nil {
		tracer.RecordError(span, err)
		return err
	}
	return nil
}

// recordSend 记录发送次数
func (c *VolcengineClient) recordSend(ctx context.Context, smsType Type, phone string) error {
	ctx, span := tracer.Start(ctx, "sms.recordSend")
	defer span.End()

	if c.rdb == nil {
		return nil
	}

	minuteKey := c.getRedisKey(minuteKeyFormat, phone, smsType)
	dayKey := c.getRedisKey(dayKeyFormat, phone, smsType)

	pipe := c.rdb.Pipeline()
	pipe.Incr(ctx, minuteKey)
	pipe.Expire(ctx, minuteKey, time.Minute)
	pipe.Incr(ctx, dayKey)
	pipe.Expire(ctx, dayKey, 24*time.Hour)
	_, err := pipe.Exec(ctx)

	if err != nil {
		tracer.RecordError(span, err)
	}
	return err
}

// sendSMS 调用火山引擎API发送短信
func (c *VolcengineClient) sendSMS(ctx context.Context, phone, templateID, code string) error {
	ctx, span := tracer.Start(ctx, "sms.volcengine.API")
	defer span.End()

	span.SetAttributes(
		tracer.String("sms.template_id", templateID),
	)

	req := &volcsms.SmsRequest{
		SmsAccount:    c.config.SmsAccount,
		Sign:          c.config.SignName,
		TemplateID:    templateID,
		TemplateParam: fmt.Sprintf(`{"code":"%s"}`, code),
		PhoneNumbers:  phone,
	}

	result, statusCode, err := volcsms.DefaultInstance.Send(req)

	span.SetAttributes(tracer.Int(tracer.AttrHTTPStatusCode, statusCode))

	if err != nil {
		tracer.RecordError(span, err)
		return fmt.Errorf("sms send err: %s, statusCode: %d", err.Error(), statusCode)
	}

	// API 成功只记录到 span
	var messageID string
	if result.Result != nil && len(result.Result.MessageID) > 0 {
		messageID = result.Result.MessageID[0]
	}
	tracer.AddEvent(span, "api.success",
		tracer.Int("status_code", statusCode),
		tracer.String("message_id", messageID),
	)
	return nil
}

// getRedisKey 生成 Redis Key
func (c *VolcengineClient) getRedisKey(format, phone string, smsType Type) string {
	switch smsType {
	case TypeRegister, TypeLogin, TypeForget:
		return fmt.Sprintf(format, smsType, phone)
	}
	return fmt.Sprintf(format, phone)
}

// generateCode 生成6位验证码
func (c *VolcengineClient) generateCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// maskPhone 脱敏手机号，只显示前3位和后4位
// 例如：13812345678 -> 138****5678
func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
