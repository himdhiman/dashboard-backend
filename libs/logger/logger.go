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
	ILogger
	config         *Config
	contextManager *ContextManager
	slogLogger     *slog.Logger
	hooks          []IHook
}

// New creates a new Logger instance
func New(config *Config) ILogger {
	// Use default config if nil
	if config == nil {
		config = DefaultConfig("DefaultService")
	}

	// Create context manager
	contextManager := NewContextManager(config.ServiceName)

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

// Add WithError method to Logger struct
func (l *Logger) WithError(err error) ILogger {
	return l.WithFields(Fields{
		"error": err.Error(),
	})
}

// WithContext adds correlation ID to the logger
func (l *Logger) WithContext(ctx context.Context) ILogger {
	newCtx := context.Background()

	// Create a new logger
	return &ContextualLogger{
		base: l,
		ctx:  newCtx,
	}
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields Fields) ILogger {
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

	// Extract correlation ID
	fields := extractFields(args...)

	// Extract correlation ID
	correlationID := ""
	if id, ok := fields["correlation_id"]; ok {
		correlationID = id.(string)
	}

	// Prepare log entry
	entry := &LogEntry{
		Level:         level,
		Message:       msg,
		CorrelationID: correlationID,
		Timestamp:     time.Now().Unix(),
		Caller:        getCallerInfo(),
	}

	entry.Fields = fields
	// Convert arguments to fields

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

func (l *Logger) WithField(key string, value interface{}) ILogger {
	return l.WithFields(Fields{
		key: value,
	})
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
	base *Logger
	ctx  context.Context
}

func (cl *ContextualLogger) Debug(msg string, args ...interface{}) {
	cl.base.log(LevelDebug, msg, args...)
}

func (cl *ContextualLogger) Info(msg string, args ...interface{}) {
	cl.base.log(LevelInfo, msg, args...)
}

func (cl *ContextualLogger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	cl.Info(msg)
}

func (cl *ContextualLogger) Warn(msg string, args ...interface{}) {
	cl.base.log(LevelWarn, msg, args...)
}

func (cl *ContextualLogger) Error(msg string, args ...interface{}) {
	cl.base.log(LevelError, msg, args...)
}

func (cl *ContextualLogger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	cl.Error(msg)
}

func (cl *ContextualLogger) Fatal(msg string, args ...interface{}) {
	cl.base.log(LevelFatal, msg, args...)
}

func (cl *ContextualLogger) WithContext(ctx context.Context) ILogger {
	return cl.base.WithContext(ctx)
}

func (l *ContextualLogger) WithField(key string, value interface{}) ILogger {
	return l.WithFields(Fields{
		key: value,
	})
}

func (cl *ContextualLogger) WithFields(fields Fields) ILogger {
	return cl.base.WithFields(fields)
}
func (l *ContextualLogger) WithError(err error) ILogger {
	return l.WithFields(Fields{
		"error": err.Error(),
	})
}

func (cl *ContextualLogger) Log(level LogLevel, msg string, args ...interface{}) {
	cl.base.log(level, msg, args...)
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

func (fl *FieldLogger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fl.Info(msg)
}

func (fl *FieldLogger) Warn(msg string, args ...interface{}) {
	mergedFields := mergeFields(fl.fields, args)
	fl.base.log(LevelWarn, msg, convertFieldsToArgs(mergedFields)...)
}

func (fl *FieldLogger) Error(msg string, args ...interface{}) {
	mergedFields := mergeFields(fl.fields, args)
	fl.base.log(LevelError, msg, convertFieldsToArgs(mergedFields)...)
}

func (fl *FieldLogger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fl.Error(msg)
}

func (fl *FieldLogger) Fatal(msg string, args ...interface{}) {
	mergedFields := mergeFields(fl.fields, args)
	fl.base.log(LevelFatal, msg, convertFieldsToArgs(mergedFields)...)
}

func (fl *FieldLogger) WithContext(ctx context.Context) ILogger {
	return fl.base.WithContext(ctx)
}

func (l *FieldLogger) WithField(key string, value interface{}) ILogger {
	return l.WithFields(Fields{
		key: value,
	})
}

func (fl *FieldLogger) WithFields(fields Fields) ILogger {
	return fl.base.WithFields(mergeFields(fl.fields, fields))
}

func (l *FieldLogger) WithError(err error) ILogger {
	return l.WithFields(Fields{
		"error": err.Error(),
	})
}

func (fl *FieldLogger) Log(level LogLevel, msg string, args ...interface{}) {
	mergedFields := mergeFields(fl.fields, args)
	fl.base.log(level, msg, convertFieldsToArgs(mergedFields)...)
}
