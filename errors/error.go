package errors

import (
	"net/http"

	"github.com/pkg/errors"
)

// Alias
// noinspection GoUnusedGlobalVariable
var (
	Is           = errors.Is
	As           = errors.As
	New          = errors.New
	Unwrap       = errors.Unwrap
	Wrap         = errors.Wrap
	Wrapf        = errors.Wrapf
	WithStack    = errors.WithStack
	WithMessage  = errors.WithMessage
	WithMessagef = errors.WithMessagef
)

// Database
var (
	DatabaseInternalError  = errors.New("database internal error")
	DatabaseRecordNotFound = errors.New("database record not found")
)

// Redis
var (
	RedisKeyNoExist = errors.New("redis key does not exist")
)

// Captcha
var (
	CaptchaAnswerCodeNoMatch = errors.New("captcha answer code no match")
)

// Auth
var (
	AuthTokenInvalid      = errors.New("auth token is invalid")
	AuthTokenExpired      = errors.New("auth token is expired")
	AuthTokenNotValidYet  = errors.New("auth token not active yet")
	AuthTokenMalformed    = errors.New("auth token is malformed")
	AuthTokenGenerateFail = errors.New("failed to generate auth token")
)

// errorHTTPStatus 错误到 HTTP 状态码的映射表
var errorHTTPStatus = []struct {
	err    error
	status int
}{
	// 500 Internal Server Error
	{DatabaseInternalError, http.StatusInternalServerError},
	{AuthTokenGenerateFail, http.StatusInternalServerError},

	// 404 Not Found
	{DatabaseRecordNotFound, http.StatusNotFound},

	// 401 Unauthorized
	{AuthTokenInvalid, http.StatusUnauthorized},
	{AuthTokenExpired, http.StatusUnauthorized},
	{AuthTokenNotValidYet, http.StatusUnauthorized},
	{AuthTokenMalformed, http.StatusUnauthorized},
}

// RegisterHTTPStatus 注册错误到 HTTP 状态码的映射（供各模块 init 时调用）
func RegisterHTTPStatus(err error, status int) {
	errorHTTPStatus = append(errorHTTPStatus, struct {
		err    error
		status int
	}{err, status})
}

// HTTPStatusCode 根据错误类型返回对应的 HTTP 状态码，未匹配返回 0
func HTTPStatusCode(err error) int {
	for _, mapping := range errorHTTPStatus {
		if Is(err, mapping.err) {
			return mapping.status
		}
	}
	return 0
}
