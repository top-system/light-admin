package route

import (
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type UserRoutes struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	userController controller.UserController
	permMiddleware middlewares.PermissionMiddleware
}

// NewUserRoutes creates new user routes
func NewUserRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	userController controller.UserController,
	permMiddleware middlewares.PermissionMiddleware,
) UserRoutes {
	return UserRoutes{
		handler:        handler,
		logger:         logger,
		userController: userController,
		permMiddleware: permMiddleware,
	}
}

// Setup user routes
func (a UserRoutes) Setup() {
	api := a.handler.RouterV1.Group("/users")
	{
		api.GET("/me", a.userController.Me)                  // 获取当前用户信息，无需权限
		api.GET("/profile", a.userController.Me)             // 兼容 /profile 路径
		api.PUT("/profile", a.userController.UpdateProfile)  // 更新当前用户资料，无需权限
		api.GET("/options", a.userController.GetOptions)     // 用户下拉选项，无需权限
		api.GET("", a.userController.Query, a.permMiddleware.RequirePerm("sys:user:query"))
		api.POST("", a.userController.Create, a.permMiddleware.RequirePerm("sys:user:add"))
		api.GET("/:id/form", a.userController.GetForm, a.permMiddleware.RequirePerm("sys:user:query"))
		api.PUT("/:id", a.userController.Update, a.permMiddleware.RequirePerm("sys:user:edit"))
		api.DELETE("/:id", a.userController.Delete, a.permMiddleware.RequirePerm("sys:user:delete"))
		api.PUT("/:id/password/reset", a.userController.ResetPassword, a.permMiddleware.RequirePerm("sys:user:reset-password"))
	}
}
