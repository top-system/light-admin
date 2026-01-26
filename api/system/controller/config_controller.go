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
)

type ConfigController struct {
	configService service.ConfigService
	logger        lib.Logger
}

// NewConfigController creates new config controller
func NewConfigController(
	configService service.ConfigService,
	logger lib.Logger,
) ConfigController {
	return ConfigController{
		configService: configService,
		logger:        logger,
	}
}

// Query 分页查询系统配置
// @Tags Config
// @Summary 分页查询系统配置
// @Produce application/json
// @Param keywords query string false "关键词"
// @Param current query int false "当前页"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} echox.Response{data=system.ConfigQueryResult} "ok"
// @Router /api/v1/configs [get]
func (a ConfigController) Query(ctx echo.Context) error {
	param := new(system.ConfigQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	qr, err := a.configService.Query(param)
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

// GetForm 获取系统配置表单数据
// @Tags Config
// @Summary 获取系统配置表单数据
// @Produce application/json
// @Param id path int true "配置ID"
// @Success 200 {object} echox.Response{data=system.ConfigForm} "ok"
// @Router /api/v1/configs/{id}/form [get]
func (a ConfigController) GetForm(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form, err := a.configService.GetForm(id)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: form}.JSON(ctx)
}

// Create 创建系统配置
// @Tags Config
// @Summary 创建系统配置
// @Produce application/json
// @Param data body system.ConfigForm true "配置信息"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/configs [post]
func (a ConfigController) Create(ctx echo.Context) error {
	form := new(system.ConfigForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var createdBy uint64
	if claims != nil {
		createdBy = claims.ID
	}

	if err := a.configService.Create(form, createdBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// Update 更新系统配置
// @Tags Config
// @Summary 更新系统配置
// @Produce application/json
// @Param id path int true "配置ID"
// @Param data body system.ConfigForm true "配置信息"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/configs/{id} [put]
func (a ConfigController) Update(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form := new(system.ConfigForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var updatedBy uint64
	if claims != nil {
		updatedBy = claims.ID
	}

	if err := a.configService.Update(id, form, updatedBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// Delete 删除系统配置
// @Tags Config
// @Summary 删除系统配置
// @Produce application/json
// @Param id path int true "配置ID"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/configs/{id} [delete]
func (a ConfigController) Delete(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var deletedBy uint64
	if claims != nil {
		deletedBy = claims.ID
	}

	if err := a.configService.Delete(id, deletedBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// RefreshCache 刷新配置缓存
// @Tags Config
// @Summary 刷新配置缓存
// @Produce application/json
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/configs/refresh [put]
func (a ConfigController) RefreshCache(ctx echo.Context) error {
	if err := a.configService.RefreshCache(); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
