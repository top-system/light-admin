package route

import (
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type DeptRoutes struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	deptController controller.DeptController
	permMiddleware middlewares.PermissionMiddleware
}

// NewDeptRoutes creates new dept routes
func NewDeptRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	deptController controller.DeptController,
	permMiddleware middlewares.PermissionMiddleware,
) DeptRoutes {
	return DeptRoutes{
		handler:        handler,
		logger:         logger,
		deptController: deptController,
		permMiddleware: permMiddleware,
	}
}

// Setup dept routes
func (a DeptRoutes) Setup() {
	a.logger.Zap.Info("Setting up dept routes")
	api := a.handler.RouterV1.Group("/depts")
	{
		api.GET("", a.deptController.Query, a.permMiddleware.RequirePerm("sys:dept:query"))
		api.GET("/options", a.deptController.GetOptions) // 下拉选项，无需权限
		api.GET("/:deptId/form", a.deptController.GetForm, a.permMiddleware.RequirePerm("sys:dept:query"))
		api.POST("", a.deptController.Create, a.permMiddleware.RequirePerm("sys:dept:add"))
		api.PUT("/:deptId", a.deptController.Update, a.permMiddleware.RequirePerm("sys:dept:edit"))
		api.DELETE("/:ids", a.deptController.Delete, a.permMiddleware.RequirePerm("sys:dept:delete"))
	}
}
