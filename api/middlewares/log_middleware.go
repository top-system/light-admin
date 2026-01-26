package middlewares

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/models/dto"
	"github.com/labstack/echo/v4"
	"github.com/mssola/useragent"
)

// LogMiddleware 日志中间件
type LogMiddleware struct {
	handler    lib.HttpHandler
	logger     lib.Logger
	logService service.LogService
}

// NewLogMiddleware creates new log middleware
func NewLogMiddleware(
	handler lib.HttpHandler,
	logger lib.Logger,
	logService service.LogService,
) LogMiddleware {
	return LogMiddleware{
		handler:    handler,
		logger:     logger,
		logService: logService,
	}
}

// Setup sets up the log middleware
func (m LogMiddleware) Setup() {
	m.logger.Zap.Info("Setting up log middleware")
	m.handler.Engine.Use(m.Handle())
}

// 定义需要记录日志的模块和操作
var logModules = map[string]string{
	"POST:/api/v1/users":             "用户管理",
	"PUT:/api/v1/users":              "用户管理",
	"DELETE:/api/v1/users":           "用户管理",
	"POST:/api/v1/roles":             "角色管理",
	"PUT:/api/v1/roles":              "角色管理",
	"DELETE:/api/v1/roles":           "角色管理",
	"POST:/api/v1/menus":             "菜单管理",
	"PUT:/api/v1/menus":              "菜单管理",
	"DELETE:/api/v1/menus":           "菜单管理",
	"POST:/api/v1/depts":             "部门管理",
	"PUT:/api/v1/depts":              "部门管理",
	"DELETE:/api/v1/depts":           "部门管理",
	"POST:/api/v1/dicts":             "字典管理",
	"PUT:/api/v1/dicts":              "字典管理",
	"DELETE:/api/v1/dicts":           "字典管理",
	"POST:/api/v1/notices":           "公告管理",
	"PUT:/api/v1/notices":            "公告管理",
	"DELETE:/api/v1/notices":         "公告管理",
	"POST:/api/v1/configs":           "配置管理",
	"PUT:/api/v1/configs":            "配置管理",
	"DELETE:/api/v1/configs":         "配置管理",
	"POST:/api/v1/auth/login":        "登录",
	"DELETE:/api/v1/auth/logout":     "登出",
}

// 定义操作内容
var logContents = map[string]string{
	"POST":   "新增",
	"PUT":    "修改",
	"DELETE": "删除",
}

// responseWriter 包装响应写入器以捕获响应内容
type responseWriter struct {
	io.Writer
	echo.Response
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Handle 处理日志记录
func (m LogMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 只记录 POST, PUT, DELETE 请求
			method := c.Request().Method
			if method != "POST" && method != "PUT" && method != "DELETE" {
				return next(c)
			}

			// 检查是否需要记录日志
			path := c.Request().URL.Path
			// 移除路径中的数字ID，用于匹配
			basePath := removePathIDs(path)
			key := method + ":" + basePath
			module, ok := logModules[key]
			if !ok {
				return next(c)
			}

			startTime := time.Now()

			// 读取请求体
			var requestBody []byte
			if c.Request().Body != nil {
				requestBody, _ = io.ReadAll(c.Request().Body)
				c.Request().Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			// 捕获响应
			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(c.Response().Writer, resBody)
			writer := &responseWriter{Writer: mw, Response: *c.Response()}
			c.Response().Writer = writer

			// 执行下一个处理器
			err := next(c)

			// 计算执行时间
			executionTime := time.Since(startTime).Milliseconds()

			// 获取用户信息
			var createBy uint64
			if claims, ok := c.Get(constants.CurrentUser).(*dto.JwtClaims); ok && claims != nil {
				createBy = claims.ID
			}

			// 解析 User-Agent
			ua := useragent.New(c.Request().UserAgent())
			browserName, browserVersion := ua.Browser()
			osInfo := ua.OS()

			// 获取操作内容
			content := logContents[method]
			if content == "" {
				content = method
			}
			content = content + module

			// 创建日志记录
			log := &system.Log{
				Module:          module,
				RequestMethod:   method,
				RequestParams:   string(requestBody),
				ResponseContent: resBody.String(),
				Content:         content,
				RequestURI:      path,
				Method:          c.Path(),
				IP:              c.RealIP(),
				ExecutionTime:   executionTime,
				Browser:         browserName,
				BrowserVersion:  browserVersion,
				OS:              osInfo,
				CreateBy:        createBy,
			}

			// 异步保存日志
			go func() {
				if saveErr := m.logService.Create(log); saveErr != nil {
					m.logger.Zap.Errorf("Failed to save log: %v", saveErr)
				}
			}()

			return err
		}
	}
}

// removePathIDs 移除路径中的数字ID
func removePathIDs(path string) string {
	parts := strings.Split(path, "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		// 如果是纯数字，跳过
		if isNumeric(part) {
			continue
		}
		result = append(result, part)
	}
	return strings.Join(result, "/")
}

// isNumeric 检查字符串是否为纯数字
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
