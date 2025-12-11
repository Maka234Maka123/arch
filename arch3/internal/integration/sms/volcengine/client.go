package volcengine

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"arch3/internal/integration/sms"
	userservice "arch3/internal/service/user"
	"arch3/pkg/tracer"

	volcsms "github.com/volcengine/volc-sdk-golang/service/sms"
	"go.opentelemetry.io/otel/trace"
)

const (
	codeLength     = 6
	codeTTL        = 5 * time.Minute
	maxPerMin      = 1
	maxPerDay      = 6
	maxVerifyFails = 5 // 最大验证失败次数
)

// Client 火山引擎短信客户端
type Client struct {
	config *sms.Config
	repo   sms.CodeRepository
}

// New 创建火山引擎短信客户端
func New(cfg *sms.Config, repo sms.CodeRepository) (*Client, error) {
	// 配置火山引擎SDK
	volcsms.DefaultInstance.Client.SetAccessKey(cfg.AccessKey)
	volcsms.DefaultInstance.Client.SetSecretKey(cfg.SecretKey)

	return &Client{
		config: cfg,
		repo:   repo,
	}, nil
}

// Send 发送短信验证码
func (c *Client) Send(ctx context.Context, smsType userservice.SMSType, phone string) error {
	ctx, span := tracer.Start(ctx, "sms.volcengine.Send")
	defer span.End()

	span.SetAttributes(
		tracer.String(tracer.AttrSMSProvider, "volcengine"),
		tracer.String(tracer.AttrSMSType, string(smsType)),
	)

	// 先检查模板是否存在（避免存储验证码后发现模板不存在）
	templateID, ok := c.config.Templates[smsType]
	if !ok {
		tracer.RecordError(span, sms.ErrTemplateNotFound)
		return sms.ErrTemplateNotFound
	}

	// 检查发送限制
	if err := c.checkSendLimit(ctx, span, smsType, phone); err != nil {
		return err
	}

	// 生成验证码
	code := generateCode()

	// 存储验证码
	if err := c.repo.StoreCode(ctx, smsType, phone, code, codeTTL); err != nil {
		tracer.RecordError(span, err)
		return err
	}

	// 发送短信
	if err := c.sendSMS(ctx, phone, templateID, code); err != nil {
		tracer.RecordError(span, err)
		// 发送失败时删除已存储的验证码
		if delErr := c.repo.DeleteCode(ctx, smsType, phone); delErr != nil {
			tracer.AddEvent(span, "delete_code_on_send_fail_failed", tracer.String("error", delErr.Error()))
		}
		return sms.ErrSendFailed
	}

	// 记录发送次数（失败不影响主流程）
	if err := c.repo.IncrSendCount(ctx, smsType, phone); err != nil {
		tracer.AddEvent(span, "record_send_failed", tracer.String("error", err.Error()))
	}

	// 重新发送验证码成功后重置验证失败计数
	if err := c.repo.ResetVerifyFailCount(ctx, smsType, phone); err != nil {
		tracer.AddEvent(span, "reset_verify_fail_count_failed", tracer.String("error", err.Error()))
	}

	return nil
}

// Verify 验证短信验证码
func (c *Client) Verify(ctx context.Context, smsType userservice.SMSType, phone, code string) error {
	ctx, span := tracer.Start(ctx, "sms.volcengine.Verify")
	defer span.End()

	span.SetAttributes(
		tracer.String(tracer.AttrSMSProvider, "volcengine"),
		tracer.String(tracer.AttrSMSType, string(smsType)),
	)

	// 检查验证失败次数
	failCount, err := c.repo.GetVerifyFailCount(ctx, smsType, phone)
	if err != nil {
		tracer.RecordError(span, err)
		return err
	}
	if failCount >= maxVerifyFails {
		tracer.RecordError(span, sms.ErrVerifyTooMany)
		return sms.ErrVerifyTooMany
	}

	storedCode, err := c.repo.GetCode(ctx, smsType, phone)
	if err != nil {
		tracer.RecordError(span, err)
		return sms.ErrCodeInvalid
	}

	if storedCode != code {
		// 增加验证失败次数
		if err := c.repo.IncrVerifyFailCount(ctx, smsType, phone); err != nil {
			tracer.AddEvent(span, "incr_verify_fail_count_failed", tracer.String("error", err.Error()))
		}
		tracer.RecordError(span, sms.ErrCodeInvalid)
		return sms.ErrCodeInvalid
	}

	// 验证成功后删除验证码和失败计数
	if err := c.repo.DeleteCode(ctx, smsType, phone); err != nil {
		tracer.AddEvent(span, "delete_code_failed", tracer.String("error", err.Error()))
	}
	if err := c.repo.ResetVerifyFailCount(ctx, smsType, phone); err != nil {
		tracer.AddEvent(span, "reset_verify_fail_count_failed", tracer.String("error", err.Error()))
	}

	return nil
}

// checkSendLimit 检查发送限制
func (c *Client) checkSendLimit(ctx context.Context, span trace.Span, smsType userservice.SMSType, phone string) error {
	minuteCount, dayCount, err := c.repo.GetSendCount(ctx, smsType, phone)
	if err != nil {
		tracer.RecordError(span, err)
		return err
	}

	span.SetAttributes(
		tracer.Int("sms.minute_count", minuteCount),
		tracer.Int("sms.day_count", dayCount),
	)

	if minuteCount >= maxPerMin {
		tracer.RecordError(span, sms.ErrSendTooFrequent)
		return sms.ErrSendTooFrequent
	}

	if dayCount >= maxPerDay {
		tracer.RecordError(span, sms.ErrDailyLimitExceeded)
		return sms.ErrDailyLimitExceeded
	}

	return nil
}

// sendSMS 调用火山引擎API发送短信
func (c *Client) sendSMS(ctx context.Context, phone, templateID, code string) error {
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

// generateCode 生成6位验证码 (使用 crypto/rand)
func generateCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		// fallback: 使用时间戳生成
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", n.Int64())
}
