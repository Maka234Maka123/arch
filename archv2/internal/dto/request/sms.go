package request

// SendSMSRequest 发送短信验证码请求
type SendSMSRequest struct {
	PhoneNumber string `json:"phone_number"`
	From        string `json:"from"`
}
