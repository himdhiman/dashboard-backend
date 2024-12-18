package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type TestDocument struct {
	ID    string `bson:"_id,omitempty"`
	Name  string `bson:"name"`
	Email string `bson:"email"`
}

func setupTestRepository() (*Repository, *MongoDB, func()) {
	mongoURI := "mongodb://localhost:27017"
	dbName := "testDB"

	// Initialize MongoDB
	mongoClient, _ := NewMongoDB(mongoURI, dbName)

	// Create test repository
	repo := NewRepository(mongoClient.Db, "testCollection")

	// Cleanup function to drop the collection
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		mongoClient.Db.Collection("testCollection").Drop(ctx)
		mongoClient.Close()
	}

	return repo, mongoClient, cleanup
}

func TestRepositoryCRUD(t *testing.T) {
	repo, _, cleanup := setupTestRepository()
	defer cleanup()

	ctx := context.Background()

	// Test Create
	t.Run("Create Document", func(t *testing.T) {
		doc := TestDocument{Name: "John Doe", Email: "john.doe@example.com"}
		result, err := repo.Create(ctx, doc)
		assert.NoError(t, err)
		assert.NotNil(t, result.InsertedID)
	})

	// Test Read
	t.Run("Read Documents", func(t *testing.T) {
		var results []TestDocument
		err := repo.Read(ctx, bson.M{}, &results)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
	})

	// Test Update
	t.Run("Update Document", func(t *testing.T) {
		filter := bson.M{"name": "John Doe"}
		update := bson.M{"email": "updated.email@example.com"}
		_, err := repo.Update(ctx, filter, update)
		assert.NoError(t, err)

		var updatedDoc TestDocument
		err = repo.FindOne(ctx, filter, &updatedDoc)
		assert.NoError(t, err)
		assert.Equal(t, "updated.email@example.com", updatedDoc.Email)
	})

	// Test FindOne
	t.Run("FindOne Document", func(t *testing.T) {
		filter := bson.M{"name": "John Doe"}
		var result TestDocument
		err := repo.FindOne(ctx, filter, &result)
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", result.Name)
	})

	// Test Count
	t.Run("Count Documents", func(t *testing.T) {
		count, err := repo.Count(ctx, bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	// Test Delete
	t.Run("Delete Document", func(t *testing.T) {
		filter := bson.M{"name": "John Doe"}
		_, err := repo.Delete(ctx, filter)
		assert.NoError(t, err)

		var results []TestDocument
		err = repo.Read(ctx, bson.M{}, &results)
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})
}
