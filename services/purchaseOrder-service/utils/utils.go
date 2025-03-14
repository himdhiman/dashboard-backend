package utils

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func IsAllowedField(fieldPath string) bool {

	var allowedFields = map[string]bool{
		"OrderStatus":           true,
		"TotalAmount":           true,
		"TentativeDispatchDate": true,
		"Remarks":               true,
		"Deposits":              true,
		"OrderType":             true,
	}

	var allowedProductFields = map[string]bool{
		"SkuCode":         true,
		"ImageURL":        true,
		"Quantity":        true,
		"CurrentRMBPrice": true,
		"Status":          true,
		"Remarks":         true,
		"ShippingMark":    true,
		"OrderDate":       true,
	}

	fields := strings.Split(fieldPath, ".")
	if len(fields) == 1 {
		return allowedFields[fields[0]]
	} else if len(fields) == 3 && fields[0] == "Products" {
		return allowedProductFields[fields[2]]
	}
	return false
}

// setField navigates through obj based on the dot-separated fieldPath.
// It supports nested fields and has special handling for "Products" where the second token is the SKU.
// If the SKU is new, a new product entry is created, its SkuCode is set, and it is appended to the purchase order.
func SetField(obj interface{}, fieldPath string, value interface{}) error {
	fields := strings.Split(fieldPath, ".")
	// Start with the base object; we assume obj is a pointer.
	v := reflect.ValueOf(obj).Elem()

	// Process the field tokens one by one.
	for len(fields) > 0 {
		field := fields[0]
		fields = fields[1:]

		if field == "Products" {
			// Next token must be the SKU.
			if len(fields) < 2 {
				return fmt.Errorf("invalid field path for Products: %s", fieldPath)
			}
			sku := fields[0]
			// Consume the SKU token.
			fields = fields[1:]
			productsField := v.FieldByName("Products")
			if !productsField.IsValid() {
				return fmt.Errorf("no such field: Products")
			}
			if productsField.Kind() != reflect.Slice {
				return fmt.Errorf("products field is not a slice")
			}
			var product reflect.Value
			found := false
			// Search for an existing product with the given SKU.
			for i := 0; i < productsField.Len(); i++ {
				candidate := productsField.Index(i)
				if candidate.FieldByName("SkuCode").String() == sku {
					product = candidate
					found = true
					break
				}
			}
			if !found {
				// Create a new product instance.
				elemType := productsField.Type().Elem()
				newProduct := reflect.New(elemType).Elem()
				// Set the SkuCode on the new product.
				skuField := newProduct.FieldByName("SkuCode")
				if skuField.IsValid() && skuField.CanSet() && skuField.Kind() == reflect.String {
					skuField.SetString(sku)
				} else {
					return fmt.Errorf("cannot set SkuCode on new product for SKU %s", sku)
				}
				// Append the new product to the Products slice.
				newSlice := reflect.Append(productsField, newProduct)
				productsField.Set(newSlice)
				// Retrieve the newly added product.
				product = newSlice.Index(newSlice.Len() - 1)
			}
			// Now continue updating within the found or newly created product.
			v = product
		} else {
			// If no further tokens, then this field should be set.
			if len(fields) == 0 {
				f := v.FieldByName(field)
				if !f.IsValid() {
					return fmt.Errorf("no such field: %s in object", field)
				}
				if !f.CanSet() {
					return fmt.Errorf("cannot set field %s", field)
				}
				val := reflect.ValueOf(value)
				// Special handling for time.Time fields.
				if f.Type() == reflect.TypeOf(time.Time{}) {
					str, ok := value.(string)
					if !ok {
						return fmt.Errorf("expected string value for time field %s", field)
					}
					parsedTime, err := time.Parse(time.RFC3339, str)
					if err != nil {
						return fmt.Errorf("error parsing time for field %s: %v", field, err)
					}
					val = reflect.ValueOf(parsedTime)
				} else if f.Type() != val.Type() {
					return fmt.Errorf("provided value type didn't match field %s type: expected %s but got %s", field, f.Type(), val.Type())
				}
				f.Set(val)
				return nil
			} else {
				// Not the final field: move deeper into the object.
				v = v.FieldByName(field)
				if !v.IsValid() {
					return fmt.Errorf("no such field: %s in object", field)
				}
				// If v is a pointer, ensure it is non-nil.
				if v.Kind() == reflect.Ptr {
					if v.IsNil() {
						v.Set(reflect.New(v.Type().Elem()))
					}
					v = v.Elem()
				}
			}
		}
	}

	return nil
}
