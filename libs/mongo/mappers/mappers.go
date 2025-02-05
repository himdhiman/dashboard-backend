package mappers

import (
	"github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MapFindOptions maps custom FindOptions to mongo.FindOptions
func MapFindOptions(opts ...*models.FindOptions) *options.FindOptions {
	mongoFindOptions := options.Find()
	if len(opts) > 0 {
		opt := opts[0]
		mapstructure.Decode(opt, mongoFindOptions)
	}
	return mongoFindOptions
}

func MapFindOneOptions(opts ...*models.FindOptions) *options.FindOneOptions {
	mongoFindOneOptions := options.FindOne()
	if len(opts) > 0 {
		opt := opts[0]
		mapstructure.Decode(opt, mongoFindOneOptions)
	}
	return mongoFindOneOptions
}

// MapUpdateResult maps mongo.UpdateResult to custom UpdateResult
func MapUpdateResult(mongoResult *mongo.UpdateResult) *models.UpdateResult {
	var updateResult models.UpdateResult
	mapstructure.Decode(mongoResult, &updateResult)
	return &updateResult
}

// MapToBson converts a map[string]interface{} to bson.M
func MapToBson(filter map[string]interface{}) bson.M {
	if id, ok := filter["_id"]; ok {
		if oid, ok := id.(string); ok {
			if objID, err := primitive.ObjectIDFromHex(oid); err == nil {
				filter["_id"] = objID
			}
		}
	}
	return bson.M(filter)
}
