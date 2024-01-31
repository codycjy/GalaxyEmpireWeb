package taskservice

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"GalaxyEmpireWeb/utils"
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskGenerator struct {
	db          *gorm.DB
	mq          *queue.RabbitMQConnection
	taskService *taskService
}

func initTaskGenerator(db *gorm.DB, mq *queue.RabbitMQConnection, taskService *taskService) *TaskGenerator {
	return &TaskGenerator{
		db:          db,
		mq:          mq,
		taskService: taskService,
	}
}

func (generator *TaskGenerator) setAccountInfo(ctx context.Context, task models.TaskModel) error {
	traceID := utils.TraceIDFromContext(ctx)
	var account models.Account
	log.Info("[service]Get account",
		zap.Uint("accountID", task.GetAccountID()),
		zap.String("traceID", traceID),
	)

	result := generator.db.Where("id = ?", task.GetAccountID()).First(&account)
	if err := result.Error; err != nil {
		log.Error("[service]Get account error",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return err

	}
	task.SetAccountInfo(*account.ToInfo())
	return nil
}

func generateTasks[T models.TaskModel](ctx context.Context, now time.Time, tasks []T, generator TaskGenerator) {
	traceID := utils.TraceIDFromContext(ctx)

	var accountIDs []uint
	generator.db.Model(&models.Account{}).Where("expire_at > ?", now).Pluck("id", &accountIDs)

	generator.db.
		Where("account_id IN ?", accountIDs).
		Where("next_start < ?", now).
		Where("enabled = ?", true).
		Find(&tasks)

	for _, task := range tasks {
		if err := generator.setAccountInfo(ctx, task); err != nil {
			continue
		}
		if err := generator.taskService.SendTask(task); err != nil {
			log.Warn("[service]Send task error",
				zap.String("traceID", traceID),
				zap.Error(err),
				zap.String("task_type", task.TaskType()),
			)
			continue
		}
		log.Info("[service]Send task success",
			zap.Uint("TaskID", task.GetID()),
			zap.String("traceID", traceID),
		)
	}
}

func (generator *TaskGenerator) generateRouteTask(ctx context.Context) {
	var routeTasks []*models.RouteTask
	now := time.Now()
	generateTasks(ctx, now, routeTasks, *generator)
}

func (generator *TaskGenerator) generatePlanTask(ctx context.Context) {
	var planTasks []*models.PlanTask
	now := time.Now()
	generateTasks(ctx, now, planTasks, *generator)
}

func (generator *TaskGenerator) FindAllTasks() {
	for {
		routeCTX := utils.NewContextWithTraceID()
		generator.generateRouteTask(routeCTX)
		generator.generatePlanTask(routeCTX)
		time.Sleep(15 * time.Second)
	}

}
