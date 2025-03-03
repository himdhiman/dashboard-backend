package models

import (
	"context"
	"io"
	"time"
)

type AuthenticationType string
type AuthenticationMethodType string

const (
	BASIC_AUTH AuthenticationType = "BasicAuth"
	OAUTH      AuthenticationType = "OAuth2"
)

const (
	GET    AuthenticationMethodType = "GET"
	POST   AuthenticationMethodType = "POST"
	PUT    AuthenticationMethodType = "PUT"
	DELETE AuthenticationMethodType = "DELETE"
	PATCH  AuthenticationMethodType = "PATCH"
)

// APIConfig is the model for the API configuration
type APIConfig struct {
	Code          string      `bson:"code" json:"code" validate:"required"`
	BaseURL       string      `bson:"base_url" json:"base_url" validate:"required,url"`
	Authorization AuthConfig  `bson:"authorization" json:"authorization" validate:"required"`
	Endpoints     []Endpoints `bson:"endpoints" json:"endpoints" validate:"required,dive"`
}

// AuthConfig is the model for the authentication configuration
type AuthConfig struct {
	Type        AuthenticationType `bson:"type" json:"type" validate:"required,oneof=OAuth2 BasicAuth"`
	Path        string             `bson:"path" json:"path" validate:"required"`
	Credentials Credentials        `bson:"credentials" json:"credentials" validate:"required"`
}

type Credentials struct {
	Username     string `json:"username" validate:"required"`
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
}

// Endpoints is the model for the endpoints configuration
type Endpoints struct {
	Code      string                   `bson:"code" json:"code" validate:"required"`
	Path      string                   `bson:"path" json:"path" validate:"required"`
	Method    AuthenticationMethodType `bson:"method" json:"method" validate:"required,oneof=GET POST"`
	RateLimit int                      `bson:"rate_limit" json:"rate_limit" validate:"required,min=1"`
	Timeout   int                      `bson:"timeout" json:"timeout" validate:"required,min=1"`
}

type TokenResponse struct {
	AccessToken  string    `json:"access_token" validate:"required"`
	TokenType    string    `json:"token_type" validate:"required,oneof=Bearer"`
	RefreshToken string    `json:"refresh_token" validate:"required"`
	ExpiresIn    int       `json:"expires_in" validate:"required,min=1"`
	Scope        string    `json:"scope" validate:"required"`
	CreatedAt    time.Time `json:"created_at" validate:"required"`
}

type TokenMetadata struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type APIRequest struct {
	Method  string            `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH"`
	URL     string            `json:"url" validate:"required,url"`
	Headers map[string]string `json:"headers"`
	Body    io.Reader         `json:"body"`
}

type APIResponse struct {
	StatusCode int    `json:"status_code" validate:"required,min=100,max=599"`
	Body       []byte `json:"body" validate:"required"`
}

type ApiClientConfig struct {
	BaseURL     string
	Timeout     int
	MaxRetries  int
	RetryDelay  int
	BearerToken string
}

type ApiClient interface {
	ApiClientConfig
	Get(ctx context.Context, path string, headers map[string]string) (*APIResponse, error)
	Post(ctx context.Context, path string, body io.Reader, headers map[string]string) (*APIResponse, error)
	Put(ctx context.Context, path string, body io.Reader, headers map[string]string) (*APIResponse, error)
	Delete(ctx context.Context, path string, headers map[string]string) (*APIResponse, error)
	Patch(ctx context.Context, path string, body io.Reader, headers map[string]string) (*APIResponse, error)
	Do(ctx context.Context, request *APIRequest) (*APIResponse, error)
}
