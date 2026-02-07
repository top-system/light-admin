package route

import (
	"github.com/top-system/light-admin/api/platform/controller"
	"github.com/top-system/light-admin/lib"
)

// WebSocketRoute WebSocket路由
type WebSocketRoute struct {
	logger              lib.Logger
	handler             lib.HttpHandler
	websocketController controller.WebSocketController
}

// NewWebSocketRoute 创建WebSocket路由
func NewWebSocketRoute(
	logger lib.Logger,
	handler lib.HttpHandler,
	websocketController controller.WebSocketController,
) WebSocketRoute {
	return WebSocketRoute{
		logger:              logger,
		handler:             handler,
		websocketController: websocketController,
	}
}

// SetupWebSocket 设置WebSocket连接端点（需要在中间件之前调用）
func (r WebSocketRoute) SetupWebSocket() {
	// WebSocket 连接端点 (与原Java项目一致: /ws)
	// 在中间件之前注册，避免中间件干扰 WebSocket 连接
	r.handler.Engine.GET("/ws", r.websocketController.Connect)
}

// Setup 设置WebSocket HTTP API路由
func (r WebSocketRoute) Setup() {
	// HTTP API 接口 (用于管理和测试)
	api := r.handler.RouterV1.Group("/websocket")
	{
		// 广播发送消息
		api.POST("/sendToAll", r.websocketController.SendToAll)

		// 点对点发送消息
		api.POST("/sendToUser", r.websocketController.SendToUser)

		// 获取在线用户列表
		api.GET("/online-users", r.websocketController.GetOnlineUsers)

		// 获取在线用户数量
		api.GET("/online-count", r.websocketController.GetOnlineCount)

		// 广播字典变更
		api.POST("/dict-change", r.websocketController.BroadcastDictChange)
	}
}
