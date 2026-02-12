package middlewares

import (
	"net/http"

	"github.com/top-system/light-admin/lib"
	"github.com/labstack/echo/v4/middleware"
)

// CorsMiddleware middleware for cors
type CorsMiddleware struct {
	handler lib.HttpHandler
	logger  lib.Logger
	config  lib.Config
}

// NewCorsMiddleware creates new cors middleware
func NewCorsMiddleware(handler lib.HttpHandler, logger lib.Logger, config lib.Config) CorsMiddleware {
	return CorsMiddleware{
		handler: handler,
		logger:  logger,
		config:  config,
	}
}

func (a CorsMiddleware) Setup() {
	allowOrigins := a.config.Http.AllowOrigins

	a.handler.Engine.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) {
			// 如果未配置白名单，允许所有来源（开发环境）
			if len(allowOrigins) == 0 {
				return true, nil
			}
			// 生产环境：只允许配置的域名
			for _, allowed := range allowOrigins {
				if origin == allowed {
					return true, nil
				}
			}
			return false, nil
		},
		AllowCredentials: true,
		AllowHeaders: []string{
			"Authorization", "Content-Type", "Accept", "Origin",
			"X-Requested-With", "X-Request-ID",
		},
		AllowMethods: []string{
			http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodDelete, http.MethodOptions, http.MethodPatch,
		},
	}))
}
