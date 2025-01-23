package taskservice

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/models"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (ts *taskService) HandleSingleResult(response *models.SingleTaskResponse) (*models.Task, error) {
	if response == nil {
		log.Error("[TaskService::HandleSingleResult] received nil response")
		return nil, errors.New("response is nil")
	}
	log.Info("[TaskService::HandleSingleResult] handling single result",
		zap.String("uuid", response.UUID),
		zap.Uint("task_id", response.TaskID),
		zap.Int("status", response.Status),
		zap.Int("BackTimestamp", int(response.BackTimestamp)))

	var task models.Task
	task.ID = response.TaskID

	tx := ts.DB.Begin()
	if err := tx.Error; err != nil {
		log.Error("[TaskService::HandleSingleResult] failed to begin transaction",
			zap.Error(err),
			zap.String("uuid", response.UUID))
		return nil, err
	}

	var err error
	switch response.TaskType {
	case models.TASKTYPE_LOGIN:
		err = ts.handleLoginTask(tx, response)
	case models.TASKTYPE_QUERY_PLANET_ID:
		err = ts.handleQueryPlanetIDTask(tx, response)
	default:
		if response.Status != models.TASK_RESULT_SUCCESS {
			err = ts.handleFailedTask(tx, response, &task)
		} else {
			err = ts.handleSuccessfulTask(tx, response, &task)
		}
	}

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		log.Error("[TaskService::HandleSingleResult] failed to commit transaction",
			zap.String("uuid", response.UUID),
			zap.Uint("task_id", task.ID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &task, nil
}

func (ts *taskService) handleLoginTask(tx *gorm.DB, response *models.SingleTaskResponse) error {
	succeed := response.Status == models.TASK_RESULT_SUCCESS
	if err := tx.Model(&models.TaskLog{}).
		Where("uuid = ?", response.UUID).
		Update("status", succeed).Error; err != nil {
		return fmt.Errorf("failed to update login task log: %w", err)
	}
	return nil
}

func (ts *taskService) handleQueryPlanetIDTask(tx *gorm.DB, response *models.SingleTaskResponse) error {
	succeed := response.Status == models.TASK_RESULT_SUCCESS
	if err := tx.Model(&models.TaskLog{}).
		Where("uuid = ?", response.UUID).
		Update("status", succeed).
		Update("msg", response.Msg).
		Update("err_msg", response.ErrMsg).Error; err != nil {
		return fmt.Errorf("failed to update query planet ID task log: %w", err)
	}
	return nil
}

func (ts *taskService) handleFailedTask(tx *gorm.DB, response *models.SingleTaskResponse, task *models.Task) error {
	if err := tx.Model(&models.TaskLog{}).
		Where("uuid = ?", response.UUID).
		Update("status", models.TASK_RESULT_FAILED).Error; err != nil {
		return fmt.Errorf("failed to update failed task log: %w", err)
	}

	if err := tx.Model(&task).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"status":     models.TaskStatusMap[models.TASK_STATUS_READY],
		"next_start": time.Now().Unix() + config.FAILED_TASK_DELAY,
	}).Error; err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

func (ts *taskService) handleSuccessfulTask(tx *gorm.DB, response *models.SingleTaskResponse, task *models.Task) error {
	if err := tx.Model(&task).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"status":     models.TaskStatusMap[models.TASK_STATUS_READY],
		"next_start": response.BackTimestamp + config.TASK_DELAY,
	}).Error; err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if err := tx.Model(&models.TaskLog{}).
		Where("uuid = ?", response.UUID).
		Update("status", models.TASK_RESULT_SUCCESS).Error; err != nil {
		return fmt.Errorf("failed to update success task log: %w", err)
	}

	return nil
}

func (ts *taskService) ListenFromResultQueue(queueName string) {
	const reconnectDelay = 5 * time.Second

	log.Info("Listening from result queue", zap.String("queueName", queueName))

	for {
		resultQueue, err := ts.MQ.ConsumeNormalMessage(queueName)
		if err != nil {
			log.Error("Failed to consume message from result queue",
				zap.Error(err),
				zap.String("queue", queueName))
			time.Sleep(reconnectDelay)
			continue
		}

		for msg := range resultQueue {
			var response models.SingleTaskResponse
			if err := json.Unmarshal(msg.Body, &response); err != nil {
				log.Error("Failed to unmarshal message",
					zap.Error(err),
					zap.ByteString("body", msg.Body))
				continue
			}

			// 异步处理消息，避免阻塞消息接收
			go func(resp models.SingleTaskResponse) {
				_, err := ts.HandleSingleResult(&resp)
				if err != nil {
					log.Error("Failed to handle single result",
						zap.Error(err),
						zap.String("uuid", resp.UUID),
						zap.Uint("task_id", resp.TaskID))
				}
			}(response)
		}

		log.Warn("Message channel closed, attempting to reconnect...",
			zap.String("queue", queueName))
		time.Sleep(reconnectDelay)
	}
}
