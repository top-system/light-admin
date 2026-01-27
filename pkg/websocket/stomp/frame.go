package stomp

import (
	"bytes"
	"fmt"
	"strings"
)

// STOMP 命令常量
const (
	// 客户端命令
	CmdConnect     = "CONNECT"
	CmdStomp       = "STOMP"
	CmdSubscribe   = "SUBSCRIBE"
	CmdUnsubscribe = "UNSUBSCRIBE"
	CmdSend        = "SEND"
	CmdDisconnect  = "DISCONNECT"
	CmdAck         = "ACK"
	CmdNack        = "NACK"
	CmdBegin       = "BEGIN"
	CmdCommit      = "COMMIT"
	CmdAbort       = "ABORT"

	// 服务端命令
	CmdConnected = "CONNECTED"
	CmdMessage   = "MESSAGE"
	CmdReceipt   = "RECEIPT"
	CmdError     = "ERROR"
)

// STOMP 头常量
const (
	HdrAcceptVersion = "accept-version"
	HdrVersion       = "version"
	HdrHost          = "host"
	HdrLogin         = "login"
	HdrPasscode      = "passcode"
	HdrHeartBeat     = "heart-beat"
	HdrSession       = "session"
	HdrServer        = "server"
	HdrDestination   = "destination"
	HdrID            = "id"
	HdrAck           = "ack"
	HdrSubscription  = "subscription"
	HdrMessageID     = "message-id"
	HdrContentType   = "content-type"
	HdrContentLength = "content-length"
	HdrReceipt       = "receipt"
	HdrReceiptID     = "receipt-id"
	HdrMessage       = "message"
	HdrAuthorization = "Authorization"
)

// NULL 字符，用于标记帧结束
const NULL = '\x00'

// Frame STOMP 帧
type Frame struct {
	Command string
	Headers map[string]string
	Body    []byte
}

// NewFrame 创建新帧
func NewFrame(command string) *Frame {
	return &Frame{
		Command: command,
		Headers: make(map[string]string),
	}
}

// SetHeader 设置头部
func (f *Frame) SetHeader(key, value string) *Frame {
	f.Headers[key] = value
	return f
}

// GetHeader 获取头部（大小写不敏感）
func (f *Frame) GetHeader(key string) string {
	// 先尝试精确匹配
	if v, ok := f.Headers[key]; ok {
		return v
	}
	// 再尝试大小写不敏感匹配
	keyLower := strings.ToLower(key)
	for k, v := range f.Headers {
		if strings.ToLower(k) == keyLower {
			return v
		}
	}
	return ""
}

// SetBody 设置消息体
func (f *Frame) SetBody(body []byte) *Frame {
	f.Body = body
	return f
}

// SetBodyString 设置字符串消息体
func (f *Frame) SetBodyString(body string) *Frame {
	f.Body = []byte(body)
	return f
}

// Marshal 序列化帧为 STOMP 格式
func (f *Frame) Marshal() []byte {
	var buf bytes.Buffer

	// 命令
	buf.WriteString(f.Command)
	buf.WriteByte('\n')

	// 头部
	for key, value := range f.Headers {
		buf.WriteString(key)
		buf.WriteByte(':')
		buf.WriteString(encodeHeaderValue(value))
		buf.WriteByte('\n')
	}

	// 如果有 body，添加 content-length
	if len(f.Body) > 0 {
		if _, ok := f.Headers[HdrContentLength]; !ok {
			buf.WriteString(fmt.Sprintf("%s:%d\n", HdrContentLength, len(f.Body)))
		}
	}

	// 空行分隔头部和 body
	buf.WriteByte('\n')

	// Body
	if len(f.Body) > 0 {
		buf.Write(f.Body)
	}

	// NULL 字符结束
	buf.WriteByte(NULL)

	return buf.Bytes()
}

// ParseFrame 解析 STOMP 帧
func ParseFrame(data []byte) (*Frame, error) {
	// 移除末尾的 NULL 字符
	data = bytes.TrimSuffix(data, []byte{NULL})

	// 查找空行（分隔头部和 body）
	headerEnd := bytes.Index(data, []byte("\n\n"))
	if headerEnd == -1 {
		headerEnd = bytes.Index(data, []byte("\r\n\r\n"))
	}

	var headerPart, bodyPart []byte
	if headerEnd == -1 {
		// 没有 body
		headerPart = data
		bodyPart = nil
	} else {
		headerPart = data[:headerEnd]
		bodyPart = data[headerEnd+2:] // 跳过 \n\n
		if len(bodyPart) > 0 && bodyPart[0] == '\n' {
			bodyPart = bodyPart[1:] // 处理 \r\n\r\n 的情况
		}
	}

	// 解析头部
	lines := bytes.Split(headerPart, []byte("\n"))
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty frame")
	}

	// 第一行是命令
	command := strings.TrimSpace(string(lines[0]))
	command = strings.TrimSuffix(command, "\r") // 确保移除 \r
	if command == "" {
		return nil, fmt.Errorf("empty command")
	}

	frame := NewFrame(command)

	// 解析头部行
	for i := 1; i < len(lines); i++ {
		line := string(lines[i])
		line = strings.TrimSuffix(line, "\r") // 处理 \r\n

		if line == "" {
			continue
		}

		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		key := line[:colonIdx]
		value := decodeHeaderValue(line[colonIdx+1:])
		frame.Headers[key] = value
	}

	// 设置 body
	if len(bodyPart) > 0 {
		frame.Body = bodyPart
	}

	return frame, nil
}

// encodeHeaderValue 编码头部值（转义特殊字符）
func encodeHeaderValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, ":", "\\c")
	return value
}

// decodeHeaderValue 解码头部值
func decodeHeaderValue(value string) string {
	value = strings.ReplaceAll(value, "\\r", "\r")
	value = strings.ReplaceAll(value, "\\n", "\n")
	value = strings.ReplaceAll(value, "\\c", ":")
	value = strings.ReplaceAll(value, "\\\\", "\\")
	return value
}

// 创建常用帧的辅助函数

// NewConnectedFrame 创建 CONNECTED 帧
func NewConnectedFrame(sessionID string) *Frame {
	return NewFrame(CmdConnected).
		SetHeader(HdrVersion, "1.2").
		SetHeader(HdrSession, sessionID).
		SetHeader(HdrServer, "echo-admin/1.0").
		SetHeader(HdrHeartBeat, "0,0")
}

// NewMessageFrame 创建 MESSAGE 帧
func NewMessageFrame(destination, subscriptionID, messageID string, body []byte) *Frame {
	frame := NewFrame(CmdMessage).
		SetHeader(HdrDestination, destination).
		SetHeader(HdrMessageID, messageID).
		SetHeader(HdrContentType, "application/json")

	if subscriptionID != "" {
		frame.SetHeader(HdrSubscription, subscriptionID)
	}

	if len(body) > 0 {
		frame.SetBody(body)
	}

	return frame
}

// NewErrorFrame 创建 ERROR 帧
func NewErrorFrame(message string) *Frame {
	return NewFrame(CmdError).
		SetHeader(HdrMessage, message).
		SetHeader(HdrContentType, "text/plain").
		SetBodyString(message)
}

// NewReceiptFrame 创建 RECEIPT 帧
func NewReceiptFrame(receiptID string) *Frame {
	return NewFrame(CmdReceipt).
		SetHeader(HdrReceiptID, receiptID)
}
