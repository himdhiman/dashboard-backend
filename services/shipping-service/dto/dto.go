package dto

type ListShippingMarksDTO struct {
	ID                     string  `json:"id"`
	ShippingMark           string  `json:"shippingMark"`
	VendorName             string  `json:"vendorName"`
	NumberOfCartons        int     `json:"numberOfCartons"`
	Weight                 float64 `json:"weight"`
	ShippingAgent          string  `json:"shippingAgent"`
	DispatchDate           string  `json:"dispatchDate"`
	WarehouseRecievingDate string  `json:"warehouseRecievingDate"`
	FlightNumber           string  `json:"flightNumber"`
	DeliveryDate           string  `json:"deliveryDate"`
	TotalUnits             int     `json:"totalUnits"`
	TotalRMBCost           float64 `json:"totalRMBCost"`
	TotalINRCost           float64 `json:"totalINRCost"`
	Status                 string  `json:"status"`
}

type ShippingMarkProductDTO struct {
	ID              string  `json:"id" binding:"required"`
	SKUCode         string  `json:"skuCode" binding:"required"`
	ImageURL        string  `json:"imageUrl" binding:"required"`
	ProductID       string  `json:"productID" binding:"required"`
	Quantity        int     `json:"quantity" binding:"required"`
	CurrentRMBPrice float64 `json:"currentRMBPrice" binding:"required"`
	Status          string  `json:"status" binding:"required,oneof=pending final"`
	Remarks         string  `json:"remarks" binding:"omitempty"`
	ShippingMark    string  `json:"shippingMark" binding:"omitempty"`
	OrderDate       string  `json:"orderDate" binding:"omitempty"`
}

type ShippingMarkDTO struct {
	ID                     string                   `json:"id"`
	ShippingMark           string                   `json:"shippingMark"`
	NumberOfCartons        int                      `json:"numberOfCartons"`
	Weight                 float64                  `json:"weight"`
	ShippingAgent          string                   `json:"shippingAgent"`
	DispatchDate           string                   `json:"dispatchDate"`
	WarehouseRecievingDate string                   `json:"warehouseRecievingDate"`
	FlightNumber           string                   `json:"flightNumber"`
	DeliveryDate           string                   `json:"deliveryDate"`
	Products               []ShippingMarkProductDTO `json:"products"`
	TotalUnits             int                      `json:"totalUnits"`
	TotalRMBCost           float64                  `json:"totalRMBCost"`
	TotalINRCost           float64                  `json:"totalINRCost"`
}
