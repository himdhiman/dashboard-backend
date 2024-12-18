package logger

import (
	"io"
	"os"
)

// Config represents the configuration for the logger
type Config struct {
	// Logging level
	Level LogLevel

	// Output destinations
	Outputs []io.Writer

	// Output format (json, text)
	Format string

	// Correlation ID key
	CorrelationKey string

	// Hooks for additional logging actions
	Hooks []HookInterface

	// Custom formatters
	Formatter FormatterInterface
}

// DefaultConfig provides a standard configuration
func DefaultConfig() *Config {
	return &Config{
		Level:          LevelInfo,
		Outputs:        []io.Writer{os.Stdout},
		Format:         "text",
		CorrelationKey: "correlation_id",
		Hooks:          []HookInterface{},
	}
}

// WithLevel sets the logging level
func (c *Config) WithLevel(level LogLevel) *Config {
	c.Level = level
	return c
}

// WithOutput adds additional output writers
func (c *Config) WithOutput(writers ...io.Writer) *Config {
	c.Outputs = append(c.Outputs, writers...)
	return c
}

// WithFormat sets the log output format
func (c *Config) WithFormat(format string) *Config {
	c.Format = format
	return c
}
