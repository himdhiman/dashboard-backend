package main

import (
	"context"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/libs/mongo/models"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/worker"
)

func main() {

	logger := logger.New(logger.DefaultConfig())

	ctx := context.Background()

	mongoConfig := models.Config{
		MongoURL:     "mongodb://localhost:27017",
		DatabaseName: "Dashboard",
	}

	mongoClient, err := mongo.NewMongoClient(mongoConfig, logger)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		return
	}

	defer mongoClient.Disconnect(ctx)

	collectionName := "sentinel_apis"
	collection, err := mongoClient.GetCollection(context.Background(), collectionName)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}

	worker.StartConfigSync(collection, logger, 10)
}
