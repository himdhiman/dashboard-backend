package scheduler

import (
	"time"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type JobConfig struct {
	Name        string                 `bson:"name" json:"name"`
	CronExpr    string                 `bson:"cron_expr" json:"cron_expr"`
	Params      map[string]interface{} `bson:"params" json:"params"`
	MaxRetries  int                    `bson:"max_retries" json:"max_retries"`
	IsRecurring bool                   `bson:"is_recurring" json:"is_recurring"`
}

type Job struct {
	ID          string                 `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string                 `bson:"name" json:"name"`
	Status      JobStatus              `bson:"status" json:"status"`
	CronExpr    string                 `bson:"cron_expr" json:"cron_expr"`
	Params      map[string]interface{} `bson:"params" json:"params"`
	LastRunAt   time.Time              `bson:"last_run_at" json:"last_run_at"`
	NextRunAt   time.Time              `bson:"next_run_at" json:"next_run_at"`
	RetryCount  int                    `bson:"retry_count" json:"retry_count"`
	MaxRetries  int                    `bson:"max_retries" json:"max_retries"`
	Error       string                 `bson:"error,omitempty" json:"error,omitempty"`
	IsRecurring bool                   `bson:"is_recurring" json:"is_recurring"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}
