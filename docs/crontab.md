# 定时任务 (Crontab)

基于 [robfig/cron](https://github.com/robfig/cron) 封装的定时任务调度器，提供任务注册、动态管理、日志追踪等功能。

## 特性

- **Cron 表达式**: 支持标准 5 字段和扩展 6 字段（含秒）格式
- **动态管理**: 运行时添加、删除、启用、禁用任务
- **任务追踪**: 每次执行自动生成关联 ID，便于日志追踪
- **Panic 恢复**: 任务执行异常自动恢复，不影响其他任务
- **全局注册**: 支持在 init() 中预注册任务
- **执行统计**: 查看任务下次执行时间、上次执行时间

## 快速开始

### 1. 基础用法

```go
package main

import (
    "context"
    "fmt"

    "github.com/top-system/light-admin/pkg/crontab"
)

func main() {
    // 创建日志记录器
    logger := crontab.NewDefaultLogger()

    // 创建定时任务调度器（支持秒级精度）
    c := crontab.New(logger)

    // 添加任务：每分钟执行
    c.AddTask("my-task", crontab.EveryMinute, func(ctx context.Context) {
        fmt.Println("Task executed!")
    })

    // 启动调度器
    c.Start()

    // 程序退出时停止
    defer c.Stop()

    // 保持运行
    select {}
}
```

### 2. 标准 Cron 格式（5 字段）

```go
// 使用标准 5 字段格式（分 时 日 月 周）
c := crontab.NewWithStandardParser(logger)

// 每天凌晨 2 点执行
c.AddTask("daily-backup", "0 2 * * *", func(ctx context.Context) {
    fmt.Println("Backup started")
})

c.Start()
```

### 3. 全局预注册任务

```go
package tasks

import (
    "context"
    "github.com/top-system/light-admin/pkg/crontab"
)

func init() {
    // 在 init 中注册，所有 Crontab 实例都会包含此任务
    crontab.Register("cleanup", crontab.EveryHour, func(ctx context.Context) {
        // 清理逻辑
    })

    crontab.RegisterWithType(crontab.CronTypeBackup, crontab.EveryDay, func(ctx context.Context) {
        // 备份逻辑
    })
}
```

### 4. 动态管理任务

```go
c := crontab.New(logger)
c.Start()

// 添加任务
c.AddTask("report", "0 0 9 * * *", generateReport)

// 禁用任务
c.DisableTask("report")

// 启用任务
c.EnableTask("report")

// 修改执行时间
c.UpdateTaskSpec("report", "0 0 10 * * *")

// 删除任务
c.RemoveTask("report")

// 立即执行任务（不影响定时调度）
c.RunTask("report")
```

### 5. 查看任务信息

```go
// 获取所有任务
tasks := c.GetTasks()
for _, task := range tasks {
    fmt.Printf("Task: %s, Spec: %s, Enabled: %v, Next: %v\n",
        task.Name, task.Spec, task.Enable, task.Next)
}

// 获取单个任务
task, err := c.GetTask("cleanup")
if err == nil {
    fmt.Printf("Next run: %v, Last run: %v\n", task.Next, task.Prev)
}

// 统计信息
fmt.Printf("Total: %d, Active: %d, Running: %v\n",
    c.TaskCount(), c.ActiveTaskCount(), c.IsRunning())
```

### 6. 使用上下文

```go
c := crontab.New(logger,
    crontab.WithContextData("db", database),
    crontab.WithContextData("config", config),
)

c.AddTask("sync", crontab.EveryHour, func(ctx context.Context) {
    // 获取关联 ID
    cid := crontab.CorrelationIDFromContext(ctx)
    fmt.Printf("Correlation ID: %s\n", cid)

    // 获取日志记录器
    log := crontab.LoggerFromContext(ctx)
    log.Info("Sync started")

    // 获取自定义数据
    db := ctx.Value("db")
    // ...
})
```

### 7. 自定义日志记录器

```go
import "go.uber.org/zap"

type ZapLogger struct {
    logger *zap.SugaredLogger
    prefix string
}

func (l *ZapLogger) Debug(format string, args ...interface{}) {
    l.logger.Debugf(l.prefix+format, args...)
}

func (l *ZapLogger) Info(format string, args ...interface{}) {
    l.logger.Infof(l.prefix+format, args...)
}

func (l *ZapLogger) Warning(format string, args ...interface{}) {
    l.logger.Warnf(l.prefix+format, args...)
}

func (l *ZapLogger) Error(format string, args ...interface{}) {
    l.logger.Errorf(l.prefix+format, args...)
}

func (l *ZapLogger) CopyWithPrefix(prefix string) crontab.Logger {
    return &ZapLogger{logger: l.logger, prefix: prefix + " "}
}

// 使用
zapLogger, _ := zap.NewProduction()
logger := &ZapLogger{logger: zapLogger.Sugar()}
c := crontab.New(logger)
```

## Cron 表达式

### 6 字段格式（含秒，默认）

```
秒 分 时 日 月 周
*  *  *  *  *  *
```

### 5 字段格式（标准，使用 NewWithStandardParser）

```
分 时 日 月 周
*  *  *  *  *
```

### 特殊字符

| 字符 | 说明 | 示例 |
|------|------|------|
| `*` | 任意值 | `* * * * * *` 每秒 |
| `,` | 列表 | `0,30 * * * * *` 第 0 和 30 秒 |
| `-` | 范围 | `0-30 * * * * *` 第 0-30 秒 |
| `/` | 步长 | `*/10 * * * * *` 每 10 秒 |

### 预定义常量

```go
// 6 字段格式（含秒）
crontab.EveryMinute      // "0 * * * * *"     每分钟
crontab.EveryFiveMinute  // "0 */5 * * * *"   每 5 分钟
crontab.EveryTenMinute   // "0 */10 * * * *"  每 10 分钟
crontab.EveryHour        // "0 0 * * * *"     每小时
crontab.EveryDay         // "0 0 0 * * *"     每天 0 点
crontab.EveryWeek        // "0 0 0 * * 0"     每周日 0 点
crontab.EveryMonth       // "0 0 0 1 * *"     每月 1 日 0 点

// 5 字段格式（标准）
crontab.StandardEveryMinute      // "* * * * *"
crontab.StandardEveryFiveMinute  // "*/5 * * * *"
crontab.StandardEveryHour        // "0 * * * *"
crontab.StandardEveryDay         // "0 0 * * *"
```

### 常用示例

```go
// 每天凌晨 3 点
"0 0 3 * * *"

// 每周一到周五上午 9 点
"0 0 9 * * 1-5"

// 每月 1 日和 15 日中午 12 点
"0 0 12 1,15 * *"

// 工作日每小时的第 30 分钟
"0 30 * * * 1-5"

// 每 15 分钟
"0 */15 * * * *"
```

## API 参考

### Crontab 方法

```go
// 创建实例
func New(logger Logger, opts ...Option) *Crontab
func NewWithStandardParser(logger Logger, opts ...Option) *Crontab

// 生命周期
func (c *Crontab) Start() error
func (c *Crontab) Stop() context.Context

// 任务管理
func (c *Crontab) AddTask(name string, spec string, fn CronTaskFunc) error
func (c *Crontab) AddTaskWithType(t CronType, spec string, fn CronTaskFunc) error
func (c *Crontab) RemoveTask(name string) error
func (c *Crontab) EnableTask(name string) error
func (c *Crontab) DisableTask(name string) error
func (c *Crontab) UpdateTaskSpec(name string, newSpec string) error
func (c *Crontab) RunTask(name string) error

// 查询
func (c *Crontab) GetTasks() []TaskInfo
func (c *Crontab) GetTask(name string) (*TaskInfo, error)
func (c *Crontab) IsRunning() bool
func (c *Crontab) TaskCount() int
func (c *Crontab) ActiveTaskCount() int
```

### 全局注册

```go
func Register(name string, spec string, fn CronTaskFunc)
func RegisterWithType(t CronType, spec string, fn CronTaskFunc)
```

### 上下文工具

```go
func LoggerFromContext(ctx context.Context) Logger
func CorrelationIDFromContext(ctx context.Context) uuid.UUID
```

## 最佳实践

1. **任务命名**: 使用有意义的名称，便于日志追踪和管理。

2. **错误处理**: 任务内部应处理好错误，记录日志而不是 panic。

3. **执行时间**: 任务应尽快完成，长时间任务考虑异步处理。

4. **重叠执行**: 默认不会阻止任务重叠执行，如需防止请自行加锁。

5. **优雅停止**: 调用 `Stop()` 会等待当前执行的任务完成。

## 与项目集成

### 与 Uber fx 集成

```go
// module.go
var Module = fx.Options(
    fx.Provide(NewCrontab),
    fx.Invoke(StartCrontab),
)

func NewCrontab(logger lib.Logger) *crontab.Crontab {
    // 适配 logger
    cronLogger := &LoggerAdapter{logger: logger}
    return crontab.New(cronLogger)
}

func StartCrontab(lc fx.Lifecycle, c *crontab.Crontab) {
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            return c.Start()
        },
        OnStop: func(ctx context.Context) error {
            c.Stop()
            return nil
        },
    })
}
```

### 注册业务任务

```go
// tasks/cleanup.go
func init() {
    crontab.Register("cleanup-temp-files", crontab.EveryHour, CleanupTempFiles)
}

func CleanupTempFiles(ctx context.Context) {
    log := crontab.LoggerFromContext(ctx)
    log.Info("Starting cleanup...")
    // 清理逻辑
}
```
