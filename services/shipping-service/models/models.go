package models

import "time"

type ShippingProvider struct {
	ProviderName string `json:"providerName" bson:"providerName" validate:"required"`
	ProviderCode string `json:"providerCode" bson:"providerCode" validate:"required"`
}

type ShippingMarkProduct struct {
}

type ShippingMarkDetails struct {
	ShippingMark           string           `json:"shippingMark" bson:"shippingMark" validate:"required"`
	NumberOfCartons        int              `json:"numberOfCartons" bson:"numberOfCartons" validate:"min=1"`
	Weight                 float64          `json:"weight" bson:"weight" validate:"min=0"`
	PackingList            string           `json:"packingList" bson:"packingList"`
	ShippingProvider       ShippingProvider `json:"shippingProvider" bson:"shippingProvider"`
	DispatchDate           time.Time        `json:"dispatchDate" bson:"dispatchDate"`
	WarehouseRecievingDate time.Time        `json:"warehouseRecievingDate" bson:"warehouseRecievingDate"`
	FlightNumber           string           `json:"flightNumber" bson:"flightNumber"`
	DeliveryDate           time.Time        `json:"deliveryDate" bson:"deliveryDate"`
}
