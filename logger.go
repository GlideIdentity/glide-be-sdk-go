package glide

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message.
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger interface for custom logging implementations.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

// Field represents a structured logging field.
type Field struct {
	Key   string
	Value interface{}
}

// defaultLogger is the built-in logger implementation.
type defaultLogger struct {
	level  LogLevel
	logger *log.Logger
	json   bool
}

// NewDefaultLogger creates a logger with the specified level.
func NewDefaultLogger(level LogLevel) Logger {
	return &defaultLogger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
		json:   false,
	}
}

// NewJSONLogger creates a logger that outputs JSON.
func NewJSONLogger(level LogLevel) Logger {
	return &defaultLogger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
		json:   true,
	}
}

func (l *defaultLogger) Debug(msg string, fields ...Field) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", msg, fields...)
	}
}

func (l *defaultLogger) Info(msg string, fields ...Field) {
	if l.level <= LogLevelInfo {
		l.log("INFO", msg, fields...)
	}
}

func (l *defaultLogger) Warn(msg string, fields ...Field) {
	if l.level <= LogLevelWarn {
		l.log("WARN", msg, fields...)
	}
}

func (l *defaultLogger) Error(msg string, fields ...Field) {
	if l.level <= LogLevelError {
		l.log("ERROR", msg, fields...)
	}
}

func (l *defaultLogger) log(level, msg string, fields ...Field) {
	timestamp := time.Now().Format(time.RFC3339)

	if l.json {
		obj := map[string]interface{}{
			"timestamp": timestamp,
			"level":     level,
			"message":   msg,
		}
		for _, f := range fields {
			obj[f.Key] = sanitizeValue(f.Key, f.Value)
		}
		if data, err := json.Marshal(obj); err == nil {
			l.logger.Println(string(data))
		}
		return
	}

	// Text format
	logMsg := fmt.Sprintf("[Glide] %s [%s] %s", timestamp, level, msg)
	if len(fields) > 0 {
		parts := make([]string, 0, len(fields))
		for _, f := range fields {
			parts = append(parts, fmt.Sprintf("%s=%v", f.Key, sanitizeValue(f.Key, f.Value)))
		}
		logMsg += " " + strings.Join(parts, " ")
	}
	l.logger.Println(logMsg)
}

// sensitiveFields contains field names that should be redacted.
var sensitiveFields = []string{
	"apikey", "api_key", "api-key", "token", "password", "secret",
	"credential", "authorization", "bearer", "key",
}

// sanitizeValue redacts sensitive information.
func sanitizeValue(key string, value interface{}) interface{} {
	lowerKey := strings.ToLower(key)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(lowerKey, sensitive) {
			return "***"
		}
	}
	return value
}

// ParseLogLevel converts a string to LogLevel.
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// String returns the string representation of LogLevel.
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	default:
		return "info"
	}
}

// noopLogger discards all log messages.
type noopLogger struct{}

func (n *noopLogger) Debug(msg string, fields ...Field) {}
func (n *noopLogger) Info(msg string, fields ...Field)  {}
func (n *noopLogger) Warn(msg string, fields ...Field)  {}
func (n *noopLogger) Error(msg string, fields ...Field) {}

// NewNoopLogger returns a logger that discards all messages.
func NewNoopLogger() Logger {
	return &noopLogger{}
}
