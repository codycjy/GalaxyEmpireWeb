package taskservice

import (
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"GalaxyEmpireWeb/utils"
	"encoding/json"
	"net/http"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var taskServiceInstance *taskService
var log = logger.GetLogger()

type taskService struct {
	DB *gorm.DB
	MQ *queue.RabbitMQConnection
}
type QueueConfig struct {
	Test []string `yaml:"test"`
	Prod []string `yaml:"prod"`
}

func InitService(db *gorm.DB, mq *queue.RabbitMQConnection) *taskService {
	if taskServiceInstance == nil {
		taskServiceInstance = NewService(db, mq)
	}
	for _, processor := range taskProcessors {
		processor.InitService(db, mq)
	}

	taskGenerator := initTaskGenerator(db, mq, taskServiceInstance)
	go taskGenerator.FindAllTasks()

	return taskServiceInstance
}

func NewService(db *gorm.DB, mq *queue.RabbitMQConnection) *taskService { // newService ?
	return &taskService{
		DB: db,
		MQ: mq,
	}
}

func (s *taskService) SetupQueue(queue []string) {
	for _, name := range queue {
		_, err := s.MQ.Channel.QueueDeclare(
			name,  // 队列名称
			true,  // 是否持久化
			false, // 是否自动删除
			false, // 是否独占
			false, // 是否阻塞
			nil,   // 其他属性
		)

		if err != nil {
			log.Fatal("Failed to declare a queue: %v",
				zap.Error(err),
			)
		}

	}

}

func (s *taskService) sendMessage(message string, queue string) *utils.ServiceError {
	err := s.MQ.Channel.Publish(
		"",    // 交换机
		queue, // 队列名称
		false, // 是否返回
		false, // 是否强制
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to send message", err)
	}
	return nil
}

func (s *taskService) SendTask(task models.TaskModel) *utils.ServiceError {
	queue := task.QueueName()
	jsonStr, err := json.Marshal(task)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "failed to encode task to JSON", err)
	}
	err = s.sendMessage(string(jsonStr), queue)
	if err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to send message", err)
	}
	return nil
}
