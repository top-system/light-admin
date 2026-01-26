package controller

import (
	"net/http"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/pkg/echox"
	"github.com/labstack/echo/v4"
)

type CaptchaController struct {
	captcha lib.Captcha
}

// NewUserController creates new user controller
func NewCaptchaController(captcha lib.Captcha) CaptchaController {
	return CaptchaController{captcha: captcha}
}

// @tags Auth
// @summary 获取验证码
// @produce application/json
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/auth/captcha [get]
func (a CaptchaController) GetCaptcha(ctx echo.Context) error {
	id, b64s, _, err := a.captcha.Generate()
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: echo.Map{"captchaId": id, "captchaBase64": b64s}}.JSON(ctx)
}

// @tags Auth
// @summary 验证验证码
// @produce application/json
// @param data body dto.CaptchaVerify true "CaptchaVerify"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @router /api/v1/auth/captcha/verify [post]
func (a CaptchaController) VerifyCaptcha(ctx echo.Context) error {
	verify := new(dto.CaptchaVerify)
	if err := ctx.Bind(verify); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	ok := a.captcha.Verify(verify.ID, verify.Code, false)
	if !ok {
		return echox.Response{Code: http.StatusBadRequest, Message: errors.CaptchaAnswerCodeNoMatch}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
