package controller

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/pkg/echox"
)

type DictController struct {
	dictService     service.DictService
	dictItemService service.DictItemService
	logger          lib.Logger
}

// NewDictController creates new dict controller
func NewDictController(
	dictService service.DictService,
	dictItemService service.DictItemService,
	logger lib.Logger,
) DictController {
	return DictController{
		dictService:     dictService,
		dictItemService: dictItemService,
		logger:          logger,
	}
}

// ========== 字典相关接口 ==========

// GetDictPage 字典分页列表
// @Tags Dict
// @Summary 字典分页列表
// @Produce application/json
// @Param keywords query string false "关键字"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} echox.Response{data=[]system.DictPageVO} "ok"
// @Router /api/v1/dicts [get]
func (a DictController) GetDictPage(ctx echo.Context) error {
	param := new(system.DictQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	qr, err := a.dictService.GetDictPage(param)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{
		Code: http.StatusOK,
		Data: qr.List.ToPageVOList(),
		Page: &echox.PageInfo{
			Total:    qr.Pagination.Total,
			PageNum:  qr.Pagination.PageNum,
			PageSize: qr.Pagination.PageSize,
		},
	}.JSON(ctx)
}

// GetDictForm 获取字典表单数据
// @Tags Dict
// @Summary 获取字典表单数据
// @Produce application/json
// @Param id path int true "字典ID"
// @Success 200 {object} echox.Response{data=system.DictForm} "ok"
// @Router /api/v1/dicts/{id}/form [get]
func (a DictController) GetDictForm(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form, err := a.dictService.GetDictForm(id)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: form}.JSON(ctx)
}

// SaveDict 新增字典
// @Tags Dict
// @Summary 新增字典
// @Produce application/json
// @Param data body system.DictForm true "字典信息"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/dicts [post]
func (a DictController) SaveDict(ctx echo.Context) error {
	form := new(system.DictForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var createdBy uint64
	if claims != nil {
		createdBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.dictService.WithTrx(trxHandle).SaveDict(form, createdBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// UpdateDict 修改字典
// @Tags Dict
// @Summary 修改字典
// @Produce application/json
// @Param id path int true "字典ID"
// @Param data body system.DictForm true "字典信息"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/dicts/{id} [put]
func (a DictController) UpdateDict(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form := new(system.DictForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var updatedBy uint64
	if claims != nil {
		updatedBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.dictService.WithTrx(trxHandle).UpdateDict(id, form, updatedBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// DeleteDict 删除字典
// @Tags Dict
// @Summary 删除字典
// @Produce application/json
// @Param ids path string true "字典ID，多个以英文逗号分割"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/dicts/{ids} [delete]
func (a DictController) DeleteDict(ctx echo.Context) error {
	ids := ctx.Param("ids")

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var deletedBy uint64
	if claims != nil {
		deletedBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.dictService.WithTrx(trxHandle).DeleteDictByIds(ids, deletedBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// ========== 字典项相关接口 ==========

// GetDictItems 字典项列表（支持分页）
// @Tags Dict
// @Summary 字典项列表
// @Produce application/json
// @Param dictCode path string true "字典编码"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} echox.Response{data=system.DictItemQueryResult} "ok"
// @Router /api/v1/dicts/{dictCode}/items [get]
func (a DictController) GetDictItems(ctx echo.Context) error {
	dictCode := ctx.Param("dictCode")

	param := new(system.DictItemQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	param.DictCode = dictCode

	qr, err := a.dictItemService.GetDictItemPage(param)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{
		Code: http.StatusOK,
		Data: qr.List.ToPageVOList(),
		Page: &echox.PageInfo{
			Total:    qr.Pagination.Total,
			PageNum:  qr.Pagination.PageNum,
			PageSize: qr.Pagination.PageSize,
		},
	}.JSON(ctx)
}

// GetDictItemOptions 字典项下拉选项（无分页）
// @Tags Dict
// @Summary 字典项下拉选项
// @Produce application/json
// @Param dictCode path string true "字典编码"
// @Success 200 {object} echox.Response{data=[]system.DictItemOptionVO} "ok"
// @Router /api/v1/dicts/{dictCode}/items/options [get]
func (a DictController) GetDictItemOptions(ctx echo.Context) error {
	dictCode := ctx.Param("dictCode")

	list, err := a.dictItemService.GetDictItems(dictCode)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: list}.JSON(ctx)
}

// GetDictItemForm 获取字典项表单数据
// @Tags Dict
// @Summary 获取字典项表单数据
// @Produce application/json
// @Param dictCode path string true "字典编码"
// @Param itemId path int true "字典项ID"
// @Success 200 {object} echox.Response{data=system.DictItemForm} "ok"
// @Router /api/v1/dicts/{dictCode}/items/{itemId}/form [get]
func (a DictController) GetDictItemForm(ctx echo.Context) error {
	itemId, err := strconv.ParseUint(ctx.Param("itemId"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form, err := a.dictItemService.GetDictItemForm(itemId)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: form}.JSON(ctx)
}

// SaveDictItem 新增字典项
// @Tags Dict
// @Summary 新增字典项
// @Produce application/json
// @Param dictCode path string true "字典编码"
// @Param data body system.DictItemForm true "字典项信息"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/dicts/{dictCode}/items [post]
func (a DictController) SaveDictItem(ctx echo.Context) error {
	dictCode := ctx.Param("dictCode")

	form := new(system.DictItemForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	form.DictCode = dictCode

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var createdBy uint64
	if claims != nil {
		createdBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.dictItemService.WithTrx(trxHandle).SaveDictItem(form, createdBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// UpdateDictItem 修改字典项
// @Tags Dict
// @Summary 修改字典项
// @Produce application/json
// @Param dictCode path string true "字典编码"
// @Param itemId path int true "字典项ID"
// @Param data body system.DictItemForm true "字典项信息"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/dicts/{dictCode}/items/{itemId} [put]
func (a DictController) UpdateDictItem(ctx echo.Context) error {
	dictCode := ctx.Param("dictCode")
	itemId, err := strconv.ParseUint(ctx.Param("itemId"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form := new(system.DictItemForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	form.DictCode = dictCode

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var updatedBy uint64
	if claims != nil {
		updatedBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.dictItemService.WithTrx(trxHandle).UpdateDictItem(itemId, form, updatedBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// DeleteDictItem 删除字典项
// @Tags Dict
// @Summary 删除字典项
// @Produce application/json
// @Param dictCode path string true "字典编码"
// @Param itemIds path string true "字典项ID，多个以英文逗号分割"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/dicts/{dictCode}/items/{itemIds} [delete]
func (a DictController) DeleteDictItem(ctx echo.Context) error {
	itemIds := ctx.Param("itemIds")

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var deletedBy uint64
	if claims != nil {
		deletedBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.dictItemService.WithTrx(trxHandle).DeleteDictItemByIds(itemIds, deletedBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
