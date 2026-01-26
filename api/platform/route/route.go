package route

import "go.uber.org/fx"

// Module exports dependency to container
var Module = fx.Options(
	fx.Provide(NewFileRoute),
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
	fileRoute FileRoute,
) Routes {
	return Routes{
		fileRoute,
	}
}

// Setup all the route
func (a Routes) Setup() {
	for _, route := range a {
		route.Setup()
	}
}
