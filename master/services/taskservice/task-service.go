package taskservice

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"GalaxyEmpireWeb/services/casbinservice"
	"GalaxyEmpireWeb/utils"
	"context"
	"net/http"
	"strconv"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var taskServiceInstance *taskService
var log = logger.GetLogger()

type taskService struct {
	DB       *gorm.DB
	MQ       *queue.RabbitMQConnection
	Enforcer casbinservice.Enforcer
}

func GetService() *taskService {
	if taskServiceInstance == nil {
		log.Fatal("Task Service Not Initialized")
	}
	return taskServiceInstance
}
func InitService(db *gorm.DB, mq *queue.RabbitMQConnection, enforcer casbinservice.Enforcer) {
	taskServiceInstance = NewService(db, mq, enforcer)
	go taskServiceInstance.GenerateTaskLoop()
	go taskServiceInstance.ListenFromResultQueue(config.RESULT_QUEUE_NAME)
	db.AutoMigrate(&models.Task{}, &models.TaskLog{})
}

func NewService(db *gorm.DB, mq *queue.RabbitMQConnection, enforcer casbinservice.Enforcer) *taskService {
	return &taskService{
		DB:       db,
		MQ:       mq,
		Enforcer: enforcer,
	}
}

func (ts *taskService) AddTask(ctx context.Context, task *models.Task) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	userID := utils.UserIDFromContext(ctx)
	log.Info("[TaskService] AddTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Any("task", task), zap.Int("AccountID", int(task.AccountID)))

	tx := ts.DB.Begin()
	if err := tx.Create(task).Error; err != nil {
		log.Error("[TaskService] AddTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(err))
		return utils.NewServiceError(http.StatusInternalServerError, "Create Task Error", err)
	}
	sub := models.UserEntityPrefix + strconv.Itoa(int(userID)) // Changed from account prefix to user prefix
	obj := task.GetEntityPrefix() + strconv.Itoa(int(task.ID))
	act := "write"
	_, err1 := ts.Enforcer.AddPolicy(ctx, tx, sub, obj, act)
	if err1 != nil {
		log.Error("[TaskService] AddTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(err1))
		tx.Rollback()
		return utils.NewServiceError(http.StatusInternalServerError, "Casbin AddPolicy Error", err1)
	}
	act = "read"
	_, err2 := ts.Enforcer.AddPolicy(ctx, tx, sub, obj, act)
	if err2 != nil {
		log.Error("[TaskService] AddTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(err2))
		tx.Rollback()
		return utils.NewServiceError(http.StatusInternalServerError, "Casbin AddPolicy Error", err2)
	}
	log.Info("[TaskService] AddTask Succeed", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Any("task", task), zap.Int("AccountID", int(task.AccountID)))

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error("[TaskService] AddTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(err))
		return utils.NewServiceError(http.StatusInternalServerError, "Commit Transaction Error", err)
	}

	go ts.Enforcer.ReloadPolicy()

	return nil
}

func (ts *taskService) GetTaskByAccountID(ctx context.Context, accountID uint) ([]models.Task, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	userID := utils.UserIDFromContext(ctx)
	log.Info("[TaskService] GetTaskByAccountID", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Uint("accountID", accountID))
	var tasks []models.Task
	if err := ts.DB.Where("account_id = ?", accountID).Find(&tasks); err.Error != nil {
		if err.RowsAffected == 0 {
			log.Warn("[TaskService] GetTaskByAccountID", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Uint("accountID", accountID), zap.Error(err.Error))
			return nil, utils.NewServiceError(http.StatusNotFound, "Task Not Found", err.Error)
		}
		log.Error("[TaskService] GetTaskByAccountID", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Uint("accountID", accountID), zap.Error(err.Error))
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Get Task Error", err.Error)
	}
	return tasks, nil
}
func (ts *taskService) GetTaskByID(ctx context.Context, taskID uint) (*models.Task, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	userID := utils.UserIDFromContext(ctx)
	var task models.Task
	task.ID = taskID // Set ID first so we can use it in prefix
	log.Info("[TaskService] GetTaskByID", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Uint("taskID", taskID))
	allowed, err := ts.Enforcer.Enforce(ctx, models.UserEntityPrefix+strconv.Itoa(int(userID)), task.GetEntityPrefix()+strconv.Itoa(int(taskID)), "read")

	if err != nil {
		log.Error("[TaskService] GetTaskByID", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(err))
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Casbin Enforce Error", err)
	}
	if !allowed {
		log.Warn("[TaskService] GetTaskByID", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Uint("taskID", taskID))
		return nil, utils.NewServiceError(http.StatusForbidden, "Permission Denied", nil)
	}

	if err := ts.DB.First(&task, taskID); err.Error != nil {
		if err.RowsAffected == 0 {
			log.Warn("[TaskService] GetTaskByID", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Uint("taskID", taskID), zap.Error(err.Error))
			return nil, utils.NewServiceError(http.StatusNotFound, "Task Not Found", err.Error)
		}
		log.Error("[TaskService] GetTaskByID", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Uint("taskID", taskID), zap.Error(err.Error))
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Get Task Error", err.Error)
	}
	return &task, nil
}
func (ts *taskService) UpdateTask(ctx context.Context, task *models.Task) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	userID := utils.UserIDFromContext(ctx)
	log.Info("[TaskService] UpdateTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Any("task", task), zap.Int("AccountID", int(task.AccountID)))

	allowed, err := ts.Enforcer.Enforce(ctx, models.UserEntityPrefix+strconv.Itoa(int(userID)), task.GetEntityPrefix()+strconv.Itoa(int(task.ID)), "write")
	if err != nil {
		log.Error("[TaskService] UpdateTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(err))
		return utils.NewServiceError(http.StatusInternalServerError, "Casbin Enforce Error", err)
	}
	if !allowed {
		log.Warn("[TaskService] UpdateTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Any("task", task), zap.Int("AccountID", int(task.AccountID)))
		return utils.NewServiceError(http.StatusForbidden, "Permission Denied", nil)
	}

	tx := ts.DB.Begin()
	if err := tx.Save(task).Error; err != nil {
		log.Error("[TaskService] UpdateTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(err))
		tx.Rollback()
		return utils.NewServiceError(http.StatusInternalServerError, "Update Task Error", err)
	}
	log.Info("[TaskService] UpdateTask Succeed", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Any("task", task), zap.Int("AccountID", int(task.AccountID)))

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error("[TaskService] UpdateTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(err))
		return utils.NewServiceError(http.StatusInternalServerError, "Commit Transaction Error", err)
	}
	return nil
}
func (ts *taskService) DeleteTask(ctx context.Context, taskID uint) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	userID := utils.UserIDFromContext(ctx)
	log.Info("[TaskService] DeleteTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Uint("taskID", taskID))
	var task models.Task
	task.ID = taskID
	allowed, err := ts.Enforcer.Enforce(ctx, models.UserEntityPrefix+strconv.Itoa(int(userID)), task.GetEntityPrefix()+strconv.Itoa(int(taskID)), "write")
	if err != nil {
		log.Error("[TaskService] DeleteTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(err))
		return utils.NewServiceError(http.StatusInternalServerError, "Casbin Enforce Error", err)
	}
	if !allowed {
		log.Warn("[TaskService] DeleteTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Uint("taskID", taskID))
		return utils.NewServiceError(http.StatusForbidden, "Permission Denied", nil)
	}
	result := ts.DB.Delete(&task, taskID)
	if result.Error != nil {
		log.Error("[TaskService] DeleteTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(result.Error))
		return utils.NewServiceError(http.StatusInternalServerError, "Delete Task Error", result.Error)
	}
	if result.RowsAffected == 0 {
		log.Warn("[TaskService] DeleteTask", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Uint("taskID", taskID))
		return utils.NewServiceError(http.StatusNotFound, "Task Not Found", nil)
	}

	log.Info("[TaskService] DeleteTask Succeed", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Uint("taskID", taskID))
	return nil

}

func (ts *taskService) UpdateTaskEnabled(ctx context.Context, taskID uint, enabled bool) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	userID := utils.UserIDFromContext(ctx)
	log.Info("[TaskService] UpdateTaskEnabled",
		zap.String("traceID", traceID),
		zap.Uint("userID", userID),
		zap.Uint("taskID", taskID),
		zap.Bool("enabled", enabled))

	var task models.Task
	task.ID = taskID // Set ID first so we can use it in prefix
	allowed, err := ts.Enforcer.Enforce(ctx, models.UserEntityPrefix+strconv.Itoa(int(userID)), task.GetEntityPrefix()+strconv.Itoa(int(taskID)), "write")
	if err != nil {
		log.Error("[TaskService] UpdateTaskEnabled", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(err))
		return utils.NewServiceError(http.StatusInternalServerError, "Casbin Enforce Error", err)
	}
	if !allowed {
		log.Warn("[TaskService] UpdateTaskEnabled", zap.String("traceID", traceID), zap.Uint("userID", userID))
		return utils.NewServiceError(http.StatusForbidden, "Permission Denied", nil)
	}

	// 更新状态
	result := ts.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("enabled", enabled)
	if result.Error != nil {
		log.Error("[TaskService] UpdateTaskEnabled", zap.String("traceID", traceID), zap.Uint("userID", userID), zap.Error(result.Error))
		return utils.NewServiceError(http.StatusInternalServerError, "Update Task Error", result.Error)
	}
	if result.RowsAffected == 0 {
		log.Warn("[TaskService] UpdateTaskEnabled", zap.String("traceID", traceID), zap.Uint("userID", userID))
		return utils.NewServiceError(http.StatusNotFound, "Task Not Found", nil)
	}

	log.Info("[TaskService] UpdateTaskEnabled Succeed",
		zap.String("traceID", traceID),
		zap.Uint("userID", userID),
		zap.Uint("taskID", taskID),
		zap.Bool("enabled", enabled))
	return nil
}
