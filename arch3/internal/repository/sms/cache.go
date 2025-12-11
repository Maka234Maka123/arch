package sms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"arch3/internal/integration/sms"

	"github.com/redis/go-redis/v9"
)

const (
	// Redis key 格式
	codeKeyFormat       = "sms:code:%s:%s"        // sms:code:{type}:{phone}
	minuteKeyFormat     = "sms:minute:%s:%s"      // sms:minute:{type}:{phone}
	dayKeyFormat        = "sms:day:%s:%s"         // sms:day:{type}:{phone}
	verifyFailKeyFormat = "sms:verify_fail:%s:%s" // sms:verify_fail:{type}:{phone}

	// 验证失败计数过期时间（1小时，重新发送验证码成功后会重置）
	verifyFailTTL = 1 * time.Hour
)

// CacheRepository Redis 实现的短信验证码存储
type CacheRepository struct {
	rdb *redis.Client
}

// NewCacheRepository 创建短信验证码存储
func NewCacheRepository(rdb *redis.Client) *CacheRepository {
	return &CacheRepository{rdb: rdb}
}

// StoreCode 存储验证码
func (r *CacheRepository) StoreCode(ctx context.Context, smsType sms.Type, phone, code string, ttl time.Duration) error {
	key := r.codeKey(smsType, phone)
	return r.rdb.Set(ctx, key, code, ttl).Err()
}

// GetCode 获取验证码
func (r *CacheRepository) GetCode(ctx context.Context, smsType sms.Type, phone string) (string, error) {
	key := r.codeKey(smsType, phone)
	code, err := r.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", sms.ErrCodeInvalid
	}
	return code, err
}

// DeleteCode 删除验证码
func (r *CacheRepository) DeleteCode(ctx context.Context, smsType sms.Type, phone string) error {
	key := r.codeKey(smsType, phone)
	return r.rdb.Del(ctx, key).Err()
}

// GetSendCount 获取发送次数
func (r *CacheRepository) GetSendCount(ctx context.Context, smsType sms.Type, phone string) (minuteCount, dayCount int, err error) {
	minuteKey := r.minuteKey(smsType, phone)
	dayKey := r.dayKey(smsType, phone)

	minuteCount, err = r.rdb.Get(ctx, minuteKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, 0, err
	}

	dayCount, err = r.rdb.Get(ctx, dayKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, 0, err
	}

	return minuteCount, dayCount, nil
}

// IncrSendCount 增加发送次数
func (r *CacheRepository) IncrSendCount(ctx context.Context, smsType sms.Type, phone string) error {
	minuteKey := r.minuteKey(smsType, phone)
	dayKey := r.dayKey(smsType, phone)

	pipe := r.rdb.Pipeline()
	pipe.Incr(ctx, minuteKey)
	pipe.Expire(ctx, minuteKey, time.Minute)
	pipe.Incr(ctx, dayKey)
	pipe.Expire(ctx, dayKey, 24*time.Hour)
	_, err := pipe.Exec(ctx)

	return err
}

func (r *CacheRepository) codeKey(smsType sms.Type, phone string) string {
	return fmt.Sprintf(codeKeyFormat, smsType, phone)
}

func (r *CacheRepository) minuteKey(smsType sms.Type, phone string) string {
	return fmt.Sprintf(minuteKeyFormat, smsType, phone)
}

func (r *CacheRepository) dayKey(smsType sms.Type, phone string) string {
	return fmt.Sprintf(dayKeyFormat, smsType, phone)
}

func (r *CacheRepository) verifyFailKey(smsType sms.Type, phone string) string {
	return fmt.Sprintf(verifyFailKeyFormat, smsType, phone)
}

// GetVerifyFailCount 获取验证失败次数
func (r *CacheRepository) GetVerifyFailCount(ctx context.Context, smsType sms.Type, phone string) (int, error) {
	key := r.verifyFailKey(smsType, phone)
	count, err := r.rdb.Get(ctx, key).Int()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return count, err
}

// IncrVerifyFailCount 增加验证失败次数
func (r *CacheRepository) IncrVerifyFailCount(ctx context.Context, smsType sms.Type, phone string) error {
	key := r.verifyFailKey(smsType, phone)
	pipe := r.rdb.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, verifyFailTTL)
	_, err := pipe.Exec(ctx)
	return err
}

// ResetVerifyFailCount 重置验证失败次数
func (r *CacheRepository) ResetVerifyFailCount(ctx context.Context, smsType sms.Type, phone string) error {
	key := r.verifyFailKey(smsType, phone)
	return r.rdb.Del(ctx, key).Err()
}
