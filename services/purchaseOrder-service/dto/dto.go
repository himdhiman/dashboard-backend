package dto

type CreatePurchaseOrderDTO struct {
	Vendor                string  `json:"vendor" binding:"required"`
	Deposits              float64 `json:"deposits" binding:"gte=0"`
	TentativeDispatchDate string  `json:"tentativeDispatchDate" binding:"required"`
	OrderType             string  `json:"orderType" binding:"required,oneof=new repeat"`
	Remarks               string  `json:"remarks" binding:"omitempty"`
}

type CreatePurchaseOrderResponse struct {
	ID       string `json:"id"`
	PONumber string `json:"poNumber"`
}

type CreatePurchaseOrderProductResponse struct {
	ID string `json:"id"`
}

type CreatePurchaseOrderProductDTO struct {
	ProductID       string  `json:"productID" binding:"required"`
	Quantity        int     `json:"quantity" binding:"required"`
	CurrentRMBPrice float64 `json:"currentRMBPrice" binding:"required"`
	Status          string  `json:"status" binding:"required,oneof=pending final"`
	Remarks         string  `json:"remarks" binding:"omitempty"`
	ShippingMark    string  `json:"shippingMark" binding:"omitempty"`
	OrderDate       string  `json:"orderDate" binding:"omitempty"`
}

type GetPurchaseOrderDTO struct {
	ID       string   `json:"id" binding:"required"`
	Products []string `json:"products"`
	CreatePurchaseOrderDTO
}

type ListPurchaseOrdersDTO struct {
	ID       string `json:"id"`
	PONumber string `json:"poNumber"`
	Vendor   string `json:"vendor"`
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
