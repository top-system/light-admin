package queue

import (
	"fmt"
)

// Logger interface for queue logging
type Logger interface {
	// Debug logs debug level message
	Debug(format string, args ...interface{})
	// Info logs info level message
	Info(format string, args ...interface{})
	// Warning logs warning level message
	Warning(format string, args ...interface{})
	// Error logs error level message
	Error(format string, args ...interface{})
	// CopyWithPrefix creates a new logger with a prefix
	CopyWithPrefix(prefix string) Logger
}

// DefaultLogger is a simple logger implementation
type DefaultLogger struct {
	prefix string
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() Logger {
	return &DefaultLogger{}
}

func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	fmt.Printf("[DEBUG] %s%s\n", l.prefix, fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] %s%s\n", l.prefix, fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) Warning(format string, args ...interface{}) {
	fmt.Printf("[WARN] %s%s\n", l.prefix, fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) Error(format string, args ...interface{}) {
	fmt.Printf("[ERROR] %s%s\n", l.prefix, fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) CopyWithPrefix(prefix string) Logger {
	return &DefaultLogger{
		prefix: prefix + " ",
	}
}

// LoggerAdapter adapts any logger with basic methods to queue.Logger interface
type LoggerAdapter struct {
	debugFn   func(format string, args ...interface{})
	infoFn    func(format string, args ...interface{})
	warningFn func(format string, args ...interface{})
	errorFn   func(format string, args ...interface{})
	prefix    string
}

// NewLoggerAdapter creates a new logger adapter
func NewLoggerAdapter(
	debugFn func(format string, args ...interface{}),
	infoFn func(format string, args ...interface{}),
	warningFn func(format string, args ...interface{}),
	errorFn func(format string, args ...interface{}),
) Logger {
	return &LoggerAdapter{
		debugFn:   debugFn,
		infoFn:    infoFn,
		warningFn: warningFn,
		errorFn:   errorFn,
	}
}

func (l *LoggerAdapter) Debug(format string, args ...interface{}) {
	if l.debugFn != nil {
		l.debugFn(l.prefix+format, args...)
	}
}

func (l *LoggerAdapter) Info(format string, args ...interface{}) {
	if l.infoFn != nil {
		l.infoFn(l.prefix+format, args...)
	}
}

func (l *LoggerAdapter) Warning(format string, args ...interface{}) {
	if l.warningFn != nil {
		l.warningFn(l.prefix+format, args...)
	}
}

func (l *LoggerAdapter) Error(format string, args ...interface{}) {
	if l.errorFn != nil {
		l.errorFn(l.prefix+format, args...)
	}
}

func (l *LoggerAdapter) CopyWithPrefix(prefix string) Logger {
	return &LoggerAdapter{
		debugFn:   l.debugFn,
		infoFn:    l.infoFn,
		warningFn: l.warningFn,
		errorFn:   l.errorFn,
		prefix:    prefix + " ",
	}
}

// ContextKey type for context keys
type ContextKey struct{}

// LoggerCtx is the context key for logger
type LoggerCtx struct{}

// CorrelationIDCtx is the context key for correlation ID
type CorrelationIDCtx struct{}

// UserCtx is the context key for user
type UserCtx struct{}
