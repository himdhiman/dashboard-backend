package oauth

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OAuthVersion represents supported OAuth versions
type OAuthVersion string

const (
	OAuth1 OAuthVersion = "1.0a"
	OAuth2 OAuthVersion = "2.0"
)

// GrantType represents different OAuth grant types
type GrantType string

const (
	PasswordGrant     GrantType = "password"
	ClientCredentials GrantType = "client_credentials"
	AuthorizationCode GrantType = "authorization_code"
	ImplicitGrant     GrantType = "implicit"
	RefreshToken      GrantType = "refresh_token"
	// OAuth 1.0a specific
	RequestToken GrantType = "request_token"
	AccessToken  GrantType = "access_token"
)

// Metadata holds common fields for all documents
type Metadata struct {
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
	CreatedBy string    `bson:"createdBy,omitempty" json:"createdBy,omitempty"`
	UpdatedBy string    `bson:"updatedBy,omitempty" json:"updatedBy,omitempty"`
}

// OAuthProvider represents the configuration for an OAuth provider
type OAuthProvider struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name"`
	Version         OAuthVersion       `bson:"version" json:"version"`
	Endpoints       ProviderEndpoints  `bson:"endpoints" json:"endpoints"`
	SupportedGrants []GrantType        `bson:"supportedGrants" json:"supportedGrants"`
	Config          ProviderConfig     `bson:"config" json:"config"`
	IsActive        bool               `bson:"isActive" json:"isActive"`
	Metadata        Metadata           `bson:"metadata" json:"metadata"`
}

// ProviderEndpoints contains all possible OAuth endpoints
type ProviderEndpoints struct {
	BaseURL          string `bson:"baseUrl" json:"baseUrl"`
	AuthorizationURL string `bson:"authorizationUrl,omitempty" json:"authorizationUrl,omitempty"`
	TokenURL         string `bson:"tokenUrl" json:"tokenUrl"`
	RequestTokenURL  string `bson:"requestTokenUrl,omitempty" json:"requestTokenUrl,omitempty"`
	RevokeTokenURL   string `bson:"revokeTokenUrl,omitempty" json:"revokeTokenUrl,omitempty"`
	UserInfoURL      string `bson:"userInfoUrl,omitempty" json:"userInfoUrl,omitempty"`
}

// ProviderConfig contains version-specific configurations
type ProviderConfig struct {
	// OAuth 2.0 specific
	RequirePKCE          bool          `bson:"requirePkce,omitempty" json:"requirePkce,omitempty"`
	SupportedScopes      []string      `bson:"supportedScopes,omitempty" json:"supportedScopes,omitempty"`
	TokenLifetime        time.Duration `bson:"tokenLifetime" json:"tokenLifetime"`
	RefreshTokenLifetime time.Duration `bson:"refreshTokenLifetime,omitempty" json:"refreshTokenLifetime,omitempty"`
	// OAuth 1.0a specific
	SignatureMethod string `bson:"signatureMethod,omitempty" json:"signatureMethod,omitempty"`
	RequireCallback bool   `bson:"requireCallback,omitempty" json:"requireCallback,omitempty"`
}

// Application represents a client application
type Application struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProviderID     primitive.ObjectID `bson:"providerId" json:"providerId"`
	ClientID       string             `bson:"clientId" json:"clientId"`
	ClientSecret   string             `bson:"clientSecret" json:"clientSecret"`
	Name           string             `bson:"name" json:"name"`
	Description    string             `bson:"description,omitempty" json:"description,omitempty"`
	RedirectURLs   []string           `bson:"redirectUrls,omitempty" json:"redirectUrls,omitempty"`
	AllowedGrants  []GrantType        `bson:"allowedGrants" json:"allowedGrants"`
	AllowedScopes  []string           `bson:"allowedScopes,omitempty" json:"allowedScopes,omitempty"`
	IsConfidential bool               `bson:"isConfidential" json:"isConfidential"`
	IsActive       bool               `bson:"isActive" json:"isActive"`
	Metadata       Metadata           `bson:"metadata" json:"metadata"`
}

// Token represents a generic OAuth token
type Token struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ProviderID     primitive.ObjectID     `bson:"providerId" json:"providerId"`
	ApplicationID  primitive.ObjectID     `bson:"applicationId" json:"applicationId"`
	UserID         primitive.ObjectID     `bson:"userId,omitempty" json:"userId,omitempty"`
	TokenType      string                 `bson:"tokenType" json:"tokenType"` // access_token, refresh_token, request_token
	TokenValue     string                 `bson:"tokenValue" json:"tokenValue"`
	TokenSecret    string                 `bson:"tokenSecret,omitempty" json:"tokenSecret,omitempty"` // For OAuth 1.0a
	Scopes         []string               `bson:"scopes,omitempty" json:"scopes,omitempty"`
	ExpiresAt      time.Time              `bson:"expiresAt" json:"expiresAt"`
	IsRevoked      bool                   `bson:"isRevoked" json:"isRevoked"`
	RevokedAt      *time.Time             `bson:"revokedAt,omitempty" json:"revokedAt,omitempty"`
	AdditionalData map[string]interface{} `bson:"additionalData,omitempty" json:"additionalData,omitempty"`
	Metadata       Metadata               `bson:"metadata" json:"metadata"`
}

// Authorization represents an OAuth authorization grant
type Authorization struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProviderID    primitive.ObjectID `bson:"providerId" json:"providerId"`
	ApplicationID primitive.ObjectID `bson:"applicationId" json:"applicationId"`
	UserID        primitive.ObjectID `bson:"userId" json:"userId"`
	Code          string             `bson:"code,omitempty" json:"code,omitempty"`
	CodeChallenge string             `bson:"codeChallenge,omitempty" json:"codeChallenge,omitempty"`
	CodeMethod    string             `bson:"codeMethod,omitempty" json:"codeMethod,omitempty"`
	RedirectURI   string             `bson:"redirectUri,omitempty" json:"redirectUri,omitempty"`
	Scopes        []string           `bson:"scopes,omitempty" json:"scopes,omitempty"`
	ExpiresAt     time.Time          `bson:"expiresAt" json:"expiresAt"`
	IsUsed        bool               `bson:"isUsed" json:"isUsed"`
	Metadata      Metadata           `bson:"metadata" json:"metadata"`
}

// Helper methods

func (p *OAuthProvider) SupportsVersion(version OAuthVersion) bool {
	return p.Version == version
}

func (p *OAuthProvider) SupportsGrant(grant GrantType) bool {
	for _, g := range p.SupportedGrants {
		if g == grant {
			return true
		}
	}
	return false
}

func (t *Token) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (m *Metadata) BeforeCreate(user string) {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	m.CreatedBy = user
	m.UpdatedBy = user
}

func (m *Metadata) BeforeUpdate(user string) {
	m.UpdatedAt = time.Now()
	m.UpdatedBy = user
}
