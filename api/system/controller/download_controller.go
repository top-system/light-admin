package controller

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/echox"
)

type DownloadController struct {
	downloadService service.DownloadService
	logger          lib.Logger
}

// NewDownloadController creates new download controller
func NewDownloadController(
	logger lib.Logger,
	downloadService service.DownloadService,
) DownloadController {
	return DownloadController{
		logger:          logger,
		downloadService: downloadService,
	}
}

// Query 查询下载任务列表
// @tags Download
// @summary Download Task Query
// @produce application/json
// @param data query system.DownloadTaskQueryParam true "DownloadTaskQueryParam"
// @success 200 {object} echox.Response{data=[]system.DownloadTaskPageVO} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/downloads [get]
func (a DownloadController) Query(ctx echo.Context) error {
	param := new(system.DownloadTaskQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	qr, err := a.downloadService.Query(param)
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

// Get 获取下载任务详情
// @tags Download
// @summary Get Download Task by ID
// @produce application/json
// @param id path int true "Task ID"
// @success 200 {object} echox.Response{data=system.DownloadTaskDetailVO} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/downloads/{id} [get]
func (a DownloadController) Get(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: "Invalid task ID"}.JSON(ctx)
	}

	detail, err := a.downloadService.GetDetail(context.Background(), id)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: detail}.JSON(ctx)
}

// Create 创建下载任务
// @tags Download
// @summary Create Download Task
// @accept application/json
// @produce application/json
// @param data body system.DownloadTaskCreateForm true "DownloadTaskCreateForm"
// @success 200 {object} echox.Response{data=system.DownloadTaskPageVO} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/downloads [post]
func (a DownloadController) Create(ctx echo.Context) error {
	form := new(system.DownloadTaskCreateForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	// 获取当前用户ID
	var ownerID uint64 = 0
	if userID := ctx.Get("userID"); userID != nil {
		ownerID = userID.(uint64)
	}

	task, err := a.downloadService.Create(context.Background(), form, ownerID)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	var progress float64
	if task.Total > 0 {
		progress = float64(task.Downloaded) / float64(task.Total) * 100
	}

	vo := &system.DownloadTaskPageVO{
		ID:            task.ID,
		TaskID:        task.TaskID,
		Hash:          task.Hash,
		Name:          task.Name,
		URL:           task.URL,
		Downloader:    task.Downloader,
		Status:        task.Status,
		Total:         task.Total,
		Downloaded:    task.Downloaded,
		DownloadSpeed: task.DownloadSpeed,
		Uploaded:      task.Uploaded,
		UploadSpeed:   task.UploadSpeed,
		SavePath:      task.SavePath,
		ErrorMessage:  task.ErrorMessage,
		Progress:      progress,
		CreatedAt:     task.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     task.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return echox.Response{Code: http.StatusOK, Data: vo}.JSON(ctx)
}

// Cancel 取消下载任务
// @tags Download
// @summary Cancel Download Task
// @produce application/json
// @param id path int true "Task ID"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/downloads/{id}/cancel [post]
func (a DownloadController) Cancel(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: "Invalid task ID"}.JSON(ctx)
	}

	if err := a.downloadService.Cancel(context.Background(), id); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// SetFiles 设置要下载的文件
// @tags Download
// @summary Set Files to Download
// @accept application/json
// @produce application/json
// @param id path int true "Task ID"
// @param data body system.SetFileDownloadForm true "SetFileDownloadForm"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/downloads/{id}/files [put]
func (a DownloadController) SetFiles(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: "Invalid task ID"}.JSON(ctx)
	}

	form := new(system.SetFileDownloadForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	if err := a.downloadService.SetFilesToDownload(context.Background(), id, form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// Sync 同步任务状态
// @tags Download
// @summary Sync Download Task Status
// @produce application/json
// @param id path int true "Task ID"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/downloads/{id}/sync [post]
func (a DownloadController) Sync(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: "Invalid task ID"}.JSON(ctx)
	}

	if err := a.downloadService.SyncTaskStatus(context.Background(), id); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// Delete 删除下载任务
// @tags Download
// @summary Delete Download Task
// @produce application/json
// @param id path string true "Task IDs (comma separated)"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/downloads/{id} [delete]
func (a DownloadController) Delete(ctx echo.Context) error {
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

	if err := a.downloadService.BatchDelete(context.Background(), ids); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// GetStats 获取任务统计信息
// @tags Download
// @summary Get Download Task Stats
// @produce application/json
// @success 200 {object} echox.Response{data=system.DownloadTaskStatsVO} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/downloads/stats [get]
func (a DownloadController) GetStats(ctx echo.Context) error {
	stats, err := a.downloadService.GetStats()
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: stats}.JSON(ctx)
}

// GetDownloaders 获取可用的下载器列表
// @tags Download
// @summary Get Available Downloaders
// @produce application/json
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/downloads/downloaders [get]
func (a DownloadController) GetDownloaders(ctx echo.Context) error {
	downloaders := a.downloadService.GetAvailableDownloaders()
	return echox.Response{Code: http.StatusOK, Data: downloaders}.JSON(ctx)
}

// TestDownloader 测试下载器连接
// @tags Download
// @summary Test Downloader Connection
// @produce application/json
// @param name path string true "Downloader Name"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/downloads/test/{name} [get]
func (a DownloadController) TestDownloader(ctx echo.Context) error {
	name := ctx.Param("name")
	version, err := a.downloadService.TestDownloader(context.Background(), name)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: map[string]string{"version": version}}.JSON(ctx)
}
