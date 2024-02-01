package models

import (
	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

type Task interface {
	PrepareMessage() ([]byte, error)
	GetDelay() int64
}

type TaskItem struct {
	ID    uuid.UUID `json:"id"`    //uuid
	Delay int64     `json:"delay"` // timestamp millisecond
	TaskModel
}

func NewTaskItem(taskModel TaskModel, delay int64) *TaskItem {
	return &TaskItem{
		ID:        uuid.New(),
		TaskModel: taskModel,
		Delay:     delay,
	}
}

func (t TaskItem) PrepareMessage() ([]byte, error) {
	return json.Marshal(t.TaskModel)
}
func (t TaskItem) GetDelay() int64 {
	return t.Delay
}

type TaskResponse struct {
	ID       uuid.UUID       `json:"id"`
	Success  bool            `json:"success"`
	Message  string          `json:"message"`
	TaskID   uint            `json:"task_id"`
	TaskType string          `json:"task_type"`
	Delay    int64           `json:"delay"`
	Data     json.RawMessage `json:"data"` // 用于存储特定任务类型的数据
}
