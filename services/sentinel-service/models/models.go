package models

import (
	"time"
)

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
	Code          string      `bson:"code" json:"code" validate:"required"`
	BaseURL       string      `bson:"base_url" json:"base_url" validate:"required,url"`
	Authorization AuthConfig  `bson:"authorization" json:"authorization" validate:"required"`
	Endpoints     []Endpoints `bson:"endpoints" json:"endpoints" validate:"required,dive"`
}

// AuthConfig is the model for the authentication configuration
type AuthConfig struct {
	Type        AuthenticationType `bson:"type" json:"type" validate:"required,oneof=OAuth2 BasicAuth"`
	Path        string             `bson:"path" json:"path" validate:"required"`
	Credentials string             `bson:"credentials" json:"credentials" validate:"required"`
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
	AccessToken  string `json:"access_token" validate:"required"`
	TokenType    string `json:"token_type" validate:"required,oneof=Bearer"`
	RefreshToken string `json:"refresh_token" validate:"required"`
	ExpiresIn    int    `json:"expires_in" validate:"required,min=1"`
	Scope        string `json:"scope" validate:"required"`
}

type Product struct {
	SKUCode              string    `json:"skuCode" bson:"skuCode" validate:"required"`
	Name                 string    `json:"name" bson:"name" validate:"required"`
	ImageURL             string    `json:"imageUrl" bson:"imageUrl" validate:"required,url"`
	PrimaryVendor        string    `json:"primaryVendor" bson:"primaryVendor" validate:"required"`
	LastProcuredRmbPrice float64   `json:"lastProcuredRmbPrice" bson:"lastProcuredRmbPrice" validate:"required,min=0"`
	CreatedAt            time.Time `json:"createdAt" bson:"createdAt" validate:"required"`
	UpdatedAt            time.Time `json:"updatedAt" bson:"updatedAt" validate:"required"`
}

type PurchaseOrderProducts struct {
	SkuCode          string  `json:"skuCode" bson:"skuCode" validate:"required"`
	ImageURL         string  `json:"imageUrl" bson:"imageUrl" validate:"required,url"`
	Quantity         float64 `json:"quantity" bson:"quantity" validate:"required,min=1"`
	LastBestRMBPrice float64 `json:"lastBestRMBPrice" bson:"lastBestRMBPrice"`
	CurrentRMBPrice  float64 `json:"currentRMBPrice" bson:"currentRMBPrice" validate:"required,min=0"`
	Status           string  `json:"status" bson:"status" validate:"required,oneof=pending final"`
	Remarks          string  `json:"remarks" bson:"remarks"`
	ShippingMark     string  `json:"shippingMark" bson:"shippingMark"`
}

type PurchaseOrder struct {
	PONumber              string                  `json:"poNumber" bson:"poNumber" validate:"required"`
	Vendor                string                  `json:"vendor" bson:"vendor" validate:"required"`
	OrderDate             time.Time               `json:"orderDate" bson:"orderDate" validate:"required"`
	TotalAmount           float64                 `json:"totalAmount" bson:"totalAmount" validate:"required,min=0"`
	Products              []PurchaseOrderProducts `json:"products" bson:"products" validate:"required,min=1,dive"`
	Deposits              float64                 `json:"deposits" bson:"deposits" validate:"min=0"`
	OrderStatus           string                  `json:"orderStatus" bson:"orderStatus" validate:"required,oneof=pending partially_pending finalized"`
	TentativeDispatchDate time.Time               `json:"tentativeDispatchDate" bson:"tentativeDispatchDate" validate:"required"`
	OrderType             string                  `json:"orderType" bson:"orderType" validate:"required,oneof=new repeat"`
	Remarks               string                  `json:"remarks" bson:"remarks"`
	UpdatedAt             time.Time               `json:"updatedAt" bson:"updatedAt" validate:"required"`
}
