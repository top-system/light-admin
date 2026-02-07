package middlewares

import (
	"net/http"

	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/pkg/echox"
	"github.com/labstack/echo/v4"
)

// PermissionMiddleware middleware for permission checking
type PermissionMiddleware struct {
	handler           lib.HttpHandler
	logger            lib.Logger
	config            lib.Config
	permissionService service.PermissionService
	userService       service.UserService
}

// NewPermissionMiddleware creates new permission middleware
func NewPermissionMiddleware(
	handler lib.HttpHandler,
	logger lib.Logger,
	config lib.Config,
	permissionService service.PermissionService,
	userService service.UserService,
) PermissionMiddleware {
	return PermissionMiddleware{
		handler:           handler,
		logger:            logger,
		config:            config,
		permissionService: permissionService,
		userService:       userService,
	}
}

func (a PermissionMiddleware) core() echo.MiddlewareFunc {
	prefixes := a.config.Casbin.IgnorePathPrefixes

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			request := ctx.Request()
			if isIgnorePath(request.URL.Path, prefixes...) {
				return next(ctx)
			}

			claims, ok := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
			if !ok {
				return echox.Response{Code: http.StatusUnauthorized}.JSON(ctx)
			}

			// For now, just pass through if user is authenticated
			// The permission check can be done at the controller level
			// by checking the perm identifier
			_ = claims

			return next(ctx)
		}
	}
}

// RequirePerm returns a middleware that checks for a specific permission
func (a PermissionMiddleware) RequirePerm(perm string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			claims, ok := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
			if !ok {
				return echox.Response{Code: http.StatusUnauthorized, Message: "未授权"}.JSON(ctx)
			}

			// 超级管理员跳过权限检查
			if a.userService.IsSuperAdmin(claims.Username) {
				return next(ctx)
			}

			// 获取用户角色
			roleIDs, err := a.permissionService.GetUserRoleIDs(claims.ID)
			if err != nil {
				return echox.Response{Code: http.StatusInternalServerError, Message: err}.JSON(ctx)
			}

			// 检查权限
			hasPerm, err := a.permissionService.HasPerm(roleIDs, perm)
			if err != nil {
				return echox.Response{Code: http.StatusInternalServerError, Message: err}.JSON(ctx)
			}

			if !hasPerm {
				return echox.Response{Code: http.StatusForbidden, Message: "没有操作权限"}.JSON(ctx)
			}

			return next(ctx)
		}
	}
}

func (a PermissionMiddleware) Setup() {
	if !a.config.Casbin.Enable {
		return
	}

	a.handler.Engine.Use(a.core())
}

// CasbinMiddleware alias for backward compatibility
type CasbinMiddleware = PermissionMiddleware

// NewCasbinMiddleware alias for backward compatibility
func NewCasbinMiddleware(
	handler lib.HttpHandler,
	logger lib.Logger,
	config lib.Config,
	permissionService service.PermissionService,
	userService service.UserService,
) CasbinMiddleware {
	return NewPermissionMiddleware(handler, logger, config, permissionService, userService)
}
