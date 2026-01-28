package service

import (
	"context"
	"fmt"
	"sync"

	"gorm.io/gorm"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/downloader"
	"github.com/top-system/light-admin/pkg/downloader/aria2"
	"github.com/top-system/light-admin/pkg/downloader/qbittorrent"
)

// downloaderLogger 是一个适配器，将 lib.Logger 转换为 downloader 需要的 Logger 接口
type downloaderLogger struct {
	logger lib.Logger
}

func (l *downloaderLogger) Info(format string, args ...interface{}) {
	l.logger.Zap.Infof(format, args...)
}

func (l *downloaderLogger) Debug(format string, args ...interface{}) {
	l.logger.Zap.Debugf(format, args...)
}

func (l *downloaderLogger) Warning(format string, args ...interface{}) {
	l.logger.Zap.Warnf(format, args...)
}

func (l *downloaderLogger) Error(format string, args ...interface{}) {
	l.logger.Zap.Errorf(format, args...)
}

// DownloadService service layer
type DownloadService struct {
	logger             lib.Logger
	config             lib.Config
	downloadRepository repository.DownloadRepository
	downloaders        map[string]downloader.Downloader
	mu                 sync.RWMutex
}

// NewDownloadService creates a new download service
func NewDownloadService(
	logger lib.Logger,
	config lib.Config,
	downloadRepository repository.DownloadRepository,
) DownloadService {
	svc := DownloadService{
		logger:             logger,
		config:             config,
		downloadRepository: downloadRepository,
		downloaders:        make(map[string]downloader.Downloader),
	}

	// 初始化下载器
	svc.initDownloaders()

	return svc
}

// initDownloaders 初始化下载器实例
func (a *DownloadService) initDownloaders() {
	// 检查下载器配置是否存在
	if a.config.Downloader == nil {
		return
	}

	dlLogger := &downloaderLogger{logger: a.logger}

	// 从配置文件读取下载器配置
	if a.config.Downloader.Aria2 != nil && a.config.Downloader.Aria2.Server != "" {
		aria2Downloader := aria2.New(dlLogger, &aria2.Settings{
			Server:   a.config.Downloader.Aria2.Server,
			Token:    a.config.Downloader.Aria2.Token,
			TempPath: a.config.Downloader.Aria2.TempPath,
			Options:  a.config.Downloader.Aria2.Options,
		})
		a.downloaders["aria2"] = aria2Downloader
		a.logger.Zap.Info("Aria2 downloader initialized")
	}

	if a.config.Downloader.QBittorrent != nil && a.config.Downloader.QBittorrent.Server != "" {
		qbDownloader, err := qbittorrent.New(dlLogger, &qbittorrent.Settings{
			Server:   a.config.Downloader.QBittorrent.Server,
			User:     a.config.Downloader.QBittorrent.User,
			Password: a.config.Downloader.QBittorrent.Password,
			TempPath: a.config.Downloader.QBittorrent.TempPath,
			Options:  a.config.Downloader.QBittorrent.Options,
		})
		if err != nil {
			a.logger.Zap.Errorf("Failed to initialize qBittorrent downloader: %v", err)
		} else {
			a.downloaders["qbittorrent"] = qbDownloader
			a.logger.Zap.Info("qBittorrent downloader initialized")
		}
	}
}

// WithTrx delegates transaction to repository database
func (a DownloadService) WithTrx(trxHandle *gorm.DB) DownloadService {
	a.downloadRepository = a.downloadRepository.WithTrx(trxHandle)
	return a
}

// Query 分页查询下载任务
func (a DownloadService) Query(param *system.DownloadTaskQueryParam) (*system.DownloadTaskQueryResult, error) {
	return a.downloadRepository.Query(param)
}

// Get 获取下载任务详情
func (a DownloadService) Get(id uint64) (*system.DownloadTask, error) {
	return a.downloadRepository.Get(id)
}

// GetDetail 获取下载任务详情（包含文件列表）
func (a DownloadService) GetDetail(ctx context.Context, id uint64) (*system.DownloadTaskDetailVO, error) {
	task, err := a.downloadRepository.Get(id)
	if err != nil {
		return nil, err
	}

	var progress float64
	if task.Total > 0 {
		progress = float64(task.Downloaded) / float64(task.Total) * 100
	}

	detail := &system.DownloadTaskDetailVO{
		DownloadTaskPageVO: system.DownloadTaskPageVO{
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
		},
		Files: make([]system.DownloadTaskFileVO, 0),
	}

	// 从下载器获取实时文件列表
	dl, ok := a.downloaders[task.Downloader]
	if ok {
		handle := &downloader.TaskHandle{
			ID:   task.TaskID,
			Hash: task.Hash,
		}
		status, err := dl.Info(ctx, handle)
		if err == nil && status != nil {
			for _, f := range status.Files {
				detail.Files = append(detail.Files, system.DownloadTaskFileVO{
					Index:    f.Index,
					Name:     f.Name,
					Size:     f.Size,
					Progress: f.Progress * 100,
					Selected: f.Selected,
				})
			}
		}
	}

	return detail, nil
}

// Create 创建下载任务
func (a DownloadService) Create(ctx context.Context, form *system.DownloadTaskCreateForm, ownerID uint64) (*system.DownloadTask, error) {
	a.mu.RLock()
	dl, ok := a.downloaders[form.Downloader]
	a.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("downloader %s not found or not configured", form.Downloader)
	}

	// 调用下载器创建任务
	handle, err := dl.CreateTask(ctx, form.URL, form.Options)
	if err != nil {
		return nil, fmt.Errorf("failed to create download task: %w", err)
	}

	// 获取任务初始状态
	status, _ := dl.Info(ctx, handle)

	task := &system.DownloadTask{
		TaskID:     handle.ID,
		Hash:       handle.Hash,
		Name:       form.URL, // 初始名称为URL，后续同步时更新
		URL:        form.URL,
		Downloader: form.Downloader,
		Status:     "downloading",
		OwnerID:    ownerID,
	}

	if status != nil {
		task.Name = status.Name
		task.Status = string(status.State)
		task.Total = status.Total
		task.Downloaded = status.Downloaded
		task.SavePath = status.SavePath
	}

	// 保存到数据库
	if err := a.downloadRepository.Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

// Cancel 取消下载任务
func (a DownloadService) Cancel(ctx context.Context, id uint64) error {
	task, err := a.downloadRepository.Get(id)
	if err != nil {
		return err
	}

	a.mu.RLock()
	dl, ok := a.downloaders[task.Downloader]
	a.mu.RUnlock()

	if ok {
		handle := &downloader.TaskHandle{
			ID:   task.TaskID,
			Hash: task.Hash,
		}
		if err := dl.Cancel(ctx, handle); err != nil {
			a.logger.Zap.Warnf("Failed to cancel task in downloader: %v", err)
		}
	}

	// 更新数据库状态
	return a.downloadRepository.UpdateStatus(id, "canceled", task.Downloaded, task.Total, 0, task.Uploaded, 0, "")
}

// SetFilesToDownload 设置要下载的文件
func (a DownloadService) SetFilesToDownload(ctx context.Context, id uint64, form *system.SetFileDownloadForm) error {
	task, err := a.downloadRepository.Get(id)
	if err != nil {
		return err
	}

	a.mu.RLock()
	dl, ok := a.downloaders[task.Downloader]
	a.mu.RUnlock()

	if !ok {
		return fmt.Errorf("downloader %s not found", task.Downloader)
	}

	handle := &downloader.TaskHandle{
		ID:   task.TaskID,
		Hash: task.Hash,
	}

	args := make([]*downloader.SetFileToDownloadArgs, 0, len(form.Files))
	for _, f := range form.Files {
		args = append(args, &downloader.SetFileToDownloadArgs{
			Index:    f.Index,
			Download: f.Download,
		})
	}

	return dl.SetFilesToDownload(ctx, handle, args...)
}

// Delete 删除下载任务
func (a DownloadService) Delete(id uint64) error {
	return a.downloadRepository.Delete(id)
}

// BatchDelete 批量删除下载任务
func (a DownloadService) BatchDelete(ids []uint64) error {
	return a.downloadRepository.BatchDelete(ids)
}

// GetStats 获取任务统计信息
func (a DownloadService) GetStats() (*system.DownloadTaskStatsVO, error) {
	return a.downloadRepository.GetStatusCounts()
}

// SyncTaskStatus 同步任务状态（从下载器同步到数据库）
func (a DownloadService) SyncTaskStatus(ctx context.Context, id uint64) error {
	task, err := a.downloadRepository.Get(id)
	if err != nil {
		return err
	}

	a.mu.RLock()
	dl, ok := a.downloaders[task.Downloader]
	a.mu.RUnlock()

	if !ok {
		return fmt.Errorf("downloader %s not found", task.Downloader)
	}

	handle := &downloader.TaskHandle{
		ID:   task.TaskID,
		Hash: task.Hash,
	}

	status, err := dl.Info(ctx, handle)
	if err != nil {
		return fmt.Errorf("failed to get task status: %w", err)
	}

	// 更新数据库
	return a.downloadRepository.UpdateStatus(
		id,
		string(status.State),
		status.Downloaded,
		status.Total,
		status.DownloadSpeed,
		status.Uploaded,
		status.UploadSpeed,
		status.ErrorMessage,
	)
}

// GetAvailableDownloaders 获取可用的下载器列表
func (a DownloadService) GetAvailableDownloaders() []map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]map[string]string, 0)
	for name := range a.downloaders {
		label := name
		if name == "aria2" {
			label = "Aria2"
		} else if name == "qbittorrent" {
			label = "qBittorrent"
		}
		result = append(result, map[string]string{
			"label": label,
			"value": name,
		})
	}
	return result
}

// TestDownloader 测试下载器连接
func (a DownloadService) TestDownloader(ctx context.Context, name string) (string, error) {
	a.mu.RLock()
	dl, ok := a.downloaders[name]
	a.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("downloader %s not found", name)
	}

	return dl.Test(ctx)
}
