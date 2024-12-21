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

type APIConfig struct {
	ApiName       string                   `bson:"api_name" json:"api_name"`
	Endpoint      string                   `bson:"endpoint" json:"endpoint"`
	Path          string                   `bson:"path" json:"path"`
	Method        AuthenticationMethodType `bson:"method" json:"method"`
	RateLimit     int                      `bson:"rate_limit" json:"rate_limit"`
	Authorization AuthConfig               `bson:"authorization" json:"authorization"`
}

type AuthConfig struct {
	Type        AuthenticationType `bson:"type" json:"type"` // e.g., "OAuth2", "BasicAuth"
	OAuthConfig *OAuthConfig       `bson:"oauth_config,omitempty" json:"oauth_config,omitempty"`
}

type OAuthConfig struct {
	Username     string `bson:"username" json:"username"`
	ClientID     string `bson:"client_id" json:"client_id"`         // OAuth2 Client ID
	ClientSecret string `bson:"client_secret" json:"client_secret"` // OAuth2 Client Secret
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // Access token expiry in seconds
	Scope        string `json:"scope"`
}
