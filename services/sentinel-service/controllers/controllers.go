package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/task"
)

type Controller struct {
	Logger      logger.ILogger
	TaskManager *task.TaskManager
}

func NewController(logger logger.ILogger, taskManager *task.TaskManager) *Controller {
	return &Controller{
		Logger:      logger,
		TaskManager: taskManager,
	}
}

// GetTaskStatus fetches the status of a task by its ID
func (uc *Controller) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	task, err := uc.TaskManager.GetTaskByID(taskID)
	if err != nil {
		uc.Logger.Error("Error fetching task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		return
	}
	c.JSON(http.StatusOK, task)
}




