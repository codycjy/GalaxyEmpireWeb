package models

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

// Enum TaskType
const (
	TASKTYPE_ATTACK = iota
	TASKTYPE_EXPLORE
	TASKTYPE_LOGIN
	MISSIONTYPE_ATTACK
	MISSIONTYPE_EXPLORE
)

const (
	TASK_STATUS_RUNNING = iota
	TASK_STATUS_WAITING
	TASK_STATUS_READY
)

var TaskTypeMap = map[int]int{
	TASKTYPE_ATTACK:  1,
	TASKTYPE_EXPLORE: 4,
	TASKTYPE_LOGIN:   99,
}
var MissionTypeMap = map[int]int{
	MISSIONTYPE_ATTACK:  1,
	MISSIONTYPE_EXPLORE: 15,
}
var TaskStatusMap = map[int]string{
	TASK_STATUS_RUNNING: "running",
	TASK_STATUS_WAITING: "waiting",
	TASK_STATUS_READY:   "ready",
} // TODO: need to rethink the status

type Task struct {
	gorm.Model
	Name      string    `json:"name"`
	NextStart time.Time `json:"next_start"`
	Enabled   bool      `json:"enabled"`
	AccountID uint      `json:"account_id"`
	TaskType  int       `json:"task_type"`
	Status    string    `json:"status"`
	Targets   []Target  `json:"targets" gorm:"foreignKey:TaskID"`
	Repeat    int       `json:"repeat"`
	NextIndex int       `json:"next_index"`
	TargetNum int       `json:"target_num"`
	Fleet     Fleet     `json:"fleet" gorm:"foreignKey:TaskID"`
}

func (t Task) ToDTO() *TaskDTO {
	log.Fatal("Task ToDTO not implemented")
	return &TaskDTO{}
}

func (t Task) GetEntityPrefix() string {
	return "task_"
}
func (t *Task) ToSingleTaskRequest() (*SingleTaskRequest, error) {
	if t.NextIndex >= len(t.Targets) {
		return nil, errors.New("NextIndex out of range")
	}

	return &SingleTaskRequest{
		TaskID:    t.ID,
		Name:      t.Name,
		NextStart: t.NextStart,
		Enabled:   t.Enabled,
		Account:   AccountInfo{},
		TaskType:  t.TaskType,
		Target:    t.Targets[t.NextIndex],
		Repeat:    t.Repeat,
		Fleet:     Fleet{},
	}, nil
}
func (t *Task) UpdateNextIndex() {
	t.NextIndex = (t.NextIndex + 1) % t.TargetNum
}

type TaskDTO struct { // TODO: finish func
	gorm.Model
	Name      string    `json:"name"`
	NextStart time.Time `json:"next_start"`
	Enabled   bool      `json:"enabled"`
	AccountID uint      `json:"account_id"`
	TaskType  int       `json:"task_type"`
	Targets   []Target  `json:"targets" gorm:"foreignKey:TaskID"`
	Repeat    int       `json:"repeat"`
	NextIndex int       `json:"next_index"`
	TargetNum int       `json:"target_num"`
	Fleet     Fleet     `json:"fleet" gorm:"foreignKey:TaskID"`
}

type SingleTaskRequest struct {
	TaskID    uint        `json:"task_id"`
	Name      string      `json:"name"`
	NextStart time.Time   `json:"next_start"`
	Enabled   bool        `json:"enabled"`
	Account   AccountInfo `json:"account"`
	TaskType  int         `json:"task_type"`
	Target    Target      `json:"target"`
	Repeat    int         `json:"repeat"`
	Fleet     Fleet       `json:"fleet"`
}
type SingleTaskResponse struct {
	TaskID        uint  `json:"task_id"`
	Status        int   `json:"status"` // 0 success, -1 failed
	TaskType      int   `json:"task_type"`
	BackTimestamp int64 `json:"back_timestamp"`
}

type TaskResponse struct {
	TaskType string          `json:"task_type"`
	Success  bool            `json:"success"`
	Message  string          `json:"message"`
	TaskID   int             `json:"task_id"`
	Data     json.RawMessage `json:"data"` // 用于存储特定任务类型的数据
}
