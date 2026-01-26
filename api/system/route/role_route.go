package route

import (
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type RoleRoutes struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	roleController controller.RoleController
	permMiddleware middlewares.PermissionMiddleware
}

// NewRoleRoutes creates new role routes
func NewRoleRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	roleController controller.RoleController,
	permMiddleware middlewares.PermissionMiddleware,
) RoleRoutes {
	return RoleRoutes{
		handler:        handler,
		logger:         logger,
		roleController: roleController,
		permMiddleware: permMiddleware,
	}
}

// Setup role routes
func (a RoleRoutes) Setup() {
	a.logger.Zap.Info("Setting up role routes")
	api := a.handler.RouterV1.Group("/roles")
	{
		api.GET("", a.roleController.Query, a.permMiddleware.RequirePerm("sys:role:query"))
		api.GET("/options", a.roleController.GetOptions) // 下拉选项，无需权限

		api.POST("", a.roleController.Create, a.permMiddleware.RequirePerm("sys:role:add"))
		api.GET("/:id/form", a.roleController.GetForm, a.permMiddleware.RequirePerm("sys:role:query"))
		api.PUT("/:id", a.roleController.Update, a.permMiddleware.RequirePerm("sys:role:edit"))
		api.DELETE("/:id", a.roleController.Delete, a.permMiddleware.RequirePerm("sys:role:delete"))
		api.GET("/:id/menuIds", a.roleController.GetMenuIds, a.permMiddleware.RequirePerm("sys:role:query"))
		api.PUT("/:id/menus", a.roleController.AssignMenus, a.permMiddleware.RequirePerm("sys:role:edit"))
	}
}
