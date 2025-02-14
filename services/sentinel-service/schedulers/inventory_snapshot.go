package schedulers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/scheduler"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/constants"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/services"
)

type InventorySnapShotScheduler struct {
	scheduler *scheduler.Scheduler
	service   *services.UnicommerceService
	logger    logger.ILogger
}

func NewInventorySnapShotScheduler(collection *models.MongoCollection, service *services.UnicommerceService, logger logger.ILogger) *InventorySnapShotScheduler {
	config := scheduler.SchedulerConfig{
		RetentionPeriod: 24 * time.Hour,
		Collection:      collection,
		Logger:          logger,
	}

	return &InventorySnapShotScheduler{
		scheduler: scheduler.NewScheduler(config),
		service:   service,
		logger:    logger,
	}
}

func (e *InventorySnapShotScheduler) Start(ctx context.Context) error {
	config := scheduler.JobConfig{
		Name:        "snapshot-inventory",
		CronExpr:    "0 */30 * * * *", // For every 30 minutes
		Params:      map[string]interface{}{},
		MaxRetries:  3,
		IsRecurring: true,
	}

	e.logger.Info("Configuring scheduled job", "jobName", config.Name, "cronExpr", config.CronExpr)

	err := e.scheduler.Schedule(ctx, config, func(ctx context.Context, params map[string]interface{}) error {
		correlationID := uuid.New().String()
		e.logger.Info("Starting scheduled job", "correlationID", correlationID)
		ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

		e.logger.Info("Calling UpdateInventoryFromGoogleSheet", "correlationID", correlationID)
		err := e.service.UpdateInventoryFromGoogleSheet(ctx)
		if err != nil {
			e.logger.Error("UpdateInventoryFromGoogleSheet failed", "correlationID", correlationID, "error", err)
			return err
		}

		e.logger.Info("UpdateInventoryFromGoogleSheet succeeded", "correlationID", correlationID)
		return nil
	})

	if err != nil {
		e.logger.Error("Failed to schedule export job", "error", err)
		return err
	}

	e.logger.Info("Starting scheduler")
	e.scheduler.Start()
	e.logger.Info("Scheduler started successfully")
	return nil
}

func (e *InventorySnapShotScheduler) Stop() {
	e.scheduler.Stop()
}
