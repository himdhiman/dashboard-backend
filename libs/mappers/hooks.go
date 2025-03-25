package mappers

import (
	"fmt"
	"reflect"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/mitchellh/mapstructure"
)

// DecodeTimeHookFunc is a hook function to decode time strings into time.Time.
func DecodeTimeHookFunc() mapstructure.DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f == reflect.TypeOf(time.Time{}) && t == reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		var parsedTime time.Time
		var err error

		switch f.Kind() {
		case reflect.String:
			parsedTime, err = time.Parse(time.RFC3339, data.(string))
			if err != nil {
				return nil, err
			}
			return parsedTime, nil
		default:
			return data, nil
		}
	}
}

func EncodeTimeToStringHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		// Handle empty maps that sometimes appear instead of time.Time
		if f.Kind() == reflect.Map && f.Key().Kind() == reflect.String {
			m, ok := data.(map[string]interface{})
			if ok && len(m) == 0 {
				// Return a default value – here, we use the zero time formatted
				return time.Time{}.Format(time.RFC3339), nil
			}
		}

		// Handle pointers to time.Time by dereferencing them
		if f.Kind() == reflect.Ptr && f.Elem() == reflect.TypeOf(time.Time{}) {
			val := reflect.ValueOf(data)
			if val.IsNil() {
				return "", nil
			}
			// Dereference and update the source type
			data = val.Elem().Interface()
			f = reflect.TypeOf(data)
		}

		// If the source is not a time.Time, skip conversion
		if f != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		// Ensure target type is string
		if t != reflect.TypeOf("") {
			return data, nil
		}

		// Perform the conversion
		timeValue, ok := data.(time.Time)
		if !ok {
			return nil, fmt.Errorf("expected time.Time but got %T", data)
		}
		return timeValue.Format(time.RFC3339), nil
	}
}


// DecodeObjectIDHookFunc is a hook function to decode mongo.ObjectID into a string.
func DecodeObjectIDHookFunc() mapstructure.DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		// Check if the target type is string
		if t != reflect.TypeOf("") {
			return data, nil
		}

		// Check if the source type is mongo.ObjectID
		if f == reflect.TypeOf(mongo.ObjectID{}) {
			// Convert ObjectID to its hex string representation
			objectID, ok := data.(mongo.ObjectID)
			if !ok {
				return nil, fmt.Errorf("expected mongo.ObjectID but got %T", data)
			}
			return objectID.Hex(), nil
		}

		return data, nil
	}
}
