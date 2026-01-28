package lib

import "go.uber.org/fx"

// Module exports dependency
var Module = fx.Options(
	fx.Provide(NewHttpHandler),
	fx.Provide(NewConfig),
	fx.Provide(NewLogger),
	fx.Provide(NewDatabase),
	fx.Provide(NewDBCompat),
	fx.Provide(NewCache),
	fx.Provide(NewCaptcha),
	fx.Provide(NewWebSocket),
)
