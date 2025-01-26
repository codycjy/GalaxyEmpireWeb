package taskservice

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/accountservice"
	"GalaxyEmpireWeb/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
		Status: models.TASK_RESULT_RUNNING,
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
	if tx.Commit().Error != nil {
		log.Error("[TaskService::CheckAccouuntLogin] failed to commit transaction", zap.Error(tx.Commit().Error))
		tx.Rollback()
		return "", utils.NewServiceError(http.StatusInternalServerError, "Commit Transaction Error", tx.Commit().Error)
	}
	log.Info("[TaskService::CheckAccouuntLogin] task published", zap.String("uuid", uuid), zap.String("routingKey", routingKey))

	// Wait for the task to be done
	// Get info at another func

	return uuid, nil
}

type TaskStatus struct {
	Succeed    bool
	Processing bool
}

func (ts *taskService) GetLoginInfo(ctx context.Context, uuid string) (*TaskStatus, error) {
	var taskLog models.TaskLog
	if err := ts.DB.Where("uuid = ?", uuid).First(&taskLog).Error; err != nil {
		log.Error("[TaskService::GetLoginInfo] failed to get task log", zap.Error(err))
		return nil, err
	}

	status := &TaskStatus{}
	switch taskLog.Status {
	case models.TASK_RESULT_RUNNING:
		status.Processing = true
		status.Succeed = false
	case models.TASK_RESULT_SUCCESS:
		status.Processing = false
		status.Succeed = true
	case models.TASK_RESULT_FAILED:
		status.Processing = false
		status.Succeed = false
	default:
		status.Processing = true
		status.Succeed = false
	}

	return status, nil
}

func (ts *taskService) QueryPlanetID(ctx context.Context, target *models.Target, account *models.Account) (string, *utils.ServiceError) {
	uuid := uuid.New().String()
	tx := ts.DB.Begin()
	log.Info("[TaskService::QueryPlanetID] start to query planet id", zap.String("uuid", uuid), zap.String("target", target.String()), zap.String("traceID", utils.TraceIDFromContext(ctx)))
	taskLog := models.TaskLog{
		TaskID: 0, // Not in DB
		UUID:   uuid,
		Status: models.TASK_RESULT_RUNNING,
	}
	account, err := accountservice.GetService().GetById(ctx, account.ID)
	if err != nil {
		log.Error("[TaskService::QueryPlanetID] failed to get account", zap.Error(err))
		return "", utils.NewServiceError(http.StatusInternalServerError, "Get Account Error", err)
	}

	if err1 := tx.Create(&taskLog).Error; err1 != nil {
		log.Error("[TaskService::QueryPlanetID] failed to create task log", zap.Error(err1))
		return "", utils.NewServiceError(http.StatusInternalServerError, "Create Task Log Error", err1)
	}
	queryTask := models.SingleTaskRequest{
		UUID:        uuid,
		Account:     *account.ToInfo(),
		StartPlanet: *target,
		TaskType:    models.TASKTYPE_QUERY_PLANET_ID,
		NextStart:   time.Now().Unix(),
	}
	taskJSON, err2 := json.Marshal(queryTask)
	if err2 != nil {
		log.Error("[TaskService::QueryPlanetID] failed to marshal task", zap.Error(err2))
		tx.Rollback()
		return "", utils.NewServiceError(http.StatusInternalServerError, "Marshal Task Error", err2)
	}
	routingKey := config.TASK_QUEUE_NAME // TODO: whether to use INSTANT_QUEUE_NAME
	log.Info("[TaskService::QueryPlanetID] start to publish task", zap.String("uuid", uuid), zap.String("routingKey", routingKey), zap.String("taskJSON", string(taskJSON)))
	if err3 := ts.MQ.SendNormalMessage(string(taskJSON), routingKey); err3 != nil {
		log.Error("[TaskService::QueryPlanetID] failed to publish task", zap.Error(err3))
		tx.Rollback()
		return "", utils.NewServiceError(http.StatusInternalServerError, "Publish Task Error", err3)
	}
	tx.Commit()
	log.Info("[TaskService::QueryPlanetID] task published", zap.String("uuid", uuid), zap.String("routingKey", routingKey))
	return uuid, nil
}

func (ts *taskService) GetPlanetID(ctx context.Context, uuid string) (*TaskStatus, int, error) {
	var taskLog models.TaskLog
	if err := ts.DB.Where("uuid = ?", uuid).First(&taskLog).Error; err != nil {
		log.Error("[TaskService::GetPlanetID] failed to get task log", zap.Error(err))
		return nil, 0, err
	}

	status := &TaskStatus{}
	switch taskLog.Status {
	case models.TASK_RESULT_RUNNING:
		status.Processing = true
		status.Succeed = false
		return status, 0, nil
	case models.TASK_RESULT_SUCCESS:
		status.Processing = false
		status.Succeed = true
	case models.TASK_RESULT_FAILED:
		status.Processing = false
		status.Succeed = false
		return status, 0, nil
	default:
		status.Processing = true
		status.Succeed = false
		return status, 0, nil
	}

	// Only process data if status is SUCCESS
	data := taskLog.Msg
	var mapData map[string]string
	if len(data) == 0 {
		return status, 0, fmt.Errorf("empty data")
	}

	if err := json.Unmarshal([]byte(data), &mapData); err != nil {
		return status, 0, err
	}

	planetID, ok := mapData["planet_id"]
	if !ok {
		return status, 0, fmt.Errorf("planet id not found")
	}

	planetIDInt, err := strconv.Atoi(planetID)
	if err != nil {
		return status, 0, err
	}

	return status, planetIDInt, nil
}
