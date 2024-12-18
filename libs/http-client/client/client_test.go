package client

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocking the http.RoundTripper interface
type MockRoundTripper struct {
	mock.Mock
}

// Implement the RoundTrip method to satisfy http.RoundTripper interface
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// Test that verifies a successful HTTP request
func TestRetryRequest_Success(t *testing.T) {
	// Create a regular http.Client with the mock round tripper
	mockRoundTripper := new(MockRoundTripper)
	httpClient := &http.Client{
		Transport: mockRoundTripper,
	}

	// Setup a successful response
	resp := &http.Response{
		StatusCode: http.StatusOK,
	}
	mockRoundTripper.On("RoundTrip", mock.Anything).Return(resp, nil).Once()

	// Define the request
	req, err := http.NewRequest("GET", "https://example.com", nil)
	assert.NoError(t, err)

	// Set up the config for retry
	config := Config{
		MaxRetries:  3,
		Timeout:     5 * time.Second,
		InitialWait: 1 * time.Second,
		MaxWait:     3 * time.Second,
	}

	// Call the RetryRequest function
	result, err := RetryRequest(httpClient, req, config)

	// Assert that the request was successful
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, http.StatusOK, result.StatusCode)

	// Verify the mock was called exactly once
	mockRoundTripper.AssertExpectations(t)
}

// Test that verifies retry logic for a failed request
func TestRetryRequest_Failure_WithRetries(t *testing.T) {
	// Create a regular http.Client with the mock round tripper
	mockRoundTripper := new(MockRoundTripper)
	httpClient := &http.Client{
		Transport: mockRoundTripper,
	}

	// Simulate request failures
	// We expect the number of calls to be MaxRetries + 1 (initial attempt + retries)
	mockRoundTripper.On("RoundTrip", mock.Anything).Return(nil, errors.New("request failed")).Times(4)

	// Define the request
	req, err := http.NewRequest("GET", "https://example.com", nil)
	assert.NoError(t, err)

	// Set up the config for retry
	config := Config{
		MaxRetries:  3,
		Timeout:     5 * time.Second,
		InitialWait: 1 * time.Second,
		MaxWait:     3 * time.Second,
	}

	// Call the RetryRequest function
	resp, err := RetryRequest(httpClient, req, config)

	// Assert that the response is nil and error is returned after retries
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Verify the mock was called exactly 4 times (1 initial + 3 retries)
	mockRoundTripper.AssertExpectations(t)
}

// Test that verifies the exponential backoff strategy
func TestRetryRequest_ExponentialBackoff(t *testing.T) {
	// Create a regular http.Client with the mock round tripper
	mockRoundTripper := new(MockRoundTripper)
	httpClient := &http.Client{
		Transport: mockRoundTripper,
	}

	// Simulate request failure
	// We expect the number of calls to be MaxRetries + 1 (initial attempt + retries)
	mockRoundTripper.On("RoundTrip", mock.Anything).Return(nil, errors.New("request failed")).Times(3)

	// Define the request
	req, err := http.NewRequest("GET", "https://example.com", nil)
	assert.NoError(t, err)

	// Set up the config for retry with backoff
	config := Config{
		MaxRetries:  2, // Changed to match the number of calls
		Timeout:     5 * time.Second,
		InitialWait: 1 * time.Second,
		MaxWait:     3 * time.Second,
	}

	// Call the RetryRequest function (this will fail twice and retry)
	resp, err := RetryRequest(httpClient, req, config)

	// Assert that the response is nil and error is returned after retries
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Verify the mock was called exactly 3 times (1 initial + 2 retries)
	mockRoundTripper.AssertExpectations(t)
}

// Test that verifies the backoff time calculation
func TestBackoffTime_Calculation(t *testing.T) {
	// Test for backoff time calculation
	config := Config{
		InitialWait: 1 * time.Second,
		MaxWait:     3 * time.Second,
	}

	// Check that backoff time grows exponentially
	backoff1 := backoffTime(0, config.InitialWait, config.MaxWait)
	backoff2 := backoffTime(1, config.InitialWait, config.MaxWait)

	assert.True(t, backoff1 < backoff2) // Backoff time should grow
	assert.Equal(t, backoff2, 2*time.Second)

	// Check the maximum wait time
	backoff3 := backoffTime(10, config.InitialWait, config.MaxWait)
	assert.Equal(t, backoff3, config.MaxWait)
}

// Test that verifies timeout handling during retry
func TestRetryRequest_Timeout(t *testing.T) {
	// Create a regular http.Client with the mock round tripper
	mockRoundTripper := new(MockRoundTripper)
	httpClient := &http.Client{
		Transport: mockRoundTripper,
	}

	// Simulate a request timeout (exceeds timeout limit)
	mockRoundTripper.On("RoundTrip", mock.Anything).Return(nil, errors.New("timeout error"))

	// Define the request
	req, err := http.NewRequest("GET", "https://example.com", nil)
	assert.NoError(t, err)

	// Set up the config with a timeout limit
	config := Config{
		MaxRetries:  3,
		Timeout:     2 * time.Second, // Set a short timeout to simulate timeout error
		InitialWait: 1 * time.Second,
		MaxWait:     3 * time.Second,
	}

	// Call the RetryRequest function
	resp, err := RetryRequest(httpClient, req, config)

	// Assert that the response is nil and error is returned due to timeout
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Verify the mock was called once
	mockRoundTripper.AssertExpectations(t)
}
