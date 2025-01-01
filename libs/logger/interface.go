package logger

import (
	"context"
)

// LogLevel represents the different logging levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Level         LogLevel
	Message       string
	Fields        Fields
	Timestamp     int64
	CorrelationID string
	Caller        string
}

// Fields is a type alias for log fields
type Fields map[string]interface{}

// LoggerInterface defines the contract for logging
type ILogger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})

	// Context-based logging
	WithContext(ctx context.Context) ILogger
	WithFields(fields Fields) ILogger

	// Advanced logging methods
	Log(level LogLevel, msg string, args ...interface{})
}

// HookInterface allows for custom logging hooks
type IHook interface {
	Fire(entry *LogEntry) error
	Levels() []LogLevel
}

// FormatterInterface allows custom log formatting
type IFormatter interface {
	Format(entry *LogEntry) ([]byte, error)
}
