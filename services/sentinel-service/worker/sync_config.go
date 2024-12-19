package worker

import (
	"context"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
)

func StartConfigSync(mongoCollection *mongo_models.MongoCollection, logger logger.LoggerInterface, interval time.Duration) {
	// ticker := time.NewTicker(interval)

	mongoRepo := repository.Repository[models.APIConfig]{Collection: mongoCollection}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := map[string]interface{}{}
	configs, err := mongoRepo.Find(ctx, filter)
	if err != nil {
		// log error
		return
	}

	for _, config := range configs {
		logger.Info("Config", "config", config)
	}

}
