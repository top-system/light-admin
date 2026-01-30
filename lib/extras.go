package lib

import (
	"context"

	"go.uber.org/fx"

	"github.com/top-system/light-admin/pkg/crontab"
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
//   3. 不需要的功能可以注释掉对应的 fx.Provide 行
//
// 在服务中使用：
//   - 将 lib.Crontab 注入到你的服务中
//   - 例如：func NewMyService(cron lib.Crontab) *MyService
// ============================================================================

// ExtrasModule 扩展功能模块
// 在 lib/lib.go 的 Module 中添加此模块即可启用
var ExtrasModule = fx.Options(
	// ====== 定时任务 ======
	// 用于定时执行任务，如数据清理、报表生成等
	// 如不需要，注释下面这行
	fx.Provide(NewCrontab),
)

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
