package models

import (
	"time"

	"github.com/himdhiman/dashboard-backend/libs/mongo"
)

type PurchaseOrderProducts struct {
	ID              mongo.ObjectID `json:"id" bson:"_id"`
	ProductID       string         `json:"productID" bson:"productID" validate:"required"`
	Quantity        float64        `json:"quantity" bson:"quantity" validate:"required,min=1"`
	CurrentRMBPrice float64        `json:"currentRMBPrice" bson:"currentRMBPrice" validate:"required,min=0"`
	Status          string         `json:"status" bson:"status" validate:"required,oneof=pending final"`
	Remarks         string         `json:"remarks" bson:"remarks"`
	ShippingMark    string         `json:"shippingMark" bson:"shippingMark"`
	OrderDate       time.Time      `json:"orderDate" bson:"orderDate" validate:"required"`
}

type PurchaseOrder struct {
	ID                    mongo.ObjectID   `json:"id" bson:"_id"`
	PONumber              string           `json:"poNumber" bson:"poNumber" validate:"required"`
	Vendor                string           `json:"vendor" bson:"vendor" validate:"required"`
	OrderDate             time.Time        `json:"orderDate" bson:"orderDate" validate:"required"`
	TotalAmount           float64          `json:"totalAmount" bson:"totalAmount" validate:"required,min=0"`
	Products              []mongo.ObjectID `json:"products" bson:"products" validate:"required,min=1,dive"`
	Deposits              float64          `json:"deposits" bson:"deposits" validate:"min=0"`
	OrderStatus           string           `json:"orderStatus" bson:"orderStatus" validate:"required,oneof=pending partially_pending finalized"`
	ShippingStatus        string           `json:"shippingStatus" bson:"shippingStatus" validate:"required,oneof=pending complete partly_shipped"`
	TentativeDispatchDate time.Time        `json:"tentativeDispatchDate" bson:"tentativeDispatchDate" validate:"required"`
	OrderType             string           `json:"orderType" bson:"orderType" validate:"required,oneof=new repeat"`
	Remarks               string           `json:"remarks" bson:"remarks"`
	UpdatedAt             time.Time        `json:"updatedAt" bson:"updatedAt" validate:"required"`
}
