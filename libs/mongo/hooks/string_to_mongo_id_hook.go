package hooks

import (
	"context"
	"errors"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StringToMongoIDHook struct {
	IHook
}

func (hook *StringToMongoIDHook) BeforeCreate(ctx context.Context, doc interface{}) error {
	return convertStringToMongoID(doc)
}

func (hook *StringToMongoIDHook) BeforeUpdate(ctx context.Context, filter, update interface{}) error {
	return convertStringToMongoID(update)
}

func convertStringToMongoID(doc interface{}) error {
	v := reflect.ValueOf(doc)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors.New("expected a struct or pointer to struct")
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && field.Type().Name() == "string" {
			if bsonID, err := primitive.ObjectIDFromHex(field.String()); err == nil {
				field.Set(reflect.ValueOf(bsonID))
			}
		}
	}

	return nil
}
