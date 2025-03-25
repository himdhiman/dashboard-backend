package utils

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

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

// this function will simply set the path of the object to the value, path will be in the format of "field1.field2.field3"
// if the path is not found in the object, it will return an error
func SetField(obj interface{}, fieldPath string, value interface{}, allowedFields map[string]bool) error {
	// Check if the fieldPath is in the allowed fields
	if !allowedFields[fieldPath] {
		return fmt.Errorf("field %s is not allowed to be updated", fieldPath)
	}

	// Get the value of the object
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("object must be a non-nil pointer")
	}

	// Dereference the pointer to get the underlying value
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("object must point to a struct")
	}

	// Split the field path into parts
	fields := strings.Split(fieldPath, ".")

	// Traverse the fields to get to the desired field
	for i, field := range fields {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				// Initialize the pointer if it's nil
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return fmt.Errorf("field %s is not a struct", strings.Join(fields[:i], "."))
		}

		v = v.FieldByName(field)
		if !v.IsValid() {
			return fmt.Errorf("field %s not found in struct", strings.Join(fields[:i+1], "."))
		}
	}

	// Set the value of the field
	if !v.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldPath)
	}

	// Handle time conversion if the field is of type time.Time
	if v.Type() == reflect.TypeOf(time.Time{}) {
		timeStr, ok := value.(string)
		if !ok {
			return fmt.Errorf("field %s expects a string for time conversion", fieldPath)
		}

		parsedTime, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return fmt.Errorf("field %s has invalid time format: %v", fieldPath, err)
		}

		v.Set(reflect.ValueOf(parsedTime))
		return nil
	}

	val := reflect.ValueOf(value)
	if val.Type() != v.Type() {
		return fmt.Errorf("value type %s does not match field type %s", val.Type(), v.Type())
	}

	v.Set(val)
	return nil
}
