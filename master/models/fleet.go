package models

import "gorm.io/gorm"

type Fleet struct {
	gorm.Model
	LightFighter int  `json:"lf"`
	HeavyFighter int  `json:"hf"`
	Cruiser      int  `json:"cr"`
	Battleship   int  `json:"bs"`
	Dreadnought  int  `json:"dr"`
	Destroyer    int  `json:"de"`
	Deathstar    int  `json:"ds"`
	Bomber       int  `json:"bomb"`
	Guardian     int  `json:"guard"`
	Satellite    int  `json:"satellite"`
	Cargo        int  `json:"cargo"`
	TaskID       uint `json:"task_id"`
}

func (fleet Fleet) GetEntityPrefix() string {
	return "fleet_"
}
