package schedulers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/scheduler"
	"github.com/himdhiman/dashboard-backend/services/products-service/services"
)

type ExportJobScheduler struct {
	scheduler *scheduler.Scheduler
	service   *services.UnicommerceProductsService
	logger    logger.ILogger
	jobCodes  []string
	cronExpr  string
}

func NewExportJobScheduler(collection *models.MongoCollection, service *services.UnicommerceProductsService, logger logger.ILogger, jobCodes []string, cronExpr string) *ExportJobScheduler {
	config := scheduler.SchedulerConfig{
		RetentionPeriod: 24 * time.Hour,
		Collection:      collection,
		Logger:          logger,
	}

	return &ExportJobScheduler{
		scheduler: scheduler.NewScheduler(config),
		service:   service,
		logger:    logger,
		jobCodes:  jobCodes,
		cronExpr:  cronExpr,
	}
}

func (e *ExportJobScheduler) Start(ctx context.Context) error {
	for _, jobCode := range e.jobCodes {
		jobConfig := scheduler.JobConfig{
			Name:        "check-export-status-" + jobCode,
			CronExpr:    e.cronExpr,
			Params:      map[string]interface{}{"exportJobCode": jobCode},
			MaxRetries:  3,
			IsRecurring: true,
		}

		e.logger.Info("Configuring scheduled job", "jobName", jobConfig.Name, "cronExpr", jobConfig.CronExpr)

		err := e.scheduler.Schedule(ctx, jobConfig, func(ctx context.Context, params map[string]interface{}) error {
			correlationID := uuid.New().String()
			e.logger.Info("Starting scheduled job", "correlationID", correlationID)
			ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

			exportJobCode, _ := params["exportJobCode"].(string)
			e.logger.Info("Calling CheckExportJobStatus", "correlationID", correlationID, "exportJobCode", exportJobCode)
			err := e.service.CheckExportJobStatus(ctx, exportJobCode)
			if err != nil {
				e.logger.Error("CheckExportJobStatus failed", "correlationID", correlationID, "error", err)
				return err
			}

			e.logger.Info("CheckExportJobStatus succeeded", "correlationID", correlationID)
			return nil
		})

		if err != nil {
			e.logger.Error("Failed to schedule export job", "error", err)
			return err
		}
	}

	e.scheduler.Start()
	return nil
}

func (e *ExportJobScheduler) Stop() {
	e.scheduler.Stop()
}
