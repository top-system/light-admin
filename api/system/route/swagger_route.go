package route

import (
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/docs"
	"github.com/top-system/light-admin/lib"
)

// @securityDefinitions.apikey Authorization
// @in header
// @name Authorization
// @schemes http https
// @basePath /
// @contact.name LiuSha
// @contact.email liusha@email.cn
type SwaggerRoutes struct {
	config  lib.Config
	logger  lib.Logger
	handler lib.HttpHandler
}

// NewUserRoutes creates new swagger routes
func NewSwaggerRoutes(
	config lib.Config,
	logger lib.Logger,
	handler lib.HttpHandler,
) SwaggerRoutes {
	return SwaggerRoutes{
		config:  config,
		logger:  logger,
		handler: handler,
	}
}

// Setup swagger routes
func (a SwaggerRoutes) Setup() {
	docs.SwaggerInfo.Title = a.config.Name
	docs.SwaggerInfo.Version = constants.Version

	a.handler.Engine.GET("/swagger/*", echoSwagger.WrapHandler)
}
