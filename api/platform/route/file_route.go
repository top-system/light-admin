package route

import (
	"github.com/top-system/light-admin/api/platform/controller"
	"github.com/top-system/light-admin/lib"
)

// FileRoute 文件路由
type FileRoute struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	fileController controller.FileController
}

// NewFileRoute 创建文件路由
func NewFileRoute(
	logger lib.Logger,
	handler lib.HttpHandler,
	fileController controller.FileController,
) FileRoute {
	return FileRoute{
		logger:         logger,
		handler:        handler,
		fileController: fileController,
	}
}

// Setup 设置文件路由
func (r FileRoute) Setup() {
	r.logger.Zap.Info("Setting up file routes")

	api := r.handler.RouterV1.Group("/files")
	{
		api.POST("", r.fileController.Upload)
		api.DELETE("", r.fileController.Delete)
	}
}
