package models

type APIConfig struct {
	Endpoint      string     `bson:"endpoint" json:"endpoint"`
	RateLimit     int        `bson:"rate_limit" json:"rate_limit"`
	Authorization AuthConfig `bson:"authorization" json:"authorization"`
}

type AuthConfig struct {
	Type      string           `bson:"type" json:"type"` // e.g., "OAuth2", "BasicAuth"
	OAuth2    *OAuth2Config    `bson:"oauth2,omitempty" json:"oauth2,omitempty"`
	BasicAuth *BasicAuthConfig `bson:"basic_auth,omitempty" json:"basic_auth,omitempty"`
}

type OAuth2Config struct {
	TokenURL string `bson:"token_url" json:"token_url"` // Token endpoint
	ClientID string `bson:"client_id" json:"client_id"` // OAuth2 Client ID
	Secret   string `bson:"secret" json:"secret"`       // OAuth2 Client Secret
}

type BasicAuthConfig struct {
	Credentials string `bson:"credentials" json:"credentials"` // JSON string of {username, password}
}
