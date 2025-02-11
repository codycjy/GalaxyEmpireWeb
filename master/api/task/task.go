package task

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/taskservice"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type taskResponse struct {
	Succeed bool            `json:"succeed"`
	Data    *models.TaskDTO `json:"data"`
	TraceID string          `json:"traceID"`
}

type accountTaskResponse struct {
	Succeed bool               `json:"succeed"`
	Data    *models.AccountDTO `json:"data"`
	TraceID string             `json:"traceID"`
}

var log = logger.GetLogger()

// GetTaskByID godoc
// @Summary Get task by ID
// @Description Get Task by ID
// @Tags task
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} taskResponse "Successful response with task data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 404 {object} api.ErrorResponse "Not Found with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /task/{id} [get]
func GetTaskByID(c *gin.Context) {
	traceID := c.GetString("traceID")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Wrong Task ID",
			TraceID: traceID,
		})
		return
	}
	taskService := taskservice.GetService()
	task, err1 := taskService.GetTaskByID(c, uint(id))
	if err1 != nil {
		c.JSON(err1.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   err1.Error(),
			Message: "Get Task Error",
			TraceID: traceID,
		})
		return
	}
	c.JSON(http.StatusOK, taskResponse{
		Succeed: true,
		Data:    task.ToDTO(),
		TraceID: traceID,
	})
}

// GetTaskByAccountID godoc
// @Summary Get task by Account ID
// @Description Get Task by Account ID
// @Tags task
// @Accept json
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} accountTaskResponse "Successful response with account data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 404 {object} api.ErrorResponse "Not Found with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /task/account/{id} [get]
func GetTaskByAccountID(c *gin.Context) {
	traceID := c.GetString("traceID")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Wrong Account ID",
			TraceID: traceID,
		})
		return
	}
	taskService := taskservice.GetService()
	tasks, err1 := taskService.GetTaskByAccountID(c, uint(id))
	if err1 != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err1.Error(),
			Message: "Get Task Error",
			TraceID: traceID,
		})
		return
	}
	taskDTOs := make([]*models.TaskDTO, len(tasks))
	for i, task := range tasks {
		taskDTOs[i] = task.ToDTO()
	}

	account := models.AccountDTO{
		Tasks: taskDTOs,
		ID:    uint(id),
	}

	c.JSON(http.StatusOK, accountTaskResponse{
		Succeed: true,
		Data:    &account,
		TraceID: traceID,
	})
}

// AddTask godoc
// @Summary Add a task
// @Description Add a task
// @Tags task
// @Accept json
// @Produce json
// @Param task body models.Task true "Task"
// @Success 200 {object} taskResponse "Successful response with task ID"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /task [post]
func AddTask(c *gin.Context) {
	traceID := c.GetString("traceID")
	task := models.Task{}
	if err := c.BindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Bind Task Error",
			TraceID: traceID,
		})
		return
	}
	if task.AccountID == 0 {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   "Account ID is required",
			TraceID: traceID,
		})
		return
	}
	task.ID = 0 // set ID to 0 to let gorm auto increment
	taskService := taskservice.GetService()
	err1 := taskService.AddTask(c, &task)
	if err1 != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err1.Error(),
			Message: "Add Task Error",
			TraceID: traceID,
		})
		return
	}
	c.JSON(http.StatusOK, taskResponse{
		Succeed: true,
		TraceID: traceID,
		Data:    task.ToDTO(),
	})
}

// UpdateTask godoc
// @Summary Update a task
// @Description Update a task
// @Tags task
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param task body models.TaskUpdateDTO true "Task updates"
// @Success 200 {object} taskResponse "Successful response with task ID"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /task/{id} [put]
func UpdateTask(c *gin.Context) {
	traceID := c.GetString("traceID")

	// 从URL参数获取taskID
	idStr := c.Param("id")
	taskID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Invalid Task ID",
			TraceID: traceID,
		})
		return
	}

	// 解析更新请求
	updates := &models.TaskUpdateDTO{}
	if err := c.BindJSON(updates); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Invalid Request Body",
			TraceID: traceID,
		})
		return
	}

	taskService := taskservice.GetService()
	if err := taskService.UpdateTask(c, uint(taskID), updates); err != nil {
		c.JSON(err.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: err.Msg(),
			TraceID: traceID,
		})
		return
	}

	c.JSON(http.StatusOK, taskResponse{
		Succeed: true,
		TraceID: traceID,
	})
}

// DeleteTask godoc
// @Summary Delete a Task
// @Description Delete a Task
// @Tags Task
// @Accept json
// @Produce json
// @Param id body models.Task true "Task"
// @Success 200 {object} taskResponse "Successful response with task ID"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 404 {object} api.ErrorResponse "Not Found with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /task/{id} [delete]

func DeleteTask(c *gin.Context) {
	traceID := c.GetString("traceID")
	task := models.Task{}
	if err := c.BindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Bind Task Error",
			TraceID: traceID,
		})
		return
	}
	taskService := taskservice.GetService()
	err1 := taskService.DeleteTask(c, task.ID)
	if err1 != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err1.Error(),
			Message: "Delete Task Error",
			TraceID: traceID,
		})
		return
	}
	c.JSON(http.StatusOK, taskResponse{
		Succeed: true,
		Data:    task.ToDTO(),
		TraceID: traceID,
	})
}

type updateTaskEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

// UpdateTaskEnabled godoc
// @Summary Update task enabled status
// @Description Update the enabled status of a task
// @Tags task
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param request body updateTaskEnabledRequest true "Enable status"
// @Success 200 {object} api.Response "Successfully updated task status"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 403 {object} api.ErrorResponse "Forbidden with error message"
// @Failure 404 {object} api.ErrorResponse "Not Found with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /task/{id}/enabled [put]
func UpdateTaskEnabled(c *gin.Context) {
	traceID := c.GetString("traceID")

	// Parse task ID
	idStr := c.Param("id")
	taskID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Invalid Task ID",
			TraceID: traceID,
		})
		return
	}

	// Parse request body
	var req updateTaskEnabledRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Invalid Request Body",
			TraceID: traceID,
		})
		return
	}

	// Update task enabled status
	taskService := taskservice.GetService()
	if err := taskService.UpdateTaskEnabled(c, uint(taskID), req.Enabled); err != nil {
		c.JSON(err.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: err.Msg(),
			TraceID: traceID,
		})
		return
	}

	c.JSON(http.StatusOK, taskResponse{
		Succeed: true,
		TraceID: traceID,
	})
}
