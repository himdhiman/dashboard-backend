package dto

type CreatePurchaseOrderDTO struct {
	Vendor                string  `json:"vendor" binding:"required"`
	TotalAmount           float64 `json:"totalAmount" binding:"gte=0"`
	Deposits              float64 `json:"deposits" binding:"gte=0"`
	OrderStatus           string  `json:"orderStatus" binding:"required"`
	TentativeDispatchDate string  `json:"tentativeDispatchDate" bson:"tentativeDispatchDate" binding:"required"`
	OrderType             string  `json:"orderType" binding:"required"`
	Remarks               string  `json:"remarks" binding:"omitempty"`
}

type CreatePurchaseOrderProductDTO struct {
	ProductID       string  `json:"productID" binding:"required"`
	Quantity        int     `json:"quantity" binding:"required"`
	CurrentRMBPrice float64 `json:"currentRMBPrice" binding:"required"`
	Status          string  `json:"status" binding:"required"`
	Remarks         string  `json:"remarks" binding:"omitempty"`
	ShippingMark    string  `json:"shippingMark" binding:"omitempty"`
	OrderDate       string  `json:"orderDate" binding:"omitempty"`
}

type PurchaseOrderProductDTO struct {
	ID       string `json:"id" binding:"required"`
	SKUCode  string `json:"skuCode" binding:"required"`
	ImageURL string `json:"imageUrl" binding:"required"`
	CreatePurchaseOrderProductDTO
}

type PurchaseOrderDTO struct {
	ID       string                    `json:"id" binding:"required"`
	PONumber string                    `json:"poNumber" binding:"required"`
	Products []PurchaseOrderProductDTO `json:"products" binding:"required"`
	CreatePurchaseOrderDTO
}
