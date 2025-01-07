package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	"github.com/robfig/cron/v3"
)

type SchedulerConfig struct {
	RetentionPeriod time.Duration
	Collection      *models.MongoCollection
	Logger          logger.ILogger
}

type Scheduler struct {
	cron            *cron.Cron
	jobs            map[string]cron.EntryID
	jobRepo         repository.IRepository[Job]
	logger          logger.ILogger
	mu              sync.RWMutex
	retentionPeriod time.Duration
}

func NewScheduler(config SchedulerConfig) *Scheduler {
	jobRepo := repository.Repository[Job]{Collection: config.Collection}
	scheduler := &Scheduler{
		cron:            cron.New(cron.WithSeconds()),
		jobs:            make(map[string]cron.EntryID),
		jobRepo:         &jobRepo,
		logger:          config.Logger,
		retentionPeriod: config.RetentionPeriod,
	}

	// Schedule cleanup job at midnight
	scheduler.cron.AddFunc("0 0 * * *", func() {
		scheduler.cleanup(context.Background())
	})

	return scheduler
}

func (s *Scheduler) cleanup(ctx context.Context) {
	cutoff := time.Now().Add(-s.retentionPeriod)
	filter := map[string]interface{}{
		"updated_at": map[string]interface{}{
			"$lt": cutoff,
		},
	}

	_, err := s.jobRepo.Delete(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to cleanup old jobs", "error", err)
	}
}

func (s *Scheduler) Schedule(ctx context.Context, config JobConfig, jobFunc func(context.Context, map[string]interface{}) error) error {
	job := &Job{
		Name:        config.Name,
		Status:      JobStatusPending,
		CronExpr:    config.CronExpr,
		Params:      config.Params,
		MaxRetries:  config.MaxRetries,
		IsRecurring: config.IsRecurring,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save job to database
	id, err := s.jobRepo.Create(ctx, job)
	if err != nil {
		return err
	}

	// Create cron job
	entryID, err := s.cron.AddFunc(config.CronExpr, func() {
		s.executeJob(ctx, id, jobFunc)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.jobs[id] = entryID
	s.mu.Unlock()

	return nil
}

func (s *Scheduler) executeJob(ctx context.Context, jobID string, jobFunc func(context.Context, map[string]interface{}) error) {
	job, err := s.GetJobStatus(ctx, jobID)
	if err != nil {
		s.logger.Error("Failed to get job details", "error", err)
		return
	}

	_, err = s.jobRepo.Update(ctx, map[string]interface{}{"_id": jobID}, map[string]interface{}{
		"status":      JobStatusRunning,
		"last_run_at": time.Now(),
		"updated_at":  time.Now(),
	})
	if err != nil {
		s.logger.Error("Failed to update job status", "error", err)
		return
	}

	err = jobFunc(ctx, job.Params)
	updateFields := map[string]interface{}{
		"updated_at":  time.Now(),
		"next_run_at": s.getNextRunTime(job.CronExpr),
	}

	if err != nil {
		updateFields["status"] = JobStatusFailed
		updateFields["error"] = err.Error()
		updateFields["retry_count"] = job.RetryCount + 1

		if job.RetryCount >= job.MaxRetries && !job.IsRecurring {
			s.removeJob(ctx, jobID)
			return
		}
	} else {
		updateFields["status"] = JobStatusCompleted
		updateFields["error"] = ""
		updateFields["retry_count"] = 0
	}

	_, err = s.jobRepo.Update(ctx, map[string]interface{}{"_id": jobID}, updateFields)
	if err != nil {
		s.logger.Error("Failed to update job status", "error", err)
	}
}

func (s *Scheduler) removeJob(ctx context.Context, jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.jobs[jobID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, jobID)
	}

	_, err := s.jobRepo.Delete(ctx, map[string]interface{}{"_id": jobID})
	return err
}

func (s *Scheduler) GetJobStatus(ctx context.Context, jobID string) (*Job, error) {
	jobs, err := s.jobRepo.Find(ctx, map[string]interface{}{"_id": jobID}, nil)
	if err != nil || len(jobs) == 0 {
		return nil, err
	}
	return jobs[0], nil
}

func (s *Scheduler) ListJobs(ctx context.Context) ([]*Job, error) {
	return s.jobRepo.Find(ctx, map[string]interface{}{}, nil)
}

func (s *Scheduler) getNextRunTime(cronExpr string) time.Time {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return time.Time{}
	}
	return schedule.Next(time.Now())
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() context.Context {
	return s.cron.Stop()
}
