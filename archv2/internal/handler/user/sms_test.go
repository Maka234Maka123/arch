package user

import (
	"context"
	"testing"

	"archv2/internal/dto/request"
	"archv2/internal/integration/sms"
	userservice "archv2/internal/service/user"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func setupTestHandler() (*UserHandler, *sms.MockClient) {
	mockClient := sms.NewMockClient()
	service := userservice.NewUserService(mockClient)
	handler := NewUserHandler(service)
	return handler, mockClient
}

func TestUserHandler_SendSMS_Success(t *testing.T) {
	handler, mockClient := setupTestHandler()

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "登录验证码发送成功",
			body:    `{"phone_number":"13800138000","from":"login"}`,
			wantErr: false,
		},
		{
			name:    "注册验证码发送成功",
			body:    `{"phone_number":"13800138001","from":"register"}`,
			wantErr: false,
		},
		{
			name:    "忘记密码验证码发送成功",
			body:    `{"phone_number":"13800138002","from":"forget"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.Reset()

			ctx := context.Background()
			c := &app.RequestContext{}
			c.Request.SetMethod(consts.MethodPost)
			c.Request.SetRequestURI("/api/v1/user/sms")
			c.Request.Header.SetContentTypeBytes([]byte("application/json"))
			c.Request.SetBodyRaw([]byte(tt.body))

			handler.SendSMS(ctx, c)

			resp := c.Response
			statusCode := resp.StatusCode()

			if !tt.wantErr && statusCode != consts.StatusOK {
				t.Errorf("Expected status 200, got %d, body: %s", statusCode, string(resp.Body()))
			}

			// 验证 Mock 客户端收到请求
			if !tt.wantErr {
				records := mockClient.GetRecords()
				if len(records) != 1 {
					t.Errorf("Expected 1 send record, got %d", len(records))
				}
			}
		})
	}
}

func TestUserHandler_SendSMS_InvalidType(t *testing.T) {
	handler, _ := setupTestHandler()

	ctx := context.Background()
	c := &app.RequestContext{}
	c.Request.SetMethod(consts.MethodPost)
	c.Request.SetRequestURI("/api/v1/user/sms")
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	c.Request.SetBodyRaw([]byte(`{"phone_number":"13800138000","from":"invalid"}`))

	handler.SendSMS(ctx, c)

	resp := c.Response
	statusCode := resp.StatusCode()

	if statusCode != consts.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", statusCode)
	}
}

func TestUserHandler_SendSMS_MissingPhone(t *testing.T) {
	handler, _ := setupTestHandler()

	ctx := context.Background()
	c := &app.RequestContext{}
	c.Request.SetMethod(consts.MethodPost)
	c.Request.SetRequestURI("/api/v1/user/sms")
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	c.Request.SetBodyRaw([]byte(`{"from":"login"}`))

	handler.SendSMS(ctx, c)

	resp := c.Response
	statusCode := resp.StatusCode()

	// 缺少必填字段应该返回 400
	if statusCode != consts.StatusBadRequest {
		t.Errorf("Expected status 400, got %d, body: %s", statusCode, string(resp.Body()))
	}
}

func TestUserHandler_SendSMS_EmptyBody(t *testing.T) {
	handler, _ := setupTestHandler()

	ctx := context.Background()
	c := &app.RequestContext{}
	c.Request.SetMethod(consts.MethodPost)
	c.Request.SetRequestURI("/api/v1/user/sms")
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	c.Request.SetBodyRaw([]byte(`{}`))

	handler.SendSMS(ctx, c)

	resp := c.Response
	statusCode := resp.StatusCode()

	// 空请求体应该返回 400
	if statusCode != consts.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", statusCode)
	}
}

func TestUserHandler_SendSMS_ClientError(t *testing.T) {
	handler, mockClient := setupTestHandler()

	// 设置客户端返回错误
	mockClient.SetSendError(context.DeadlineExceeded)

	ctx := context.Background()
	c := &app.RequestContext{}
	c.Request.SetMethod(consts.MethodPost)
	c.Request.SetRequestURI("/api/v1/user/sms")
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	c.Request.SetBodyRaw([]byte(`{"phone_number":"13800138000","from":"login"}`))

	handler.SendSMS(ctx, c)

	resp := c.Response
	statusCode := resp.StatusCode()

	if statusCode != consts.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", statusCode)
	}
}

func TestUserHandler_SendSMS_ResponseFormat(t *testing.T) {
	handler, mockClient := setupTestHandler()
	mockClient.Reset()

	ctx := context.Background()
	c := &app.RequestContext{}
	c.Request.SetMethod(consts.MethodPost)
	c.Request.SetRequestURI("/api/v1/user/sms")
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	c.Request.SetBodyRaw([]byte(`{"phone_number":"13800138000","from":"login"}`))

	handler.SendSMS(ctx, c)

	resp := c.Response
	body := string(resp.Body())

	// 验证响应格式包含 code 和 message
	if !contains(body, "code") {
		t.Error("Response should contain 'code' field")
	}
	if !contains(body, "message") {
		t.Error("Response should contain 'message' field")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 确保编译时检查 protocol 包被正确导入
var _ = protocol.Request{}

// 确保编译时检查 request 包被正确导入
var _ = request.SendSMSRequest{}
