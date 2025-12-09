package user

import (
	"context"
	"net/http"

	"archv2/internal/dto/request"
	"archv2/pkg/tracer"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

// SendSMS 发送短信验证码
// @Summary 发送短信验证码
// @Description 发送短信验证码，支持登录、注册、忘记密码三种类型
// @Tags users
// @Accept json
// @Produce json
// @Param request body request.SendSMSRequest true "发送短信请求"
// @Success 200 {object} response.SendSMSResponse
// @Router /api/v1/user/sms [post]
func (h *UserHandler) SendSMS(ctx context.Context, c *app.RequestContext) {
	ctx, span := tracer.Start(ctx, "handler.SendSMS")
	defer span.End()

	var req request.SendSMSRequest
	if err := c.BindJSON(&req); err != nil {
		tracer.RecordError(span, err)
		_ = c.Error(err) // 注册错误供中间件记录
		c.JSON(http.StatusBadRequest, utils.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 记录关键属性（手机号已脱敏）
	span.SetAttributes(
		tracer.String(tracer.AttrSMSType, req.From),
	)

	resp, err := h.userService.SendSMS(ctx, &req)
	if err != nil {
		tracer.RecordError(span, err)
		_ = c.Error(err) // 注册错误供中间件记录
		c.JSON(http.StatusBadRequest, utils.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
