package hooks

import (
	"context"
)

type IHook interface {
	BeforeCreate(ctx context.Context, doc interface{}) error
	BeforeUpdate(ctx context.Context, filter, update interface{}) error
}

// MongoIDToStringHook converts primitive.ObjectID to string
// func MongoIDToStringHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
// 	if f == reflect.TypeOf(primitive.ObjectID{}) && t == reflect.TypeOf("") {
// 		objID, ok := data.(primitive.ObjectID)
// 		if !ok {
// 			return nil, errors.New("failed to cast to ObjectID")
// 		}
// 		return objID.Hex(), nil
// 	}
// 	return data, nil
// }

// StringToMongoIDHook converts string to primitive.ObjectID
// func StringToMongoIDHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
// 	if f == reflect.TypeOf("") && t == reflect.TypeOf(primitive.ObjectID{}) {
// 		id, ok := data.(string)
// 		if !ok {
// 			return nil, errors.New("failed to cast to string")
// 		}
// 		objID, err := primitive.ObjectIDFromHex(id)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return objID, nil
// 	}
// 	return data, nil
// }
