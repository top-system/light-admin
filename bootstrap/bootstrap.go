package bootstrap

import (
	"context"
	"net/http"
	"time"

	"github.com/top-system/light-admin/api"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/api/platform/controller"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"

	"go.uber.org/fx"
)

// Module exported for initializing application
var Module = fx.Options(
	lib.Module,
	api.Module,
	fx.Invoke(bootstrap),
)

func bootstrap(
	lifecycle fx.Lifecycle,
	handler lib.HttpHandler,
	routes api.Routes,
	logger lib.Logger,
	config lib.Config,
	middlewares middlewares.Middlewares,
	database lib.Database,
	websocketController controller.WebSocketController, // 注入 WebSocket 控制器
) {
	db, err := database.ORM.DB()
	if err != nil {
		logger.Zap.Fatalf("Error to get database connection: %v", err)
	}

	// server 提升到外层，供 OnStop 使用
	var server *http.Server

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err := db.Ping(); err != nil {
				logger.Zap.Fatalf("Error to ping database connection: %v", err)
			}

			// set conn
			db.SetMaxOpenConns(config.Database.MaxOpenConns)
			db.SetMaxIdleConns(config.Database.MaxIdleConns)
			db.SetConnMaxLifetime(time.Duration(config.Database.MaxLifetime) * time.Second)
			db.SetConnMaxIdleTime(10 * time.Minute)

			go func() {
				middlewares.Setup()
				routes.Setup()

				// 创建自定义 HTTP 处理器，在 Echo 之前拦截 WebSocket 请求
				wsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// 拦截 /ws 请求，直接处理 WebSocket，不经过 Echo 中间件
					if r.URL.Path == "/ws" {
						websocketController.HandleWebSocket(w, r)
						return
					}
					// 其他请求交给 Echo 处理
					handler.Engine.ServeHTTP(w, r)
				})

				server = &http.Server{
					Addr:           config.Http.ListenAddr(),
					Handler:        wsHandler,
					ReadTimeout:    15 * time.Second,
					WriteTimeout:   15 * time.Second,
					IdleTimeout:    60 * time.Second,
					MaxHeaderBytes: 1 << 20, // 1MB
				}

				logger.Zap.Infof("Server started on %s", config.Http.ListenAddr())
				if err := server.ListenAndServe(); err != nil {
					if errors.Is(err, http.ErrServerClosed) {
						logger.Zap.Debug("Shutting down the Application")
					} else {
						logger.Zap.Fatalf("Error to Start Application: %v", err)
					}
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Zap.Info("Stopping Application")

			// Graceful shutdown: 等待在途请求完成（最多30秒）
			if server != nil {
				shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				defer cancel()
				if err := server.Shutdown(shutdownCtx); err != nil {
					logger.Zap.Errorf("Server forced shutdown: %v", err)
				}
			}

			db.Close()
			return nil
		},
	})
}
