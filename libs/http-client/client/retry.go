package client

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// backoffTime calculates the exponential backoff time
func backoffTime(retryCount int, initialWait, maxWait time.Duration) time.Duration {
	// Ensure retryCount is non-negative
	if retryCount < 0 {
		retryCount = 0
	}

	// Calculate exponential backoff base
	backoff := initialWait * time.Duration(math.Pow(2, float64(retryCount)))

	// Add some jitter to prevent thundering herd problem
	jitter := time.Duration(rand.Int63n(int64(initialWait.Seconds()))) * time.Second
	backoff += jitter

	// Ensure the backoff doesn't exceed maxWait
	if backoff > maxWait {
		return maxWait
	}

	return backoff
}

// RetryRequest implements the retry mechanism with exponential backoff
func RetryRequest(client *http.Client, req *http.Request, config Config) (*http.Response, error) {
	// Seed the random number generator to ensure different jitter each time
	rand.Seed(time.Now().UnixNano())

	// Set a timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Create a copy of the request for each retry
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Create a new request for each attempt
		reqCopy := req.Clone(ctx)

		// Perform the request
		resp, err = client.Do(reqCopy)

		// Check if request was successful
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		// If not the last attempt, wait before retrying
		if attempt < config.MaxRetries {
			// Calculate backoff time
			waitTime := backoffTime(attempt, config.InitialWait, config.MaxWait)
			time.Sleep(waitTime)
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %v", config.MaxRetries, err)
}
