package route

import (
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

// LogRoute struct
type LogRoute struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	logController  controller.LogController
	permMiddleware middlewares.PermissionMiddleware
}

// NewLogRoute creates new log route
func NewLogRoute(
	logger lib.Logger,
	handler lib.HttpHandler,
	logController controller.LogController,
	permMiddleware middlewares.PermissionMiddleware,
) LogRoute {
	return LogRoute{
		logger:         logger,
		handler:        handler,
		logController:  logController,
		permMiddleware: permMiddleware,
	}
}

// Setup log routes
func (a LogRoute) Setup() {
	api := a.handler.RouterV1.Group("/logs")
	{
		api.GET("", a.logController.Query, a.permMiddleware.RequirePerm("sys:log:query"))
	}
}
