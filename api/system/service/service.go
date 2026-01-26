package service

import "go.uber.org/fx"

// Module exports services present
var Module = fx.Options(
	fx.Provide(NewUserService),
	fx.Provide(NewRoleService),
	fx.Provide(NewMenuService),
	fx.Provide(NewPermissionService),
	fx.Provide(NewAuthService),
	fx.Provide(NewConfigService),
	fx.Provide(NewNoticeService),
	fx.Provide(NewDeptService),
	fx.Provide(NewDictService),
	fx.Provide(NewDictItemService),
	fx.Provide(NewLogService),
)
