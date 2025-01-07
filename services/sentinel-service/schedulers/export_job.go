package schedulers

import (
	"context"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/scheduler"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/services"
)

type ExportJobScheduler struct {
	scheduler *scheduler.Scheduler
	service   *services.UnicommerceService
	logger    logger.ILogger
}

func NewExportJobScheduler(collection *models.MongoCollection, service *services.UnicommerceService, logger logger.ILogger) *ExportJobScheduler {
	config := scheduler.SchedulerConfig{
		RetentionPeriod: 24 * time.Hour,
		Collection:      collection,
		Logger:          logger,
	}

	return &ExportJobScheduler{
		scheduler: scheduler.NewScheduler(config),
		service:   service,
		logger:    logger,
	}
}

func (e *ExportJobScheduler) Start(ctx context.Context) error {
	config := scheduler.JobConfig{
		Name:        "check-export-status",
		CronExpr:    "0 */5 * * * *", // Every 30 minutes
		Params:      map[string]interface{}{},
		MaxRetries:  3,
		IsRecurring: true,
	}

	err := e.scheduler.Schedule(ctx, config, func(ctx context.Context, params map[string]interface{}) error {
		return e.service.CheckExportJobStatus(ctx)
	})

	if err != nil {
		e.logger.Error("Failed to schedule export job", "error", err)
		return err
	}

	e.scheduler.Start()
	return nil
}

func (e *ExportJobScheduler) Stop() {
	e.scheduler.Stop()
}
