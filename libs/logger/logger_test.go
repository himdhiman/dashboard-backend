package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	config := &Config{
		Format: "json",
	}
	logger := New(config)
	assert.NotNil(t, logger)
	assert.Equal(t, config, logger.config)
}

func TestLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Format: "text",
	}
	logger := New(config)
	logger.slogLogger = slog.New(slog.NewTextHandler(&buf, nil))

	logger.Info("Test info message")
	assert.Contains(t, buf.String(), "Test info message")
}

func TestLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Format: "text",
	}
	logger := New(config)
	logger.slogLogger = slog.New(slog.NewTextHandler(&buf, nil))

	logger.Warn("Test warn message")
	assert.Contains(t, buf.String(), "Test warn message")
}

func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Format: "text",
	}
	logger := New(config)
	logger.slogLogger = slog.New(slog.NewTextHandler(&buf, nil))

	logger.Error("Test error message")
	assert.Contains(t, buf.String(), "Test error message")
}

func TestLoggerWithContext(t *testing.T) {
	config := &Config{
		Format: "json",
	}
	logger := New(config)
	ctx := context.Background()
	newLogger := logger.WithContext(ctx)
	assert.NotNil(t, newLogger)
}

func TestLoggerWithFields(t *testing.T) {
	config := &Config{
		Format: "json",
	}
	logger := New(config)
	fields := Fields{"key": "value"}
	newLogger := logger.WithFields(fields)
	assert.NotNil(t, newLogger)
}
