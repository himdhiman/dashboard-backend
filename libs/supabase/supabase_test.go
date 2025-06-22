package supabase

import (
	"context"
	"testing"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/supabase/config"
	"github.com/himdhiman/dashboard-backend/libs/supabase/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUser represents a test user
type TestUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func TestNewSupabaseClient(t *testing.T) {
	// Create test configuration
	cfg := config.NewConfig().
		WithURL("https://test-project.supabase.co").
		WithKey("test-anon-key").
		WithServiceKey("test-service-key").
		WithTimeout(30 * time.Second)

	// Create test logger (mock)
	logger := &mockLogger{}

	// Create client
	client, err := NewSupabaseClient(cfg, logger)
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Test client methods
	assert.NotNil(t, client.Auth())
	assert.NotNil(t, client.Storage())
	assert.NotNil(t, client.Realtime())

	// Test database operations
	usersRepo := client.From("users")
	assert.NotNil(t, usersRepo)

	// Test typed repository
	typedRepo := client.FromTyped("users")
	assert.NotNil(t, typedRepo)

	// Close client
	err = client.Close()
	assert.NoError(t, err)
}

func TestNewSupabaseClientFromEnv(t *testing.T) {
	// Create test logger (mock)
	logger := &mockLogger{}

	// This will fail without environment variables, but we can test the function exists
	_, err := NewSupabaseClientFromEnv(logger)
	// We expect this to fail because environment variables are not set
	assert.Error(t, err)
}

func TestSupabaseClient_Ping(t *testing.T) {
	// Create test configuration
	cfg := config.NewConfig().
		WithURL("https://test-project.supabase.co").
		WithKey("test-anon-key")

	// Create test logger (mock)
	logger := &mockLogger{}

	// Create client
	client, err := NewSupabaseClient(cfg, logger)
	require.NoError(t, err)
	defer client.Close()

	// Test ping (this will fail with test credentials, but we can test the method exists)
	err = client.Ping(context.Background())
	// We expect this to fail because we're using test credentials
	assert.Error(t, err)
}

func TestSupabaseClient_Auth(t *testing.T) {
	// Create test configuration
	cfg := config.NewConfig().
		WithURL("https://test-project.supabase.co").
		WithKey("test-anon-key")

	// Create test logger (mock)
	logger := &mockLogger{}

	// Create client
	client, err := NewSupabaseClient(cfg, logger)
	require.NoError(t, err)
	defer client.Close()

	// Test auth client
	authClient := client.Auth()
	assert.NotNil(t, authClient)

	// Test auth methods exist (they will fail with test credentials)

	// Test sign up
	signUpReq := &models.SignUpRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err = authClient.SignUp(context.Background(), signUpReq)
	assert.Error(t, err) // Expected to fail with test credentials

	// Test sign in
	signInReq := &models.SignInRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err = authClient.SignIn(context.Background(), signInReq)
	assert.Error(t, err) // Expected to fail with test credentials
}

func TestSupabaseClient_Storage(t *testing.T) {
	// Create test configuration
	cfg := config.NewConfig().
		WithURL("https://test-project.supabase.co").
		WithKey("test-anon-key")

	// Create test logger (mock)
	logger := &mockLogger{}

	// Create client
	client, err := NewSupabaseClient(cfg, logger)
	require.NoError(t, err)
	defer client.Close()

	// Test storage client
	storageClient := client.Storage()
	assert.NotNil(t, storageClient)

	// Test storage methods exist (they will fail with test credentials)

	// Test list buckets
	_, err = storageClient.ListBuckets(context.Background())
	assert.Error(t, err) // Expected to fail with test credentials

	// Test bucket-specific operations
	bucket := storageClient.From("test-bucket")
	assert.NotNil(t, bucket)
}

func TestSupabaseClient_Realtime(t *testing.T) {
	// Create test configuration
	cfg := config.NewConfig().
		WithURL("https://test-project.supabase.co").
		WithKey("test-anon-key")

	// Create test logger (mock)
	logger := &mockLogger{}

	// Create client
	client, err := NewSupabaseClient(cfg, logger)
	require.NoError(t, err)
	defer client.Close()

	// Test realtime client
	realtimeClient := client.Realtime()
	assert.NotNil(t, realtimeClient)

	// Test realtime methods exist (they will fail with test credentials)

	// Test connect
	err = realtimeClient.Connect(context.Background())
	assert.Error(t, err) // Expected to fail with test credentials

	// Test channel operations
	channel := realtimeClient.Channel("test-channel")
	assert.NotNil(t, channel)
}

func TestSupabaseClient_Functions(t *testing.T) {
	// Create test configuration
	cfg := config.NewConfig().
		WithURL("https://test-project.supabase.co").
		WithKey("test-anon-key")

	// Create test logger (mock)
	logger := &mockLogger{}

	// Create client
	client, err := NewSupabaseClient(cfg, logger)
	require.NoError(t, err)
	defer client.Close()

	// Test function calls (they will fail with test credentials)

	// Test edge function
	_, err = client.EdgeFunction("test-function", map[string]interface{}{
		"param": "value",
	})
	assert.Error(t, err) // Expected to fail with test credentials

	// Test regular function
	_, err = client.Function("test-function", map[string]interface{}{
		"param": "value",
	})
	assert.Error(t, err) // Expected to fail with test credentials
}

// Mock logger for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...interface{})     {}
func (m *mockLogger) Info(msg string, args ...interface{})      {}
func (m *mockLogger) Warn(msg string, args ...interface{})      {}
func (m *mockLogger) Error(msg string, args ...interface{})     {}
func (m *mockLogger) Fatal(msg string, args ...interface{})     {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {}
func (m *mockLogger) Infof(format string, args ...interface{})  {}
func (m *mockLogger) WithContext(ctx context.Context) logger.ILogger {
	return m
}
func (m *mockLogger) WithFields(fields logger.Fields) logger.ILogger { return m }
func (m *mockLogger) WithField(key string, value interface{}) logger.ILogger {
	return m
}
func (m *mockLogger) Log(level logger.LogLevel, msg string, args ...interface{}) {}
func (m *mockLogger) WithError(err error) logger.ILogger                         { return m }
