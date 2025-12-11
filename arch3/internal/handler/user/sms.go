package user

import (
	"context"

	"arch3/pkg/response"
	"arch3/pkg/tracer"

	"github.com/cloudwego/hertz/pkg/app"
)

// SendSMS 发送短信验证码
// @Summary 发送短信验证码
// @Description 发送短信验证码，支持登录、注册、忘记密码三种类型
// @Tags users
// @Accept json
// @Produce json
// @Param request body SendSMSRequest true "发送短信请求"
// @Success 200 {object} response.Result
// @Router /api/v1/user/sms [post]
func (h *Handler) SendSMS(ctx context.Context, c *app.RequestContext) error {
	ctx, span := tracer.Start(ctx, "handler.SendSMS")
	defer span.End()

	var req SendSMSRequest
	if err := c.BindAndValidate(&req); err != nil {
		tracer.RecordError(span, err)
		return response.Validation(err.Error())
	}

	// 记录关键属性（手机号已脱敏）
	span.SetAttributes(
		tracer.String(tracer.AttrSMSType, req.From),
	)

	if err := h.userService.SendSMS(ctx, req.PhoneNumber, req.From); err != nil {
		tracer.RecordError(span, err)
		return err
	}

	return response.Success(c, nil)
}
