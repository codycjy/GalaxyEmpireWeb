package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Target struct {
	gorm.Model
	Galaxy  int  `json:"galaxy"`
	System  int  `json:"system"`
	Planet  int  `json:"planet"`
	Is_moon bool `json:"is_moon"`
	TaskID  uint `json:"task_id"`
}
type StartPlanet struct {
	Target
}

func (t Target) String() string {
	is_moon_int := 0
	if t.Is_moon {
		is_moon_int = 1
	}
	return fmt.Sprintf("%d:%d:%d:%d", t.Galaxy, t.System, t.Planet, is_moon_int)
}
