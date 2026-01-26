package lib

import (
	"context"

	"go.uber.org/fx"

	"github.com/top-system/light-admin/pkg/crontab"
	"github.com/top-system/light-admin/pkg/downloader"
	"github.com/top-system/light-admin/pkg/downloader/aria2"
	"github.com/top-system/light-admin/pkg/downloader/qbittorrent"
	"github.com/top-system/light-admin/pkg/queue"
)

// ============================================================================
// 扩展模块 - 根据需要启用或禁用
//
// 使用方法：
//   1. 在 lib/lib.go 的 Module 中添加: ExtrasModule
//      例如：var Module = fx.Options(
//              fx.Provide(NewHttpHandler),
//              ...
//              ExtrasModule,  // 添加这一行
//            )
//
//   2. 在 config.yaml 中添加对应配置（参考 config/extras.yaml.example）
//
//   3. 如果使用任务队列，需要运行数据库迁移：
//      ./light-admin migrate
//
//   4. 不需要的功能可以注释掉对应的 fx.Provide 行
//
// 在服务中使用：
//   - 将 lib.TaskQueue / lib.Crontab / lib.Downloader 注入到你的服务中
//   - 例如：func NewMyService(queue lib.TaskQueue, cron lib.Crontab) *MyService
// ============================================================================

// ExtrasModule 扩展功能模块
// 在 lib/lib.go 的 Module 中添加此模块即可启用
var ExtrasModule = fx.Options(
	// ====== 任务队列 ======
	// 用于异步任务处理，如发送邮件、处理文件等
	// 如不需要，注释下面这行
	fx.Provide(NewTaskQueue),

	// ====== 定时任务 ======
	// 用于定时执行任务，如数据清理、报表生成等
	// 如不需要，注释下面这行
	fx.Provide(NewCrontab),

	// ====== 下载器 ======
	// 用于管理 aria2/qBittorrent 下载任务
	// 如不需要，注释下面这行
	fx.Provide(NewDownloader),
)

// ============================================================================
// 任务队列 (Queue)
// ============================================================================

// TaskQueue 任务队列封装
type TaskQueue struct {
	Queue    queue.Queue
	Registry queue.TaskRegistry
}

// queueLogger 适配器 - 实现 queue.Logger 接口
type queueLogger struct {
	logger Logger
	prefix string
}

func (l *queueLogger) Info(format string, args ...interface{})  { l.logger.Zap.Infof(l.prefix+format, args...) }
func (l *queueLogger) Debug(format string, args ...interface{}) { l.logger.Zap.Debugf(l.prefix+format, args...) }
func (l *queueLogger) Warning(format string, args ...interface{}) {
	l.logger.Zap.Warnf(l.prefix+format, args...)
}
func (l *queueLogger) Error(format string, args ...interface{}) {
	l.logger.Zap.Errorf(l.prefix+format, args...)
}
func (l *queueLogger) CopyWithPrefix(prefix string) queue.Logger {
	return &queueLogger{logger: l.logger, prefix: prefix + " "}
}

// NewTaskQueue 创建任务队列
func NewTaskQueue(lc fx.Lifecycle, config Config, logger Logger, db Database) TaskQueue {
	cfg := config.Queue
	if cfg == nil || !cfg.Enable {
		logger.Zap.Info("Queue is disabled")
		return TaskQueue{}
	}

	// 创建任务注册表
	registry := queue.NewTaskRegistry()

	// 创建任务仓库（用于持久化）
	taskRepo := queue.NewGormTaskRepository(db.ORM)

	// 配置选项
	opts := []queue.Option{
		queue.WithWorkerCount(cfg.WorkerNum),
	}

	if cfg.MaxRetry > 0 {
		opts = append(opts, queue.WithMaxRetry(cfg.MaxRetry))
	}

	if cfg.Name != "" {
		opts = append(opts, queue.WithName(cfg.Name))
	}

	// 创建队列
	q := queue.New(&queueLogger{logger: logger}, taskRepo, registry, opts...)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Zap.Info("Starting Task Queue")
			q.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Zap.Info("Stopping Task Queue")
			q.Shutdown()
			return nil
		},
	})

	logger.Zap.Infof("Task Queue initialized with %d workers", cfg.WorkerNum)
	return TaskQueue{Queue: q, Registry: registry}
}

// IsEnabled 检查任务队列是否启用
func (q *TaskQueue) IsEnabled() bool {
	return q.Queue != nil
}

// QueueTask 提交任务到队列
// 示例:
//
//	task := queue.NewSimpleTask("send_email", payload, owner)
//	err := taskQueue.QueueTask(ctx, task)
func (q *TaskQueue) QueueTask(ctx context.Context, t queue.Task) error {
	if q.Queue != nil {
		return q.Queue.QueueTask(ctx, t)
	}
	return nil
}

// Stats 获取队列统计信息
func (q *TaskQueue) Stats() map[string]int {
	if q.Queue == nil {
		return nil
	}
	return map[string]int{
		"busy_workers":     q.Queue.BusyWorkers(),
		"success_tasks":    q.Queue.SuccessTasks(),
		"failure_tasks":    q.Queue.FailureTasks(),
		"submitted_tasks":  q.Queue.SubmittedTasks(),
		"suspending_tasks": q.Queue.SuspendingTasks(),
	}
}

// ============================================================================
// 定时任务 (Crontab)
// ============================================================================

// Crontab 定时任务封装
type Crontab struct {
	Cron *crontab.Crontab
}

// crontabLogger 适配器 - 实现 crontab.Logger 接口
type crontabLogger struct {
	logger Logger
	prefix string
}

func (l *crontabLogger) Info(format string, args ...interface{}) {
	l.logger.Zap.Infof(l.prefix+format, args...)
}
func (l *crontabLogger) Debug(format string, args ...interface{}) {
	l.logger.Zap.Debugf(l.prefix+format, args...)
}
func (l *crontabLogger) Warning(format string, args ...interface{}) {
	l.logger.Zap.Warnf(l.prefix+format, args...)
}
func (l *crontabLogger) Error(format string, args ...interface{}) {
	l.logger.Zap.Errorf(l.prefix+format, args...)
}
func (l *crontabLogger) CopyWithPrefix(prefix string) crontab.Logger {
	return &crontabLogger{logger: l.logger, prefix: prefix + " "}
}

// NewCrontab 创建定时任务管理器
func NewCrontab(lc fx.Lifecycle, config Config, logger Logger) Crontab {
	cfg := config.Crontab
	if cfg == nil || !cfg.Enable {
		logger.Zap.Info("Crontab is disabled")
		return Crontab{}
	}

	c := crontab.New(&crontabLogger{logger: logger})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Zap.Info("Starting Crontab")
			return c.Start()
		},
		OnStop: func(ctx context.Context) error {
			logger.Zap.Info("Stopping Crontab")
			c.Stop()
			return nil
		},
	})

	logger.Zap.Info("Crontab initialized")
	return Crontab{Cron: c}
}

// AddTask 添加定时任务
// 示例：
//
//	crontab.AddTask("cleanup", "0 0 * * * *", func(ctx context.Context) {
//	    // 每天凌晨执行清理
//	})
func (c *Crontab) AddTask(name, spec string, fn crontab.CronTaskFunc) error {
	if c.Cron != nil {
		return c.Cron.AddTask(name, spec, fn)
	}
	return nil
}

// RemoveTask 移除定时任务
func (c *Crontab) RemoveTask(name string) error {
	if c.Cron != nil {
		return c.Cron.RemoveTask(name)
	}
	return nil
}

// IsEnabled 检查定时任务是否启用
func (c *Crontab) IsEnabled() bool {
	return c.Cron != nil
}

// ============================================================================
// 下载器 (Downloader)
// ============================================================================

// Downloader 下载器封装
type Downloader struct {
	Client downloader.Downloader
	Type   string // "aria2" 或 "qbittorrent"
}

// downloaderLogger 适配器
type downloaderLogger struct {
	logger Logger
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

// NewDownloader 创建下载器
func NewDownloader(config Config, logger Logger) Downloader {
	cfg := config.Downloader
	if cfg == nil || !cfg.Enable {
		logger.Zap.Info("Downloader is disabled")
		return Downloader{}
	}

	dl := &downloaderLogger{logger}

	switch cfg.Type {
	case "aria2":
		if cfg.Aria2 == nil {
			logger.Zap.Error("Aria2 config is missing")
			return Downloader{}
		}
		client := aria2.New(dl, &aria2.Settings{
			Server:   cfg.Aria2.Server,
			Token:    cfg.Aria2.Token,
			TempPath: cfg.Aria2.TempPath,
			Options:  cfg.Aria2.Options,
		})
		logger.Zap.Infof("Aria2 downloader initialized: %s", cfg.Aria2.Server)
		return Downloader{Client: client, Type: "aria2"}

	case "qbittorrent":
		if cfg.QBittorrent == nil {
			logger.Zap.Error("QBittorrent config is missing")
			return Downloader{}
		}
		client, err := qbittorrent.New(dl, &qbittorrent.Settings{
			Server:   cfg.QBittorrent.Server,
			User:     cfg.QBittorrent.User,
			Password: cfg.QBittorrent.Password,
			TempPath: cfg.QBittorrent.TempPath,
			Options:  cfg.QBittorrent.Options,
		})
		if err != nil {
			logger.Zap.Errorf("Failed to create qBittorrent client: %v", err)
			return Downloader{}
		}
		logger.Zap.Infof("QBittorrent downloader initialized: %s", cfg.QBittorrent.Server)
		return Downloader{Client: client, Type: "qbittorrent"}

	default:
		logger.Zap.Errorf("Unknown downloader type: %s", cfg.Type)
		return Downloader{}
	}
}

// CreateTask 创建下载任务
func (d *Downloader) CreateTask(ctx context.Context, url string, options map[string]interface{}) (*downloader.TaskHandle, error) {
	if d.Client != nil {
		return d.Client.CreateTask(ctx, url, options)
	}
	return nil, nil
}

// Info 获取任务状态
func (d *Downloader) Info(ctx context.Context, handle *downloader.TaskHandle) (*downloader.TaskStatus, error) {
	if d.Client != nil {
		return d.Client.Info(ctx, handle)
	}
	return nil, nil
}

// Cancel 取消任务
func (d *Downloader) Cancel(ctx context.Context, handle *downloader.TaskHandle) error {
	if d.Client != nil {
		return d.Client.Cancel(ctx, handle)
	}
	return nil
}

// Test 测试连接
func (d *Downloader) Test(ctx context.Context) (string, error) {
	if d.Client != nil {
		return d.Client.Test(ctx)
	}
	return "", nil
}

// IsEnabled 检查下载器是否启用
func (d *Downloader) IsEnabled() bool {
	return d.Client != nil
}
