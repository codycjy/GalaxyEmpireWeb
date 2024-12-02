package taskservice

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/models"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func (ts *taskService) GenerateAllTask() {
	var accounts []*models.Account
	if err := ts.DB.Preload("Tasks").Where("ExpireAt > ?", time.Now()).Find(&accounts).Error; err != nil {
		log.Error("[TaskService::GenerateTask] failed to fetch accounts", zap.Error(err))
		return
	}

	for _, account := range accounts {
		currentAccount := account // avoid closure problem
		go func() {
			err := ts.GenerateTaskForAccount(currentAccount)
			if err != nil {
				log.Error("[TaskService::GenerateTask] failed to generate task for account", zap.Error(err))
			}
		}()
	}
}
func (ts *taskService) GenerateSingleTask(task *models.Task) *models.SingleTaskRequest {
	if task.Enabled &&
		task.Status == models.TaskStatusMap[models.TASK_STATUS_READY] &&
		// At most add before 1 hour
		task.NextStart.Before(time.Now().Add(time.Hour)) {
		singleTask, err := task.ToSingleTaskRequest()
		if err != nil {
			log.Error("[TaskService::GenerateSingleTask] failed to convert task to single task", zap.Error(err))
			return nil
		}

		return singleTask
	}

	return nil
}

func (ts *taskService) GenerateTaskForAccount(account *models.Account) error {
	for _, task := range account.Tasks {
		if singleTask := ts.GenerateSingleTask(&task); singleTask != nil {

			// 转换为JSON
			taskJson, err := json.Marshal(singleTask)
			if err != nil {
				return fmt.Errorf("failed to marshal task: %v", err)
			}
			delay := singleTask.NextStart.Sub(time.Now()) / time.Millisecond
			log.Debug("[TaskService::GenerateTaskForAccount] delay", zap.Int64("delay", int64(delay)), zap.String("task", string(taskJson)))

			// 发送延迟消息
			routingKey := config.TASK_QUEUE_NAME
			err = ts.MQ.SendDelayedMessage(string(taskJson), routingKey, delay)
			if err != nil {
				return fmt.Errorf("failed to send delayed message: %v", err)
			}

			// 消息发送成功后保存任务状态
			task.UpdateNextIndex()
			if err := ts.DB.Save(&task).Error; err != nil {
				return fmt.Errorf("failed to save task: %v", err)
			}
		}
	}
	return nil
}

func (ts *taskService) GenerateTaskLoop() {
	time.Sleep(5 * time.Second)
	log.Info("[TaskService::GenerateTaskLoop] start task generator loop")
	for {
		ts.GenerateAllTask()
		time.Sleep(config.TASK_GENERATOR_INTERVAL)
	}
}
