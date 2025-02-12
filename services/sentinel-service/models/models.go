package models

import "time"

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
	SKUCode              string    `json:"skuCode" bson:"skuCode"`
	Name                 string    `json:"name" bson:"name"`
	ImageURL             string    `json:"imageUrl" bson:"imageUrl"`
	PrimaryVendor        string    `json:"primaryVendor" bson:"primaryVendor"`
	LastProcuredRmbPrice float64   `json:"lastProcuredRmbPrice" bson:"lastProcuredRmbPrice"`
	CreatedAt            time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt" bson:"updatedAt"`
}

type PurchaseOrderProducts struct {
	ProductSKUCode   string  `json:"productSKUCode" bson:"productSKUCode"`
	ImageURL         string  `json:"imageUrl" bson:"imageUrl"`
	Quantity         int     `json:"quantity" bson:"quantity"`
	LastBestRMBPrice float64 `json:"lastBestRMBPrice" bson:"lastBestRMBPrice"`
	CurrentRMBPrice  float64 `json:"currentRMBPrice" bson:"currentRMBPrice"`
	Status           string  `json:"status" bson:"status"`
	Remarks          string  `json:"remarks" bson:"remarks"`
	ShippingMark     string  `json:"shippingMark" bson:"shippingMark"`
}

type PurchaseOrder struct {
	OrderNumber string                  `json:"orderNumber" bson:"orderNumber"`
	Vendor      string                  `json:"vendor" bson:"vendor"`
	OrderDate   time.Time               `json:"orderDate" bson:"orderDate"`
	TotalAmount float64                 `json:"totalAmount" bson:"totalAmount"`
	Products    []PurchaseOrderProducts `json:"products" bson:"products"`
	Deposits    float64                 `json:"deposits" bson:"deposits"`
	OrderStatus string                  `json:"orderStatus" bson:"orderStatus"`
}
