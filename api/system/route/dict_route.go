package route

import (
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type DictRoutes struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	dictController controller.DictController
	permMiddleware middlewares.PermissionMiddleware
}

// NewDictRoutes creates new dict routes
func NewDictRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	dictController controller.DictController,
	permMiddleware middlewares.PermissionMiddleware,
) DictRoutes {
	return DictRoutes{
		handler:        handler,
		logger:         logger,
		dictController: dictController,
		permMiddleware: permMiddleware,
	}
}

// Setup dict routes
func (a DictRoutes) Setup() {
	a.logger.Zap.Info("Setting up dict routes")
	api := a.handler.RouterV1.Group("/dicts")
	{
		// 字典相关接口
		api.GET("", a.dictController.GetDictPage, a.permMiddleware.RequirePerm("sys:dict:query"))
		api.GET("/:id/form", a.dictController.GetDictForm, a.permMiddleware.RequirePerm("sys:dict:query"))
		api.POST("", a.dictController.SaveDict, a.permMiddleware.RequirePerm("sys:dict:add"))
		api.PUT("/:id", a.dictController.UpdateDict, a.permMiddleware.RequirePerm("sys:dict:edit"))
		api.DELETE("/:ids", a.dictController.DeleteDict, a.permMiddleware.RequirePerm("sys:dict:delete"))

		// 字典项相关接口
		api.GET("/:dictCode/items", a.dictController.GetDictItems, a.permMiddleware.RequirePerm("sys:dict-item:query"))
		api.GET("/:dictCode/items/options", a.dictController.GetDictItemOptions) // 无需权限，用于下拉选项
		api.GET("/:dictCode/items/:itemId/form", a.dictController.GetDictItemForm, a.permMiddleware.RequirePerm("sys:dict-item:query"))
		api.POST("/:dictCode/items", a.dictController.SaveDictItem, a.permMiddleware.RequirePerm("sys:dict-item:add"))
		api.PUT("/:dictCode/items/:itemId", a.dictController.UpdateDictItem, a.permMiddleware.RequirePerm("sys:dict-item:edit"))
		api.DELETE("/:dictCode/items/:itemIds", a.dictController.DeleteDictItem, a.permMiddleware.RequirePerm("sys:dict-item:delete"))
	}
}
