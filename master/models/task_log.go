package models

import "gorm.io/gorm"

type TaskLog struct {
	gorm.Model
	TaskID uint   `json:"task_id"`
	UUID   string `json:"uuid"`
	Status int    `json:"status"`
}
