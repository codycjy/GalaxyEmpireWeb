package taskservice

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/utils"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (ts *taskService) CheckAccountLogin(ctx context.Context, account *models.Account) (string, *utils.ServiceError) {
	uuid := uuid.New().String()
	tx := ts.DB.Begin()
	log.Info("[TaskService::CheckAccouuntLogin] start to check account login", zap.String("uuid", uuid))
	taskLog := models.TaskLog{
		TaskID: 0, // Not in DB
		UUID:   uuid,
		Status: models.TASK_STATUS_READY,
	}
	if err1 := tx.Create(&taskLog).Error; err1 != nil {
		log.Error("[TaskService::CheckAccouuntLogin] failed to create task log", zap.Error(err1))
		return "", utils.NewServiceError(http.StatusInternalServerError, "Create Task Log Error", err1)
	}
	loginTask := models.SingleTaskRequest{
		UUID:      uuid,
		Account:   *account.ToInfo(),
		TaskType:  models.TASKTYPE_LOGIN,
		NextStart: time.Now().Unix(),
	}
	taskJSON, err2 := json.Marshal(loginTask)
	if err2 != nil {
		log.Error("[TaskService::CheckAccouuntLogin] failed to marshal task", zap.Error(err2))
		tx.Rollback()
		return "", utils.NewServiceError(http.StatusInternalServerError, "Marshal Task Error", err2)
	}
	routingKey := config.TASK_QUEUE_NAME // TODO: whether to use INSTANT_QUEUE_NAME
	if err3 := ts.MQ.SendNormalMessage(string(taskJSON), routingKey); err3 != nil {
		log.Error("[TaskService::CheckAccouuntLogin] failed to publish task", zap.Error(err3))
		tx.Rollback()
		return "", utils.NewServiceError(http.StatusInternalServerError, "Publish Task Error", err3)
	}
	tx.Commit()
	log.Info("[TaskService::CheckAccouuntLogin] task published", zap.String("uuid", uuid), zap.String("routingKey", routingKey))

	// Wait for the task to be done
	// Get info at another func

	return uuid, nil
}

func (ts *taskService) GetLoginInfo(ctx context.Context, uuid string) bool {
	var taskLog models.TaskLog
	// try to get the task log until succeed at most 3 times
	for i := 0; i < 3; i++ {
		if err := ts.DB.Where("uuid = ?", uuid).First(&taskLog).Error; err != nil {
			log.Error("[TaskService::GetLoginInfo] failed to get task log", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue

		}
		if taskLog.Status == models.TASK_RESULT_FAILED {
			log.Warn("[TaskService::GetLoginInfo] login failed", zap.String("uuid", uuid))
			return false
		}
		if taskLog.Status == models.TASK_RESULT_SUCCESS {
			log.Info("[TaskService::GetLoginInfo] login success", zap.String("uuid", uuid))
			return true
		}

		time.Sleep(1 * time.Second)

	}
	log.Warn("[TaskService::GetLoginInfo] login timeout", zap.String("uuid", uuid))
	return false

}
