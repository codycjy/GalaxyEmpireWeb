package taskservice

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/models"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
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

	if response.Status != models.TASK_RESULT_SUCCESS {
		// Update task log to failed status
		if err := tx.Model(&models.TaskLog{}).
			Where("uuid = ?", response.UUID).
			Update("status", models.TASK_RESULT_FAILED).Error; err != nil {
			tx.Rollback()
			log.Error("[TaskService::HandleSingleResult] failed to update failed task log",
				zap.String("uuid", response.UUID),
				zap.Uint("task_id", task.ID),
				zap.Error(err))
			return nil, err
		}
		if response.TaskType == models.TASKTYPE_LOGIN {
			log.Warn("[TaskService::HandleSingleResult] login task failed, no need to update task status",
				zap.String("uuid", response.UUID),
				zap.Uint("task_id", task.ID),
				zap.Int("status", response.Status))
			return nil, nil
		}

		// 即使任务失败也要更新任务状态和下次执行时间
		if err := tx.Model(&task).Updates(map[string]interface{}{
			"status":     models.TaskStatusMap[models.TASK_STATUS_READY],
			"next_start": time.Now().Unix() + config.FAILED_TASK_DELAY,
		}).Error; err != nil {
			tx.Rollback()
			log.Error("[TaskService::HandleSingleResult] failed to update task status",
				zap.String("uuid", response.UUID),
				zap.Uint("task_id", task.ID),
				zap.Error(err))
			return nil, err
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			log.Error("[TaskService::HandleSingleResult] failed to commit transaction",
				zap.String("uuid", response.UUID),
				zap.Uint("task_id", task.ID),
				zap.Error(err))
			return nil, err
		}

		log.Warn("[TaskService::HandleSingleResult] task execution failed but status updated",
			zap.String("uuid", response.UUID),
			zap.Uint("task_id", task.ID),
			zap.Int("status", response.Status))

		return &task, nil // 返回更新后的任务，而不是错误
	}

	// Handle login task
	if response.TaskType == models.TASKTYPE_LOGIN {
		if err := tx.Model(&models.TaskLog{}).
			Where("uuid = ?", response.UUID).
			Update("status", models.TASK_RESULT_SUCCESS).Error; err != nil {
			tx.Rollback()
			log.Error("[TaskService::HandleSingleResult] failed to update login task log",
				zap.String("uuid", response.UUID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to update login task log: %w", err)
		}
		return nil, tx.Commit().Error
	}

	// Handle other tasks
	log.Info("[TaskService::HandleSingleResult] task succeeded",
		zap.String("uuid", response.UUID),
		zap.Uint("task_id", task.ID),
		zap.Int("task_type", response.TaskType))

	// Update task status
	task.Status = models.TaskStatusMap[models.TASK_STATUS_READY]
	task.NextStart = response.BackTimestamp + config.TASK_DELAY

	if err := tx.Model(&task).Select("status", "next_start").Updates(models.Task{
		Status:    models.TaskStatusMap[models.TASK_STATUS_READY],
		NextStart: response.BackTimestamp + config.TASK_DELAY,
	}).Error; err != nil {
		tx.Rollback()
		log.Error("[TaskService::HandleSingleResult] failed to update task",
			zap.String("uuid", response.UUID),
			zap.Uint("task_id", task.ID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Update task log
	if err := tx.Model(&models.TaskLog{}).
		Where("uuid = ?", response.UUID).
		Update("status", models.TASK_RESULT_SUCCESS).Error; err != nil {
		tx.Rollback()
		log.Error("[TaskService::HandleSingleResult] failed to update success task log",
			zap.String("uuid", response.UUID),
			zap.Uint("task_id", task.ID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update task log: %w", err)
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
