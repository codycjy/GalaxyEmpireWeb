package models

import (
	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

type Task interface {
	PrepareMessage() ([]byte, error)
}

type taskItem struct {
	ID    uuid.UUID `json:"id"`    //uuid
	Delay int64     `json:"delay"` // timestamp millisecond
	TaskModel
}

func NewTaskItem(taskModel TaskModel, delay int64) *taskItem {
	return &taskItem{
		ID:        uuid.New(),
		TaskModel: taskModel,
		Delay:     delay,
	}
}

func (t taskItem) PrepareMessage() ([]byte, error) {
	return json.Marshal(t.TaskModel)
}

type TaskResponse struct {
	ID      uuid.UUID       `json:"id"`
	Success bool            `json:"success"`
	Message string          `json:"message"`
	TaskID  int             `json:"task_id"`
	Delay   int64 		`json:"delay"`
	Data    json.RawMessage `json:"data"` // 用于存储特定任务类型的数据
}
