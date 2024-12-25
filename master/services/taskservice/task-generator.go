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

	// Generate single task request without DB operations
	singleTask, err := task.ToSingleTaskRequest(account)
	if err != nil {
		log.Error("[TaskService::GenerateSingleTask] failed to convert task to single task", zap.Error(err))
		return nil
	}

	return singleTask
}

func (ts *taskService) GenerateTaskForAccount(account *models.Account) error {
	fourHoursAgo := time.Now().Add(-4 * time.Hour)

	for _, task := range account.Tasks {
		// Reset long-running tasks to ready status
		if task.Status == models.TaskStatusMap[models.TASK_STATUS_RUNNING] &&
			time.Unix(task.NextStart, 0).Before(fourHoursAgo) {
			log.Warn("[TaskService::GenerateTaskForAccount] task stuck in running state, resetting to ready",
				zap.Uint("task_id", task.ID),
				zap.String("task_name", task.Name),
				zap.Time("next_start", time.Unix(task.NextStart, 0)))

			if err := ts.DB.Model(&task).Update("status", models.TaskStatusMap[models.TASK_STATUS_READY]).Error; err != nil {
				log.Error("[TaskService::GenerateTaskForAccount] failed to reset task status",
					zap.Error(err))
				continue
			}
			task.Status = models.TaskStatusMap[models.TASK_STATUS_READY]
		}

		// Check and reset NextIndex if it's invalid
		if task.NextIndex >= len(task.Targets) {
			log.Warn("[TaskService::GenerateTaskForAccount] invalid next_index, resetting to 0",
				zap.Uint("task_id", task.ID),
				zap.Int("next_index", task.NextIndex),
				zap.Int("targets_length", len(task.Targets)))
			task.NextIndex = 0
			if err := ts.DB.Model(&task).Update("next_index", 0).Error; err != nil {
				log.Error("[TaskService::GenerateTaskForAccount] failed to reset next_index",
					zap.Error(err))
				continue
			}
		}

		if singleTask := ts.GenerateSingleTask(&task, account); singleTask != nil {
			// Start transaction
			tx := ts.DB.Begin()
			if err := tx.Error; err != nil {
				return fmt.Errorf("failed to begin transaction: %v", err)
			}

			// Lock task record
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&models.Task{}, task.ID).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to lock task: %v", err)
			}

			// Create task log
			taskLog := models.TaskLog{
				TaskID: task.ID,
				UUID:   singleTask.UUID,
				Status: models.TASK_RESULT_RUNNING,
			}
			if err := tx.Create(&taskLog).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create task log: %v", err)
			}

			// Update task's NextIndex
			if err := tx.Model(&task).Update("next_index", task.NextIndex).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update next_index: %v", err)
			}

			// Commit transaction
			if err := tx.Commit().Error; err != nil {
				return fmt.Errorf("failed to commit transaction: %v", err)
			}

			// Convert to JSON and send message
			nextStart := time.Unix(singleTask.NextStart, 0)
			taskJson, err := json.Marshal(singleTask)
			if err != nil {
				return fmt.Errorf("failed to marshal task: %v", err)
			}

			delay := time.Until(nextStart)
			if delay < 0 {
				delay = time.Duration(config.TASK_DELAY) * time.Second
			}
			log.Debug("[TaskService::GenerateTaskForAccount] delay",
				zap.Int64("delay", delay.Milliseconds()),
				zap.String("task", string(taskJson)))

			// Send delayed message
			routingKey := config.TASK_QUEUE_NAME
			if err := ts.MQ.SendDelayedMessage(string(taskJson), routingKey, delay); err != nil {
				return fmt.Errorf("failed to send delayed message: %v", err)
			}

			// Update task status
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
