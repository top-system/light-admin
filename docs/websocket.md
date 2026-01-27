# WebSocket 实时通信

这是一个功能完整的 WebSocket 实时通信模块，采用标准 STOMP 协议，与原 Java (Spring WebSocket) 项目完全兼容。

## 特性

- **标准 STOMP 协议**: 完全兼容 `@stomp/stompjs` 等标准 STOMP 客户端
- **STOMP 风格消息路由**: 支持 `/app/*`、`/topic/*`、`/user/*/queue/*` 消息路径
- **用户会话管理**: 支持多设备同时在线，自动管理会话生命周期
- **消息广播**: 向所有在线用户或订阅者发送消息
- **点对点通信**: 向指定用户发送私信
- **主题订阅**: 支持订阅特定主题接收消息
- **在线状态**: 实时获取在线用户数量和列表

## 架构对比

| Java (Spring WebSocket) | Go 实现 |
|------------------------|---------|
| `/ws` 端点 | `/ws` 端点 |
| `@MessageMapping("/sendToAll")` | `RegisterHandler("/app/sendToAll", ...)` |
| `@SendTo("/topic/notice")` | `Broker.Publish("/topic/notice", ...)` |
| `SimpMessagingTemplate.convertAndSendToUser()` | `Broker.SendToUser(username, destination, ...)` |

## 端点和路由

### WebSocket 连接端点

```
GET /ws
```

客户端通过此端点建立 WebSocket 连接（与原 Java 项目一致）。

### 认证流程

1. 客户端建立 WebSocket 连接到 `/ws`
2. 发送 STOMP CONNECT 帧，在 `Authorization` 头或 `login` 头中携带 Bearer Token
3. 服务端验证 Token，成功返回 CONNECTED 帧，失败返回 ERROR 帧

### 消息目标前缀

| 前缀 | 说明 | 示例 |
|------|------|------|
| `/app/*` | 客户端发送消息到服务端 | `/app/sendToAll` |
| `/topic/*` | 广播消息（发布/订阅模式） | `/topic/notice` |
| `/user/{username}/queue/*` | 点对点消息 | `/user/zhangsan/queue/greeting` |

### HTTP API（服务端主动推送）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/websocket/sendToAll` | 广播消息 |
| POST | `/api/v1/websocket/sendToUser` | 点对点消息 |
| GET | `/api/v1/websocket/online-users` | 获取在线用户列表 |
| GET | `/api/v1/websocket/online-count` | 获取在线用户数 |
| POST | `/api/v1/websocket/dict-change` | 广播字典变更 |

## 前端使用示例（@stomp/stompjs）

推荐使用 `@stomp/stompjs` 库，与 Spring WebSocket 完全兼容。

### 安装依赖

```bash
npm install @stomp/stompjs
```

### 基本连接

```typescript
import { Client } from '@stomp/stompjs'

const client = new Client({
  brokerURL: 'ws://localhost:2222/ws',

  // 在 CONNECT 帧中携带 Token 认证
  connectHeaders: {
    Authorization: `Bearer ${getToken()}`
  },

  // 自动重连
  reconnectDelay: 5000,
  heartbeatIncoming: 4000,
  heartbeatOutgoing: 4000,

  onConnect: (frame) => {
    console.log('Connected:', frame)

    // 订阅广播通知
    client.subscribe('/topic/notice', (message) => {
      console.log('收到通知:', message.body)
    })

    // 订阅在线用户数
    client.subscribe('/topic/online-count', (message) => {
      console.log('在线用户数:', message.body)
    })

    // 订阅字典变更
    client.subscribe('/topic/dict', (message) => {
      const data = JSON.parse(message.body)
      console.log('字典变更:', data.dictCode)
    })

    // 订阅个人消息（点对点）
    client.subscribe('/user/admin/queue/greeting', (message) => {
      const data = JSON.parse(message.body)
      console.log('收到私信:', data)
    })
  },

  onStompError: (frame) => {
    console.error('STOMP error:', frame.headers['message'])
  },

  onDisconnect: () => {
    console.log('Disconnected')
  }
})

// 连接
client.activate()
```

### 发送消息

```typescript
// 广播消息 (对应 Java @MessageMapping("/sendToAll"))
client.publish({
  destination: '/app/sendToAll',
  body: JSON.stringify('这是一条广播消息')
})

// 发送私信 (对应 Java @MessageMapping("/sendToUser/{username}"))
client.publish({
  destination: '/app/sendToUser',
  body: JSON.stringify({
    username: 'zhangsan',
    message: '你好，张三'
  })
})
```

### Vue 3 封装示例

```typescript
// src/utils/websocket.ts
import { ref, onUnmounted } from 'vue'
import { Client, IMessage } from '@stomp/stompjs'

export function useWebSocket() {
  const connected = ref(false)
  const onlineCount = ref(0)
  let client: Client | null = null

  function connect(token: string) {
    client = new Client({
      brokerURL: import.meta.env.VITE_WS_URL || 'ws://localhost:2222/ws',
      connectHeaders: {
        Authorization: `Bearer ${token}`
      },
      reconnectDelay: 5000,
      heartbeatIncoming: 4000,
      heartbeatOutgoing: 4000,

      onConnect: () => {
        connected.value = true

        // 订阅在线用户数
        client?.subscribe('/topic/online-count', (msg: IMessage) => {
          onlineCount.value = parseInt(msg.body)
        })

        // 订阅通知
        client?.subscribe('/topic/notice', (msg: IMessage) => {
          ElNotification({ title: '通知', message: msg.body })
        })

        // 订阅字典变更
        client?.subscribe('/topic/dict', (msg: IMessage) => {
          const data = JSON.parse(msg.body)
          // 触发字典刷新事件
          window.dispatchEvent(new CustomEvent('dict-change', { detail: data }))
        })
      },

      onDisconnect: () => {
        connected.value = false
      },

      onStompError: (frame) => {
        console.error('WebSocket error:', frame.headers['message'])
        connected.value = false
      }
    })

    client.activate()
  }

  function disconnect() {
    client?.deactivate()
  }

  function sendToAll(message: string) {
    client?.publish({
      destination: '/app/sendToAll',
      body: JSON.stringify(message)
    })
  }

  function sendToUser(username: string, message: string) {
    client?.publish({
      destination: '/app/sendToUser',
      body: JSON.stringify({ username, message })
    })
  }

  function subscribe(destination: string, callback: (message: IMessage) => void) {
    return client?.subscribe(destination, callback)
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    connected,
    onlineCount,
    connect,
    disconnect,
    sendToAll,
    sendToUser,
    subscribe
  }
}
```

### 在组件中使用

```vue
<script setup lang="ts">
import { onMounted } from 'vue'
import { useWebSocket } from '@/utils/websocket'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const { connected, onlineCount, connect, sendToAll } = useWebSocket()

onMounted(() => {
  connect(userStore.token)
})

function broadcastMessage() {
  sendToAll('Hello everyone!')
}
</script>

<template>
  <div>
    <span>状态: {{ connected ? '已连接' : '未连接' }}</span>
    <span>在线用户: {{ onlineCount }}</span>
    <button @click="broadcastMessage">发送广播</button>
  </div>
</template>
```

## 服务端使用示例

### 在其他服务中使用 WebSocket

```go
package service

import (
    ws "github.com/top-system/light-admin/pkg/websocket"
)

type DictService struct {
    websocket *ws.WebSocket
}

func NewDictService(websocket *ws.WebSocket) *DictService {
    return &DictService{websocket: websocket}
}

// 字典更新后广播通知
func (s *DictService) UpdateDict(dictCode string) error {
    // ... 更新字典逻辑

    // 广播字典变更通知 (对应 Java webSocketService.broadcastDictChange)
    s.websocket.BroadcastDictChange(dictCode)

    return nil
}

// 发送通知给指定用户
func (s *DictService) NotifyUser(username, message string) {
    s.websocket.SendNotification(username, message)
}

// 检查用户是否在线
func (s *DictService) IsUserOnline(username string) bool {
    return s.websocket.IsUserOnline(username)
}
```

## STOMP 协议格式

### CONNECT 帧（客户端认证）

```
CONNECT
accept-version:1.2
host:localhost
Authorization:Bearer <token>

^@
```

或使用 login 头：

```
CONNECT
accept-version:1.2
host:localhost
login:<token>

^@
```

### CONNECTED 帧（认证成功）

```
CONNECTED
version:1.2
session:<session-id>
server:echo-admin/1.0
heart-beat:0,0

^@
```

### SUBSCRIBE 帧（订阅主题）

```
SUBSCRIBE
id:sub-0
destination:/topic/notice

^@
```

### MESSAGE 帧（接收消息）

```
MESSAGE
destination:/topic/notice
message-id:msg-1
subscription:sub-0
content-type:application/json

{"content":"Hello"}^@
```

### SEND 帧（发送消息）

```
SEND
destination:/app/sendToAll
content-type:application/json

"Hello everyone"^@
```

### ERROR 帧（错误响应）

```
ERROR
message:Token validation failed
content-type:text/plain

Token validation failed^@
```

## 目录结构

```
pkg/websocket/
├── stomp/
│   ├── frame.go      # STOMP 帧解析和序列化
│   └── broker.go     # 消息代理（会话管理、消息路由）
└── websocket.go      # WebSocket 管理器（对外接口）

api/platform/
├── controller/
│   └── websocket_controller.go  # WebSocket 控制器
└── route/
    └── websocket_route.go       # 路由配置
```

## 与 Java 项目对应关系

| Java 类/方法 | Go 对应 |
|-------------|---------|
| `WebSocketConfig` | `api/platform/route/websocket_route.go` |
| `WebsocketController` | `api/platform/controller/websocket_controller.go` |
| `@MessageMapping("/sendToAll")` | `ws.RegisterHandler("/app/sendToAll", handler)` |
| `@SendTo("/topic/notice")` | `ws.Broker.Publish("/topic/notice", msg)` |
| `SimpMessagingTemplate.convertAndSendToUser()` | `ws.Broker.SendToUser(username, dest, msg)` |
| `WebSocketService.broadcastDictChange()` | `ws.BroadcastDictChange(dictCode)` |
| `UserSessionRegistry` | `stomp.Broker` (内置会话管理) |

## 注意事项

1. **认证**: STOMP CONNECT 帧需要携带有效的 Bearer Token（在 `Authorization` 或 `login` 头中）
2. **跨域**: 默认允许所有来源，生产环境应配置白名单
3. **心跳**: 支持 STOMP 心跳协议，建议客户端配置心跳
4. **重连**: `@stomp/stompjs` 内置自动重连功能
5. **订阅**: 客户端需要先订阅主题才能收到 `/topic/*` 的消息
6. **点对点**: 用户队列格式为 `/user/{username}/queue/*`
