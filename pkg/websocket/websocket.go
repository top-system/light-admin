package websocket

import (
	"time"

	"github.com/top-system/light-admin/pkg/websocket/stomp"
	"go.uber.org/zap"
)

// 主题常量 (与原Java项目 WebSocketTopics 一致)
const (
	// 广播主题
	TopicDict        = "/topic/dict"
	TopicOnlineCount = "/topic/online-count"
	TopicPublic      = "/topic/public"
	TopicNotice      = "/topic/notice"

	// 用户队列
	UserQueueMessages = "/queue/messages"
	UserQueueMessage  = "/queue/message"
	UserQueueGreeting = "/queue/greeting"

	// 应用目标前缀 (客户端发送消息用)
	AppSendToAll  = "/app/sendToAll"
	AppSendToUser = "/app/sendToUser"
)

// WebSocket WebSocket管理器
type WebSocket struct {
	Broker *stomp.Broker
	logger *zap.Logger
}

// New 创建WebSocket管理器
func New(logger *zap.Logger) *WebSocket {
	broker := stomp.NewBroker(logger)

	ws := &WebSocket{
		Broker: broker,
		logger: logger.With(zap.String("module", "websocket")),
	}

	// 设置连接/断开回调，用于广播在线用户数
	broker.OnConnect = func(session *stomp.Session) {
		ws.broadcastOnlineCount()
	}
	broker.OnDisconnect = func(session *stomp.Session) {
		ws.broadcastOnlineCount()
	}

	// 设置订阅回调，当用户订阅特定主题时推送初始数据
	broker.OnSubscribe = func(session *stomp.Session, destination string) {
		switch destination {
		case TopicOnlineCount:
			// 用户订阅在线人数主题时，立即发送当前在线连接数
			// 这样用户登录后能立即看到包含自己的在线人数
			count := broker.GetTotalSessionCount()
			ws.logger.Info("User subscribed to online-count, sending current count",
				zap.String("username", session.Username),
				zap.Int("count", count))
			broker.SendToSession(session.ID, TopicOnlineCount, count)
		}
	}

	return ws
}

// RegisterHandler 注册消息处理器
// destination: /app/sendToAll, /app/sendToUser 等
func (ws *WebSocket) RegisterHandler(destination string, handler stomp.MessageHandler) {
	ws.Broker.RegisterHandler(destination, handler)
}

// broadcastOnlineCount 广播在线连接数
// 使用 Broadcast 而不是 Publish，因为：
// 1. OnConnect 触发时，新用户还没订阅 /topic/online-count
// 2. 需要确保所有已认证用户都能收到在线人数更新
func (ws *WebSocket) broadcastOnlineCount() {
	// 使用会话数（每个连接都计数，同一用户多设备登录会增加）
	// 如果想用唯一用户数（去重），改用 GetOnlineUserCount()
	count := ws.Broker.GetTotalSessionCount()
	ws.Broker.Broadcast(TopicOnlineCount, count)
}

// === 服务方法 (对应 Java WebSocketService) ===

// BroadcastDictChange 广播字典变更
func (ws *WebSocket) BroadcastDictChange(dictCode string) {
	if dictCode == "" {
		return
	}
	event := map[string]interface{}{
		"dictCode":  dictCode,
		"timestamp": time.Now().UnixMilli(),
	}
	ws.Broker.Publish(TopicDict, event)
}

// SendNotification 发送通知给指定用户
func (ws *WebSocket) SendNotification(username string, message interface{}) {
	if username == "" || message == nil {
		return
	}
	ws.Broker.SendToUser(username, UserQueueMessages, message)
}

// BroadcastSystemMessage 广播系统消息
func (ws *WebSocket) BroadcastSystemMessage(message string) {
	if message == "" {
		return
	}
	msg := map[string]interface{}{
		"sender":    "System",
		"content":   message,
		"timestamp": time.Now().UnixMilli(),
	}
	ws.Broker.Publish(TopicPublic, msg)
}

// SendToUser 发送点对点消息
func (ws *WebSocket) SendToUser(sender, receiver, message string) {
	if receiver == "" {
		return
	}
	msg := map[string]interface{}{
		"sender":    sender,
		"content":   message,
		"timestamp": time.Now().UnixMilli(),
	}
	ws.Broker.SendToUser(receiver, UserQueueGreeting, msg)
}

// BroadcastNotice 广播通知
func (ws *WebSocket) BroadcastNotice(message string) {
	ws.Broker.Broadcast(TopicNotice, "Server Notice: "+message)
}

// GetOnlineUserCount 获取在线用户数
func (ws *WebSocket) GetOnlineUserCount() int {
	return ws.Broker.GetOnlineUserCount()
}

// GetOnlineUsers 获取在线用户列表
func (ws *WebSocket) GetOnlineUsers() []stomp.OnlineUser {
	return ws.Broker.GetOnlineUsers()
}

// IsUserOnline 检查用户是否在线
func (ws *WebSocket) IsUserOnline(username string) bool {
	return ws.Broker.IsUserOnline(username)
}
