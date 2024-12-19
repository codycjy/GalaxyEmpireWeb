package taskservice

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/models"
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"
)

func (ts *taskService) HandleSingleResult(response *models.SingleTaskResponse) (*models.Task, error) {
	if response == nil {
		return nil, errors.New("response is nil")
	}
	var task models.Task
	task.ID = response.TaskID
	if response.Status != models.TASK_RESULT_SUCCESS {
		log.Error("Task failed", zap.String("uuid", response.UUID))
		return nil, errors.New("response status is not 0")
	}
	if response.TaskType != models.TASKTYPE_LOGIN {
		log.Info("Task succeeded", zap.String("uuid", response.UUID))
		task.Status = models.TaskStatusMap[models.TASK_STATUS_READY]
		task.NextStart = response.BackTimestamp + config.TASK_DELAY
		return &task, nil
	}
	if response.TaskType == models.TASKTYPE_LOGIN {
		ts.DB.Model(&models.TaskLog{}).Where("uuid = ?", response.UUID).Update("status", models.TASK_RESULT_SUCCESS)
		return nil, nil // We don't need to save login task
	}

	return nil, errors.New("Failed to handle single result")

}

func (ts *taskService) ListenFromResultQueue(queueName string) {
	time.Sleep(5 * time.Second)
	log.Info("Listening from result queue", zap.String("queueName", queueName))
	for {
		resultQueue, err := ts.MQ.ConsumeNormalMessage(queueName)
		if err != nil {
			log.Error("Failed to consume message from result queue", zap.Error(err))
			// 添加重试延迟
			time.Sleep(5 * time.Second)
			continue
		}

		for msg := range resultQueue {
			var response models.SingleTaskResponse
			err := json.Unmarshal(msg.Body, &response)
			if err != nil {
				log.Error("Failed to unmarshal message from result queue", zap.Error(err))
				continue
			}

			task, err := ts.HandleSingleResult(&response)
			if err != nil {
				log.Error("Failed to handle single result", zap.Error(err))
				continue
			}

			if task != nil {
				ts.DB.Save(task)
			}
		}

		log.Warn("Message channel closed, attempting to reconnect...")
	}
}
