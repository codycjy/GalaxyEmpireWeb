package taskservice

import (
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"GalaxyEmpireWeb/utils"
	"encoding/json"

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

func (s *taskService) sendMessage(message string, queue string) error {
	err := s.MQ.Channel.Publish(
		"",    // 交换机
		queue, // 队列名称
		false, // 是否返回
		false, // 是否强制
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil{
		return utils.NewServiceError(500,"Failed to send message",err)
	}
	return nil
}

func (s *taskService) SendTask(task models.Task) error {
	queue := task.QueueName()
	jsonStr, err := json.Marshal(task)
	if err != nil {
		return utils.NewServiceError(500,"failed to encode task to JSON",err)
	}
	err = s.sendMessage(string(jsonStr), queue)
	if err != nil{
		return utils.NewServiceError(500,"Failed to send message",err)
	}
	return nil
}
func (s *taskService) ConsumeResponseQueue() {
	queueName := ""

	msgs, err := s.MQ.Channel.Consume(
		queueName, // 队列名称
		"",        // 消费者标签 - 不指定则由服务器生成
		true,      // 自动应答
		false,     // 独占模式
		false,     // 不发送给同一连接的消费者
		false,     // 阻塞
		nil,       // 参数
	)
	if err != nil {
		log.Fatal("Failed to register a consumer",
			zap.Error(err),
		)
	}

	go func() {
		for msg := range msgs {
			s.processResponseMessage(msg)
		}
	}()
}

func (s *taskService) processResponseMessage(msg amqp.Delivery) {
	// TODO: process response message
	jsonStr := string(msg.Body)
	log.Info("[service]Received a message: %s",
		zap.String("json body", jsonStr),
	)
	var taskResponse models.TaskResponse
	err := json.Unmarshal([]byte(jsonStr), &taskResponse)
	if err != nil {
		log.Error("[service]Unmarshal task response failed",
			zap.Error(err),
		)
		msg.Ack(false)
	}
	s.processTaskResponse(&taskResponse)
	msg.Ack(false)

}

func (s *taskService) processTaskResponse(taskResponse *models.TaskResponse) {
	if !taskResponse.Success {
		s.taskRetry(taskResponse.TaskID)
		return
	}
	switch taskResponse.TaskType {
	case models.RouteTaskName:
		{
			taskProcessors[models.RouteTaskName].ProcessTask(&taskResponse.Data)
		}
	case models.PlanTaskName:
		{
			taskProcessors[models.PlanTaskName].ProcessTask(&taskResponse.Data)

		}

	}
}
func (s *taskService) taskRetry(taskID int) {

}
