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

type MenuController struct {
	menuService service.MenuService
	userService service.UserService
	logger      lib.Logger
}

// NewMenuController creates new menu controller
func NewMenuController(
	logger lib.Logger,
	menuService service.MenuService,
	userService service.UserService,
) MenuController {
	return MenuController{
		logger:      logger,
		menuService: menuService,
		userService: userService,
	}
}

// @tags Menu
// @summary Menu Query (Tree Structure)
// @produce application/json
// @param keywords query string false "搜索关键词"
// @success 200 {object} echox.Response{data=[]system.MenuTree} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/menus [get]
func (a MenuController) Query(ctx echo.Context) error {
	param := new(system.MenuQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	// 菜单列表返回树形结构，不分页
	param.PaginationParam.PageSize = 9999
	param.PaginationParam.PageNum = 1
	param.OrderParam.Key = "sort"
	param.OrderParam.Direction = "ASC"

	qr, err := a.menuService.Query(param)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	// 如果有搜索关键词，返回扁平列表；否则返回树形结构
	if param.Keywords != "" {
		return echox.Response{Code: http.StatusOK, Data: qr.List.ToFlatMenuTrees()}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: qr.List.ToMenuTrees()}.JSON(ctx)
}

// @tags Menu
// @summary Get Menu Form Data
// @produce application/json
// @param id path int true "menu id"
// @success 200 {object} echox.Response{data=system.MenuForm} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/menus/{id}/form [get]
func (a MenuController) GetForm(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	menu, err := a.menuService.Get(id)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: menu.ToForm()}.JSON(ctx)
}

// @tags Menu
// @summary Menu Create
// @produce application/json
// @param data body system.MenuForm true "Menu"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/menus [post]
func (a MenuController) Create(ctx echo.Context) error {
	form := new(system.MenuForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)

	id, err := a.menuService.WithTrx(trxHandle).Create(form.ToMenu())
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: echo.Map{"id": id}}.JSON(ctx)
}

// @tags Menu
// @summary Menu Update By ID
// @produce application/json
// @param id path int true "menu id"
// @param data body system.MenuForm true "Menu"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/menus/{id} [put]
func (a MenuController) Update(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form := new(system.MenuForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.menuService.WithTrx(trxHandle).Update(id, form.ToMenu()); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// @tags Menu
// @summary Menu Delete By ID
// @produce application/json
// @param id path int true "menu id"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/menus/{id} [delete]
func (a MenuController) Delete(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.menuService.WithTrx(trxHandle).Delete(id); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// @tags Menu
// @summary Get Menu Options
// @produce application/json
// @param onlyParent query bool false "only parent menus"
// @success 200 {object} echox.Response{data=[]dto.MenuOption} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/menus/options [get]
func (a MenuController) GetOptions(ctx echo.Context) error {
	onlyParent := ctx.QueryParam("onlyParent") == "true"

	options, err := a.menuService.ListMenuOptions(onlyParent)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: options}.JSON(ctx)
}

// @tags Menu
// @summary Get Current User Routes
// @produce application/json
// @success 200 {object} echox.Response{data=[]dto.RouteVO} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 401 {object} echox.Response "unauthorized"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/menus/routes [get]
func (a MenuController) Routes(ctx echo.Context) error {
	claims, ok := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	if !ok || claims == nil {
		return echox.Response{Code: http.StatusUnauthorized, Message: "unauthorized"}.JSON(ctx)
	}

	// 获取用户角色ID列表
	roleIDs, err := a.userService.GetUserRoleIDs(claims.ID)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	// 检查是否是超级管理员
	isSuperAdmin := claims.Username == "root" || claims.Username == "admin"

	// 使用 menuService.GetUserRoutes 获取正确格式的路由
	routes, err := a.menuService.GetUserRoutes(roleIDs, isSuperAdmin)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: routes}.JSON(ctx)
}
