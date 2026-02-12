package stomp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// 模块标识，用于日志
const moduleTag = "stomp"

// Session WebSocket会话
type Session struct {
	ID            string
	Username      string
	Conn          *websocket.Conn
	Subscriptions map[string]string // subscriptionID -> destination
	ConnectTime   int64
	Authenticated bool // 是否已认证
	mu            sync.RWMutex
}

// Subscribe 订阅主题
func (s *Session) Subscribe(subscriptionID, destination string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Subscriptions[subscriptionID] = destination
}

// Unsubscribe 取消订阅
func (s *Session) Unsubscribe(subscriptionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Subscriptions, subscriptionID)
}

// GetSubscriptionID 根据 destination 获取 subscriptionID
func (s *Session) GetSubscriptionID(destination string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id, dest := range s.Subscriptions {
		if dest == destination {
			return id
		}
	}
	return ""
}

// IsSubscribed 检查是否订阅了某个目标
func (s *Session) IsSubscribed(destination string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, dest := range s.Subscriptions {
		if dest == destination {
			return true
		}
	}
	return false
}

// MessageHandler 消息处理器函数
type MessageHandler func(session *Session, destination string, body []byte)

// TokenValidator Token验证器函数
// 返回用户名和错误，如果验证失败返回错误
type TokenValidator func(token string) (username string, err error)

// Broker STOMP消息代理
type Broker struct {
	mu             sync.RWMutex
	sessions       map[string]*Session            // sessionID -> Session
	users          map[string]map[string]*Session // username -> sessionID -> Session
	handlers       map[string]MessageHandler      // destination pattern -> handler
	logger         *zap.Logger
	tokenValidator TokenValidator // Token验证器
	messageCounter uint64         // 消息计数器

	// 回调
	OnConnect    func(session *Session)
	OnDisconnect func(session *Session)
	OnSubscribe  func(session *Session, destination string) // 订阅回调
}

// NewBroker 创建消息代理
func NewBroker(logger *zap.Logger) *Broker {
	return &Broker{
		sessions: make(map[string]*Session),
		users:    make(map[string]map[string]*Session),
		handlers: make(map[string]MessageHandler),
		logger:   logger.With(zap.String("module", moduleTag)),
	}
}

// RegisterHandler 注册消息处理器
// destination 支持 /app/sendToAll 格式
func (b *Broker) RegisterHandler(destination string, handler MessageHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[destination] = handler
}

// SetTokenValidator 设置Token验证器
func (b *Broker) SetTokenValidator(validator TokenValidator) {
	b.tokenValidator = validator
}

// AddSession 添加会话（未认证状态）
func (b *Broker) AddSession(session *Session) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.sessions[session.ID] = session

	b.logger.Info("Session added (pending authentication)",
		zap.String("sessionID", session.ID))

	// 注意：OnConnect 回调在认证成功后触发，而不是在这里
}

// RemoveSession 移除会话
func (b *Broker) RemoveSession(sessionID string) {
	b.mu.Lock()
	session, ok := b.sessions[sessionID]
	if !ok {
		b.mu.Unlock()
		return
	}

	delete(b.sessions, sessionID)

	if session.Username != "" {
		if userSessions, ok := b.users[session.Username]; ok {
			delete(userSessions, sessionID)
			if len(userSessions) == 0 {
				delete(b.users, session.Username)
			}
		}
	}
	b.mu.Unlock()

	b.logger.Info("Session removed",
		zap.String("sessionID", sessionID),
		zap.String("username", session.Username))

	if b.OnDisconnect != nil && session.Authenticated {
		b.OnDisconnect(session)
	}
}

// GetSession 获取会话
func (b *Broker) GetSession(sessionID string) *Session {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.sessions[sessionID]
}

// GetUserSessions 获取用户的所有会话
func (b *Broker) GetUserSessions(username string) []*Session {
	b.mu.RLock()
	defer b.mu.RUnlock()

	userSessions, ok := b.users[username]
	if !ok {
		return nil
	}

	result := make([]*Session, 0, len(userSessions))
	for _, session := range userSessions {
		result = append(result, session)
	}
	return result
}

// GetOnlineUserCount 获取在线用户数
func (b *Broker) GetOnlineUserCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.users)
}

// GetTotalSessionCount 获取总会话数
func (b *Broker) GetTotalSessionCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.sessions)
}

// IsUserOnline 检查用户是否在线
func (b *Broker) IsUserOnline(username string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	sessions, ok := b.users[username]
	return ok && len(sessions) > 0
}

// HandleMessage 处理客户端消息（标准 STOMP 协议）
func (b *Broker) HandleMessage(session *Session, data []byte) {
	// 防止 panic 导致连接断开
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("Panic in HandleMessage",
				zap.Any("panic", r),
				zap.String("sessionID", session.ID))
		}
	}()

	// 忽略心跳帧（空数据或只有换行符）
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		b.logger.Debug("Received heartbeat frame")
		return
	}

	b.logger.Info("Received raw data",
		zap.String("data", string(data)),
		zap.Int("length", len(data)),
		zap.String("sessionID", session.ID))

	frame, err := ParseFrame(data)
	if err != nil {
		b.logger.Error("Failed to parse STOMP frame",
			zap.Error(err),
			zap.String("data", string(data)))
		b.sendError(session, "Failed to parse frame: "+err.Error())
		return
	}

	b.logger.Info("Parsed STOMP frame",
		zap.String("command", frame.Command),
		zap.Any("headers", frame.Headers),
		zap.String("sessionID", session.ID))

	switch frame.Command {
	case CmdConnect, CmdStomp:
		b.handleConnect(session, frame)
	case CmdSubscribe:
		if !session.Authenticated {
			b.sendError(session, "Not authenticated. Please send CONNECT first.")
			return
		}
		b.handleSubscribe(session, frame)
	case CmdUnsubscribe:
		if !session.Authenticated {
			b.sendError(session, "Not authenticated. Please send CONNECT first.")
			return
		}
		b.handleUnsubscribe(session, frame)
	case CmdSend:
		if !session.Authenticated {
			b.sendError(session, "Not authenticated. Please send CONNECT first.")
			return
		}
		b.handleSend(session, frame)
	case CmdDisconnect:
		b.handleDisconnect(session, frame)
	case CmdAck, CmdNack:
		// ACK/NACK 暂不处理
		b.logger.Debug("Received ACK/NACK (ignored)",
			zap.String("sessionID", session.ID))
	default:
		b.logger.Warn("Unknown STOMP command",
			zap.String("command", frame.Command))
	}
}

// handleConnect 处理 CONNECT 命令
func (b *Broker) handleConnect(session *Session, frame *Frame) {
	// 获取 Authorization 头（尝试多种形式）
	// GetHeader 已经是大小写不敏感的
	auth := frame.GetHeader("Authorization")
	if auth == "" {
		auth = frame.GetHeader("Authentication") // 某些客户端使用这个
	}
	if auth == "" {
		// 尝试从 login 头获取 token
		auth = frame.GetHeader("login")
		if auth != "" && !strings.HasPrefix(auth, "Bearer ") {
			auth = "Bearer " + auth
		}
	}

	b.logger.Info("CONNECT authentication attempt",
		zap.String("sessionID", session.ID),
		zap.String("authHeader", auth),
		zap.Any("allHeaders", frame.Headers))

	// 检查 Authorization 头格式
	const prefix = "Bearer "
	if auth == "" || !strings.HasPrefix(auth, prefix) {
		b.logger.Warn("Invalid Authorization header",
			zap.String("sessionID", session.ID))
		b.sendError(session, "Missing or invalid Authorization header. Use 'Authorization: Bearer <token>' or 'login: <token>'")
		return
	}

	token := auth[len(prefix):]
	if token == "" {
		b.sendError(session, "Token is empty")
		return
	}

	// 验证 Token
	if b.tokenValidator == nil {
		b.logger.Error("Token validator not set")
		b.sendError(session, "Server configuration error")
		return
	}

	username, err := b.tokenValidator(token)
	if err != nil {
		b.logger.Warn("Token validation failed",
			zap.String("sessionID", session.ID),
			zap.Error(err))
		b.sendError(session, "Token validation failed: "+err.Error())
		return
	}

	// 认证成功，更新会话信息
	session.Username = username
	session.Authenticated = true

	// 将会话添加到用户映射
	b.mu.Lock()
	if _, ok := b.users[username]; !ok {
		b.users[username] = make(map[string]*Session)
	}
	b.users[username][session.ID] = session
	b.mu.Unlock()

	b.logger.Info("Session authenticated",
		zap.String("sessionID", session.ID),
		zap.String("username", username))

	// 发送 CONNECTED 帧
	b.sendConnected(session)

	// 触发连接回调
	if b.OnConnect != nil {
		b.OnConnect(session)
	}
}

// handleSubscribe 处理 SUBSCRIBE 命令
func (b *Broker) handleSubscribe(session *Session, frame *Frame) {
	destination := frame.GetHeader(HdrDestination)
	subscriptionID := frame.GetHeader(HdrID)

	if destination == "" {
		b.sendError(session, "Missing destination header")
		return
	}
	if subscriptionID == "" {
		subscriptionID = destination // 使用 destination 作为默认 ID
	}

	session.Subscribe(subscriptionID, destination)

	b.logger.Debug("Subscribed",
		zap.String("sessionID", session.ID),
		zap.String("destination", destination),
		zap.String("subscriptionID", subscriptionID))

	// 如果请求了回执，发送 RECEIPT
	if receiptID := frame.GetHeader(HdrReceipt); receiptID != "" {
		b.sendReceipt(session, receiptID)
	}

	// 触发订阅回调（用于在订阅时推送初始数据）
	if b.OnSubscribe != nil {
		b.OnSubscribe(session, destination)
	}
}

// handleUnsubscribe 处理 UNSUBSCRIBE 命令
func (b *Broker) handleUnsubscribe(session *Session, frame *Frame) {
	subscriptionID := frame.GetHeader(HdrID)
	if subscriptionID == "" {
		b.sendError(session, "Missing id header")
		return
	}

	session.Unsubscribe(subscriptionID)

	b.logger.Debug("Unsubscribed",
		zap.String("sessionID", session.ID),
		zap.String("subscriptionID", subscriptionID))

	// 如果请求了回执，发送 RECEIPT
	if receiptID := frame.GetHeader(HdrReceipt); receiptID != "" {
		b.sendReceipt(session, receiptID)
	}
}

// handleSend 处理 SEND 命令
func (b *Broker) handleSend(session *Session, frame *Frame) {
	destination := frame.GetHeader(HdrDestination)
	if destination == "" {
		b.sendError(session, "Missing destination header")
		return
	}

	b.logger.Debug("Received SEND",
		zap.String("sessionID", session.ID),
		zap.String("destination", destination))

	// 查找处理器
	b.mu.RLock()
	handler, ok := b.handlers[destination]
	b.mu.RUnlock()

	if ok {
		handler(session, destination, frame.Body)
	} else {
		b.logger.Warn("No handler for destination",
			zap.String("destination", destination))
	}

	// 如果请求了回执，发送 RECEIPT
	if receiptID := frame.GetHeader(HdrReceipt); receiptID != "" {
		b.sendReceipt(session, receiptID)
	}
}

// handleDisconnect 处理 DISCONNECT 命令
func (b *Broker) handleDisconnect(session *Session, frame *Frame) {
	b.logger.Info("Client requested disconnect",
		zap.String("sessionID", session.ID),
		zap.String("username", session.Username))

	// 如果请求了回执，发送 RECEIPT
	if receiptID := frame.GetHeader(HdrReceipt); receiptID != "" {
		b.sendReceipt(session, receiptID)
	}
}

// sendConnected 发送 CONNECTED 帧
func (b *Broker) sendConnected(session *Session) {
	frame := NewConnectedFrame(session.ID)
	b.sendFrame(session, frame)
}

// sendError 发送 ERROR 帧
func (b *Broker) sendError(session *Session, message string) {
	frame := NewErrorFrame(message)
	b.sendFrame(session, frame)
}

// sendReceipt 发送 RECEIPT 帧
func (b *Broker) sendReceipt(session *Session, receiptID string) {
	frame := NewReceiptFrame(receiptID)
	b.sendFrame(session, frame)
}

// sendFrame 发送帧到会话
func (b *Broker) sendFrame(session *Session, frame *Frame) error {
	data := frame.Marshal()

	b.logger.Info("Sending STOMP frame",
		zap.String("command", frame.Command),
		zap.String("sessionID", session.ID),
		zap.String("data", string(data)))

	session.mu.Lock()
	defer session.mu.Unlock()

	// 设置写超时，防止慢客户端导致 goroutine 阻塞
	session.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err := session.Conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		b.logger.Error("Failed to send frame",
			zap.String("sessionID", session.ID),
			zap.Error(err))
	}
	return err
}

// nextMessageID 生成下一个消息ID
func (b *Broker) nextMessageID() string {
	id := atomic.AddUint64(&b.messageCounter, 1)
	return fmt.Sprintf("msg-%d", id)
}

// SendToSession 发送消息给指定会话
func (b *Broker) SendToSession(sessionID string, destination string, body interface{}) error {
	b.mu.RLock()
	session, ok := b.sessions[sessionID]
	b.mu.RUnlock()

	if !ok {
		return nil
	}

	return b.sendMessage(session, destination, body)
}

// SendToUser 发送消息给指定用户（所有会话）
// 对应 Java 的 /user/{username}/queue/* 模式
func (b *Broker) SendToUser(username, destination string, body interface{}) {
	sessions := b.GetUserSessions(username)
	if len(sessions) == 0 {
		b.logger.Debug("User not online",
			zap.String("username", username),
			zap.String("destination", destination))
		return
	}

	// 构造用户专属目标地址: /user/{username}{destination}
	userDestination := "/user/" + username + destination

	for _, session := range sessions {
		if err := b.sendMessage(session, userDestination, body); err != nil {
			b.logger.Error("Failed to send to user",
				zap.String("username", username),
				zap.String("sessionID", session.ID),
				zap.Error(err))
		}
	}

	b.logger.Debug("Sent to user",
		zap.String("username", username),
		zap.String("destination", userDestination))
}

// Publish 发布消息到主题（广播给所有订阅者）
// 对应 Java 的 /topic/* 模式
func (b *Broker) Publish(destination string, body interface{}) {
	b.mu.RLock()
	sessions := make([]*Session, 0)
	for _, session := range b.sessions {
		if session.Authenticated && session.IsSubscribed(destination) {
			sessions = append(sessions, session)
		}
	}
	b.mu.RUnlock()

	for _, session := range sessions {
		if err := b.sendMessage(session, destination, body); err != nil {
			b.logger.Error("Failed to publish",
				zap.String("sessionID", session.ID),
				zap.String("destination", destination),
				zap.Error(err))
		}
	}

	b.logger.Debug("Published message",
		zap.String("destination", destination),
		zap.Int("subscribers", len(sessions)))
}

// Broadcast 广播消息给所有已认证用户（不管是否订阅）
func (b *Broker) Broadcast(destination string, body interface{}) {
	b.mu.RLock()
	sessions := make([]*Session, 0, len(b.sessions))
	for _, session := range b.sessions {
		if session.Authenticated {
			sessions = append(sessions, session)
		}
	}
	b.mu.RUnlock()

	for _, session := range sessions {
		if err := b.sendMessage(session, destination, body); err != nil {
			b.logger.Error("Failed to broadcast",
				zap.String("sessionID", session.ID),
				zap.String("destination", destination),
				zap.Error(err))
		}
	}

	b.logger.Debug("Broadcast message",
		zap.String("destination", destination),
		zap.Int("sessions", len(sessions)))
}

// sendMessage 发送 MESSAGE 帧
func (b *Broker) sendMessage(session *Session, destination string, body interface{}) error {
	// 序列化 body
	var bodyBytes []byte
	var err error

	switch v := body.(type) {
	case []byte:
		bodyBytes = v
	case string:
		bodyBytes = []byte(v)
	default:
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	// 获取订阅ID
	subscriptionID := session.GetSubscriptionID(destination)

	// 创建 MESSAGE 帧
	frame := NewMessageFrame(destination, subscriptionID, b.nextMessageID(), bodyBytes)

	return b.sendFrame(session, frame)
}

// GetOnlineUsers 获取在线用户列表
func (b *Broker) GetOnlineUsers() []OnlineUser {
	b.mu.RLock()
	defer b.mu.RUnlock()

	users := make([]OnlineUser, 0, len(b.users))
	for username, sessions := range b.users {
		var earliestTime int64 = 0
		for _, session := range sessions {
			if earliestTime == 0 || session.ConnectTime < earliestTime {
				earliestTime = session.ConnectTime
			}
		}
		users = append(users, OnlineUser{
			Username:     username,
			SessionCount: len(sessions),
			LoginTime:    earliestTime,
		})
	}
	return users
}

// OnlineUser 在线用户信息
type OnlineUser struct {
	Username     string `json:"username"`
	SessionCount int    `json:"sessionCount"`
	LoginTime    int64  `json:"loginTime"`
}
