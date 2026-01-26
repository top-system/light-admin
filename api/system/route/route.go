package route

import "go.uber.org/fx"

// Module exports dependency to container
var Module = fx.Options(
	fx.Provide(NewPprofRoutes),
	fx.Provide(NewSwaggerRoutes),
	fx.Provide(NewPublicRoutes),
	fx.Provide(NewUserRoutes),
	fx.Provide(NewRoleRoutes),
	fx.Provide(NewMenuRoutes),
	fx.Provide(NewConfigRoutes),
	fx.Provide(NewNoticeRoutes),
	fx.Provide(NewDeptRoutes),
	fx.Provide(NewDictRoutes),
	fx.Provide(NewLogRoute),
	fx.Provide(NewRoutes),
)

// Routes contains multiple routes
type Routes []Route

// Route interface
type Route interface {
	Setup()
}

// NewRoutes sets up routes
func NewRoutes(
	pprofRoutes PprofRoutes,
	swaggerRoutes SwaggerRoutes,
	publicRoutes PublicRoutes,
	userRoutes UserRoutes,
	roleRoutes RoleRoutes,
	menuRoutes MenuRoutes,
	configRoutes ConfigRoutes,
	noticeRoutes NoticeRoutes,
	deptRoutes DeptRoutes,
	dictRoutes DictRoutes,
	logRoutes LogRoute,
) Routes {
	return Routes{
		pprofRoutes,
		swaggerRoutes,
		publicRoutes,
		userRoutes,
		roleRoutes,
		menuRoutes,
		configRoutes,
		noticeRoutes,
		deptRoutes,
		dictRoutes,
		logRoutes,
	}
}

// Setup all the route
func (a Routes) Setup() {
	for _, route := range a {
		route.Setup()
	}
}
