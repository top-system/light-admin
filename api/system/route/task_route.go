package route

import (
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/lib"
)

type TaskRoutes struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	taskController controller.TaskController
	permMiddleware middlewares.PermissionMiddleware
}

// NewTaskRoutes creates new task routes
func NewTaskRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	taskController controller.TaskController,
	permMiddleware middlewares.PermissionMiddleware,
) TaskRoutes {
	return TaskRoutes{
		handler:        handler,
		logger:         logger,
		taskController: taskController,
		permMiddleware: permMiddleware,
	}
}

// Setup task routes
func (a TaskRoutes) Setup() {
	a.logger.Zap.Info("Setting up task routes")
	api := a.handler.RouterV1.Group("/tasks")
	{
		api.GET("/stats", a.taskController.GetStats)   // 获取队列统计信息，无需特定权限
		api.GET("/types", a.taskController.GetTypes)   // 获取任务类型列表，无需特定权限
		api.GET("", a.taskController.Query, a.permMiddleware.RequirePerm("sys:task:query"))
		api.GET("/:id", a.taskController.Get, a.permMiddleware.RequirePerm("sys:task:query"))
		api.DELETE("/:id", a.taskController.Delete, a.permMiddleware.RequirePerm("sys:task:delete"))
	}
}
