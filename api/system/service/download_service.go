package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"gorm.io/gorm"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/downloader"
	"github.com/top-system/light-admin/pkg/downloader/aria2"
	"github.com/top-system/light-admin/pkg/downloader/qbittorrent"
	"github.com/top-system/light-admin/pkg/queue"
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
	logger               lib.Logger
	config               lib.Config
	db                   lib.Database
	downloadRepository   repository.DownloadRepository
	downloaders          map[string]downloader.Downloader
	downloaderRegistry   *queue.DownloaderRegistry
	taskQueue            lib.TaskQueue
	mu                   sync.RWMutex
}

// NewDownloadService creates a new download service
func NewDownloadService(
	logger lib.Logger,
	config lib.Config,
	db lib.Database,
	downloadRepository repository.DownloadRepository,
	taskQueue lib.TaskQueue,
) DownloadService {
	svc := DownloadService{
		logger:             logger,
		config:             config,
		db:                 db,
		downloadRepository: downloadRepository,
		downloaders:        make(map[string]downloader.Downloader),
		downloaderRegistry: queue.NewDownloaderRegistry(),
		taskQueue:          taskQueue,
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
		a.downloaderRegistry.Register("aria2", aria2Downloader)
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
			a.downloaderRegistry.Register("qbittorrent", qbDownloader)
			a.logger.Zap.Info("qBittorrent downloader initialized")
		}
	}
}

// GetDownloaderRegistry returns the downloader registry
func (a *DownloadService) GetDownloaderRegistry() *queue.DownloaderRegistry {
	return a.downloaderRegistry
}

// getDefaultDownloader 获取默认下载器名称
func (a *DownloadService) getDefaultDownloader() string {
	// 优先使用配置中指定的默认类型
	if a.config.Downloader != nil && a.config.Downloader.Type != "" {
		return a.config.Downloader.Type
	}
	// 否则返回第一个可用的下载器
	a.mu.RLock()
	defer a.mu.RUnlock()
	for name := range a.downloaders {
		return name
	}
	return ""
}

// WithTrx delegates transaction to repository database
func (a DownloadService) WithTrx(trxHandle *gorm.DB) DownloadService {
	a.downloadRepository = a.downloadRepository.WithTrx(trxHandle)
	return a
}

// Query 分页查询下载任务（从队列任务表查询）
func (a DownloadService) Query(param *system.DownloadTaskQueryParam) (*system.DownloadTaskQueryResult, error) {
	// 先同步活跃任务状态
	ctx := context.Background()
	_ = a.SyncAllActiveTasks(ctx)

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

// Create 创建下载任务（通过队列）
func (a DownloadService) Create(ctx context.Context, form *system.DownloadTaskCreateForm, ownerID uint64) (*system.DownloadTask, error) {
	// 检查队列是否启用
	if a.taskQueue.Queue == nil {
		return nil, fmt.Errorf("task queue is not enabled, please enable it in config")
	}

	// 如果没有指定下载器，使用默认下载器
	downloaderName := form.Downloader
	if downloaderName == "" {
		downloaderName = a.getDefaultDownloader()
		if downloaderName == "" {
			return nil, fmt.Errorf("no downloader configured, please configure at least one downloader in config")
		}
	}

	a.mu.RLock()
	dl, ok := a.downloaders[downloaderName]
	a.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("downloader %s not found or not configured", downloaderName)
	}
	form.Downloader = downloaderName // 回填下载器名称

	// 创建队列任务
	owner := &queue.TaskOwner{
		ID: ownerID,
	}

	queueTask, err := queue.NewRemoteDownloadTask(ctx, form.URL, form.Downloader, form.Options, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue task: %w", err)
	}

	// 设置下载器
	if remoteTask, ok := queueTask.(*queue.RemoteDownloadTask); ok {
		remoteTask.SetDownloader(dl)
	}

	// 提交到队列
	if err := a.taskQueue.Queue.QueueTask(ctx, queueTask); err != nil {
		return nil, fmt.Errorf("failed to queue download task: %w", err)
	}

	// 同时保存到下载任务表（用于界面查询）
	task := &system.DownloadTask{
		QueueTaskID: uint64(queueTask.ID()),
		Name:        form.URL, // 初始名称为URL，后续同步时更新
		URL:         form.URL,
		Downloader:  form.Downloader,
		Status:      "queued",
		OwnerID:     ownerID,
	}

	// 保存到数据库
	if err := a.downloadRepository.Create(task); err != nil {
		return nil, err
	}

	a.logger.Zap.Infof("Download task created and queued: %d", task.ID)
	return task, nil
}

// Cancel 取消下载任务
func (a DownloadService) Cancel(ctx context.Context, id uint64) error {
	task, err := a.downloadRepository.Get(id)
	if err != nil {
		return err
	}

	// 尝试从队列注册表获取任务并取消
	if task.QueueTaskID > 0 && a.taskQueue.Registry != nil {
		if qTask, ok := a.taskQueue.Registry.Get(int(task.QueueTaskID)); ok && qTask != nil {
			if remoteTask, ok := qTask.(*queue.RemoteDownloadTask); ok {
				if err := remoteTask.CancelDownload(ctx); err != nil {
					a.logger.Zap.Warnf("Failed to cancel download in downloader: %v", err)
				}
			}
		}
	}

	// 直接从下载器取消（如果有 handle）
	a.mu.RLock()
	dl, ok := a.downloaders[task.Downloader]
	a.mu.RUnlock()

	if ok && task.TaskID != "" {
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

	// 尝试从队列任务操作
	if task.QueueTaskID > 0 && a.taskQueue.Registry != nil {
		if qTask, ok := a.taskQueue.Registry.Get(int(task.QueueTaskID)); ok && qTask != nil {
			if remoteTask, ok := qTask.(*queue.RemoteDownloadTask); ok {
				args := make([]*downloader.SetFileToDownloadArgs, 0, len(form.Files))
				for _, f := range form.Files {
					args = append(args, &downloader.SetFileToDownloadArgs{
						Index:    f.Index,
						Download: f.Download,
					})
				}
				return remoteTask.SetFilesToDownload(ctx, args...)
			}
		}
	}

	// 直接从下载器操作
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
func (a DownloadService) Delete(ctx context.Context, id uint64) error {
	// 先取消下载器中的任务
	_ = a.cancelDownloaderTask(ctx, id)
	return a.downloadRepository.Delete(id)
}

// BatchDelete 批量删除下载任务
func (a DownloadService) BatchDelete(ctx context.Context, ids []uint64) error {
	// 先取消所有任务
	for _, id := range ids {
		_ = a.cancelDownloaderTask(ctx, id)
	}
	return a.downloadRepository.BatchDelete(ids)
}

// cancelDownloaderTask 取消下载器中的任务
func (a DownloadService) cancelDownloaderTask(ctx context.Context, id uint64) error {
	task, err := a.downloadRepository.Get(id)
	if err != nil {
		return err
	}

	// 尝试从队列任务获取 handle
	var handle *downloader.TaskHandle
	if task.QueueTaskID > 0 {
		state := a.getRemoteDownloadState(int(task.QueueTaskID))
		if state != nil && state.Handle != nil {
			handle = state.Handle
		}
	}

	// 如果没有从队列获取到，使用数据库中的
	if handle == nil && (task.TaskID != "" || task.Hash != "") {
		handle = &downloader.TaskHandle{
			ID:   task.TaskID,
			Hash: task.Hash,
		}
	}

	if handle == nil {
		return nil
	}

	// 调用下载器取消
	a.mu.RLock()
	dl, ok := a.downloaders[task.Downloader]
	a.mu.RUnlock()

	if ok {
		if err := dl.Cancel(ctx, handle); err != nil {
			a.logger.Zap.Warnf("Failed to cancel task in downloader: %v", err)
		}
	}

	return nil
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

	// 尝试从队列任务获取状态（先从内存 Registry，再从数据库）
	if task.QueueTaskID > 0 {
		state := a.getRemoteDownloadState(int(task.QueueTaskID))
		if state != nil && state.Status != nil {
			var taskID, hash string
			if state.Handle != nil {
				taskID = state.Handle.ID
				hash = state.Handle.Hash
			}

			return a.downloadRepository.UpdateFromDownloader(
				id,
				taskID,
				hash,
				state.Status.Name,
				state.Status.SavePath,
				string(state.Status.State),
				state.Status.Downloaded,
				state.Status.Total,
				state.Status.DownloadSpeed,
				state.Status.Uploaded,
				state.Status.UploadSpeed,
				state.Status.ErrorMessage,
			)
		}
	}

	// 如果有 taskID，直接从下载器获取状态
	if task.TaskID != "" || task.Hash != "" {
		a.mu.RLock()
		dl, ok := a.downloaders[task.Downloader]
		a.mu.RUnlock()

		if ok {
			handle := &downloader.TaskHandle{
				ID:   task.TaskID,
				Hash: task.Hash,
			}

			status, err := dl.Info(ctx, handle)
			if err == nil {
				return a.downloadRepository.UpdateFromDownloader(
					id,
					handle.ID,
					handle.Hash,
					status.Name,
					status.SavePath,
					string(status.State),
					status.Downloaded,
					status.Total,
					status.DownloadSpeed,
					status.Uploaded,
					status.UploadSpeed,
					status.ErrorMessage,
				)
			}
		}
	}

	return nil
}

// getRemoteDownloadState 获取远程下载任务状态（从 Registry 或数据库）
func (a DownloadService) getRemoteDownloadState(queueTaskID int) *queue.RemoteDownloadTaskState {
	// 先尝试从 Registry 获取（任务还在运行中）
	if a.taskQueue.Registry != nil {
		if qTask, ok := a.taskQueue.Registry.Get(queueTaskID); ok && qTask != nil {
			if remoteTask, ok := qTask.(*queue.RemoteDownloadTask); ok {
				return remoteTask.GetState()
			}
		}
	}

	// 从数据库获取队列任务的 PrivateState
	var taskModel queue.TaskModel
	if err := a.db.ORM.First(&taskModel, queueTaskID).Error; err == nil && taskModel.PrivateState != "" {
		state := &queue.RemoteDownloadTaskState{}
		if err := json.Unmarshal([]byte(taskModel.PrivateState), state); err == nil {
			return state
		}
	}

	return nil
}

// SyncAllActiveTasks 同步所有活跃任务的状态
func (a DownloadService) SyncAllActiveTasks(ctx context.Context) error {
	tasks, err := a.downloadRepository.GetActiveTaskIDs()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if err := a.SyncTaskStatus(ctx, task.ID); err != nil {
			a.logger.Zap.Warnf("Failed to sync task %d: %v", task.ID, err)
		}
	}

	return nil
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

// GetQueueStats 获取队列统计信息
func (a DownloadService) GetQueueStats() map[string]int {
	if a.taskQueue.Queue == nil {
		return nil
	}
	return map[string]int{
		"busyWorkers":     a.taskQueue.Queue.BusyWorkers(),
		"successTasks":    a.taskQueue.Queue.SuccessTasks(),
		"failureTasks":    a.taskQueue.Queue.FailureTasks(),
		"submittedTasks":  a.taskQueue.Queue.SubmittedTasks(),
		"suspendingTasks": a.taskQueue.Queue.SuspendingTasks(),
	}
}
