package route

import (
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/lib"
)

type PublicRoutes struct {
	logger            lib.Logger
	handler           lib.HttpHandler
	publicController  controller.PublicController
	captchaController controller.CaptchaController
}

// NewUserRoutes creates new public routes
func NewPublicRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	publicController controller.PublicController,
	captchaController controller.CaptchaController,
) PublicRoutes {
	return PublicRoutes{
		handler:           handler,
		logger:            logger,
		publicController:  publicController,
		captchaController: captchaController,
	}
}

// Setup public routes
func (a PublicRoutes) Setup() {
	// /api/v1/auth 路由组
	auth := a.handler.RouterV1.Group("/auth")
	{
		auth.POST("/login", a.publicController.UserLogin)
		auth.DELETE("/logout", a.publicController.UserLogout)
		auth.GET("/captcha", a.captchaController.GetCaptcha)
		auth.POST("/captcha/verify", a.captchaController.VerifyCaptcha)
	}
}
