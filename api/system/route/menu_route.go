package route

import (
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type MenuRoutes struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	menuController controller.MenuController
	permMiddleware middlewares.PermissionMiddleware
}

// NewMenuRoutes creates new menu routes
func NewMenuRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	menuController controller.MenuController,
	permMiddleware middlewares.PermissionMiddleware,
) MenuRoutes {
	return MenuRoutes{
		handler:        handler,
		logger:         logger,
		menuController: menuController,
		permMiddleware: permMiddleware,
	}
}

// Setup menu routes
func (a MenuRoutes) Setup() {
	api := a.handler.RouterV1.Group("/menus")
	{
		api.GET("", a.menuController.Query, a.permMiddleware.RequirePerm("sys:menu:query"))
		api.GET("/routes", a.menuController.Routes)      // 获取路由，无需权限（用于动态路由）
		api.GET("/options", a.menuController.GetOptions) // 下拉选项，无需权限

		api.POST("", a.menuController.Create, a.permMiddleware.RequirePerm("sys:menu:add"))
		api.GET("/:id/form", a.menuController.GetForm, a.permMiddleware.RequirePerm("sys:menu:query"))
		api.PUT("/:id", a.menuController.Update, a.permMiddleware.RequirePerm("sys:menu:edit"))
		api.DELETE("/:id", a.menuController.Delete, a.permMiddleware.RequirePerm("sys:menu:delete"))
	}
}
