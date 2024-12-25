package models

import "gorm.io/gorm"

type TaskLog struct {
	gorm.Model
	TaskID uint   `json:"task_id"`
	UUID   string `json:"uuid" gorm:"unique"` // Unique limit
	Status int    `json:"status"`
}
