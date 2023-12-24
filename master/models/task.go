package models

import (
	"os"
	"time"

	"gorm.io/gorm"
)

type Task interface {
	QueueName() string
	TaskType() string
}
type BaseTask struct {
	gorm.Model
	Logs []TaskLog `gorm:"polymorphic:Refer"`
}
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

type DailyTask struct {
	BaseTask
	AccountID uint
}

func (t *DailyTask) QueueName() string {
	if os.Getenv("ENV") == "test" {
		return "Test_" + "HPQueue"
	}
	return "HPQueue"
}
func (t *DailyTask) TaskType() string {
	return "DailyTask"
}

type ExtraTask struct {
	BaseTask
	AccountID uint
}

func (t *ExtraTask) QueueName() string {
	if os.Getenv("ENV") == "test" {
		return "Test_" + "HPQueue"
	}
	return "HPQueue"
}
func (t *ExtraTask) TaskType() string {
	return "ExtraTask"
}

type TaskResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	TaskID      int       `json:"task_id"`
	Next        time.Time `json:"next"`
	OriginQueue string    `json:"origin_queue"`
}

func (t *TaskResponse) QueueName() string {
	if os.Getenv("ENV") == "test" {
		return "Test_" + "ResponseQueue"
	}
	return "ResponseQueue"
}

// TODO: task log
type TaskLog struct {
	gorm.Model
	ReferID   uint   // 引用的 Task ID
	ReferType string // 引用的 Task 类型
	// 其他字段...
}

