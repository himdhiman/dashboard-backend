package models

type AuthenticationType string
type AuthenticationMethodType string

const (
	BASIC_AUTH AuthenticationType = "BasicAuth"
	OAUTH      AuthenticationType = "OAuth2"
)

const (
	GET  AuthenticationMethodType = "GET"
	POST AuthenticationMethodType = "POST"
)

// APIConfig is the model for the API configuration
type APIConfig struct {
	Code          string      `bson:"code" json:"code"`
	BaseURL       string      `bson:"base_url" json:"base_url"`
	Authorization AuthConfig  `bson:"authorization" json:"authorization"`
	Endpoints     []Endpoints `bson:"endpoints" json:"endpoints"`
}

// AuthConfig is the model for the authentication configuration
type AuthConfig struct {
	Type        AuthenticationType `bson:"type" json:"type"` // e.g., "OAuth2", "BasicAuth"
	Path        string             `bson:"path" json:"path"` // e.g., "/oauth/token"
	Credentials string             `bson:"credentials" json:"credentials"`
}

// Endpoints is the model for the endpoints configuration
type Endpoints struct {
	Code      string                   `bson:"code" json:"code"`
	Path      string                   `bson:"path" json:"path"`
	Method    AuthenticationMethodType `bson:"method" json:"method"`
	RateLimit int                      `bson:"rate_limit" json:"rate_limit"`
	Timeout   int                      `bson:"timeout" json:"timeout"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // Access token expiry in seconds
	Scope        string `json:"scope"`
}

type Product struct {
	ID                   int     `json:"id"`
	SKUCode              string  `json:"skuCode"`
	Name                 string  `json:"name"`
	ImageURL             string  `json:"imageUrl"`
	PrimaryVendor        string  `json:"primaryVendor"`
	LastProcuredRmbPrice float64 `json:"lastProcuredRmbPrice"`
}
