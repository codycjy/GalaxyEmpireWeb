package models

import "gorm.io/gorm"

type Target struct {
	gorm.Model
	Galaxy  int  `json:"galaxy"`
	System  int  `json:"system"`
	Planet  int  `json:"planet"`
	Is_moon bool `json:"is_moon"`
	TaskID  uint `json:"task_id"`
}
