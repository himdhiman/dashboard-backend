package models

import (
	"time"
)

type ExportJobPayload struct {
	ExportJobTypeName string      `json:"exportJobTypeName"`
	ExportColumns     []string    `json:"exportColums"`
	ExportFilters     interface{} `json:"exportFilters"`
	Frequency         string      `json:"frequency"`
	ReportName        string      `json:"reportName"`
}
type Product struct {
	ID                   string    `bson:"_id,omitempty" json:"id,omitempty"`
	SKUCode              string    `json:"skuCode" bson:"skuCode" validate:"required"`
	Name                 string    `json:"name" bson:"name" validate:"required"`
	ImageURL             string    `json:"imageUrl" bson:"imageUrl" validate:"required,url"`
	PrimaryVendor        string    `json:"primaryVendor" bson:"primaryVendor" validate:"required"`
	LastProcuredRmbPrice float64   `json:"lastProcuredRmbPrice" bson:"lastProcuredRmbPrice" validate:"required,min=0"`
	CreatedAt            time.Time `json:"createdAt" bson:"createdAt" validate:"required"`
	UpdatedAt            time.Time `json:"updatedAt" bson:"updatedAt" validate:"required"`
}

type ProductBundle struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	BundleSKU string    `json:"bundleSku" bson:"bundleSku" validate:"required"`
	Name      string    `json:"name" bson:"name" validate:"required"`
	Products  []string  `json:"products" bson:"products" validate:"required,dive,required"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt" validate:"required"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt" validate:"required"`
}
