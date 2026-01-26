# Downloader 下载器包

`pkg/downloader` 包提供了统一的下载器接口，支持多种下载后端（aria2、qBittorrent）。

## 功能特性

- **统一接口**: 提供通用的 `Downloader` 接口，支持多种下载后端
- **任务管理**: 创建、查询、取消下载任务
- **文件选择**: 支持选择性下载（仅下载部分文件）
- **状态追踪**: 实时获取下载进度、速度、状态等信息
- **连接测试**: 测试与下载服务的连接状态

## 支持的下载后端

### 1. aria2

aria2 是一个轻量级的多协议、多源命令行下载工具，支持 HTTP/HTTPS、FTP、SFTP、BitTorrent 和 Metalink。

**特性**:
- JSON-RPC 2.0 协议
- HTTP 和 WebSocket 双传输支持
- 支持 BitTorrent 下载
- 支持断点续传

### 2. qBittorrent

qBittorrent 是一个开源的 BitTorrent 客户端，提供 Web API 接口。

**特性**:
- REST API 接口
- 完整的 BitTorrent 支持
- 支持 RSS 订阅
- 支持 Web UI 管理

## 安装

确保已安装必要的依赖：

```bash
go get github.com/gorilla/websocket
go get github.com/gofrs/uuid
go get github.com/samber/lo
```

## 基本使用

### 接口定义

```go
type Downloader interface {
    // CreateTask 使用给定的 URL 和选项创建任务，返回任务句柄
    CreateTask(ctx context.Context, url string, options map[string]interface{}) (*TaskHandle, error)

    // Info 返回任务状态
    Info(ctx context.Context, handle *TaskHandle) (*TaskStatus, error)

    // Cancel 取消任务
    Cancel(ctx context.Context, handle *TaskHandle) error

    // SetFilesToDownload 设置要下载的文件
    SetFilesToDownload(ctx context.Context, handle *TaskHandle, args ...*SetFileToDownloadArgs) error

    // Test 测试与下载器的连接
    Test(ctx context.Context) (string, error)
}
```

### 使用 aria2

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/top-system/light-admin/pkg/downloader/aria2"
)

// 实现 Logger 接口
type logger struct{}

func (l *logger) Info(format string, args ...interface{})    { log.Printf("[INFO] "+format, args...) }
func (l *logger) Debug(format string, args ...interface{})   { log.Printf("[DEBUG] "+format, args...) }
func (l *logger) Warning(format string, args ...interface{}) { log.Printf("[WARN] "+format, args...) }
func (l *logger) Error(format string, args ...interface{})   { log.Printf("[ERROR] "+format, args...) }

func main() {
    ctx := context.Background()

    // 创建 aria2 客户端
    client := aria2.New(&logger{}, &aria2.Settings{
        Server:   "http://localhost:6800",
        Token:    "your-secret-token",
        TempPath: "/tmp/downloads",
        Options: map[string]interface{}{
            "max-concurrent-downloads": 5,
            "split":                    16,
        },
    })

    // 测试连接
    version, err := client.Test(ctx)
    if err != nil {
        log.Fatalf("Failed to connect to aria2: %v", err)
    }
    fmt.Printf("Connected to aria2 version: %s\n", version)

    // 创建下载任务
    handle, err := client.CreateTask(ctx, "https://example.com/file.zip", nil)
    if err != nil {
        log.Fatalf("Failed to create task: %v", err)
    }
    fmt.Printf("Task created with ID: %s\n", handle.ID)

    // 获取任务状态
    status, err := client.Info(ctx, handle)
    if err != nil {
        log.Fatalf("Failed to get task info: %v", err)
    }
    fmt.Printf("Download progress: %.2f%%\n", status.Progress())
    fmt.Printf("Speed: %d bytes/s\n", status.DownloadSpeed)
}
```

### 使用 qBittorrent

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/top-system/light-admin/pkg/downloader/qbittorrent"
)

// Logger 实现同上

func main() {
    ctx := context.Background()

    // 创建 qBittorrent 客户端
    client, err := qbittorrent.New(&logger{}, &qbittorrent.Settings{
        Server:   "http://localhost:8080",
        User:     "admin",
        Password: "adminadmin",
        TempPath: "/tmp/downloads",
        Options: map[string]interface{}{
            "sequentialDownload": "true",
        },
    })
    if err != nil {
        log.Fatalf("Failed to create qBittorrent client: %v", err)
    }

    // 测试连接
    version, err := client.Test(ctx)
    if err != nil {
        log.Fatalf("Failed to connect to qBittorrent: %v", err)
    }
    fmt.Printf("Connected to qBittorrent version: %s\n", version)

    // 创建下载任务（使用磁力链接）
    handle, err := client.CreateTask(ctx, "magnet:?xt=urn:btih:...", nil)
    if err != nil {
        log.Fatalf("Failed to create task: %v", err)
    }
    fmt.Printf("Task created with ID: %s\n", handle.ID)

    // 获取任务状态
    status, err := client.Info(ctx, handle)
    if err != nil {
        log.Fatalf("Failed to get task info: %v", err)
    }
    fmt.Printf("Download: %s\n", status.Name)
    fmt.Printf("Progress: %.2f%%\n", status.Progress())
}
```

## 数据结构

### TaskHandle

任务句柄，用于后续操作。

```go
type TaskHandle struct {
    ID   string `json:"id"`   // 任务 ID
    Hash string `json:"hash"` // 任务哈希（BitTorrent）
}
```

### TaskStatus

任务状态信息。

```go
type TaskStatus struct {
    FollowedBy    *TaskHandle // 任务句柄变更指示
    SavePath      string      // 保存路径
    Name          string      // 任务名称
    State         Status      // 状态
    Total         int64       // 总大小（字节）
    Downloaded    int64       // 已下载（字节）
    DownloadSpeed int64       // 下载速度（字节/秒）
    Uploaded      int64       // 已上传（字节）
    UploadSpeed   int64       // 上传速度（字节/秒）
    Hash          string      // 哈希值
    Files         []TaskFile  // 文件列表
    Pieces        []byte      // 分片下载状态
    NumPieces     int         // 分片数量
    ErrorMessage  string      // 错误信息
}
```

### 任务状态常量

```go
const (
    StatusDownloading Status = "downloading" // 下载中
    StatusSeeding     Status = "seeding"     // 做种中
    StatusCompleted   Status = "completed"   // 已完成
    StatusError       Status = "error"       // 错误
    StatusUnknown     Status = "unknown"     // 未知
)
```

## 高级功能

### 选择性下载

可以选择只下载 BitTorrent 任务中的部分文件：

```go
// 获取任务状态以查看文件列表
status, _ := client.Info(ctx, handle)

// 打印文件列表
for _, file := range status.Files {
    fmt.Printf("[%d] %s (%d bytes) - Selected: %v\n",
        file.Index, file.Name, file.Size, file.Selected)
}

// 设置要下载的文件（仅下载索引为 0 和 2 的文件）
client.SetFilesToDownload(ctx, handle,
    &downloader.SetFileToDownloadArgs{Index: 0, Download: true},
    &downloader.SetFileToDownloadArgs{Index: 1, Download: false},
    &downloader.SetFileToDownloadArgs{Index: 2, Download: true},
)
```

### 监控下载进度

```go
func monitorProgress(ctx context.Context, client downloader.Downloader, handle *downloader.TaskHandle) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            status, err := client.Info(ctx, handle)
            if err != nil {
                log.Printf("Error getting status: %v", err)
                continue
            }

            fmt.Printf("\rProgress: %.2f%% | Speed: %s/s | ETA: calculating...",
                status.Progress(),
                formatBytes(status.DownloadSpeed),
            )

            if status.IsComplete() {
                fmt.Println("\nDownload completed!")
                return
            }

            if status.IsError() {
                fmt.Printf("\nDownload error: %s\n", status.ErrorMessage)
                return
            }
        }
    }
}

func formatBytes(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

## aria2 RPC 接口

`pkg/downloader/aria2/rpc` 包提供了完整的 aria2 JSON-RPC 2.0 接口实现。

### 支持的方法

| 方法 | 描述 |
|------|------|
| AddURI | 添加 HTTP/FTP 下载 |
| AddTorrent | 添加 .torrent 文件下载 |
| AddMetalink | 添加 Metalink 下载 |
| Remove | 移除任务 |
| ForceRemove | 强制移除任务 |
| Pause | 暂停任务 |
| PauseAll | 暂停所有任务 |
| Unpause | 恢复任务 |
| UnpauseAll | 恢复所有任务 |
| TellStatus | 获取任务状态 |
| GetURIs | 获取任务 URI 列表 |
| GetFiles | 获取任务文件列表 |
| GetPeers | 获取 Peer 列表 |
| TellActive | 获取活动任务 |
| TellWaiting | 获取等待任务 |
| TellStopped | 获取已停止任务 |
| ChangePosition | 修改任务位置 |
| ChangeOption | 修改任务选项 |
| GetGlobalOption | 获取全局选项 |
| ChangeGlobalOption | 修改全局选项 |
| GetGlobalStat | 获取全局统计 |
| PurgeDownloadResult | 清除下载结果 |
| GetVersion | 获取版本信息 |
| Shutdown | 关闭 aria2 |
| SaveSession | 保存会话 |

### 使用 WebSocket 通知

```go
package main

import (
    "context"
    "log"

    "github.com/top-system/light-admin/pkg/downloader/aria2/rpc"
)

// 实现 Notifier 接口
type downloadNotifier struct{}

func (n *downloadNotifier) OnDownloadStart(events []rpc.Event) {
    for _, e := range events {
        log.Printf("Download started: %s", e.Gid)
    }
}

func (n *downloadNotifier) OnDownloadPause(events []rpc.Event) {
    for _, e := range events {
        log.Printf("Download paused: %s", e.Gid)
    }
}

func (n *downloadNotifier) OnDownloadStop(events []rpc.Event) {
    for _, e := range events {
        log.Printf("Download stopped: %s", e.Gid)
    }
}

func (n *downloadNotifier) OnDownloadComplete(events []rpc.Event) {
    for _, e := range events {
        log.Printf("Download completed: %s", e.Gid)
    }
}

func (n *downloadNotifier) OnDownloadError(events []rpc.Event) {
    for _, e := range events {
        log.Printf("Download error: %s", e.Gid)
    }
}

func (n *downloadNotifier) OnBtDownloadComplete(events []rpc.Event) {
    for _, e := range events {
        log.Printf("BT download completed (seeding): %s", e.Gid)
    }
}

func main() {
    ctx := context.Background()

    // 使用 WebSocket 连接以接收通知
    client, err := rpc.New(ctx, "ws://localhost:6800/jsonrpc", "secret-token", 0, &downloadNotifier{})
    if err != nil {
        log.Fatal(err)
    }

    // ... 使用 client
}
```

## 目录结构

```
pkg/downloader/
├── downloader.go              # 主接口定义
├── aria2/
│   ├── aria2.go              # aria2 客户端实现
│   └── rpc/
│       ├── client.go         # RPC 客户端（HTTP/WebSocket）
│       ├── const.go          # RPC 方法常量
│       ├── json2.go          # JSON-RPC 2.0 编解码
│       ├── notification.go   # 通知接口
│       ├── proc.go           # 响应处理器
│       ├── proto.go          # 协议接口
│       └── types.go          # 类型定义
└── qbittorrent/
    ├── qbittorrent.go        # qBittorrent 客户端实现
    └── types.go              # 类型定义
```

## 注意事项

1. **aria2 安装**: 使用 aria2 前需要确保 aria2 已安装并启动 RPC 服务
   ```bash
   aria2c --enable-rpc --rpc-listen-all --rpc-secret=your-token
   ```

2. **qBittorrent 配置**: 使用 qBittorrent 前需要启用 Web UI
   - 设置 -> Web UI -> 启用 Web 用户界面

3. **临时文件**: 下载器会在 TempPath 下创建临时文件夹，任务取消后会自动清理

4. **并发安全**: 所有客户端方法都是并发安全的

5. **超时设置**: 建议为长时间运行的操作设置合适的 context 超时

## 错误处理

```go
import "github.com/top-system/light-admin/pkg/downloader"

status, err := client.Info(ctx, handle)
if err != nil {
    if errors.Is(err, downloader.ErrTaskNotFound) {
        // 任务不存在或已被删除
        log.Println("Task not found")
    } else {
        // 其他错误
        log.Printf("Error: %v", err)
    }
}
```
