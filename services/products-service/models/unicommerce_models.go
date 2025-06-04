package models

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

type UnicommerceError struct {
	Code        int                    `json:"code" bson:"code"`
	FieldName   string                 `json:"fieldName" bson:"fieldName"`
	Description string                 `json:"description" bson:"description"`
	Message     string                 `json:"message" bson:"message"`
	ErrorParams map[string]interface{} `json:"errorParams" bson:"errorParams"`
}

type UnicommerceWarning struct {
	Code        int    `json:"code" bson:"code"`
	Message     string `json:"message" bson:"message"`
	Description string `json:"description" bson:"description"`
}

type UnicommerceInventoryAdjustmentResponse struct {
	Successful                   bool                 `json:"successful" bson:"successful"`
	Message                      string               `json:"message" bson:"message"`
	Errors                       []UnicommerceError   `json:"errors" bson:"errors"`
	Warnings                     []UnicommerceWarning `json:"warnings" bson:"warnings"`
	InventoryAdjustmentResponses []struct {
		Successful bool               `json:"successful" bson:"successful"`
		Errors     []UnicommerceError `json:"errors" bson:"errors"`
	} `json:"inventoryAdjustmentResponses" bson:"inventoryAdjustmentResponses"`
}
