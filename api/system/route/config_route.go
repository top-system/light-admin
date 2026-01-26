package route

import (
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type ConfigRoutes struct {
	logger           lib.Logger
	handler          lib.HttpHandler
	configController controller.ConfigController
	permMiddleware   middlewares.PermissionMiddleware
}

// NewConfigRoutes creates new config routes
func NewConfigRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	configController controller.ConfigController,
	permMiddleware middlewares.PermissionMiddleware,
) ConfigRoutes {
	return ConfigRoutes{
		handler:          handler,
		logger:           logger,
		configController: configController,
		permMiddleware:   permMiddleware,
	}
}

// Setup config routes
func (a ConfigRoutes) Setup() {
	a.logger.Zap.Info("Setting up config routes")
	api := a.handler.RouterV1.Group("/configs")
	{
		api.GET("", a.configController.Query, a.permMiddleware.RequirePerm("sys:config:query"))
		api.GET("/:id/form", a.configController.GetForm, a.permMiddleware.RequirePerm("sys:config:query"))
		api.POST("", a.configController.Create, a.permMiddleware.RequirePerm("sys:config:add"))
		api.PUT("/:id", a.configController.Update, a.permMiddleware.RequirePerm("sys:config:update"))
		api.DELETE("/:id", a.configController.Delete, a.permMiddleware.RequirePerm("sys:config:delete"))
		api.PUT("/refresh", a.configController.RefreshCache, a.permMiddleware.RequirePerm("sys:config:refresh"))
	}
}
