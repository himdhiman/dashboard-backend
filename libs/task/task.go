package task

import (
	"context"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID        string                 `bson:"_id,omitempty" json:"id,omitempty"`
	TaskType  string                 `bson:"task_type" json:"task_type"`
	Status    TaskStatus             `bson:"status" json:"status"`
	Params    map[string]interface{} `bson:"params,omitempty" json:"params,omitempty"`
	CreatedAt string                 `bson:"created_at" json:"created_at"`
	UpdatedAt string                 `bson:"updated_at" json:"updated_at"`
}

type TaskManager struct {
	Logger   logger.ILogger
	TaskRepo repository.IRepository[Task]
}

func NewTaskManager(collection *models.MongoCollection, logger logger.ILogger) *TaskManager {
	taskRepo := repository.Repository[Task]{Collection: collection}

	return &TaskManager{
		Logger:   logger,
		TaskRepo: &taskRepo,
	}
}

// RunTask runs a task in the background and adds an entry in the MongoDB database for that task
func (tm *TaskManager) RunTask(taskType string, params map[string]interface{}, taskFunc func(params map[string]interface{})) (string, error) {
	task := &Task{
		TaskType:  taskType,
		Status:    TaskStatusPending,
		Params:    params,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	ctx := context.Background()
	id, err := tm.TaskRepo.Create(ctx, task)
	if err != nil {
		tm.Logger.Error("Error creating task", "error", err)
		return "", err
	}

	go func() {
		// Update the task status to running
		_, err := tm.TaskRepo.Update(ctx, map[string]interface{}{"_id": id}, map[string]interface{}{"status": TaskStatusRunning, "updated_at": time.Now().Format(time.RFC3339)})
		if err != nil {
			tm.Logger.Error("Error updating task status to running", "error", err)
			return
		}

		// Run the task function with the provided parameters
		taskFunc(params)

		// Update the task status to completed
		_, err = tm.TaskRepo.Update(ctx, map[string]interface{}{"_id": id}, map[string]interface{}{"status": TaskStatusCompleted, "updated_at": time.Now().Format(time.RFC3339)})
		if err != nil {
			tm.Logger.Error("Error updating task status to completed", "error", err)
			return
		}
	}()

	return id, nil
}

// GetTaskByID fetches a task by its ID
func (tm *TaskManager) GetTaskByID(id string) (*Task, error) {
	ctx := context.Background()
	task, err := tm.TaskRepo.FindByID(ctx, id)
	if err != nil {
		tm.Logger.Error("Error fetching task by ID", "error", err)
		return nil, err
	}
	return task, nil
}

// GetTaskStatusByID fetches the status of a task by its ID
func (tm *TaskManager) GetTaskStatusByID(id string) (TaskStatus, error) {
	ctx := context.Background()
	task, err := tm.TaskRepo.FindByID(ctx, id)
	if err != nil {
		tm.Logger.Error("Error fetching task by ID", "error", err)
		return "", err
	}
	return task.Status, nil
}
