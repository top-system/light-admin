package controller

import (
	"net/http"

	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/pkg/echox"
	"github.com/labstack/echo/v4"
)

type PublicController struct {
	userService service.UserService
	authService service.AuthService
	captcha     lib.Captcha
	config      lib.Config
	logger      lib.Logger
}

// NewPublicController creates new public controller
func NewPublicController(
	userService service.UserService,
	authService service.AuthService,
	captcha lib.Captcha,
	config lib.Config,
	logger lib.Logger,
) PublicController {
	return PublicController{
		userService: userService,
		authService: authService,
		captcha:     captcha,
		config:      config,
		logger:      logger,
	}
}

// @tags Auth
// @summary 用户登录
// @produce application/json
// @param data body dto.Login true "Login"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/auth/login [post]
func (a PublicController) UserLogin(ctx echo.Context) error {
	login := new(dto.Login)

	if err := ctx.Bind(login); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	// Only verify captcha if enabled in config
	if a.config.Captcha != nil && a.config.Captcha.Enable {
		if !a.captcha.Verify(login.CaptchaID, login.CaptchaCode, false) {
			return echox.Response{Code: http.StatusBadRequest, Message: errors.CaptchaAnswerCodeNoMatch}.JSON(ctx)
		}
	}

	user, err := a.userService.Verify(login.Username, login.Password)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	loginResp, err := a.authService.GenerateToken(user)
	if err != nil {
		return echox.Response{Code: http.StatusInternalServerError, Message: errors.AuthTokenGenerateFail}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: loginResp}.JSON(ctx)
}

// @tags Auth
// @summary 用户登出
// @produce application/json
// @success 200 {object} echox.Response "success"
// @router /api/v1/auth/logout [delete]
func (a PublicController) UserLogout(ctx echo.Context) error {
	claims, ok := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	if ok {
		a.authService.DestroyToken(claims.Username)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
