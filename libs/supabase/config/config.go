package config

import (
	"os"
	"time"

	supabaseErrors "github.com/himdhiman/dashboard-backend/libs/supabase/errors"
)

// Config holds the Supabase configuration
type Config struct {
	URL           string
	Key           string
	ServiceKey    string
	JWTSecret     string
	Timeout       time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
	EnableLogging bool
}

// NewConfig creates a new Supabase configuration
func NewConfig() *Config {
	return &Config{
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		EnableLogging: true,
	}
}

// FromEnv loads configuration from environment variables
func (c *Config) FromEnv() *Config {
	if url := os.Getenv("SUPABASE_URL"); url != "" {
		c.URL = url
	}
	if key := os.Getenv("SUPABASE_ANON_KEY"); key != "" {
		c.Key = key
	}
	if serviceKey := os.Getenv("SUPABASE_SERVICE_KEY"); serviceKey != "" {
		c.ServiceKey = serviceKey
	}
	if jwtSecret := os.Getenv("SUPABASE_JWT_SECRET"); jwtSecret != "" {
		c.JWTSecret = jwtSecret
	}
	return c
}

// WithURL sets the Supabase URL
func (c *Config) WithURL(url string) *Config {
	c.URL = url
	return c
}

// WithKey sets the Supabase anon key
func (c *Config) WithKey(key string) *Config {
	c.Key = key
	return c
}

// WithServiceKey sets the Supabase service key
func (c *Config) WithServiceKey(serviceKey string) *Config {
	c.ServiceKey = serviceKey
	return c
}

// WithJWTSecret sets the JWT secret
func (c *Config) WithJWTSecret(jwtSecret string) *Config {
	c.JWTSecret = jwtSecret
	return c
}

// WithTimeout sets the timeout for requests
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

// WithMaxRetries sets the maximum number of retries
func (c *Config) WithMaxRetries(maxRetries int) *Config {
	c.MaxRetries = maxRetries
	return c
}

// WithRetryDelay sets the delay between retries
func (c *Config) WithRetryDelay(retryDelay time.Duration) *Config {
	c.RetryDelay = retryDelay
	return c
}

// WithLogging enables or disables logging
func (c *Config) WithLogging(enable bool) *Config {
	c.EnableLogging = enable
	return c
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.URL == "" {
		return supabaseErrors.ErrMissingURL
	}
	if c.Key == "" {
		return supabaseErrors.ErrMissingKey
	}
	return nil
}
