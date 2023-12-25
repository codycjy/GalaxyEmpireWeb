package models

import (
	"os"
	"time"
)

var RouteTaskName string = "RouteTask"

type RouteTask struct {
	BaseTask
	Repeat    int
	From      Star      `gorm:"embedded;embeddedPrefix:from_"`
	To        Star      `gorm:"embedded;embeddedPrefix:to_"`
	NextStart time.Time `json:"next_start"`
	AccountID uint
	Fleets    []Fleet `gorm:"many2many:route_task_fleet;"`
}

func (routeTask *RouteTask) QueueName() string {
	if os.Getenv("ENV") == "test" {
		return "TEST_" + "NormalQueue"
	}
	return "NormalQueue"
}
func (routeTask *RouteTask) TaskType() string {
	return "RouteTask"
}
