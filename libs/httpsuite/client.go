package httpsuite

import (
	"net/http"
)

// CreateHTTPClient creates a new HTTP client with the given configuration settings.
func CreateHTTPClient(config Config) *http.Client {
	return &http.Client{
		Timeout: config.Timeout,
	}
}

// Get makes an HTTP GET request with retry logic.
func Get(url string, config Config) (*http.Response, error) {
	client := CreateHTTPClient(config)

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Call the retry logic for the request
	return RetryRequest(client, req, config)
}

// Post makes an HTTP POST request with retry logic.
func Post(url string, body []byte, config Config) (*http.Response, error) {
	client := CreateHTTPClient(config)

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	// Call the retry logic for the request
	return RetryRequest(client, req, config)
}
