package repository

import (
	"context"
	"errors"

	"github.com/himdhiman/dashboard-backend/libs/mongo/mappers"
	"github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IRepository[T any] interface {
	CreateIndex(ctx context.Context, keys bson.D, unique bool) error

	Create(ctx context.Context, data *T) (string, error)
	FindByID(ctx context.Context, id string) (*T, error)
	Find(ctx context.Context, filter map[string]interface{}, opts ...*models.FindOptions) ([]*T, error)
	Update(ctx context.Context, filter map[string]interface{}, update interface{}) (*models.UpdateResult, error)
	Delete(ctx context.Context, filter map[string]interface{}) (int64, error)
}

type Repository[T any] struct {
	IRepository[T]
	Collection *models.MongoCollection
}

// NewRepository initializes a new repository
func NewRepository[T any](collection *models.MongoCollection) IRepository[T] {
	return &Repository[T]{Collection: collection}
}

// CreateIndex creates a compound index
func (r *Repository[T]) CreateIndex(ctx context.Context, keys bson.D, unique bool) error {
	index := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(unique),
	}
	_, err := r.Collection.Collection.Indexes().CreateOne(ctx, index)
	return err
}

// Create adds a document to the collection
func (r *Repository[T]) Create(ctx context.Context, data *T) (string, error) {
	result, err := r.Collection.Collection.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("failed to parse inserted ID")
	}
	return id.Hex(), nil
}

// FindByID retrieves a document by its ID
func (r *Repository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}
	var result T
	if err := r.Collection.Collection.FindOne(ctx, filter).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Find retrieves documents matching a filter
func (r *Repository[T]) Find(ctx context.Context, filter map[string]interface{}, opts ...*models.FindOptions) ([]*T, error) {
	bsonFilters := mappers.MapToBson(filter)
	mongoFindOptions := mappers.MapFindOptions(opts...)
	cursor, err := r.Collection.Collection.Find(ctx, bsonFilters, mongoFindOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*T
	for cursor.Next(ctx) {
		var item T
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		results = append(results, &item)
	}
	return results, nil
}

// Update updates documents matching the filter
func (r *Repository[T]) Update(ctx context.Context, filter map[string]interface{}, update interface{}) (*models.UpdateResult, error) {
	bsonFilters := mappers.MapToBson(filter)
	result, err := r.Collection.Collection.UpdateMany(ctx, bsonFilters, update)
	if err != nil {
		return nil, err
	}
	return mappers.MapUpdateResult(result), nil
}

// Delete removes documents matching the filter
func (r *Repository[T]) Delete(ctx context.Context, filter map[string]interface{}) (int64, error) {
	bsonFilters := mappers.MapToBson(filter)
	result, err := r.Collection.Collection.DeleteMany(ctx, bsonFilters)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}
