package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Task interface {
	QueueName() string
	TaskType() string
}

type BaseTask struct {
	gorm.Model
	Name      string    `json:"name"`
	Logs      []taskLog `gorm:"polymorphic:Refer"`
	NextStart time.Time `json:"next_start"`
	Enabled   bool      `json:"enabled"`
}

type TaskResponse struct {
	TaskType string          `json:"task_type"`
	Success  bool            `json:"success"`
	Message  string          `json:"message"`
	TaskID   int             `json:"task_id"`
	Data     json.RawMessage `json:"data"` // 用于存储特定任务类型的数据
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
