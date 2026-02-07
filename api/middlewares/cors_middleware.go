package middlewares

import (
	"github.com/top-system/light-admin/lib"
	"github.com/labstack/echo/v4/middleware"
)

// CorsMiddleware middleware for cors
type CorsMiddleware struct {
	handler lib.HttpHandler
	logger  lib.Logger
}

// NewCorsMiddleware creates new cors middleware
func NewCorsMiddleware(handler lib.HttpHandler, logger lib.Logger) CorsMiddleware {
	return CorsMiddleware{
		handler: handler,
		logger:  logger,
	}
}

func (a CorsMiddleware) Setup() {
	a.handler.Engine.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) {
			// 允许所有来源（开发环境）
			// 生产环境应该配置具体的域名白名单
			return true, nil
		},
		AllowCredentials: true,
		AllowHeaders:     []string{"*"},
		AllowMethods:     []string{"*"},
	}))
}
