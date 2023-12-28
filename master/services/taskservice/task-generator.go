package taskservice

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"log"
	"time"

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

func (generator *TaskGenerator) generateRouteTask() {
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
			log.Printf("send task error: %v", err)
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
			log.Printf("send task error: %v", err)
			continue
		}
	}
}

func (generator *TaskGenerator) FindTasks() {
	for {
		generator.generateRouteTask()
		// generator.generatePlanTask()
		time.Sleep(15 * time.Second)
	}

}
