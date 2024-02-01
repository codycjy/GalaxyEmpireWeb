package models

import (
	"gorm.io/gorm"
)

type TaskModel interface {
	RoutingKey() string
	TaskType() string
	SetAccountInfo(accountInfo AccountInfo)
	GetAccountID() uint
	GetID() uint
}
type TaskStatus int

const TaskStatusRunning TaskStatus = 1
const TaskStatusStop TaskStatus = 0

type BaseTask struct {
	gorm.Model
	Name    string     `json:"name"`
	Logs    []taskLog  `gorm:"polymorphic:Refer"`
	Status  TaskStatus `json:"status"`
	Enabled bool       `json:"enabled"`
}

// TODO: task log
type taskLog struct {
	gorm.Model
	ReferID   uint   // 引用的 Task ID
	ReferType string // 引用的 Task 类型
	// 其他字段...
}

func NewLog() *taskLog {

	return &taskLog{}
}
