package interfaces

import (
	"context"

	"github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"go.mongodb.org/mongo-driver/bson"
)

type IMongoClient interface {
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	GetCollection(ctx context.Context, name string) (*models.MongoCollection, error)
}

type IMongoRepository[T any] interface {
	CreateIndex(ctx context.Context, keys bson.D, unique bool) error

	Create(ctx context.Context, data *T) (string, error)
	FindByID(ctx context.Context, id string) (*T, error)
	Find(ctx context.Context, filter map[string]interface{}, opts ...*models.FindOptions) ([]*T, error)
	Update(ctx context.Context, filter map[string]interface{}, update interface{}) (*models.UpdateResult, error)
	Delete(ctx context.Context, filter map[string]interface{}) (int64, error)
}
