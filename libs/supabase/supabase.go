package supabase

import (
	"context"
	"fmt"
	"net/http"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/supabase/auth"
	"github.com/himdhiman/dashboard-backend/libs/supabase/config"
	"github.com/himdhiman/dashboard-backend/libs/supabase/errors"
	"github.com/himdhiman/dashboard-backend/libs/supabase/realtime"
	"github.com/himdhiman/dashboard-backend/libs/supabase/repository"
	"github.com/himdhiman/dashboard-backend/libs/supabase/storage"
)

// ISupabaseClient defines the interface for Supabase operations
type ISupabaseClient interface {
	// Core methods
	Ping(ctx context.Context) error
	Close() error

	// Auth methods
	Auth() auth.IAuthClient

	// Database methods
	From(table string) repository.IRepository[map[string]interface{}]
	FromTyped(table string) repository.IRepository[interface{}]

	// Storage methods
	Storage() storage.IStorageClient

	// Realtime methods
	Realtime() realtime.IRealtimeClient

	// Function methods
	Function(name string, params map[string]interface{}) (interface{}, error)
	EdgeFunction(name string, params map[string]interface{}) (interface{}, error)
}

// SupabaseClient represents a Supabase client
type SupabaseClient struct {
	ISupabaseClient
	config     *config.Config
	httpClient *http.Client
	logger     logger.ILogger
	auth       auth.IAuthClient
	storage    storage.IStorageClient
	realtime   realtime.IRealtimeClient
}

// NewSupabaseClient creates a new Supabase client
func NewSupabaseClient(cfg *config.Config, logger logger.ILogger) (ISupabaseClient, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}

	client := &SupabaseClient{
		config:     cfg,
		httpClient: httpClient,
		logger:     logger,
	}

	// Initialize auth client
	authClient, err := auth.NewAuthClient(cfg, httpClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth client: %w", err)
	}
	client.auth = authClient

	// Initialize storage client
	storageClient, err := storage.NewStorageClient(cfg, httpClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage client: %w", err)
	}
	client.storage = storageClient

	// Initialize realtime client
	realtimeClient, err := realtime.NewRealtimeClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize realtime client: %w", err)
	}
	client.realtime = realtimeClient

	logger.Info("Supabase client initialized successfully")
	return client, nil
}

// NewSupabaseClientFromEnv creates a new Supabase client from environment variables
func NewSupabaseClientFromEnv(logger logger.ILogger) (ISupabaseClient, error) {
	cfg := config.NewConfig().FromEnv()
	return NewSupabaseClient(cfg, logger)
}

// Ping checks the connection to Supabase
func (s *SupabaseClient) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.config.URL+"/rest/v1/", nil)
	if err != nil {
		return errors.NewError(err, "failed to create ping request")
	}

	req.Header.Set("apikey", s.config.Key)
	req.Header.Set("Authorization", "Bearer "+s.config.Key)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return errors.NewError(err, "failed to ping Supabase")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("ping failed with status: %d", resp.StatusCode)
	}

	s.logger.Info("Supabase ping successful")
	return nil
}

// Close closes the Supabase client and all its connections
func (s *SupabaseClient) Close() error {
	var errs []error

	// Close realtime connections
	if err := s.realtime.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close realtime client: %w", err))
	}

	// Close HTTP client (if needed)
	s.httpClient.CloseIdleConnections()

	if len(errs) > 0 {
		return fmt.Errorf("errors closing client: %v", errs)
	}

	s.logger.Info("Supabase client closed successfully")
	return nil
}

// Auth returns the auth client
func (s *SupabaseClient) Auth() auth.IAuthClient {
	return s.auth
}

// From returns a repository for the specified table with generic map type
func (s *SupabaseClient) From(table string) repository.IRepository[map[string]interface{}] {
	return repository.NewRepository[map[string]interface{}](s.config, s.httpClient, s.logger, table)
}

// FromTyped returns a repository for the specified table with typed data
func (s *SupabaseClient) FromTyped(table string) repository.IRepository[interface{}] {
	return repository.NewRepository[interface{}](s.config, s.httpClient, s.logger, table)
}

// FromTypedGeneric is a helper function to create typed repositories
func FromTyped[T any](client ISupabaseClient, table string) repository.IRepository[T] {
	return repository.NewRepository[T](client.(*SupabaseClient).config, client.(*SupabaseClient).httpClient, client.(*SupabaseClient).logger, table)
}

// Storage returns the storage client
func (s *SupabaseClient) Storage() storage.IStorageClient {
	return s.storage
}

// Realtime returns the realtime client
func (s *SupabaseClient) Realtime() realtime.IRealtimeClient {
	return s.realtime
}

// Function calls a Supabase function
func (s *SupabaseClient) Function(name string, params map[string]interface{}) (interface{}, error) {
	// Implementation for calling Supabase functions
	// This would make an HTTP POST request to the function endpoint
	return nil, errors.ErrFunctionError
}

// EdgeFunction calls a Supabase edge function
func (s *SupabaseClient) EdgeFunction(name string, params map[string]interface{}) (interface{}, error) {
	// Implementation for calling Supabase edge functions
	// This would make an HTTP POST request to the edge function endpoint
	return nil, errors.ErrEdgeFunctionError
}
