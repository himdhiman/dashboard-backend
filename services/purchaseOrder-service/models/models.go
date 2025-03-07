package models

import "time"

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
