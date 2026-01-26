package controller

import (
	"net/http"

	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/echox"
	"github.com/labstack/echo/v4"
)

type LogController struct {
	logService service.LogService
	logger     lib.Logger
}

// NewLogController creates new log controller
func NewLogController(
	logger lib.Logger,
	logService service.LogService,
) LogController {
	return LogController{
		logger:     logger,
		logService: logService,
	}
}

// @tags Log
// @summary Log Query
// @produce application/json
// @param data query system.LogQueryParam true "LogQueryParam"
// @success 200 {object} echox.Response{data=[]system.LogPageVO} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/logs [get]
func (a LogController) Query(ctx echo.Context) error {
	param := new(system.LogQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	qr, err := a.logService.Query(param)
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
