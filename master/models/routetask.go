package models

import (
	"os"
)

var RouteTaskName string = "RouteTask"

type RouteTask struct {
	BaseTask
	Repeat      int
	From        Star `gorm:"embedded;embeddedPrefix:from_"`
	To          Star `gorm:"embedded;embeddedPrefix:to_"`
	AccountID   uint
	Fleets      []Fleet     `gorm:"many2many:route_task_fleet;"`
	AccountInfo AccountInfo `gorm:"-"`
}

func (routeTask *RouteTask) RoutingKey() string {
	if os.Getenv("ENV") == "test" {
		return "TEST_" + "NormalQueue"
	}
	return "NormalQueue"
}
func (routeTask *RouteTask) TaskType() string {
	return "RouteTask"
}

func (routeTask *RouteTask) SetAccountInfo(accountInfo AccountInfo) {
	routeTask.AccountInfo = accountInfo
}

func (routeTask *RouteTask) GetAccountID() uint {
	return routeTask.AccountID
}
func (routeTask *RouteTask) GetID() uint {
	return routeTask.ID
}
func (routeTask *RouteTask) SetID(id uint) {
	routeTask.ID = id
}
