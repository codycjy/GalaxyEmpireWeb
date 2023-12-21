package taskservice

import (
	"GalaxyEmpireWeb/models"

	"gorm.io/gorm"
)

type taskService struct {
	DB *gorm.DB
	// TODO: Add MQ
}

func NewTaskService(db *gorm.DB) *taskService {
	return &taskService{
		DB: db,
	}
}

func (s *taskService) SendTask(task models.Task) error {
	return task.Send()
}
