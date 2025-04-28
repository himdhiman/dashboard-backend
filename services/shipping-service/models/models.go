package models

import (
	"time"

	"github.com/himdhiman/dashboard-backend/libs/mongo"
)

type ShippingProvider struct {
	ID           mongo.ObjectID `json:"id" bson:"_id"`
	VendorName   string         `json:"vendorName" bson:"vendorName" validate:"required"`
	WeightPerKg  float64        `json:"weightPerKg" bson:"weightPerKg" validate:"required,min=0"`
	PaymentTerms string         `json:"paymentTerms" bson:"paymentTerms" validate:"required"`
	DeliveryTAT  string         `json:"deliveryTAT" bson:"deliveryTAT" validate:"required,min=0"`
}

type ShippingMark struct {
	ID                     mongo.ObjectID `json:"id" bson:"_id"`
	ShippingMark           string         `json:"shippingMark" bson:"shippingMark" validate:"required"`
	NumberOfCartons        float64        `json:"numberOfCartons" bson:"numberOfCartons"`
	Weight                 float64        `json:"weight" bson:"weight"`
	PackingList            string         `json:"packingList" bson:"packingList"`
	ShippingProvider       string         `json:"shippingProvider" bson:"shippingProvider"`
	DispatchDate           time.Time      `json:"dispatchDate" bson:"dispatchDate"`
	WarehouseRecievingDate time.Time      `json:"warehouseRecievingDate" bson:"warehouseRecievingDate"`
	FlightNumber           string         `json:"flightNumber" bson:"flightNumber"`
	DeliveryDate           time.Time      `json:"deliveryDate" bson:"deliveryDate"`
	Status                 string         `json:"status" bson:"status" validate:"required,oneof=InTransit RecievedByWH FlightBoarded Delivered"`
	InrConversionRate      float64        `json:"inrConversionRate" bson:"inrConversionRate" validate:"gte=0"`
	InternalLogisticsCost  float64        `json:"internalLogisticsCost" bson:"internalLogisticsCost" validate:"gte=0"`
}
