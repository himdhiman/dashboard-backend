package models

import (
	"time"

	"github.com/himdhiman/dashboard-backend/libs/mongo"
)

type Product struct {
	ID                   mongo.ObjectID `json:"id" bson:"_id"`
	SKUCode              string         `json:"skuCode" bson:"skuCode" validate:"required"`
	Name                 string         `json:"name" bson:"name" validate:"required"`
	ImageURL             string         `json:"imageUrl" bson:"imageUrl" validate:"required,url"`
	PrimaryVendor        string         `json:"primaryVendor" bson:"primaryVendor" validate:"required"`
	LastProcuredRmbPrice float64        `json:"lastProcuredRmbPrice" bson:"lastProcuredRmbPrice" validate:"required,min=0"`
	CreatedAt            time.Time      `json:"createdAt" bson:"createdAt" validate:"required"`
	UpdatedAt            time.Time      `json:"updatedAt" bson:"updatedAt" validate:"required"`
}
type UnicommerceInventoryAdjustment struct {
	ItemSKU        string `json:"itemSKU" bson:"itemSKU" validate:"required"`
	Quantity       int    `json:"quantity" bson:"quantity" validate:"required,min=0"`
	ShelfCode      string `json:"shelfCode" bson:"shelfCode" validate:"required"`
	InventoryType  string `json:"inventoryType" bson:"inventoryType" validate:"required,oneof=GOOD_INVENTORY DAMAGED_INVENTORY" default:"GOOD_INVENTORY"`
	AdjustmentType string `json:"adjustmentType" bson:"adjustmentType" validate:"required,oneof=ADD REMOVE"`
	Remarks        string `json:"remarks,omitempty" bson:"remarks,omitempty"`
	FacilityCode   string `json:"facilityCode" bson:"facilityCode" validate:"required"`
}
type UnicommerceInventoryAdjustmentRequest struct {
	InventoryAdjustments []UnicommerceInventoryAdjustment `json:"inventoryAdjustments" bson:"inventoryAdjustments" validate:"required,dive"`
	ForceAllocate        bool                             `json:"forceAllocate" bson:"forceAllocate" validate:"required" default:"false"`
}
