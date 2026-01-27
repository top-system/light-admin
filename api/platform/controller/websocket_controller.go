package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/pkg/echox"
	ws "github.com/top-system/light-admin/pkg/websocket"
	"github.com/top-system/light-admin/pkg/websocket/stomp"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该限制
	},
	// 支持 STOMP 子协议
	Subprotocols: []string{"v12.stomp", "v11.stomp", "v10.stomp"},
}

// WebSocketController WebSocket控制器
type WebSocketController struct {
	ws          *ws.WebSocket
	logger      lib.Logger
	authService service.AuthService
}

// NewWebSocketController 创建WebSocket控制器
func NewWebSocketController(
	websocket *ws.WebSocket,
	logger lib.Logger,
	authService service.AuthService,
) WebSocketController {
	ctrl := WebSocketController{
		ws:          websocket,
		logger:      logger,
		authService: authService,
	}

	// 设置 Token 验证器 (用于 STOMP CONNECT 认证)
	ctrl.ws.Broker.SetTokenValidator(func(token string) (string, error) {
		claims, err := authService.ParseToken(token)
		if err != nil {
			return "", err
		}
		return claims.Username, nil
	})

	// 注册消息处理器 (对应 Java @MessageMapping)
	ctrl.registerHandlers()

	return ctrl
}

// registerHandlers 注册 STOMP 消息处理器
// 对应 Java 控制器中的 @MessageMapping 注解
func (c WebSocketController) registerHandlers() {
	// @MessageMapping("/sendToAll")
	// @SendTo("/topic/notice")
	c.ws.RegisterHandler(ws.AppSendToAll, func(session *stomp.Session, destination string, body []byte) {
		var message string
		if err := json.Unmarshal(body, &message); err != nil {
			c.logger.Zap.Errorf("Failed to unmarshal message: %v", err)
			return
		}

		// 广播到 /topic/notice
		c.ws.BroadcastNotice(message)
		c.logger.Zap.Infof("Broadcast message from %s: %s", session.Username, message)
	})

	// @MessageMapping("/sendToUser/{username}")
	c.ws.RegisterHandler(ws.AppSendToUser, func(session *stomp.Session, destination string, body []byte) {
		var req struct {
			Username string `json:"username"`
			Message  string `json:"message"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			c.logger.Zap.Errorf("Failed to unmarshal message: %v", err)
			return
		}

		sender := session.Username
		c.logger.Zap.Infof("Sender: %s, Receiver: %s", sender, req.Username)

		// 发送到 /user/{username}/queue/greeting
		c.ws.SendToUser(sender, req.Username, req.Message)
	})
}

// HandleWebSocket 处理原始 HTTP WebSocket 请求（绕过 Echo 中间件）
func (c WebSocketController) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	c.logger.Zap.Infof("WebSocket upgrade request from: %s (bypassing middleware)", r.RemoteAddr)

	// 升级为WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.logger.Zap.Errorf("Failed to upgrade to websocket: %v", err)
		return
	}

	c.logger.Zap.Infof("WebSocket upgrade successful")

	sessionID := uuid.New().String()

	// 创建会话（未认证状态，Username 为空）
	session := &stomp.Session{
		ID:            sessionID,
		Username:      "", // 认证成功后会设置
		Conn:          conn,
		Subscriptions: make(map[string]string),
		ConnectTime:   time.Now().UnixMilli(),
		Authenticated: false,
	}

	// 注册会话
	c.ws.Broker.AddSession(session)

	c.logger.Zap.Infof("WebSocket connected (pending auth): session=%s", sessionID)

	// 直接处理消息
	c.handleMessages(session)
}

// Connect WebSocket连接端点 (Echo handler，保留用于兼容)
// WebSocket 握手后，客户端需要发送 STOMP CONNECT 帧进行认证
// @tags WebSocket
// @summary WebSocket Connect
// @produce json
// @success 101 "Switching Protocols"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /ws [get]
func (c WebSocketController) Connect(ctx echo.Context) error {
	c.logger.Zap.Infof("WebSocket upgrade request from: %s", ctx.RealIP())

	// 升级为WebSocket连接（此时还未认证，认证在 STOMP CONNECT 帧中处理）
	conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		c.logger.Zap.Errorf("Failed to upgrade to websocket: %v", err)
		return nil // 升级失败时不能返回HTTP响应
	}

	c.logger.Zap.Infof("WebSocket upgrade successful")

	sessionID := uuid.New().String()

	// 创建会话（未认证状态，Username 为空）
	session := &stomp.Session{
		ID:            sessionID,
		Username:      "", // 认证成功后会设置
		Conn:          conn,
		Subscriptions: make(map[string]string),
		ConnectTime:   time.Now().UnixMilli(),
		Authenticated: false,
	}

	// 注册会话
	c.ws.Broker.AddSession(session)

	c.logger.Zap.Infof("WebSocket connected (pending auth): session=%s", sessionID)

	// 直接在当前处理函数中处理消息（不使用 goroutine）
	// 这样可以保持连接活跃，直到客户端断开
	c.handleMessages(session)

	return nil
}

// handleMessages 处理WebSocket消息
func (c WebSocketController) handleMessages(session *stomp.Session) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Zap.Errorf("Panic in handleMessages: %v", r)
		}
		c.ws.Broker.RemoveSession(session.ID)
		session.Conn.Close()
		c.logger.Zap.Infof("WebSocket disconnected: user=%s, session=%s", session.Username, session.ID)
	}()

	c.logger.Zap.Infof("Starting message loop for session=%s", session.ID)

	for {
		c.logger.Zap.Infof("Waiting for message on session=%s", session.ID)
		messageType, message, err := session.Conn.ReadMessage()
		if err != nil {
			// 记录所有错误，不仅仅是 UnexpectedCloseError
			c.logger.Zap.Errorf("WebSocket ReadMessage error: %v, session=%s", err, session.ID)
			break
		}

		c.logger.Zap.Infof("Received WebSocket message: type=%d, length=%d", messageType, len(message))

		// 处理 TextMessage 和 BinaryMessage
		if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
			c.ws.Broker.HandleMessage(session, message)
		}
	}
}

// === HTTP API 接口 (用于服务端主动推送) ===

// SendToAllRequest 广播发送消息请求
type SendToAllRequest struct {
	Message string `json:"message" validate:"required"`
}

// SendToAll 广播发送消息 (HTTP API)
// @tags WebSocket
// @summary Send message to all users
// @accept json
// @produce json
// @param body body SendToAllRequest true "Message"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/websocket/sendToAll [post]
func (c WebSocketController) SendToAll(ctx echo.Context) error {
	var req SendToAllRequest
	if err := ctx.Bind(&req); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	c.ws.BroadcastNotice(req.Message)

	return echox.Response{Code: http.StatusOK, Message: "Message sent to all users"}.JSON(ctx)
}

// SendToUserRequest 点对点发送消息请求
type SendToUserRequest struct {
	Username string `json:"username" validate:"required"`
	Message  string `json:"message" validate:"required"`
}

// SendToUser 点对点发送消息 (HTTP API)
// @tags WebSocket
// @summary Send message to specific user
// @accept json
// @produce json
// @param body body SendToUserRequest true "Message"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/websocket/sendToUser [post]
func (c WebSocketController) SendToUser(ctx echo.Context) error {
	var req SendToUserRequest
	if err := ctx.Bind(&req); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	// 获取发送人（从认证中间件获取）
	senderName := "System"
	if sender := ctx.Get(constants.CurrentUser); sender != nil {
		if claims, ok := sender.(string); ok {
			senderName = claims
		}
	}

	c.logger.Zap.Infof("Sender: %s, Receiver: %s", senderName, req.Username)

	c.ws.SendToUser(senderName, req.Username, req.Message)

	return echox.Response{Code: http.StatusOK, Message: "Message sent to user"}.JSON(ctx)
}

// GetOnlineUsers 获取在线用户列表 (HTTP API)
// @tags WebSocket
// @summary Get online users
// @produce json
// @success 200 {object} echox.Response{data=[]stomp.OnlineUser} "ok"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/websocket/online-users [get]
func (c WebSocketController) GetOnlineUsers(ctx echo.Context) error {
	users := c.ws.GetOnlineUsers()
	return echox.Response{Code: http.StatusOK, Data: users}.JSON(ctx)
}

// GetOnlineCount 获取在线用户数量 (HTTP API)
// @tags WebSocket
// @summary Get online user count
// @produce json
// @success 200 {object} echox.Response{data=int} "ok"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/websocket/online-count [get]
func (c WebSocketController) GetOnlineCount(ctx echo.Context) error {
	count := c.ws.GetOnlineUserCount()
	return echox.Response{Code: http.StatusOK, Data: count}.JSON(ctx)
}

// BroadcastDictChangeRequest 广播字典变更请求
type BroadcastDictChangeRequest struct {
	DictCode string `json:"dictCode" validate:"required"`
}

// BroadcastDictChange 广播字典变更 (HTTP API)
// @tags WebSocket
// @summary Broadcast dict change
// @accept json
// @produce json
// @param body body BroadcastDictChangeRequest true "Dict code"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/websocket/dict-change [post]
func (c WebSocketController) BroadcastDictChange(ctx echo.Context) error {
	var req BroadcastDictChangeRequest
	if err := ctx.Bind(&req); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	c.ws.BroadcastDictChange(req.DictCode)

	return echox.Response{Code: http.StatusOK, Message: "Dict change notification sent"}.JSON(ctx)
}
