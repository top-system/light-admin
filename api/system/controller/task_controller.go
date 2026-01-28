package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/echox"
)

type TaskController struct {
	taskService service.TaskService
	logger      lib.Logger
}

// NewTaskController creates new task controller
func NewTaskController(
	logger lib.Logger,
	taskService service.TaskService,
) TaskController {
	return TaskController{
		logger:      logger,
		taskService: taskService,
	}
}

// Query 查询任务列表
// @tags Task
// @summary Task Query
// @produce application/json
// @param data query system.TaskQueryParam true "TaskQueryParam"
// @success 200 {object} echox.Response{data=[]system.TaskPageVO} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/tasks [get]
func (a TaskController) Query(ctx echo.Context) error {
	param := new(system.TaskQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	qr, err := a.taskService.Query(param)
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

// Get 获取任务详情
// @tags Task
// @summary Get Task by ID
// @produce application/json
// @param id path int true "Task ID"
// @success 200 {object} echox.Response{data=system.TaskPageVO} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/tasks/{id} [get]
func (a TaskController) Get(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: "Invalid task ID"}.JSON(ctx)
	}

	task, err := a.taskService.Get(id)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	vo := &system.TaskPageVO{
		ID:               task.ID,
		Type:             task.Type,
		Status:           string(task.Status),
		CorrelationID:    task.CorrelationID,
		OwnerID:          task.OwnerID,
		RetryCount:       task.RetryCount,
		ExecutedDuration: task.ExecutedDuration,
		Error:            task.Error,
		ErrorHistory:     task.ErrorHistory,
		ResumeTime:       task.ResumeTime,
		CreatedAt:        task.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        task.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return echox.Response{Code: http.StatusOK, Data: vo}.JSON(ctx)
}

// Delete 删除任务
// @tags Task
// @summary Delete Task
// @produce application/json
// @param id path string true "Task IDs (comma separated)"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/tasks/{id} [delete]
func (a TaskController) Delete(ctx echo.Context) error {
	idStr := ctx.Param("id")
	idStrs := strings.Split(idStr, ",")

	ids := make([]uint64, 0, len(idStrs))
	for _, s := range idStrs {
		id, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return echox.Response{Code: http.StatusBadRequest, Message: "Invalid task IDs"}.JSON(ctx)
	}

	if err := a.taskService.BatchDelete(ids); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// GetStats 获取任务统计信息
// @tags Task
// @summary Get Task Stats
// @produce application/json
// @success 200 {object} echox.Response{data=system.TaskStatsVO} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/tasks/stats [get]
func (a TaskController) GetStats(ctx echo.Context) error {
	stats, err := a.taskService.GetStats()
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: stats}.JSON(ctx)
}

// GetTypes 获取所有任务类型
// @tags Task
// @summary Get Task Types
// @produce application/json
// @success 200 {object} echox.Response{data=[]system.TaskTypeVO} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/tasks/types [get]
func (a TaskController) GetTypes(ctx echo.Context) error {
	types, err := a.taskService.GetTaskTypes()
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: types}.JSON(ctx)
}
