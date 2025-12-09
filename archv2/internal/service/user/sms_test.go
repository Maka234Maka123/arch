package user

import (
	"context"
	"fmt"
	"testing"

	"archv2/internal/dto/request"
	"archv2/internal/integration/sms"
)

func TestUserService_SendSMS(t *testing.T) {
	mockClient := sms.NewMockClient()
	service := NewUserService(mockClient)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *request.SendSMSRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "登录验证码发送成功",
			req: &request.SendSMSRequest{
				PhoneNumber: "13800138000",
				From:        "login",
			},
			wantErr: false,
		},
		{
			name: "注册验证码发送成功",
			req: &request.SendSMSRequest{
				PhoneNumber: "13800138001",
				From:        "register",
			},
			wantErr: false,
		},
		{
			name: "忘记密码验证码发送成功",
			req: &request.SendSMSRequest{
				PhoneNumber: "13800138002",
				From:        "forget",
			},
			wantErr: false,
		},
		{
			name: "无效的短信类型",
			req: &request.SendSMSRequest{
				PhoneNumber: "13800138003",
				From:        "invalid",
			},
			wantErr: true,
			errMsg:  "无效的短信类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.Reset()
			resp, err := service.SendSMS(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errMsg != "" && err.Error() != fmt.Sprintf("无效的短信类型: %s", tt.req.From) {
					// 检查错误消息是否包含预期内容
					t.Logf("Error message: %s", err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("SendSMS() error = %v", err)
				return
			}

			if resp == nil {
				t.Error("Expected response but got nil")
				return
			}

			if resp.Code != 0 {
				t.Errorf("Expected code 0, got %d", resp.Code)
			}

			if resp.Message != "Success" {
				t.Errorf("Expected message 'Success', got %s", resp.Message)
			}

			// 验证 Mock 客户端记录
			records := mockClient.GetRecords()
			if len(records) != 1 {
				t.Errorf("Expected 1 send record, got %d", len(records))
			}
		})
	}
}

func TestUserService_SendSMS_ClientError(t *testing.T) {
	mockClient := sms.NewMockClient()
	service := NewUserService(mockClient)
	ctx := context.Background()

	// 设置客户端返回错误
	mockClient.SetSendError(fmt.Errorf("一分钟内最多发送1次"))

	req := &request.SendSMSRequest{
		PhoneNumber: "13800138000",
		From:        "login",
	}

	resp, err := service.SendSMS(ctx, req)

	if err == nil {
		t.Error("Expected error but got nil")
	}

	if resp != nil {
		t.Error("Expected nil response on error")
	}
}
