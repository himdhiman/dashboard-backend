package utils

import (
	"context"
	"fmt"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/services/dashboard-service/pkg/config"
)

func GetMongoClient(ctx context.Context, projectConfig *config.ProjectConfig, logger logger.ILogger) (mongo.IMongoClient, error) {
	mongoConnectString := fmt.Sprintf("%s:%s@%s", projectConfig.MongoUser, projectConfig.MongoPassword, projectConfig.MongoHost)
	mongoConfig := mongo.NewMongoConfig(fmt.Sprintf("mongodb://%s:27017", mongoConnectString), "Dashboard")
	mongoClient, err := mongo.NewMongoClient(mongoConfig, logger)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		return nil, err
	}

	return mongoClient, nil
}
