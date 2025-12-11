package user

import (
	"context"

	"arch3/pkg/tracer"
)

// SendSMS 发送短信验证码
func (s *service) SendSMS(ctx context.Context, phoneNumber, smsType string) error {
	ctx, span := tracer.Start(ctx, "service.user.SendSMS")
	defer span.End()

	if err := s.smsClient.Send(ctx, SMSType(smsType), phoneNumber); err != nil {
		tracer.RecordError(span, err)
		return SMSToResponse(err)
	}

	return nil
}
