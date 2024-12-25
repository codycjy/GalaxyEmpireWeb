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
		Preload("Tasks.Targets"). // 通过 Tasks 预加载 Targets
		Preload("Tasks.Fleet").
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

func (ts *taskService) GenerateSingleTask(task *models.Task, account *models.Account) *models.SingleTaskRequest {
	nextStart := time.Unix(task.NextStart, 0)
	if !task.Enabled ||
		task.Status != models.TaskStatusMap[models.TASK_STATUS_READY] ||
		time.Until(nextStart) > config.QUEUE_THRESHOLD { // 如果距离执行时间超过1小时
		reason := "unknown"
		if !task.Enabled {
			reason = "task disabled"
		} else if task.Status != models.TaskStatusMap[models.TASK_STATUS_READY] {
			reason = "task not in ready status"
		} else if time.Until(nextStart) > config.QUEUE_THRESHOLD {
			reason = "too early to generate"
		}
		log.Info("[TaskService::GenerateSingleTask] task not ready",
			zap.String("task", task.Name),
			zap.Uint("task_id", task.ID),
			zap.Time("next_start", nextStart),
			zap.Time("now", time.Now()),
			zap.Duration("time_until_start", time.Until(nextStart)),
			zap.String("reason", reason))
		return nil
	}
	log.Info("[TaskService::GenerateSingleTask] generating single task",
		zap.String("task", task.Name),
		zap.Uint("task_id", task.ID),
		zap.Time("next_start", nextStart),
		zap.Time("now", time.Now()))

	// 开启事务
	tx := ts.DB.Begin()
	if err := tx.Error; err != nil {
		log.Error("[TaskService::GenerateSingleTask] failed to begin transaction", zap.Error(err))
		return nil
	}

	// 使用事务锁定任务记录
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&models.Task{}, task.ID).Error; err != nil {
		tx.Rollback()
		log.Error("[TaskService::GenerateSingleTask] failed to lock task", zap.Error(err))
		return nil
	}

	// 生成单次任务请求
	singleTask, err := task.ToSingleTaskRequest(account)
	if err != nil {
		tx.Rollback()
		log.Error("[TaskService::GenerateSingleTask] failed to convert task to single task", zap.Error(err))
		return nil
	}

	// 更新任务的 NextIndex
	if err := tx.Model(task).Update("next_index", task.NextIndex).Error; err != nil {
		tx.Rollback()
		log.Error("[TaskService::GenerateSingleTask] failed to update next_index",
			zap.Error(err),
			zap.Uint("task_id", task.ID))
		return nil
	}

	// 创建任务日志
	taskLog := models.TaskLog{
		TaskID: task.ID,
		UUID:   singleTask.UUID,
		Status: models.TASK_RESULT_RUNNING,
	}
	if err := tx.Create(&taskLog).Error; err != nil {
		tx.Rollback()
		log.Error("[TaskService::GenerateSingleTask] failed to create task log",
			zap.Error(err),
			zap.String("uuid", singleTask.UUID))
		return nil
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Error("[TaskService::GenerateSingleTask] failed to commit transaction",
			zap.Error(err),
			zap.String("uuid", singleTask.UUID))
		return nil
	}

	log.Info("[TaskService::GenerateSingleTask] generate single task",
		zap.String("uuid", singleTask.UUID),
		zap.String("task", task.Name))

	return singleTask
}
func (ts *taskService) GenerateTaskForAccount(account *models.Account) error {
	for _, task := range account.Tasks {
		if singleTask := ts.GenerateSingleTask(&task, account); singleTask != nil {

			// 转换为JSON
			nextStart := time.Unix(singleTask.NextStart, 0)
			taskJson, err := json.Marshal(singleTask)
			if err != nil {
				return fmt.Errorf("failed to marshal task: %v", err)
			}
			delay := nextStart.Sub(time.Now())
			if delay < 0 {
				// 如果计算出的延迟为负，使用最小延迟时间
				delay = time.Duration(config.TASK_DELAY) * time.Second
			}
			log.Debug("[TaskService::GenerateTaskForAccount] delay", zap.Int64("delay", delay.Milliseconds()), zap.String("task", string(taskJson)))

			// 发送延迟消息
			routingKey := config.TASK_QUEUE_NAME
			err = ts.MQ.SendDelayedMessage(string(taskJson), routingKey, delay)
			if err != nil {
				return fmt.Errorf("failed to send delayed message: %v", err)
			}

			// 消息发送成功后保存任务状态 并更新下一个任务索引
			task.Status = models.TaskStatusMap[models.TASK_STATUS_RUNNING]
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
