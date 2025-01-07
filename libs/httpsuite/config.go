// client/config.go

package httpsuite

import (
	"time"
)

// Config represents the configuration for the HTTP client.
type Config struct {
	// Maximum number of retries before giving up
	MaxRetries int `json:"max_retries"`
	// Timeout for each request
	Timeout time.Duration `json:"timeout"`
	// Retry wait time (starting point for backoff)
	InitialWait time.Duration `json:"initial_wait"`
	// Maximum wait time between retries
	MaxWait time.Duration `json:"max_wait"`
}

// Default configuration values.
var DefaultConfig = Config{
	MaxRetries:  3,
	Timeout:     10 * time.Second,
	InitialWait: 1 * time.Second,
	MaxWait:     5 * time.Second,
}
