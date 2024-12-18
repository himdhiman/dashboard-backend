package mongo

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/himdhiman/dashboard-backend/libs/logger"
)


// MongoDB holds the MongoDB connection instance
type MongoDB struct {
	Client *mongo.Client
	Db     *mongo.Database
	Logger logger.LoggerInterface
}

// NewMongoDB initializes the MongoDB connection
func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	// Set up a MongoDB client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the MongoDB server
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Connected to MongoDB")
	db := client.Database(dbName)

	return &MongoDB{
		Client: client,
		Db:     db,
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close() {
	if err := m.Client.Disconnect(context.Background()); err != nil {
		log.Printf("Error disconnecting MongoDB: %v", err)
	} else {
		log.Println("MongoDB connection closed")
	}
}
