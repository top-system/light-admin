package route

import (
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/lib"
)

type DownloadRoutes struct {
	logger             lib.Logger
	handler            lib.HttpHandler
	downloadController controller.DownloadController
	permMiddleware     middlewares.PermissionMiddleware
}

// NewDownloadRoutes creates new download routes
func NewDownloadRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	downloadController controller.DownloadController,
	permMiddleware middlewares.PermissionMiddleware,
) DownloadRoutes {
	return DownloadRoutes{
		handler:            handler,
		logger:             logger,
		downloadController: downloadController,
		permMiddleware:     permMiddleware,
	}
}

// Setup download routes
func (a DownloadRoutes) Setup() {
	a.logger.Zap.Info("Setting up download routes")
	api := a.handler.RouterV1.Group("/downloads")
	{
		api.GET("/stats", a.downloadController.GetStats)                // 获取统计信息
		api.GET("/downloaders", a.downloadController.GetDownloaders)    // 获取下载器列表
		api.GET("/test/:name", a.downloadController.TestDownloader)     // 测试下载器
		api.GET("", a.downloadController.Query, a.permMiddleware.RequirePerm("sys:download:query"))
		api.GET("/:id", a.downloadController.Get, a.permMiddleware.RequirePerm("sys:download:query"))
		api.POST("", a.downloadController.Create, a.permMiddleware.RequirePerm("sys:download:add"))
		api.POST("/:id/cancel", a.downloadController.Cancel, a.permMiddleware.RequirePerm("sys:download:edit"))
		api.PUT("/:id/files", a.downloadController.SetFiles, a.permMiddleware.RequirePerm("sys:download:edit"))
		api.POST("/:id/sync", a.downloadController.Sync, a.permMiddleware.RequirePerm("sys:download:query"))
		api.DELETE("/:id", a.downloadController.Delete, a.permMiddleware.RequirePerm("sys:download:delete"))
	}
}
