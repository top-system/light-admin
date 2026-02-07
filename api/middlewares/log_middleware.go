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
	m.handler.Engine.Use(m.Handle())
}

// 定义需要记录日志的模块和操作
var logModules = map[string]string{
	// 用户管理
	"POST:/api/v1/users":                "用户管理",
	"PUT:/api/v1/users":                 "用户管理",
	"DELETE:/api/v1/users":              "用户管理",
	"PUT:/api/v1/users/password/reset":  "用户密码重置",

	// 角色管理
	"POST:/api/v1/roles":       "角色管理",
	"PUT:/api/v1/roles":        "角色管理",
	"DELETE:/api/v1/roles":     "角色管理",
	"PUT:/api/v1/roles/menus":  "角色菜单分配",

	// 菜单管理
	"POST:/api/v1/menus":   "菜单管理",
	"PUT:/api/v1/menus":    "菜单管理",
	"DELETE:/api/v1/menus": "菜单管理",

	// 部门管理
	"POST:/api/v1/depts":   "部门管理",
	"PUT:/api/v1/depts":    "部门管理",
	"DELETE:/api/v1/depts": "部门管理",

	// 字典管理
	"POST:/api/v1/dicts":         "字典管理",
	"PUT:/api/v1/dicts":          "字典管理",
	"DELETE:/api/v1/dicts":       "字典管理",
	"POST:/api/v1/dicts/items":   "字典项管理",
	"PUT:/api/v1/dicts/items":    "字典项管理",
	"DELETE:/api/v1/dicts/items": "字典项管理",

	// 公告管理
	"POST:/api/v1/notices":         "公告管理",
	"PUT:/api/v1/notices":          "公告管理",
	"DELETE:/api/v1/notices":       "公告管理",
	"PUT:/api/v1/notices/publish":  "公告发布",
	"PUT:/api/v1/notices/revoke":   "公告撤回",
	"PUT:/api/v1/notices/read-all": "公告全部已读",

	// 配置管理
	"POST:/api/v1/configs":        "配置管理",
	"PUT:/api/v1/configs":         "配置管理",
	"DELETE:/api/v1/configs":      "配置管理",
	"PUT:/api/v1/configs/refresh": "配置刷新缓存",

	// 文件管理
	"POST:/api/v1/files":   "文件上传",
	"DELETE:/api/v1/files": "文件删除",

	// 登录登出
	"POST:/api/v1/auth/login":    "登录",
	"DELETE:/api/v1/auth/logout": "登出",
}

// 定义操作内容
var logContents = map[string]string{
	"POST":   "新增",
	"PUT":    "修改",
	"DELETE": "删除",
}

// 不需要添加操作前缀的模块（模块名已包含完整操作描述）
var selfDescribedModules = map[string]bool{
	"用户密码重置":  true,
	"角色菜单分配":  true,
	"公告发布":    true,
	"公告撤回":    true,
	"公告全部已读":  true,
	"配置刷新缓存":  true,
	"文件上传":    true,
	"文件删除":    true,
	"登录":      true,
	"登出":      true,
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

			// 读取请求体（跳过文件上传，避免将整个文件读入内存）
			var requestBody []byte
			contentType := c.Request().Header.Get("Content-Type")
			isMultipart := strings.HasPrefix(contentType, "multipart/form-data")
			if c.Request().Body != nil && !isMultipart {
				requestBody, _ = io.ReadAll(io.LimitReader(c.Request().Body, 4096))
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
			var content string
			if selfDescribedModules[module] {
				// 模块名已包含完整操作描述，直接使用
				content = module
			} else {
				// 普通模块，添加操作前缀
				content = logContents[method]
				if content == "" {
					content = method
				}
				content = content + module
			}

			// 截断过长的响应内容
			responseContent := resBody.String()
			if len(responseContent) > 4096 {
				responseContent = responseContent[:4096]
			}

			// 创建日志记录
			log := &system.Log{
				Module:          module,
				RequestMethod:   method,
				RequestParams:   string(requestBody),
				ResponseContent: responseContent,
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

// knownPathSegments 已知的路径段（不是参数值的）
var knownPathSegments = map[string]bool{
	"api": true, "v1": true,
	"users": true, "password": true, "reset": true,
	"roles": true, "menus": true,
	"depts": true,
	"dicts": true, "items": true,
	"notices": true, "publish": true, "revoke": true, "read-all": true,
	"configs": true, "refresh": true,
	"files": true,
	"auth": true, "login": true, "logout": true,
	"ws": true, "sendToAll": true, "sendToUser": true, "dict-change": true,
	"logs": true,
}

// removePathIDs 移除路径中的参数值（ID、编码等）
func removePathIDs(path string) string {
	parts := strings.Split(path, "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		// 空字符串保留（路径开头的/）
		if part == "" {
			result = append(result, part)
			continue
		}
		// 只保留已知的路径段
		if knownPathSegments[part] {
			result = append(result, part)
		}
	}
	return strings.Join(result, "/")
}
