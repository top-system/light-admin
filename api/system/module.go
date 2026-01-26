package system

import (
	"github.com/top-system/light-admin/api/system/controller"
	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/api/system/route"
	"github.com/top-system/light-admin/api/system/service"

	"go.uber.org/fx"
)

// Module exports system module
var Module = fx.Options(
	controller.Module,
	service.Module,
	repository.Module,
	route.Module,
)
