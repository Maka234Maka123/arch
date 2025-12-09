package sms

import (
	"context"
	"fmt"
	"sync"
)

// MockClient 测试用 Mock 客户端
type MockClient struct {
	mu       sync.Mutex
	records  []SendRecord
	codes    map[string]string // phone -> code
	sendErr  error             // 模拟发送错误
	verifyErr error            // 模拟验证错误
}

// SendRecord 发送记录
type SendRecord struct {
	SmsType Type
	Phone   string
	Code    string
}

// NewMockClient 创建 Mock 客户端
func NewMockClient() *MockClient {
	return &MockClient{
		records: make([]SendRecord, 0),
		codes:   make(map[string]string),
	}
}

// Send 模拟发送短信
func (m *MockClient) Send(ctx context.Context, smsType Type, phone string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sendErr != nil {
		return m.sendErr
	}

	// 生成测试验证码
	code := "123456"
	m.codes[fmt.Sprintf("%s:%s", smsType, phone)] = code

	m.records = append(m.records, SendRecord{
		SmsType: smsType,
		Phone:   phone,
		Code:    code,
	})

	return nil
}

// Verify 模拟验证短信
func (m *MockClient) Verify(ctx context.Context, smsType Type, phone, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.verifyErr != nil {
		return m.verifyErr
	}

	// 测试验证码
	if code == "777777" {
		return nil
	}

	key := fmt.Sprintf("%s:%s", smsType, phone)
	storedCode, exists := m.codes[key]
	if !exists {
		return fmt.Errorf("验证码已过期或不存在")
	}

	if storedCode != code {
		return fmt.Errorf("验证码不正确")
	}

	// 验证成功后删除
	delete(m.codes, key)
	return nil
}

// GetRecords 获取发送记录（测试用）
func (m *MockClient) GetRecords() []SendRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.records
}

// GetCode 获取存储的验证码（测试用）
func (m *MockClient) GetCode(smsType Type, phone string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.codes[fmt.Sprintf("%s:%s", smsType, phone)]
}

// SetSendError 设置发送错误（测试用）
func (m *MockClient) SetSendError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendErr = err
}

// SetVerifyError 设置验证错误（测试用）
func (m *MockClient) SetVerifyError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.verifyErr = err
}

// Reset 重置状态（测试用）
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = make([]SendRecord, 0)
	m.codes = make(map[string]string)
	m.sendErr = nil
	m.verifyErr = nil
}
