package taskservice

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"encoding/json"

	"gorm.io/gorm"
)

var taskProcessors = map[string]TaskProcessor{
	models.RouteTaskName: &RouteTaskProcessor{},
	models.PlanTaskName:  &DailyTaskProcessor{},
}

type TaskProcessor interface {
	ProcessTask(data *json.RawMessage) error
	InitService(db *gorm.DB, mq *queue.RabbitMQConnection)
}

type RouteTaskProcessor struct {
	db *gorm.DB
	mq *queue.RabbitMQConnection
}

func (r *RouteTaskProcessor) InitService(db *gorm.DB, mq *queue.RabbitMQConnection) {
	r.db = db
	r.mq = mq
}

func (processor *RouteTaskProcessor) ProcessTask(data *json.RawMessage) error {
	var routeTask models.RouteTask
	if err := json.Unmarshal(*data, &routeTask); err != nil {
		return err
	}

	return nil
}

type DailyTaskProcessor struct {
	db *gorm.DB
	mq *queue.RabbitMQConnection
}

func (d *DailyTaskProcessor) InitService(db *gorm.DB, mq *queue.RabbitMQConnection) {
	d.db = db
	d.mq = mq
}
func (processor *DailyTaskProcessor) ProcessTask(data *json.RawMessage) error {
	return nil
}
