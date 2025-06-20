package routes

import (
	"io"
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"
	"time"

	"github.com/gin-gonic/gin"
)

// GetTaskById godoc
// @Summary Get a task by ID
// @Description The task represents a background job that can be run by the system
// @Produce  json
// @Success 200 {object} models.ResponseSingle[services.TaskResponse]
// @Router /task/{taskId} [get]
// @Tags Task
// @Param taskId path string true "The task ID"
// @Error 400 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func GetTaskById(c *gin.Context, sess *cache.Session) {
	defer HandleError("GetTaskById")
	taskId := c.Param("taskId")
	if taskId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Task ID is required"})
		return
	}
	task_service := services.NewTaskService()
	task, err := task_service.GetTask(c, models.IDType(taskId))
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error getting task : " + err.Error()})
		return
	}
	if task == nil {
		c.JSON(404, models.ResponseError{Detail: "Task not found"})
		return
	}
	c.JSON(200, models.ResponseSingle[models.Task]{Data: *task})
}

// GetTaskByIDStream godoc
// @Summary Get a task by ID but stream the response
// @Description The task represents a background job that can be run by the system. This endpoint streams the response
// @Produce  json
// @Router /task/{taskId}/stream [get]
// @Tags Task
// @Param taskId path string true "The task ID"
// @Error 400 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func GetTaskByIDStream(c *gin.Context, sess *cache.Session) {
	defer HandleError("GetTaskById")
	taskId := c.Param("taskId")
	if taskId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Task ID is required"})
		return
	}
	task_service := services.NewTaskService()
	c.Stream(func(w io.Writer) bool {
		task, err := task_service.GetTask(c, models.IDType(taskId))
		if err != nil {
			c.SSEvent("error", err.Error())
			return false
		}
		if task == nil {
			c.SSEvent("error", "Task not found")
			return false
		}
		c.SSEvent("task", task)
		if task.Status == models.TaskStatusSuccess || task.Status == models.TaskStatusFailed {
			// No need to stream anymore
			return false
		}
		time.Sleep(5 * time.Second)
		return true
	})
}
func NewTaskRoutes(p *gin.RouterGroup) {
	task := p.Group("/task")
	task.GET("/:taskId", WithSession(GetTaskById))
	task.GET("/:taskId/stream", WithSession(GetTaskByIDStream))
}
