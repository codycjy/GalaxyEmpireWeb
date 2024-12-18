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
	if err := ts.DB.Preload("Tasks").
		Preload("Tasks.Fleet").   // 通过 Tasks 预加载 Fleet
		Preload("Tasks.Targets"). // 通过 Tasks 预加载 Targets
		Where("expire_at > ?", time.Now()).
		Find(&accounts).Error; err != nil {
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
	nextStart := time.Unix(task.NextStart, 0)
	if task.Enabled &&
		task.Status == models.TaskStatusMap[models.TASK_STATUS_READY] &&
		// At most add before 1 hour
		nextStart.Before(time.Now().Add(time.Hour)) {
		singleTask, err := task.ToSingleTaskRequest()
		if err != nil {
			log.Error("[TaskService::GenerateSingleTask] failed to convert task to single task", zap.Error(err))
			return nil
		}
		uuid := singleTask.UUID
		log.Info("[TaskService::GenerateSingleTask] generate single task", zap.String("uuid", uuid), zap.String("task", task.Name))
		taskLog := models.TaskLog{
			TaskID: task.ID,
			UUID:   uuid,
			Status: models.TASK_STATUS_READY,
		}
		if err := ts.DB.Create(&taskLog).Error; err != nil {
			log.Error("[TaskService::GenerateSingleTask] failed to create task log", zap.Error(err))
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
			nexStart := time.Unix(singleTask.NextStart, 0)
			taskJson, err := json.Marshal(singleTask)
			if err != nil {
				return fmt.Errorf("failed to marshal task: %v", err)
			}
			delay := nexStart.Sub(time.Now()) / time.Millisecond
			delay = max(delay, 0)
			log.Debug("[TaskService::GenerateTaskForAccount] delay", zap.Int64("delay", int64(delay)), zap.String("task", string(taskJson)))

			// 发送延迟消息
			routingKey := config.TASK_QUEUE_NAME
			err = ts.MQ.SendDelayedMessage(string(taskJson), routingKey, delay)
			if err != nil {
				return fmt.Errorf("failed to send delayed message: %v", err)
			}

			// 消息发送成功后保存任务状态 并更新下一个任务索引
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
		log.Info("[TaskService::GenerateTaskLoop] generate all task")
		time.Sleep(config.TASK_GENERATOR_INTERVAL)
	}
}
