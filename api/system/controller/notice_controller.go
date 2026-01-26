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

type NoticeController struct {
	noticeService service.NoticeService
	logger        lib.Logger
}

// NewNoticeController creates new notice controller
func NewNoticeController(
	noticeService service.NoticeService,
	logger lib.Logger,
) NoticeController {
	return NoticeController{
		noticeService: noticeService,
		logger:        logger,
	}
}

// Query 通知公告分页列表
// @Tags Notice
// @Summary 通知公告分页列表
// @Produce application/json
// @Param title query string false "标题"
// @Param type query int false "类型"
// @Param publishStatus query int false "发布状态"
// @Param current query int false "当前页"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} echox.Response{data=[]system.Notice} "ok"
// @Router /api/v1/notices [get]
func (a NoticeController) Query(ctx echo.Context) error {
	param := new(system.NoticeQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	qr, err := a.noticeService.Query(param)
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

// GetForm 获取通知公告表单数据
// @Tags Notice
// @Summary 获取通知公告表单数据
// @Produce application/json
// @Param id path int true "通知公告ID"
// @Success 200 {object} echox.Response{data=system.NoticeForm} "ok"
// @Router /api/v1/notices/{id}/form [get]
func (a NoticeController) GetForm(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form, err := a.noticeService.GetForm(id)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: form}.JSON(ctx)
}

// GetDetail 获取通知公告详情（阅读）
// @Tags Notice
// @Summary 获取通知公告详情
// @Produce application/json
// @Param id path int true "通知公告ID"
// @Success 200 {object} echox.Response{data=system.NoticeDetailVO} "ok"
// @Router /api/v1/notices/{id}/detail [get]
func (a NoticeController) GetDetail(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var userID uint64
	if claims != nil {
		userID = claims.ID
	}

	detail, err := a.noticeService.GetDetail(id, userID)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: detail}.JSON(ctx)
}

// Create 新增通知公告
// @Tags Notice
// @Summary 新增通知公告
// @Produce application/json
// @Param data body system.NoticeForm true "通知公告信息"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/notices [post]
func (a NoticeController) Create(ctx echo.Context) error {
	form := new(system.NoticeForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var createdBy uint64
	if claims != nil {
		createdBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.noticeService.WithTrx(trxHandle).Create(form, createdBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// Update 修改通知公告
// @Tags Notice
// @Summary 修改通知公告
// @Produce application/json
// @Param id path int true "通知公告ID"
// @Param data body system.NoticeForm true "通知公告信息"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/notices/{id} [put]
func (a NoticeController) Update(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form := new(system.NoticeForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var updatedBy uint64
	if claims != nil {
		updatedBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.noticeService.WithTrx(trxHandle).Update(id, form, updatedBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// Delete 删除通知公告
// @Tags Notice
// @Summary 删除通知公告
// @Produce application/json
// @Param ids path string true "通知公告ID，多个以英文逗号分割"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/notices/{ids} [delete]
func (a NoticeController) Delete(ctx echo.Context) error {
	ids := ctx.Param("ids")

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var deletedBy uint64
	if claims != nil {
		deletedBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.noticeService.WithTrx(trxHandle).Delete(ids, deletedBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// Publish 发布通知公告
// @Tags Notice
// @Summary 发布通知公告
// @Produce application/json
// @Param id path int true "通知公告ID"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/notices/{id}/publish [put]
func (a NoticeController) Publish(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var publisherId uint64
	if claims != nil {
		publisherId = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.noticeService.WithTrx(trxHandle).Publish(id, publisherId); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// Revoke 撤回通知公告
// @Tags Notice
// @Summary 撤回通知公告
// @Produce application/json
// @Param id path int true "通知公告ID"
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/notices/{id}/revoke [put]
func (a NoticeController) Revoke(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var updatedBy uint64
	if claims != nil {
		updatedBy = claims.ID
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	if err := a.noticeService.WithTrx(trxHandle).Revoke(id, updatedBy); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// ReadAll 全部已读
// @Tags Notice
// @Summary 全部已读
// @Produce application/json
// @Success 200 {object} echox.Response "ok"
// @Router /api/v1/notices/read-all [put]
func (a NoticeController) ReadAll(ctx echo.Context) error {
	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	var userID uint64
	if claims != nil {
		userID = claims.ID
	}

	if err := a.noticeService.ReadAll(userID); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// GetMyNoticePage 获取我的通知公告分页列表
// @Tags Notice
// @Summary 获取我的通知公告分页列表
// @Produce application/json
// @Param title query string false "标题"
// @Param type query int false "类型"
// @Param current query int false "当前页"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} echox.Response{data=[]system.UserNoticePageVO} "ok"
// @Router /api/v1/notices/my-page [get]
func (a NoticeController) GetMyNoticePage(ctx echo.Context) error {
	param := new(system.NoticeQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	if claims != nil {
		param.UserID = claims.ID
	}

	// 设置默认分页
	if param.PageNum == 0 {
		param.PageNum = 1
	}
	if param.PageSize == 0 {
		param.PageSize = 10
	}

	list, total, err := a.noticeService.GetMyNoticePage(param)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{
		Code: http.StatusOK,
		Data: list,
		Page: &echox.PageInfo{
			Total:    total,
			PageNum:  param.PageNum,
			PageSize: param.PageSize,
		},
	}.JSON(ctx)
}
