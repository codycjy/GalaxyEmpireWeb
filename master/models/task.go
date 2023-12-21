package models

import (
	"time"

	"gorm.io/gorm"
)

type Task interface {
	Send() error
}

type RouteTask struct {
	gorm.Model
	Repeat    int
	From      Star      `gorm:"embedded;embeddedPrefix:from_"`
	To        Star      `gorm:"embedded;embeddedPrefix:to_"`
	NextStart time.Time `json:"next_start"`
	AccountID uint
	Fleets    []Fleet `gorm:"many2many:route_task_fleet;"`
}

func (routeTask *RouteTask) Send() error {
	return nil //TODO: Implement
}

type DailyTask struct {
	gorm.Model
	AccountID uint
}

func (dailyTask *DailyTask) Send() error {
	return nil //TODO: Implement
}

type ExtraTask struct {
	gorm.Model
	AccountID uint
}

func (extraTask *ExtraTask) Send() error {
	return nil //TODO: Implement
}
