package user

import (
	"context"
	"fmt"

	"archv2/internal/dto/request"
	"archv2/internal/dto/response"
	"archv2/internal/integration/sms"
	"archv2/pkg/tracer"
)

// SendSMS 发送短信验证码
func (s *userService) SendSMS(ctx context.Context, req *request.SendSMSRequest) (*response.SendSMSResponse, error) {
	ctx, span := tracer.Start(ctx, "service.user.SendSMS")
	defer span.End()

	// 校验手机号
	if req.PhoneNumber == "" {
		err := fmt.Errorf("手机号不能为空")
		tracer.RecordError(span, err)
		return nil, err
	}

	// 校验短信类型
	smsType := sms.Type(req.From)
	switch smsType {
	case sms.TypeLogin, sms.TypeRegister, sms.TypeForget:
		// 有效类型
	default:
		err := fmt.Errorf("无效的短信类型: %s", req.From)
		tracer.RecordError(span, err)
		span.SetAttributes(tracer.String("sms.invalid_type", req.From))
		return nil, err
	}

	// 调用 SMS 客户端发送短信
	if err := s.smsClient.Send(ctx, smsType, req.PhoneNumber); err != nil {
		tracer.RecordError(span, err)
		return nil, err
	}

	return response.NewSendSMSSuccessResponse(), nil
}
