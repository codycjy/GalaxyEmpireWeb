package models

import (
	"GalaxyEmpireWeb/logger"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
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

var log = logger.GetLogger()

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

const (
	TASK_RESULT_RUNNING = 0
	TASK_RESULT_SUCCESS = 1
	TASK_RESULT_FAILED  = 2
)

type Task struct {
	gorm.Model
	Name      string   `json:"name"`
	NextStart int64    `json:"next_start"` // Unix timestamp seconds
	Enabled   bool     `json:"enabled"`
	AccountID uint     `json:"account_id"`
	TaskType  int      `json:"task_type"`
	Status    string   `json:"status"`
	Targets   []Target `json:"targets" gorm:"foreignKey:TaskID"`
	Repeat    int      `json:"repeat"`
	NextIndex int      `json:"next_index"`
	TargetNum int      `json:"target_num"`
	Fleet     Fleet    `json:"fleet" gorm:"foreignKey:TaskID"`
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
		log.Warn("Task::ToSingleTaskRequest: NextIndex out of range")
		if len(t.Targets) == 0 {
			return nil, errors.New("Task::ToSingleTaskRequest: No targets")
		} else {
			t.NextIndex = 0
		}
	}

	return &SingleTaskRequest{
		TaskID:    t.ID,
		UUID:      uuid.NewString(),
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
	UUID      string      `json:"uuid"`
	Name      string      `json:"name"`
	NextStart int64       `json:"next_start"` // Unix timestamp seconds
	Enabled   bool        `json:"enabled"`
	Account   AccountInfo `json:"account"`
	TaskType  int         `json:"task_type"`
	Target    Target      `json:"target"`
	Repeat    int         `json:"repeat"`
	Fleet     Fleet       `json:"fleet"`
}
type SingleTaskResponse struct {
	TaskID        uint   `json:"task_id"`
	UUID          string `json:"uuid"`
	Status        int    `json:"status"` // 0 success, -1 failed
	TaskType      int    `json:"task_type"`
	BackTimestamp int64  `json:"back_timestamp"`
	Message       string `json:"message"`
}

type TaskResponse struct {
	TaskType string          `json:"task_type"`
	Success  bool            `json:"success"`
	Message  string          `json:"message"`
	TaskID   int             `json:"task_id"`
	Data     json.RawMessage `json:"data"` // 用于存储特定任务类型的数据
}
