# 任务队列 (Task Queue)

这是一个功能完整的任务队列系统，支持 Worker Pool、任务调度、重试机制、状态持久化等功能。

## 特性

- **Worker Pool**: 多工作线程并发处理任务
- **FIFO 调度器**: 基于堆的优先级队列，支持延迟执行
- **任务状态机**: 完整的状态流转管理（Queued → Processing → Success/Failed/Suspended）
- **重试机制**: 支持指数退避的自动重试
- **任务持久化**: 支持 GORM 数据库持久化或纯内存模式
- **指标统计**: 实时统计任务成功/失败/挂起数量

## 快速开始

### 1. 基础用法（纯内存模式）

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/top-system/light-admin/pkg/queue"
)

// 定义你的任务
type MyTask struct {
    *queue.InMemoryTask
    Name string
}

func NewMyTask(name string) *MyTask {
    return &MyTask{
        InMemoryTask: &queue.InMemoryTask{
            DBTask: &queue.DBTask{
                TaskModel: &queue.TaskModel{
                    Type: "my_task",
                },
            },
        },
        Name: name,
    }
}

func (t *MyTask) Do(ctx context.Context) (queue.Status, error) {
    fmt.Printf("Processing task: %s\n", t.Name)
    // 执行你的业务逻辑
    time.Sleep(1 * time.Second)
    return queue.StatusCompleted, nil
}

func main() {
    // 创建日志记录器
    logger := queue.NewDefaultLogger()

    // 创建队列（纯内存模式，不需要数据库）
    q := queue.New(
        logger,
        nil, // 不使用持久化
        queue.NewTaskRegistry(),
        queue.WithWorkerCount(4),
        queue.WithName("my-queue"),
    )

    // 启动队列
    q.Start()

    // 提交任务
    task := NewMyTask("Hello World")
    if err := q.QueueTask(context.Background(), task); err != nil {
        fmt.Printf("Failed to queue task: %v\n", err)
    }

    // 等待任务完成（实际应用中可能需要更优雅的方式）
    time.Sleep(5 * time.Second)

    // 关闭队列
    q.Shutdown()

    // 打印统计信息
    fmt.Printf("Success: %d, Failure: %d\n", q.SuccessTasks(), q.FailureTasks())
}
```

### 2. 带数据库持久化（MySQL）

```go
package main

import (
    "context"
    "fmt"

    "github.com/top-system/light-admin/pkg/queue"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func main() {
    // 连接数据库
    dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    // 自动迁移任务表
    db.AutoMigrate(&queue.TaskModel{})

    // 创建任务仓库
    taskRepo := queue.NewGormTaskRepository(db)

    // 创建日志记录器
    logger := queue.NewDefaultLogger()

    // 创建队列
    q := queue.New(
        logger,
        taskRepo,
        queue.NewTaskRegistry(),
        queue.WithWorkerCount(4),
        queue.WithName("persistent-queue"),
        queue.WithResumeTaskType("my_task"), // 启动时恢复这些类型的任务
    )

    // 启动队列
    q.Start()

    // ... 提交任务等操作
}
```

### 2.1 使用 SQLite 持久化

```go
package main

import (
    "github.com/top-system/light-admin/pkg/queue"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func main() {
    // 连接 SQLite 数据库（开启 WAL 模式提高并发性能）
    dsn := "./data/app.db?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL"
    db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    // 自动迁移任务表
    db.AutoMigrate(&queue.TaskModel{})

    // 创建任务仓库
    taskRepo := queue.NewGormTaskRepository(db)

    // 创建日志记录器
    logger := queue.NewDefaultLogger()

    // 创建队列（SQLite 建议使用较少的 Worker）
    q := queue.New(
        logger,
        taskRepo,
        queue.NewTaskRegistry(),
        queue.WithWorkerCount(2),
        queue.WithName("sqlite-queue"),
    )

    q.Start()
    // ...
}
```

### 3. 带重试的任务

```go
type RetryableTask struct {
    *queue.DBTask
    attempts int
}

func NewRetryableTask() *RetryableTask {
    return &RetryableTask{
        DBTask: &queue.DBTask{
            TaskModel: &queue.TaskModel{
                Type: "retryable_task",
            },
        },
    }
}

func (t *RetryableTask) Do(ctx context.Context) (queue.Status, error) {
    t.attempts++

    // 模拟前两次失败
    if t.attempts < 3 {
        return queue.StatusError, fmt.Errorf("attempt %d failed", t.attempts)
    }

    fmt.Println("Task succeeded on attempt", t.attempts)
    return queue.StatusCompleted, nil
}

func (t *RetryableTask) ShouldPersist() bool {
    return true // 启用持久化
}

func main() {
    // ... 创建队列

    q := queue.New(
        logger,
        taskRepo,
        queue.NewTaskRegistry(),
        queue.WithMaxRetry(5),                      // 最多重试5次
        queue.WithRetryDelay(2 * time.Second),      // 重试间隔2秒
        queue.WithBackoffFactor(2),                 // 指数退避因子
        queue.WithBackoffMaxDuration(60 * time.Second), // 最大退避时间
    )

    // ...
}
```

### 4. 使用自定义日志记录器

```go
import "go.uber.org/zap"

// 适配 zap logger
func main() {
    zapLogger, _ := zap.NewProduction()
    sugar := zapLogger.Sugar()

    logger := queue.NewLoggerAdapter(
        func(format string, args ...interface{}) {
            sugar.Debugf(format, args...)
        },
        func(format string, args ...interface{}) {
            sugar.Infof(format, args...)
        },
        func(format string, args ...interface{}) {
            sugar.Warnf(format, args...)
        },
        func(format string, args ...interface{}) {
            sugar.Errorf(format, args...)
        },
    )

    q := queue.New(logger, nil, nil)
    // ...
}
```

## 任务状态

任务在生命周期中会经历以下状态：

```
                    ┌─────────────────┐
                    │                 │
                    ▼                 │
┌─────────┐    ┌─────────┐    ┌──────────────┐
│ Queued  │───▶│Processing│───▶│  Completed   │
└─────────┘    └─────────┘    └──────────────┘
                    │
                    │         ┌──────────────┐
                    ├────────▶│    Error     │
                    │         └──────────────┘
                    │
                    │         ┌──────────────┐
                    ├────────▶│   Canceled   │
                    │         └──────────────┘
                    │
                    │         ┌──────────────┐
                    └────────▶│  Suspending  │──┐
                              └──────────────┘  │
                                    ▲           │
                                    └───────────┘
                                   (重试后重新入队)
```

## 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithWorkerCount(n)` | Worker 数量 | CPU 核心数 |
| `WithMaxTaskExecution(d)` | 任务最大执行时间 | 60 小时 |
| `WithMaxRetry(n)` | 最大重试次数 | 0 |
| `WithRetryDelay(d)` | 固定重试延迟 | 0 (使用退避算法) |
| `WithBackoffFactor(f)` | 退避因子 | 2 |
| `WithBackoffMaxDuration(d)` | 最大退避时间 | 60 秒 |
| `WithTaskPullInterval(d)` | 任务拉取间隔 | 1 秒 |
| `WithResumeTaskType(types...)` | 启动时恢复的任务类型 | 空 |
| `WithName(name)` | 队列名称 | "default" |

## 自定义任务

创建自定义任务需要实现 `Task` 接口。最简单的方式是嵌入 `InMemoryTask` 或 `DBTask`：

### 纯内存任务

```go
type MyInMemoryTask struct {
    *queue.InMemoryTask
    // 你的字段
}

func (t *MyInMemoryTask) Do(ctx context.Context) (queue.Status, error) {
    // 你的业务逻辑
    return queue.StatusCompleted, nil
}
```

### 持久化任务

```go
type MyPersistentTask struct {
    *queue.DBTask
    // 你的字段
}

func (t *MyPersistentTask) Do(ctx context.Context) (queue.Status, error) {
    // 你的业务逻辑
    return queue.StatusCompleted, nil
}

// 返回 true 启用持久化
func (t *MyPersistentTask) ShouldPersist() bool {
    return true
}
```

### 可恢复任务

如果需要在应用重启后恢复任务，需要注册工厂函数：

```go
func init() {
    queue.RegisterResumableTaskFactory("my_task", func(model *queue.TaskModel) queue.Task {
        return &MyPersistentTask{
            DBTask: &queue.DBTask{
                TaskModel: model,
            },
        }
    })
}
```

然后在创建队列时指定需要恢复的任务类型：

```go
q := queue.New(
    logger,
    taskRepo,
    nil,
    queue.WithResumeTaskType("my_task"),
)
```

## API 参考

### Queue 接口

```go
type Queue interface {
    // 启动队列和 Worker
    Start()

    // 关闭队列
    Shutdown()

    // 提交任务
    QueueTask(ctx context.Context, t Task) error

    // 获取繁忙的 Worker 数量
    BusyWorkers() int

    // 获取成功任务数
    SuccessTasks() int

    // 获取失败任务数
    FailureTasks() int

    // 获取已提交任务数
    SubmittedTasks() int

    // 获取挂起任务数
    SuspendingTasks() int
}
```

### Task 接口

```go
type Task interface {
    // 执行任务，返回下一个状态
    Do(ctx context.Context) (Status, error)

    // 获取任务 ID
    ID() int

    // 获取任务类型
    Type() string

    // 获取任务状态
    Status() Status

    // 获取任务所有者
    Owner() *TaskOwner

    // 获取私有状态（序列化后的内部状态）
    State() string

    // 是否需要持久化
    ShouldPersist() bool

    // 是否已持久化
    Persisted() bool

    // 获取已执行时间
    Executed() time.Duration

    // 获取已重试次数
    Retried() int

    // 获取错误
    Error() error

    // 获取错误历史
    ErrorHistory() []error

    // ... 更多方法见 task.go
}
```

## 错误处理

### 可重试错误

默认情况下，任务返回的错误会触发重试（如果配置了重试）：

```go
func (t *MyTask) Do(ctx context.Context) (queue.Status, error) {
    if err := doSomething(); err != nil {
        return queue.StatusError, err // 会触发重试
    }
    return queue.StatusCompleted, nil
}
```

### 不可重试错误

使用 `CriticalErr` 标记不应重试的错误：

```go
func (t *MyTask) Do(ctx context.Context) (queue.Status, error) {
    if err := validateInput(); err != nil {
        return queue.StatusError, fmt.Errorf("validation failed: %w", queue.CriticalErr)
    }
    return queue.StatusCompleted, nil
}
```

## 数据库表结构

如果使用 GORM 持久化，任务表结构如下：

### MySQL

```sql
CREATE TABLE sys_tasks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    correlation_id CHAR(36),
    owner_id BIGINT UNSIGNED,
    private_state TEXT,
    public_retry_count INT DEFAULT 0,
    public_executed_duration BIGINT DEFAULT 0,
    public_error TEXT,
    public_error_history TEXT,
    public_resume_time BIGINT DEFAULT 0,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_correlation_id (correlation_id),
    INDEX idx_deleted_at (deleted_at)
);
```

### SQLite

```sql
CREATE TABLE sys_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    correlation_id TEXT,
    owner_id INTEGER,
    private_state TEXT,
    public_retry_count INTEGER DEFAULT 0,
    public_executed_duration INTEGER DEFAULT 0,
    public_error TEXT,
    public_error_history TEXT,
    public_resume_time INTEGER DEFAULT 0,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);
CREATE INDEX idx_sys_tasks_type ON sys_tasks(type);
CREATE INDEX idx_sys_tasks_status ON sys_tasks(status);
CREATE INDEX idx_sys_tasks_correlation_id ON sys_tasks(correlation_id);
CREATE INDEX idx_sys_tasks_deleted_at ON sys_tasks(deleted_at);
```

> **注意**: 使用 GORM AutoMigrate 会自动创建表结构，无需手动执行 SQL。

## 最佳实践

1. **任务幂等性**: 确保任务可以安全地重试，即使执行多次也不会产生副作用。

2. **合理设置超时**: 使用 `WithMaxTaskExecution` 设置合理的任务超时时间。

3. **监控指标**: 定期检查 `BusyWorkers()`、`SuccessTasks()`、`FailureTasks()` 等指标。

4. **优雅关闭**: 调用 `Shutdown()` 会等待当前正在执行的任务完成。

5. **错误分类**: 区分可重试和不可重试的错误，使用 `CriticalErr` 标记不应重试的错误。

6. **SQLite 注意事项**: 使用 SQLite 时建议减少 Worker 数量（2-4个），并开启 WAL 模式。

## 实际案例：远程下载任务

项目中的 `RemoteDownloadTask` 是一个完整的任务队列使用示例：

```go
// pkg/queue/remote_download_task.go
type RemoteDownloadTask struct {
    *DBTask
    URL        string
    SavePath   string
    Downloader downloader.Downloader
    state      RemoteDownloadTaskState
}

// 状态流转
const (
    StateNotStarted = "not_started" // -> 调用下载器创建任务
    StateMonitor    = "monitor"     // -> 监控下载进度
    StateSeeding    = "seeding"     // -> BT做种中
    StateCompleted  = "completed"   // -> 任务完成
)

func (t *RemoteDownloadTask) Do(ctx context.Context) (Status, error) {
    switch t.state.Status {
    case StateNotStarted:
        // 创建下载任务
        handle, err := t.Downloader.CreateTask(ctx, t.URL, nil)
        if err != nil {
            return StatusError, err
        }
        t.state.Handle = handle
        t.state.Status = StateMonitor
        return StatusProcessing, nil

    case StateMonitor:
        // 查询下载状态
        status, err := t.Downloader.Info(ctx, t.state.Handle)
        if err != nil {
            return StatusError, err
        }

        if status.IsComplete() {
            t.state.Status = StateCompleted
            return StatusCompleted, nil
        }

        if status.State == downloader.StatusSeeding {
            t.state.Status = StateSeeding
        }

        // 继续监控
        t.ResumeAfter(5 * time.Second)
        return StatusSuspending, nil

    case StateSeeding:
        // BT做种中，继续监控
        t.ResumeAfter(30 * time.Second)
        return StatusSuspending, nil

    case StateCompleted:
        return StatusCompleted, nil
    }

    return StatusError, errors.New("unknown state")
}

// 状态持久化
func (t *RemoteDownloadTask) State() string {
    data, _ := json.Marshal(t.state)
    return string(data)
}
```

这个例子展示了：
- 使用状态机管理任务生命周期
- 使用 `ResumeAfter()` 实现轮询监控
- 状态自动持久化到数据库
- 服务重启后自动恢复任务
