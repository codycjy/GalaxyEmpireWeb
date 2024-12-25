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

type FleetDTO struct {
	LightFighter int `json:"lf"`
	HeavyFighter int `json:"hf"`
	Cruiser      int `json:"cr"`
	Battleship   int `json:"bs"`
	Dreadnought  int `json:"dr"`
	Destroyer    int `json:"de"`
	Deathstar    int `json:"ds"`
	Bomber       int `json:"bomb"`
	Guardian     int `json:"guard"`
	Satellite    int `json:"satellite"`
	Cargo        int `json:"cargo"`
}

func (fleet Fleet) ToDTO() *FleetDTO {
	return &FleetDTO{
		LightFighter: fleet.LightFighter,
		HeavyFighter: fleet.HeavyFighter,
		Cruiser:      fleet.Cruiser,
		Battleship:   fleet.Battleship,
		Dreadnought:  fleet.Dreadnought,
		Destroyer:    fleet.Destroyer,
		Deathstar:    fleet.Deathstar,
		Bomber:       fleet.Bomber,
		Guardian:     fleet.Guardian,
		Satellite:    fleet.Satellite,
		Cargo:        fleet.Cargo,
	}
}
func (fleet Fleet) GetEntityPrefix() string {
	return "fleet_"
}
