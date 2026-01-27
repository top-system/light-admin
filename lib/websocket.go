package lib

import (
	"github.com/top-system/light-admin/pkg/websocket"
)

// NewWebSocket 创建WebSocket管理器
func NewWebSocket(logger Logger) *websocket.WebSocket {
	return websocket.New(logger.DesugarZap)
}
