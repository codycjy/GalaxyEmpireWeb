package models

import "gorm.io/gorm"

type Fleet struct {
	gorm.Model
	RouteTasks []RouteTask `gorm:"many2many:route_task_fleet;"`
}
