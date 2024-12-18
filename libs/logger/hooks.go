package logger

import (
	"fmt"
	"sync"
)

// BaseHook provides a base implementation for logging hooks
type BaseHook struct {
	levels []LogLevel
	mu     sync.Mutex
}

// NewBaseHook creates a new base hook
func NewBaseHook(levels []LogLevel) *BaseHook {
	return &BaseHook{
		levels: levels,
	}
}

// Levels returns the log levels for this hook
func (h *BaseHook) Levels() []LogLevel {
	return h.levels
}

// MetricsHook provides a hook for tracking log metrics
type MetricsHook struct {
	BaseHook
	logCounts map[LogLevel]int
}

// NewMetricsHook creates a new metrics hook
func NewMetricsHook() *MetricsHook {
	return &MetricsHook{
		BaseHook: *NewBaseHook([]LogLevel{
			LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal,
		}),
		logCounts: make(map[LogLevel]int),
	}
}

// Fire processes a log entry and tracks metrics
func (h *MetricsHook) Fire(entry *LogEntry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logCounts[entry.Level]++
	return nil
}

// GetMetrics returns the current log metrics
func (h *MetricsHook) GetMetrics() map[LogLevel]int {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Create a copy to prevent direct modification
	metrics := make(map[LogLevel]int)
	for level, count := range h.logCounts {
		metrics[level] = count
	}
	return metrics
}

// PrintMetrics prints the current log metrics
func (h *MetricsHook) PrintMetrics() {
	metrics := h.GetMetrics()
	fmt.Println("Log Metrics:")
	for level, count := range metrics {
		fmt.Printf("%s logs: %d\n", h.levelToString(level), count)
	}
}

// levelToString converts LogLevel to string representation
func (h *MetricsHook) levelToString(level LogLevel) string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}