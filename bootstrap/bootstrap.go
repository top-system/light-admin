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

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err := db.Ping(); err != nil {
				logger.Zap.Fatalf("Error to ping database connection: %v", err)
			}

			// set conn
			db.SetMaxOpenConns(config.Database.MaxOpenConns)
			db.SetMaxIdleConns(config.Database.MaxIdleConns)
			db.SetConnMaxLifetime(time.Duration(config.Database.MaxLifetime) * time.Second)

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

				server := &http.Server{
					Addr:    config.Http.ListenAddr(),
					Handler: wsHandler,
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
		OnStop: func(context.Context) error {
			logger.Zap.Info("Stopping Application")

			handler.Engine.Close()
			db.Close()
			return nil
		},
	})
}
