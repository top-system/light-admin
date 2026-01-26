package route

import (
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type NoticeRoutes struct {
	logger           lib.Logger
	handler          lib.HttpHandler
	noticeController controller.NoticeController
	permMiddleware   middlewares.PermissionMiddleware
}

// NewNoticeRoutes creates new notice routes
func NewNoticeRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	noticeController controller.NoticeController,
	permMiddleware middlewares.PermissionMiddleware,
) NoticeRoutes {
	return NoticeRoutes{
		handler:          handler,
		logger:           logger,
		noticeController: noticeController,
		permMiddleware:   permMiddleware,
	}
}

// Setup notice routes
func (a NoticeRoutes) Setup() {
	a.logger.Zap.Info("Setting up notice routes")
	api := a.handler.RouterV1.Group("/notices")
	{
		// 管理端接口
		api.GET("", a.noticeController.Query, a.permMiddleware.RequirePerm("sys:notice:query"))
		api.GET("/:id/form", a.noticeController.GetForm, a.permMiddleware.RequirePerm("sys:notice:query"))
		api.GET("/:id/detail", a.noticeController.GetDetail, a.permMiddleware.RequirePerm("sys:notice:query"))
		api.POST("", a.noticeController.Create, a.permMiddleware.RequirePerm("sys:notice:add"))
		api.PUT("/:id", a.noticeController.Update, a.permMiddleware.RequirePerm("sys:notice:edit"))
		api.DELETE("/:ids", a.noticeController.Delete, a.permMiddleware.RequirePerm("sys:notice:delete"))
		api.PUT("/:id/publish", a.noticeController.Publish, a.permMiddleware.RequirePerm("sys:notice:publish"))
		api.PUT("/:id/revoke", a.noticeController.Revoke, a.permMiddleware.RequirePerm("sys:notice:revoke"))

		// 用户端接口（无需特殊权限，登录即可）
		api.GET("/my", a.noticeController.GetMyNoticePage)
		api.PUT("/read-all", a.noticeController.ReadAll)
	}
}
