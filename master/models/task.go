package models

import (
	"GalaxyEmpireWeb/logger"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Enum TaskType
const (
	TASKTYPE_ATTACK          = 1
	TASKTYPE_EXPLORE         = 4
	TASKTYPE_LOGIN           = 99
	TASKTYPE_QUERY_PLANET_ID = 100
	MISSIONTYPE_ATTACK       = 1
	MISSIONTYPE_EXPLORE      = 15
)

const (
	TASK_STATUS_RUNNING = iota
	TASK_STATUS_WAITING
	TASK_STATUS_READY
)

var log = logger.GetLogger()

var TaskStatusMap = map[int]string{
	TASK_STATUS_RUNNING: "running",
	TASK_STATUS_WAITING: "waiting",
	TASK_STATUS_READY:   "ready",
} // TODO: need to rethink the status

const (
	TASK_RESULT_RUNNING = 0 // TODO: use var to def type
	TASK_RESULT_SUCCESS = 1
	TASK_RESULT_FAILED  = 2
)

type Task struct {
	gorm.Model
	Name          string      `json:"name"`
	NextStart     int64       `json:"next_start"`
	Enabled       bool        `json:"enabled"`
	AccountID     uint        `json:"account_id"`
	TaskType      int         `json:"task_type"`
	Status        string      `json:"status"`
	StartPlanet   StartPlanet `json:"start_planet"`
	StartPlanetID uint        `json:"start_planet_id"`
	Targets       []Target    `json:"targets" gorm:"foreignKey:TaskID"`
	Repeat        int         `json:"repeat"`
	NextIndex     int         `json:"next_index"`
	TargetNum     int         `json:"target_num"`
	Fleet         Fleet       `json:"fleet" gorm:"foreignKey:TaskID"`
}

func (t Task) ToDTO() *TaskDTO {
	return &TaskDTO{
		Model:       t.Model,
		Name:        t.Name,
		NextStart:   time.Unix(t.NextStart, 0),
		Enabled:     t.Enabled,
		AccountID:   t.AccountID,
		TaskType:    t.TaskType,
		Targets:     t.Targets,
		StartPlanet: t.StartPlanet,
		Repeat:      t.Repeat,
		TargetNum:   len(t.Targets),
		Fleet:       t.Fleet,
	}
}

func (t Task) GetEntityPrefix() string {
	return "task_"
}

func (t *Task) ToSingleTaskRequest(account *Account) (*SingleTaskRequest, error) {
	// 基础验证
	if len(t.Targets) == 0 {
		log.Error("Task::ToSingleTaskRequest: no targets",
			zap.Uint("task_id", t.ID),
			zap.String("task_name", t.Name))
		return nil, errors.New("no targets available for task")
	}

	// Validate NextIndex
	if t.NextIndex >= len(t.Targets) {
		log.Error("Task::ToSingleTaskRequest: invalid next_index",
			zap.Uint("task_id", t.ID),
			zap.Int("next_index", t.NextIndex),
			zap.Int("targets_length", len(t.Targets)))
		return nil, errors.New("invalid next_index")
	}

	// 验证账号信息
	if account == nil {
		log.Error("Task::ToSingleTaskRequest: account is nil",
			zap.Uint("task_id", t.ID))
		return nil, errors.New("account information missing")
	}

	// 验证并更新 NextIndex
	currentIndex := t.NextIndex
	t.NextIndex = (t.NextIndex + 1) % len(t.Targets)

	log.Debug("Task::ToSingleTaskRequest: preparing task",
		zap.Uint("task_id", t.ID),
		zap.String("task_name", t.Name),
		zap.Int("current_index", currentIndex),
		zap.Int("next_index", t.NextIndex),
		zap.Int("targets_count", len(t.Targets)))

	return &SingleTaskRequest{
		TaskID:        t.ID,
		UUID:          uuid.NewString(),
		Name:          t.Name,
		NextStart:     t.NextStart,
		Enabled:       t.Enabled,
		Account:       *account.ToInfo(),
		TaskType:      t.TaskType,
		StartPlanet:   t.StartPlanet,
		StartPlanetID: t.StartPlanetID,
		Target:        t.Targets[currentIndex], // 使用当前索引
		Repeat:        t.Repeat,
		Fleet:         t.Fleet.ToDTO(),
	}, nil
}

// TaskUpdateDTO 用于部分更新Task的DTO
type TaskUpdateDTO struct {
	Name          *string      `json:"name,omitempty"`
	NextStart     *int64       `json:"next_start,omitempty"`
	Enabled       *bool        `json:"enabled,omitempty"`
	TaskType      *int         `json:"task_type,omitempty"`
	StartPlanet   *StartPlanet `json:"start_planet,omitempty"`
	StartPlanetID *uint        `json:"start_planet_id,omitempty"`
	Targets       *[]Target    `json:"targets,omitempty"`
	Repeat        *int         `json:"repeat,omitempty"`
	Fleet         *Fleet       `json:"fleet,omitempty"`
}

func (t *Task) ApplyUpdates(updates *TaskUpdateDTO) map[string]interface{} {
	if updates == nil {
		return nil
	}

	updateMap := make(map[string]interface{})

	if updates.Name != nil {
		t.Name = *updates.Name
		updateMap["name"] = t.Name
	}
	if updates.NextStart != nil {
		t.NextStart = *updates.NextStart
		updateMap["next_start"] = t.NextStart
	}
	if updates.Enabled != nil {
		t.Enabled = *updates.Enabled
		updateMap["enabled"] = t.Enabled
	}
	if updates.TaskType != nil {
		t.TaskType = *updates.TaskType
		updateMap["task_type"] = t.TaskType
	}
	if updates.Repeat != nil {
		t.Repeat = *updates.Repeat
		updateMap["repeat"] = t.Repeat
	}
	if updates.StartPlanetID != nil {
		t.StartPlanetID = *updates.StartPlanetID
		updateMap["start_planet_id"] = t.StartPlanetID
	}
	if updates.Targets != nil {
		t.TargetNum = len(*updates.Targets)
		t.NextIndex = 0
		updateMap["target_num"] = t.TargetNum
		updateMap["next_index"] = t.NextIndex
	}

	return updateMap
}

type TaskDTO struct { // TODO: finish func
	gorm.Model
	Name        string      `json:"name"`
	NextStart   time.Time   `json:"next_start"`
	Enabled     bool        `json:"enabled"`
	AccountID   uint        `json:"account_id"`
	TaskType    int         `json:"task_type"`
	Targets     []Target    `json:"targets" gorm:"foreignKey:TaskID"`
	Repeat      int         `json:"repeat"`
	NextIndex   int         `json:"next_index"`
	StartPlanet StartPlanet `json:"start_planet"`
	TargetNum   int         `json:"target_num"`
	Fleet       Fleet       `json:"fleet" gorm:"foreignKey:TaskID"`
}

type SingleTaskRequest struct {
	TaskID        uint        `json:"task_id"`
	UUID          string      `json:"uuid"`
	Name          string      `json:"name"`
	NextStart     int64       `json:"next_start"` // Unix timestamp seconds
	Enabled       bool        `json:"enabled"`
	Account       AccountInfo `json:"account"`
	TaskType      int         `json:"task_type"`
	StartPlanet   StartPlanet `json:"start_planet"`
	StartPlanetID uint        `json:"start_planet_id"`
	Target        Target      `json:"target"`
	Repeat        int         `json:"repeat"`
	Fleet         *FleetDTO   `json:"fleet"`
}
type SingleTaskResponse struct {
	TaskID        uint   `json:"task_id"`
	UUID          string `json:"uuid"`
	Status        int    `json:"status"` // 0 success, -1 failed
	TaskType      int    `json:"task_type"`
	BackTimestamp int64  `json:"back_ts"`
	Msg           string `json:"msg"`
	ErrMsg        string `json:"err_msg"`
}

type TaskResponse struct {
	TaskType string          `json:"task_type"`
	Success  bool            `json:"success"`
	Message  string          `json:"message"`
	TaskID   int             `json:"task_id"`
	Data     json.RawMessage `json:"data"` // 用于存储特定任务类型的数据
}
