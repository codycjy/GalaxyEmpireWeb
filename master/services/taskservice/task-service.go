package taskservice

import (
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var taskServiceInstance *taskService
var log = logger.GetLogger()

type taskService struct {
	DB *gorm.DB
}
type QueueConfig struct {
	Test []string `yaml:"test"`
	Prod []string `yaml:"prod"`
}

func InitService(rdb *redis.Client, db *gorm.DB, taskChan chan *models.TaskItem, messageChan chan *queue.DelayedMessage, responseChan chan []byte) {
	if taskServiceInstance == nil {
		taskServiceInstance = NewService(db)
	}

	taskGenerator := getTaskGenerator(db, taskChan, messageChan)
	taskHandler := getTaskHandler(rdb, db, responseChan)

	go taskGenerator.Start()
	go taskHandler.HandleResponse()

}

func NewService(db *gorm.DB) *taskService {
	return &taskService{
		DB: db,
	}
}
