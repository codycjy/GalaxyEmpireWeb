package taskservice

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

var taskProcessors = map[string]TaskProcessor{
	models.RouteTaskName: &RouteTaskProcessor{},
	models.PlanTaskName:  &PlanTaskProcessor{},
}

type TaskProcessor interface {
	ProcessTask(data *json.RawMessage) error
	InitService(db *gorm.DB, mq *queue.RabbitMQConnection)
}

type RouteTaskProcessor struct {
	db *gorm.DB
	mq *queue.RabbitMQConnection
}

func (processor *RouteTaskProcessor) InitService(db *gorm.DB, mq *queue.RabbitMQConnection) {
	processor.db = db
	processor.mq = mq
}

func (processor *RouteTaskProcessor) ProcessTask(data *json.RawMessage) error {
	var routeTask *models.RouteTask
	if err := json.Unmarshal(*data, &routeTask); err != nil {
		return err
	}
	log := models.NewLog()
	processor.db.Save(routeTask)
	err := processor.db.Create(&log).Error
	if err != nil {
		fmt.Printf("log: %v\n", log)
	}

	return nil
}

type PlanTaskProcessor struct {
	db *gorm.DB
	mq *queue.RabbitMQConnection
}

func (processor *PlanTaskProcessor) InitService(db *gorm.DB, mq *queue.RabbitMQConnection) {
	processor.db = db
	processor.mq = mq
}
func (processor *PlanTaskProcessor) ProcessTask(data *json.RawMessage) error {
	return nil
}
