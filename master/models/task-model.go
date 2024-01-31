package models

import (
	"time"

	"gorm.io/gorm"
)

type TaskModel interface {
	QueueName() string
	TaskType() string
	SetAccountInfo(accountInfo AccountInfo)
	GetAccountID() uint
	GetID() uint
}

type BaseTask struct {
	gorm.Model
	Name      string    `json:"name"`
	Logs      []taskLog `gorm:"polymorphic:Refer"`
	NextStart time.Time `json:"next_start"`
	Enabled   bool      `json:"enabled"`
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
