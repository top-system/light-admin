package controller

import (
	"net/http"
	"strconv"

	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/pkg/echox"
	"github.com/labstack/echo/v4"

	"gorm.io/gorm"
)

type RoleController struct {
	logger      lib.Logger
	roleService service.RoleService
}

// NewRoleController creates new role controller
func NewRoleController(
	logger lib.Logger,
	roleService service.RoleService,
) RoleController {
	return RoleController{
		logger:      logger,
		roleService: roleService,
	}
}

// @tags Role
// @summary Role Query
// @produce application/json
// @param data query system.RoleQueryParam true "RoleQueryParam"
// @success 200 {object} echox.Response{data=system.Roles} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/roles [get]
func (a RoleController) Query(ctx echo.Context) error {
	param := new(system.RoleQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	qr, err := a.roleService.Query(param)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{
		Code: http.StatusOK,
		Data: qr.List,
		Page: &echox.PageInfo{
			Total:    qr.Pagination.Total,
			PageNum:  qr.Pagination.PageNum,
			PageSize: qr.Pagination.PageSize,
		},
	}.JSON(ctx)
}

// @tags Role
// @summary Role Options
// @produce application/json
// @success 200 {object} echox.Response{data=[]system.RoleOption} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/roles/options [get]
func (a RoleController) GetOptions(ctx echo.Context) error {
	options, err := a.roleService.ListRoleOptions()
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: options}.JSON(ctx)
}

// @tags Role
// @summary Get Role Form Data
// @produce application/json
// @param id path int true "role id"
// @success 200 {object} echox.Response{data=system.Role} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/roles/{id}/form [get]
func (a RoleController) GetForm(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	role, err := a.roleService.Get(id)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: role}.JSON(ctx)
}

// @tags Role
// @summary Role Create
// @produce application/json
// @param data body system.Role true "Role"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/roles [post]
func (a RoleController) Create(ctx echo.Context) error {
	role := new(system.Role)
	if err := ctx.Bind(role); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	role.CreateBy = claims.ID

	id, err := a.roleService.WithTrx(trxHandle).Create(role)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: echo.Map{"id": id}}.JSON(ctx)
}

// @tags Role
// @summary Role Update By ID
// @produce application/json
// @param id path int true "role id"
// @param data body system.Role true "Role"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/roles/{id} [put]
func (a RoleController) Update(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	role := new(system.Role)
	if err := ctx.Bind(role); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	role.UpdateBy = claims.ID

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.roleService.WithTrx(trxHandle).Update(id, role); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// @tags Role
// @summary Role Delete By ID
// @produce application/json
// @param id path int true "role id"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/roles/{id} [delete]
func (a RoleController) Delete(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.roleService.WithTrx(trxHandle).Delete(id); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// @tags Role
// @summary Get Role Menu IDs
// @produce application/json
// @param id path int true "role id"
// @success 200 {object} echox.Response{data=[]uint64} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/roles/{id}/menus [get]
func (a RoleController) GetMenuIds(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	menuIDs, err := a.roleService.GetRoleMenuIds(id)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: menuIDs}.JSON(ctx)
}

// @tags Role
// @summary Assign Menus To Role
// @produce application/json
// @param id path int true "role id"
// @param data body []uint64 true "menu ids"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/roles/{id}/menus [put]
func (a RoleController) AssignMenus(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	var menuIDs []uint64
	if err := ctx.Bind(&menuIDs); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.roleService.WithTrx(trxHandle).AssignMenusToRole(id, menuIDs); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
