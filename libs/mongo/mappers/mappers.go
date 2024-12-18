package mappers

import (
	"github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
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

// MapUpdateResult maps mongo.UpdateResult to custom UpdateResult
func MapUpdateResult(mongoResult *mongo.UpdateResult) *models.UpdateResult {
	var updateResult models.UpdateResult
	mapstructure.Decode(mongoResult, &updateResult)
	return &updateResult
}

// MapToBson converts a map[string]interface{} to bson.M
func MapToBson(filter map[string]interface{}) bson.M {
	return bson.M(filter)
}
