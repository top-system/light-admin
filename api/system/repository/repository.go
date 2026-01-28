package repository

import "go.uber.org/fx"

// Module exports dependency
var Module = fx.Options(
	fx.Provide(NewUserRepository),
	fx.Provide(NewUserRoleRepository),
	fx.Provide(NewRoleRepository),
	fx.Provide(NewRoleMenuRepository),
	fx.Provide(NewMenuRepository),
	fx.Provide(NewConfigRepository),
	fx.Provide(NewNoticeRepository),
	fx.Provide(NewUserNoticeRepository),
	fx.Provide(NewDeptRepository),
	fx.Provide(NewDictRepository),
	fx.Provide(NewDictItemRepository),
	fx.Provide(NewLogRepository),
	fx.Provide(NewTaskRepository),
	fx.Provide(NewDownloadRepository),
)
