package controller

import "go.uber.org/fx"

// Module exported for initializing application
var Module = fx.Options(
	fx.Provide(NewPublicController),
	fx.Provide(NewCaptchaController),
	fx.Provide(NewUserController),
	fx.Provide(NewRoleController),
	fx.Provide(NewMenuController),
	fx.Provide(NewConfigController),
	fx.Provide(NewNoticeController),
	fx.Provide(NewDeptController),
	fx.Provide(NewDictController),
	fx.Provide(NewLogController),
	fx.Provide(NewTaskController),
	fx.Provide(NewDownloadController),
)
