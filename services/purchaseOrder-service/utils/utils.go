package utils

var AllowedFields = map[string]bool{
	"OrderStatus":           true,
	"TentativeDispatchDate": true,
	"Remarks":               true,
	"Deposits":              true,
	"OrderType":             true,
}

var AllowedProductFields = map[string]bool{
	"Quantity":        true,
	"CurrentRMBPrice": true,
	"Status":          true,
	"Remarks":         true,
	"ShippingMark":    true,
	"OrderDate":       true,
}
