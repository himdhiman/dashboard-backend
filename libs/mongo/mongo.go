package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo/helpers"
	"github.com/himdhiman/dashboard-backend/libs/mongo/models"
)

type IMongoClient interface {
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	GetCollection(ctx context.Context, name string) (*models.MongoCollection, error)
}

type MongoClient struct {
	IMongoClient
	Client   *mongo.Client
	Database *mongo.Database
	Logger   logger.LoggerInterface
}

func NewMongoConfig(mongoURL, databaseName string) *models.Config {
	return &models.Config{
		MongoURL:     mongoURL,
		DatabaseName: databaseName,
	}
}

// NewMongoClient initializes the MongoDB connection and returns a MongoClient instance
func NewMongoClient(config *models.Config, logger logger.LoggerInterface) (IMongoClient, error) {
	client := &MongoClient{Logger: logger}
	err := client.connect(context.Background(), config.MongoURL)
	if err != nil {
		return nil, err
	}
	logger.Info("MongoDB client initialized")
	err = client.getDatabase(context.Background(), config.DatabaseName)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// connect initializes the MongoDB connection
func (m *MongoClient) connect(ctx context.Context, uri string) error {
	// Set up a MongoDB client
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		m.Logger.Error("Failed to connect to MongoDB", "error", err)
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the MongoDB server
	if err := client.Ping(ctx, nil); err != nil {
		m.Logger.Error("Failed to ping MongoDB", "error", err)
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.Logger.Info("Connected to MongoDB")
	m.Client = client
	return nil
}

// Disconnect closes the MongoDB connection
func (m *MongoClient) Disconnect(ctx context.Context) error {
	if err := m.Client.Disconnect(ctx); err != nil {
		m.Logger.Error("Error disconnecting MongoDB", "error", err)
		return err
	}
	m.Logger.Info("MongoDB connection closed")
	return nil
}

// Ping checks the connection to MongoDB
func (m *MongoClient) Ping(ctx context.Context) error {
	if err := m.Client.Ping(ctx, nil); err != nil {
		m.Logger.Error("Failed to ping MongoDB", "error", err)
		return err
	}
	m.Logger.Info("MongoDB ping successful")
	return nil
}

// getDatabase validates and returns a MongoDB database instance
func (m *MongoClient) getDatabase(ctx context.Context, name string) error {
	if !helpers.IsValidDatabaseName(name) {
		return fmt.Errorf("invalid database name")
	}

	// Check if the database exists
	databases, err := m.Client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to list databases: %w", err)
	}

	found := false
	for _, dbName := range databases {
		if dbName == name {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("database %s does not exist", name)
	}

	m.Database = m.Client.Database(name)
	m.Logger.Info("MongoDB database initialized successfully for database", "database", name)
	return nil
}

// GetCollection validates and returns a MongoCollection instance
func (m *MongoClient) GetCollection(ctx context.Context, collection string) (*models.MongoCollection, error) {
	if !helpers.IsValidCollectionName(collection) {
		return nil, fmt.Errorf("invalid collection name")
	}

	// Check if the collection exists
	collections, err := m.Database.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	found := false
	for _, collName := range collections {
		if collName == collection {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("collection %s does not exist", collection)
	}

	m.Logger.Info("MongoDB collection initialized successfully for collection", "collection", collection)
	return &models.MongoCollection{Collection: m.Database.Collection(collection)}, nil
}
