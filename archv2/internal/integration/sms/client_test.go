package sms

import (
	"context"
	"testing"
)

func TestMockClient_Send(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	tests := []struct {
		name    string
		smsType Type
		phone   string
		wantErr bool
	}{
		{
			name:    "登录验证码",
			smsType: TypeLogin,
			phone:   "13800138000",
			wantErr: false,
		},
		{
			name:    "注册验证码",
			smsType: TypeRegister,
			phone:   "13800138001",
			wantErr: false,
		},
		{
			name:    "忘记密码验证码",
			smsType: TypeForget,
			phone:   "13800138002",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.Reset()
			err := client.Send(ctx, tt.smsType, tt.phone)
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// 验证记录
				records := client.GetRecords()
				if len(records) != 1 {
					t.Errorf("Expected 1 record, got %d", len(records))
				}
				if records[0].Phone != tt.phone {
					t.Errorf("Expected phone %s, got %s", tt.phone, records[0].Phone)
				}
				if records[0].SmsType != tt.smsType {
					t.Errorf("Expected type %s, got %s", tt.smsType, records[0].SmsType)
				}

				// 验证验证码已存储
				code := client.GetCode(tt.smsType, tt.phone)
				if code == "" {
					t.Error("Expected code to be stored")
				}
			}
		})
	}
}

func TestMockClient_Verify(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// 先发送验证码
	phone := "13800138000"
	err := client.Send(ctx, TypeLogin, phone)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	code := client.GetCode(TypeLogin, phone)

	tests := []struct {
		name    string
		smsType Type
		phone   string
		code    string
		wantErr bool
	}{
		{
			name:    "正确验证码",
			smsType: TypeLogin,
			phone:   phone,
			code:    code,
			wantErr: false,
		},
		{
			name:    "测试验证码777777",
			smsType: TypeLogin,
			phone:   "13800138001",
			code:    "777777",
			wantErr: false,
		},
		{
			name:    "错误验证码",
			smsType: TypeLogin,
			phone:   "13800138002",
			code:    "000000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Verify(ctx, tt.smsType, tt.phone, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockClient_SendError(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// 设置发送错误
	client.SetSendError(context.DeadlineExceeded)

	err := client.Send(ctx, TypeLogin, "13800138000")
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded error, got %v", err)
	}
}

func TestType_Values(t *testing.T) {
	tests := []struct {
		typ  Type
		want string
	}{
		{TypeLogin, "login"},
		{TypeRegister, "register"},
		{TypeForget, "forget"},
	}

	for _, tt := range tests {
		if string(tt.typ) != tt.want {
			t.Errorf("Type %s != %s", tt.typ, tt.want)
		}
	}
}
