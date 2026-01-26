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

type DeptController struct {
	deptService service.DeptService
	logger      lib.Logger
}

// NewDeptController creates new dept controller
func NewDeptController(
	deptService service.DeptService,
	logger lib.Logger,
) DeptController {
	return DeptController{
		deptService: deptService,
		logger:      logger,
	}
}

// Query 部门列表
// @Tags Dept
// @Summary 部门列表
// @Produce application/json
// @Param keywords query string false "关键字"
// @Param status query int false "状态"
// @Success 200 {object} echox.Response{data=[]system.DeptVO} "ok"
// @Router /api/v1/depts [get]
func (a DeptController) Query(ctx echo.Context) error {
	param := new(system.DeptQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	list, err := a.deptService.GetDeptList(param)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: list}.JSON(ctx)
}

// GetOptions 部门下拉列表
// @Tags Dept
// @Summary 部门下拉列表
// @Produce application/json
// @Success 200 {object} echox.Response{data=[]system.DeptOption} "ok"
// @Router /api/v1/depts/options [get]
func (a DeptController) GetOptions(ctx echo.Context) error {
	list, err := a.deptService.ListDeptOptions()
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: list}.JSON(ctx)
}

// Create 新增部门
// @Tags Dept
// @Summary 新增部门
// @Produce application/json
// @Param data body system.DeptForm true "部门信息"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/depts [post]
func (a DeptController) Create(ctx echo.Context) error {
	form := new(system.DeptForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var createdBy uint64
	if claims != nil {
		createdBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	id, err := a.deptService.WithTrx(trxHandle).SaveDept(form, createdBy)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: id}.JSON(ctx)
}

// GetForm 获取部门表单数据
// @Tags Dept
// @Summary 获取部门表单数据
// @Produce application/json
// @Param deptId path int true "部门ID"
// @Success 200 {object} echox.Response{data=system.DeptForm} "ok"
// @Router /api/v1/depts/{deptId}/form [get]
func (a DeptController) GetForm(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("deptId"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form, err := a.deptService.GetDeptForm(id)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: form}.JSON(ctx)
}

// Update 修改部门
// @Tags Dept
// @Summary 修改部门
// @Produce application/json
// @Param deptId path int true "部门ID"
// @Param data body system.DeptForm true "部门信息"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/depts/{deptId} [put]
func (a DeptController) Update(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("deptId"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form := new(system.DeptForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var updatedBy uint64
	if claims != nil {
		updatedBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	_, err = a.deptService.WithTrx(trxHandle).UpdateDept(id, form, updatedBy)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// Delete 删除部门
// @Tags Dept
// @Summary 删除部门
// @Produce application/json
// @Param ids path string true "部门ID，多个以英文逗号分割"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/depts/{ids} [delete]
func (a DeptController) Delete(ctx echo.Context) error {
	ids := ctx.Param("ids")

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var deletedBy uint64
	if claims != nil {
		deletedBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.deptService.WithTrx(trxHandle).DeleteByIds(ids, deletedBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
