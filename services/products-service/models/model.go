package models

import "time"

type Product struct {
	SKUCode              string    `json:"skuCode" bson:"skuCode" validate:"required"`
	Name                 string    `json:"name" bson:"name" validate:"required"`
	ImageURL             string    `json:"imageUrl" bson:"imageUrl" validate:"required,url"`
	PrimaryVendor        string    `json:"primaryVendor" bson:"primaryVendor" validate:"required"`
	LastProcuredRmbPrice float64   `json:"lastProcuredRmbPrice" bson:"lastProcuredRmbPrice" validate:"required,min=0"`
	CreatedAt            time.Time `json:"createdAt" bson:"createdAt" validate:"required"`
	UpdatedAt            time.Time `json:"updatedAt" bson:"updatedAt" validate:"required"`
}
