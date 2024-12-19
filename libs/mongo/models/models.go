package models

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoCollection struct {
	Collection *mongo.Collection
}

type UpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedCount int64
	UpsertedID    interface{}
}

type FindOptions struct {
	Limit      int64
	Skip       int64
	Sort       interface{}
	Projection interface{}
}

// PaginationOptions represents pagination settings
type PaginationOptions struct {
	Page     int64
	PageSize int64
}

// FilterOptions represents a basic filter
type FilterOptions map[string]interface{}
