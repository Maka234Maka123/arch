package user

// SendSMSRequest 发送短信验证码请求
type SendSMSRequest struct {
	// 手机号：必填，11位数字，以1开头
	PhoneNumber string `json:"phone_number" vd:"len($)==11 && regexp('^1[3-9]\\d{9}$'); msg:'手机号格式无效，需要11位有效手机号'"`
	// 类型：必填，只能是 login/register/forget
	From string `json:"from" vd:"in($,'login','register','forget'); msg:'类型必须是 login、register 或 forget'"`
}

// SMSLoginRequest 短信验证码登录请求
type SMSLoginRequest struct {
	// 手机号：必填，11位数字，以1开头
	PhoneNumber string `json:"phone_number" vd:"len($)==11 && regexp('^1[3-9]\\d{9}$'); msg:'手机号格式无效，需要11位有效手机号'"`
	// 短信验证码：必填，6位数字
	SMSCode string `json:"sms_code" vd:"len($)==6 && regexp('^\\d{6}$'); msg:'验证码格式无效，需要6位数字'"`
}
