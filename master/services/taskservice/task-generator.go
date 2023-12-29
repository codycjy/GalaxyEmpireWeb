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

func (generator *TaskGenerator) generateRouteTask(ctx context.Context) {
	var routeTasks []models.RouteTask
	var accountIDs []uint
	now := time.Now()

	// Step 1: Get Account IDs
	generator.db.Model(&models.Account{}).Where("expire_at > ?", now).Pluck("id", &accountIDs)

	// Step 2: Get RouteTasks based on Account IDs
	generator.db.
		Where("account_id IN ?", accountIDs).
		Where("next_start < ?", now).
		Where("enabled = ?", true).
		Find(&routeTasks)

	for _, task := range routeTasks {
		err := generator.taskService.SendTask(&task)
		if err != nil {
			log.Warn("[service]Send route task error",
				zap.Error(err),
			)
			continue
		}
	}
}

func (generator *TaskGenerator) generatePlanTask() {
	var planTasks []models.PlanTask
	var accountIDs []uint
	now := time.Now()

	// Step 1: Get Account IDs
	generator.db.Model(&models.Account{}).Where("expire_at > ?", now).Pluck("id", &accountIDs)

	// Step 2: Get RouteTasks based on Account IDs
	generator.db.
		Where("account_id IN ?", accountIDs).
		Where("next_start < ?", now).
		Where("enabled = ?", true).
		Find(&planTasks)

	for _, task := range planTasks {
		err := generator.taskService.SendTask(&task)
		if err != nil {
			log.Warn("[service]Send plan task error",
				zap.Error(err),
			)
			continue
		}
	}
}

func (generator *TaskGenerator) FindTasks() {
	for {
		routeCTX := utils.NewContextWithTraceID()
		generator.generateRouteTask(routeCTX)
		// generator.generatePlanTask()
		time.Sleep(15 * time.Second)
	}

}
