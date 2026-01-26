package platform

import (
	"github.com/top-system/light-admin/api/platform/controller"
	"github.com/top-system/light-admin/api/platform/repository"
	"github.com/top-system/light-admin/api/platform/route"
	"github.com/top-system/light-admin/api/platform/service"

	"go.uber.org/fx"
)

// Module exports platform module
var Module = fx.Options(
	controller.Module,
	service.Module,
	repository.Module,
	route.Module,
)
