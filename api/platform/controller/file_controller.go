package controller

import (
	"net/http"

	"github.com/top-system/light-admin/api/platform/service"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/pkg/echox"
	"github.com/labstack/echo/v4"
)

// FileController 文件控制器
type FileController struct {
	fileService service.FileService
	logger      lib.Logger
}

// NewFileController 创建文件控制器
func NewFileController(
	fileService service.FileService,
	logger lib.Logger,
) FileController {
	return FileController{
		fileService: fileService,
		logger:      logger,
	}
}

// Upload 文件上传
// @tags File
// @summary Upload File
// @accept multipart/form-data
// @produce application/json
// @param file formData file true "File to upload"
// @success 200 {object} echox.Response{data=platform.FileInfo} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/files [post]
func (c FileController) Upload(ctx echo.Context) error {
	// 获取上传的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: "file is required"}.JSON(ctx)
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return echox.Response{Code: http.StatusInternalServerError, Message: err}.JSON(ctx)
	}
	defer src.Close()

	// 上传文件
	fileInfo, err := c.fileService.UploadFile(file.Filename, src, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		c.logger.Zap.Errorf("Failed to upload file: %v", err)
		return echox.Response{Code: http.StatusInternalServerError, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: fileInfo}.JSON(ctx)
}

// Delete 文件删除
// @tags File
// @summary Delete File
// @produce application/json
// @param filePath query string true "File path"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/files [delete]
func (c FileController) Delete(ctx echo.Context) error {
	filePath := ctx.QueryParam("filePath")
	if filePath == "" {
		return echox.Response{Code: http.StatusBadRequest, Message: "filePath is required"}.JSON(ctx)
	}

	if err := c.fileService.DeleteFile(filePath); err != nil {
		c.logger.Zap.Errorf("Failed to delete file: %v", err)
		return echox.Response{Code: http.StatusInternalServerError, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
