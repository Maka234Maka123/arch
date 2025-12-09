package response

// SendSMSResponse 发送短信验证码响应
type SendSMSResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewSendSMSSuccessResponse 创建成功响应
func NewSendSMSSuccessResponse() *SendSMSResponse {
	return &SendSMSResponse{
		Code:    0,
		Message: "Success",
	}
}
