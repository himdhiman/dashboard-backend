package mappers

import (
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

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
