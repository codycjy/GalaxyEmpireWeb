package taskservice

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"encoding/json"
	"log"
	"os"

	"github.com/streadway/amqp"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

var taskServiceInstance *taskService

type taskService struct {
	DB *gorm.DB
	MQ *queue.RabbitMQConnection
}
type QueueConfig struct {
	Test []string `yaml:"test"`
	Prod []string `yaml:"prod"`
}

func GetQueueNames() []string {
	// 读取环境变量
	env := os.Getenv("env")

	// 读取配置文件
	yamlFile, err := os.ReadFile("queue/queues.yaml")
	if err != nil {
		log.Fatalf("yamlFile.Get err   #%v ", err)
	}

	var config QueueConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	// 根据环境变量选择队列
	if env == "test" {
		return config.Test
	}
	return config.Prod
}

func InitService(db *gorm.DB, mq *queue.RabbitMQConnection) *taskService {
	if taskServiceInstance == nil {
		taskServiceInstance = &taskService{DB: db, MQ: mq}
	}
	return taskServiceInstance
}

func NewService(db *gorm.DB, mq *queue.RabbitMQConnection) *taskService {
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
			log.Fatalf("Failed to declare a queue: %v", err)
		}

	}

}

func (s *taskService) SendMessage(message string, queue string) error {
	err := s.MQ.Channel.Publish(
		"",    // 交换机
		queue, // 队列名称
		false, // 是否返回
		false, // 是否强制
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	return err
}

func (s *taskService) SendTask(task models.Task) error {
	queue := task.QueueName()
	jsonStr, err := json.Marshal(task)
	if err != nil {
		return err
	}
	err = s.SendMessage(string(jsonStr), queue)
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
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			s.processResponseMessage(msg)
		}
	}()
}

func (s *taskService) processResponseMessage(msg amqp.Delivery) {
	// TODO: process response message
	log.Printf("Received a message: %s", msg.Body)
	jsonStr := string(msg.Body)
	var taskResponse models.TaskResponse
	err := json.Unmarshal([]byte(jsonStr), &taskResponse)
	if err != nil {
		log.Printf("Unmarshal error: %v", err)
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

}
func (s *taskService) taskRetry(taskID int) {

}

