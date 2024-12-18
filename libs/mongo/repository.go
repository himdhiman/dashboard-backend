package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repository represents the generic CRUD operations
type Repository struct {
	Collection *mongo.Collection
}

// NewRepository creates a new repository for a MongoDB collection
func NewRepository(db *mongo.Database, collectionName string) *Repository {
	return &Repository{
		Collection: db.Collection(collectionName),
	}
}

// Create inserts a single document into the collection
func (r *Repository) Create(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	return r.Collection.InsertOne(ctx, document)
}

// Read retrieves documents based on a filter
func (r *Repository) Read(ctx context.Context, filter interface{}, result interface{}) error {
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode all documents into the result
	if err := cursor.All(ctx, result); err != nil {
		return fmt.Errorf("failed to decode documents: %w", err)
	}

	return nil
}

// Update updates documents based on a filter
func (r *Repository) Update(ctx context.Context, filter, update interface{}) (*mongo.UpdateResult, error) {
	return r.Collection.UpdateMany(ctx, filter, bson.M{"$set": update})
}

// Delete deletes documents based on a filter
func (r *Repository) Delete(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return r.Collection.DeleteMany(ctx, filter)
}

// FindOne retrieves a single document based on a filter
func (r *Repository) FindOne(ctx context.Context, filter interface{}, result interface{}) error {
	err := r.Collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return fmt.Errorf("failed to find document: %w", err)
	}
	return nil
}

// Count returns the number of documents matching a filter
func (r *Repository) Count(ctx context.Context, filter interface{}) (int64, error) {
	return r.Collection.CountDocuments(ctx, filter)
}
