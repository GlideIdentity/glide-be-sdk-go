package glide

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// LogLevelSilent disables all logging
	LogLevelSilent LogLevel = iota
	// LogLevelError logs only errors
	LogLevelError
	// LogLevelWarn logs warnings and above
	LogLevelWarn
	// LogLevelInfo logs info messages and above
	LogLevelInfo
	// LogLevelDebug logs all messages including debug
	LogLevelDebug
)

// Logger interface allows for custom logging implementations
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// defaultLogger is the built-in logger implementation
type defaultLogger struct {
	level      LogLevel
	logger     *log.Logger
	timeFormat string
}

// NewDefaultLogger creates a new default logger with the specified level
func NewDefaultLogger(level LogLevel) Logger {
	return &defaultLogger{
		level:      level,
		logger:     log.New(os.Stdout, "[Glide] ", 0),
		timeFormat: time.RFC3339,
	}
}

// Debug logs a debug message
func (l *defaultLogger) Debug(msg string, fields ...Field) {
	if l.level >= LogLevelDebug {
		l.log("DEBUG", msg, fields...)
	}
}

// Info logs an info message
func (l *defaultLogger) Info(msg string, fields ...Field) {
	if l.level >= LogLevelInfo {
		l.log("INFO", msg, fields...)
	}
}

// Warn logs a warning message
func (l *defaultLogger) Warn(msg string, fields ...Field) {
	if l.level >= LogLevelWarn {
		l.log("WARN", msg, fields...)
	}
}

// Error logs an error message
func (l *defaultLogger) Error(msg string, fields ...Field) {
	if l.level >= LogLevelError {
		l.log("ERROR", msg, fields...)
	}
}

// log formats and outputs a log message
func (l *defaultLogger) log(level, msg string, fields ...Field) {
	// Build the log message
	timestamp := time.Now().Format(l.timeFormat)
	logMsg := fmt.Sprintf("%s [%s] %s", timestamp, level, msg)

	// Add fields if present
	if len(fields) > 0 {
		fieldStrs := make([]string, 0, len(fields))
		for _, f := range fields {
			// Sanitize sensitive data
			value := sanitizeValue(f.Key, f.Value)
			fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", f.Key, value))
		}
		logMsg += " " + strings.Join(fieldStrs, " ")
	}

	l.logger.Println(logMsg)
}

// sanitizeValue redacts sensitive information from log values
func sanitizeValue(key string, value interface{}) interface{} {
	// Convert to string for pattern matching
	strValue := fmt.Sprintf("%v", value)

	// List of sensitive field names (case-insensitive)
	sensitiveFields := []string{
		"apikey", "api_key", "apiKey",
		"token", "accesstoken", "access_token",
		"password", "passwd", "pwd",
		"secret", "credential",
		"authorization", "auth",
	}

	// Check if field name contains sensitive keywords
	lowerKey := strings.ToLower(key)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(lowerKey, sensitive) {
			// Redact but show first 4 chars for debugging
			if len(strValue) > 4 {
				return strValue[:4] + "****[REDACTED]"
			}
			return "****[REDACTED]"
		}
	}

	// Phone number pattern - show area code only
	phonePattern := regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)
	if phonePattern.MatchString(strValue) {
		if len(strValue) > 6 {
			return strValue[:6] + "****"
		}
		return "****[PHONE]"
	}

	// Email pattern - show domain only
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if emailPattern.MatchString(strValue) {
		parts := strings.Split(strValue, "@")
		if len(parts) == 2 {
			return "****@" + parts[1]
		}
	}

	// URL with potential credentials
	if strings.Contains(strValue, "://") && strings.Contains(strValue, "@") {
		// Redact credentials in URLs
		urlPattern := regexp.MustCompile(`(https?://)([^:]+:[^@]+)@`)
		return urlPattern.ReplaceAllString(strValue, "${1}****:****@")
	}

	return value
}

// ParseLogLevel converts a string to a LogLevel
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
	case "silent", "none", "off":
		return LogLevelSilent
	default:
		return LogLevelSilent
	}
}

// String returns the string representation of a LogLevel
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
	case LogLevelSilent:
		return "silent"
	default:
		return "unknown"
	}
}

// noopLogger is a logger that does nothing (for when logging is disabled)
type noopLogger struct{}

func (n *noopLogger) Debug(msg string, fields ...Field) {}
func (n *noopLogger) Info(msg string, fields ...Field)  {}
func (n *noopLogger) Warn(msg string, fields ...Field)  {}
func (n *noopLogger) Error(msg string, fields ...Field) {}

// NewNoopLogger returns a logger that does nothing
func NewNoopLogger() Logger {
	return &noopLogger{}
}
