package middlewares

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"

	"github.com/labstack/echo/v4"

	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/pkg/echox"
)

// RateLimitMiddleware 请求限流中间件
type RateLimitMiddleware struct {
	handler  lib.HttpHandler
	logger   lib.Logger
	visitors sync.Map // ip -> *rate.Limiter
}

// NewRateLimitMiddleware creates new rate limit middleware
func NewRateLimitMiddleware(handler lib.HttpHandler, logger lib.Logger) RateLimitMiddleware {
	return RateLimitMiddleware{
		handler: handler,
		logger:  logger,
	}
}

// getVisitor 获取或创建访问者的限流器
func (a RateLimitMiddleware) getVisitor(ip string) *rate.Limiter {
	if v, ok := a.visitors.Load(ip); ok {
		return v.(*rate.Limiter)
	}

	// 每秒 10 个请求，突发 20 个
	limiter := rate.NewLimiter(10, 20)
	actual, _ := a.visitors.LoadOrStore(ip, limiter)
	return actual.(*rate.Limiter)
}

func (a RateLimitMiddleware) Setup() {
	a.handler.Engine.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ip := ctx.RealIP()
			limiter := a.getVisitor(ip)

			if !limiter.Allow() {
				return echox.Response{
					Code:    http.StatusTooManyRequests,
					Message: "Too many requests",
				}.JSON(ctx)
			}

			return next(ctx)
		}
	})
}
