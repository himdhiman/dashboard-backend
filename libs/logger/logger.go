package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// Logger is the main logging implementation
type Logger struct {
	config         *Config
	contextManager *ContextManager
	slogLogger     *slog.Logger
	hooks          []HookInterface
}

// New creates a new Logger instance
func New(config *Config) *Logger {
	// Use default config if nil
	if config == nil {
		config = DefaultConfig()
	}

	// Create context manager
	contextManager := NewContextManager(config.CorrelationKey)

	// Determine handler based on format
	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	return &Logger{
		config:         config,
		contextManager: contextManager,
		slogLogger:     slog.New(handler),
		hooks:          config.Hooks,
	}
}

// WithContext adds correlation ID to the logger
func (l *Logger) WithContext(ctx context.Context) LoggerInterface {
	// Ensure context has a correlation ID
	newCtx, correlationID := l.contextManager.ExtractOrCreateCorrelationID(ctx)

	// Create a new logger with the correlation ID
	return &ContextualLogger{
		base:          l,
		ctx:           newCtx,
		correlationID: correlationID,
	}
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields Fields) LoggerInterface {
	return &FieldLogger{
		base:   l,
		fields: fields,
	}
}

// getCallerInfo retrieves the caller's file and line
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown:0"
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// log is the core logging method
func (l *Logger) log(level LogLevel, msg string, args ...interface{}) {
	// Check log level
	if level < l.config.Level {
		return
	}

	// Prepare log entry
	entry := &LogEntry{
		Level:     level,
		Message:   msg,
		Timestamp: time.Now().Unix(),
		Caller:    getCallerInfo(),
	}

	// Convert arguments to fields
	fields := extractFields(args...)
	entry.Fields = fields

	// Run hooks
	for _, hook := range l.hooks {
		if containsLevel(hook.Levels(), level) {
			hook.Fire(entry)
		}
	}

	// Convert to slog level
	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	case LevelFatal:
		slogLevel = slog.LevelError
	}

	// Log the message
	l.slogLogger.Log(context.Background(), slogLevel, msg, slogAttrFromFields(fields)...)

	// Handle fatal level
	if level == LevelFatal {
		os.Exit(1)
	}
}

// Convenience logging methods
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(LevelDebug, msg, args...)
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(LevelInfo, msg, args...)
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(LevelWarn, msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(LevelError, msg, args...)
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log(LevelFatal, msg, args...)
}

func (l *Logger) Log(level LogLevel, msg string, args ...interface{}) {
	l.log(level, msg, args...)
}

// Utility functions
func containsLevel(levels []LogLevel, level LogLevel) bool {
	for _, l := range levels {
		if l == level {
			return true
		}
	}
	return false
}

// Contextual logger wrapper
type ContextualLogger struct {
	base          *Logger
	ctx           context.Context
	correlationID string
}

func (cl *ContextualLogger) Debug(msg string, args ...interface{}) {
	cl.base.log(LevelDebug, msg, append(args, "correlation_id", cl.correlationID)...)
}

func (cl *ContextualLogger) Info(msg string, args ...interface{}) {
	cl.base.log(LevelInfo, msg, append(args, "correlation_id", cl.correlationID)...)
}

func (cl *ContextualLogger) Warn(msg string, args ...interface{}) {
	cl.base.log(LevelWarn, msg, append(args, "correlation_id", cl.correlationID)...)
}

func (cl *ContextualLogger) Error(msg string, args ...interface{}) {
	cl.base.log(LevelError, msg, append(args, "correlation_id", cl.correlationID)...)
}

func (cl *ContextualLogger) Fatal(msg string, args ...interface{}) {
	cl.base.log(LevelFatal, msg, append(args, "correlation_id", cl.correlationID)...)
}

func (cl *ContextualLogger) WithContext(ctx context.Context) LoggerInterface {
	return cl.base.WithContext(ctx)
}

func (cl *ContextualLogger) WithFields(fields Fields) LoggerInterface {
	return cl.base.WithFields(fields)
}

func (cl *ContextualLogger) Log(level LogLevel, msg string, args ...interface{}) {
	cl.base.log(level, msg, append(args, "correlation_id", cl.correlationID)...)
}

// Field logger wrapper
type FieldLogger struct {
	base   *Logger
	fields Fields
}

func (fl *FieldLogger) Debug(msg string, args ...interface{}) {
	mergedFields := mergeFields(fl.fields, args)
	fl.base.log(LevelDebug, msg, convertFieldsToArgs(mergedFields)...)
}

func (fl *FieldLogger) Info(msg string, args ...interface{}) {
	mergedFields := mergeFields(fl.fields, args)
	fl.base.log(LevelInfo, msg, convertFieldsToArgs(mergedFields)...)
}

func (fl *FieldLogger) Warn(msg string, args ...interface{}) {
	mergedFields := mergeFields(fl.fields, args)
	fl.base.log(LevelWarn, msg, convertFieldsToArgs(mergedFields)...)
}

func (fl *FieldLogger) Error(msg string, args ...interface{}) {
	mergedFields := mergeFields(fl.fields, args)
	fl.base.log(LevelError, msg, convertFieldsToArgs(mergedFields)...)
}

func (fl *FieldLogger) Fatal(msg string, args ...interface{}) {
	mergedFields := mergeFields(fl.fields, args)
	fl.base.log(LevelFatal, msg, convertFieldsToArgs(mergedFields)...)
}

func (fl *FieldLogger) WithContext(ctx context.Context) LoggerInterface {
	return fl.base.WithContext(ctx)
}

func (fl *FieldLogger) WithFields(fields Fields) LoggerInterface {
	return fl.base.WithFields(mergeFields(fl.fields, fields))
}

func (fl *FieldLogger) Log(level LogLevel, msg string, args ...interface{}) {
	mergedFields := mergeFields(fl.fields, args)
	fl.base.log(level, msg, convertFieldsToArgs(mergedFields)...)
}
