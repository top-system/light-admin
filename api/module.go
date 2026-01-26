package api

import (
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/api/platform"
	platformRoute "github.com/top-system/light-admin/api/platform/route"
	"github.com/top-system/light-admin/api/system"
	systemRoute "github.com/top-system/light-admin/api/system/route"

	"go.uber.org/fx"
)

// Module exports all api modules
// 添加新模块时，只需要在这里添加即可
var Module = fx.Options(
	middlewares.Module,
	system.Module,
	platform.Module,
	fx.Provide(NewRoutes),
)

// Routes 聚合所有模块的路由
type Routes struct {
	System   systemRoute.Routes
	Platform platformRoute.Routes
}

// NewRoutes creates aggregated routes
func NewRoutes(
	system systemRoute.Routes,
	platform platformRoute.Routes,
) Routes {
	return Routes{
		System:   system,
		Platform: platform,
	}
}

// Setup sets up all routes
func (r Routes) Setup() {
	r.System.Setup()
	r.Platform.Setup()
}
