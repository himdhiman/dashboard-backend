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
