package hooks

import (
    "errors"
    "reflect"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

func MongoIDToStringHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
    if f == reflect.TypeOf(primitive.ObjectID{}) && t == reflect.TypeOf("") {
        objID, ok := data.(primitive.ObjectID)
        if !ok {
            return nil, errors.New("failed to cast to ObjectID")
        }
        return objID.Hex(), nil
    }
    return data, nil
}
