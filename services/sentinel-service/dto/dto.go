package dto

type CreatePurchaseOrderDTO struct {
	Vendor                string                    `json:"vendor" binding:"required"`
	TotalAmount           float64                   `json:"totalAmount" binding:"gte=0"`
	Deposits              float64                   `json:"deposits" binding:"gte=0"`
	OrderStatus           string                    `json:"orderStatus" binding:"required"`
	TentativeDispatchDate string                    `json:"tentativeDispatchDate" bson:"tentativeDispatchDate" binding:"required"`
	OrderType             string                    `json:"orderType" binding:"required"`
	Products              []PurchaseOrderProductDTO `json:"products" binding:"required,dive"`
	Remarks               string                    `json:"remarks" binding:"omitempty"`
}

type PurchaseOrderProductDTO struct {
	SkuCode         string  `json:"skuCode" binding:"required"`
	ImageURL        string  `json:"imageUrl" binding:"required"`
	Quantity        int     `json:"quantity" binding:"required"`
	CurrentRMBPrice float64 `json:"currentRMBPrice" binding:"required"`
	Status          string  `json:"status" binding:"required"`
	Remarks         string  `json:"remarks" binding:"omitempty"`
	ShippingMark    string  `json:"shippingMark" binding:"omitempty"`
}
