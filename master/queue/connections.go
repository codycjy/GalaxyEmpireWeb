package queue

import (
	"GalaxyEmpireWeb/config"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

// RabbitMQConnection 管理 RabbitMQ 的连接
type RabbitMQConnection struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

var rabbitMQConnection *RabbitMQConnection

// NewRabbitMQConnection 创建一个新的 RabbitMQ 连接
func NewRabbitMQConnection(cfg *config.RabbitMQConfig) *RabbitMQConnection {
	var connStr string
	if os.Getenv("env") == "test" {
		connStr = os.Getenv("RABBITMQ_STR")
		if connStr == "" {
			log.Fatalf("Failed to connect to RabbitMQ: %v", "RABBITMQ_STR is empty")
		}
	} else {
		connStr = fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
			cfg.RabbitMQ.User,
			cfg.RabbitMQ.Password,
			cfg.RabbitMQ.Host,
			cfg.RabbitMQ.Port,
			cfg.RabbitMQ.Vhost,
		)
	}
	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	return &RabbitMQConnection{
		Conn:    conn,
		Channel: ch,
	}
}

// InitConnection 初始化 RabbitMQ 连接
func InitConnection() {
	rabbitMQConnection = NewRabbitMQConnection(config.GetRabbitMQConfig())
	InitDeclare()

}
func InitDeclare() {
	log.Println("InitDeclare")
	log.Println("DeclareDelayedExchange")
	DeclareDelayedExchange(rabbitMQConnection.Channel)
	log.Println(fmt.Sprintf("DeclareQueue %s", config.TASK_QUEUE_NAME))
	DeclareQueue(rabbitMQConnection.Channel, config.TASK_QUEUE_NAME)
	log.Println(fmt.Sprintf("DeclareQueue %s", config.RESULT_QUEUE_NAME))
	DeclareQueue(rabbitMQConnection.Channel, config.RESULT_QUEUE_NAME)
	log.Println(fmt.Sprintf("DeclareQueue %s", config.INSTANT_QUEUE_NAME))
	DeclareQueue(rabbitMQConnection.Channel, config.INSTANT_QUEUE_NAME)
	log.Println("BindQueue")
	log.Println(fmt.Sprintf("BindQueue %s %s %s", config.TASK_QUEUE_NAME, config.TASK_QUEUE_NAME, config.DELAYED_EXCHANGE_NAME))
	BindQueue(rabbitMQConnection.Channel, config.TASK_QUEUE_NAME, config.TASK_QUEUE_NAME, config.DELAYED_EXCHANGE_NAME)
	// Bind task queue to delayed exchange
}

func DeclareDelayedExchange(ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(
		config.DELAYED_EXCHANGE_NAME,
		"x-delayed-message",
		true,
		false,
		false,
		false,
		map[string]interface{}{
			"x-delayed-type": "direct",
		},
	)
	if err != nil {
		log.Fatalf("Failed to declare delayed exchange: %v", err)
	}
	return nil
}

func DeclareQueue(ch *amqp.Channel, queueName string) error {
	_, err := ch.QueueDeclare(
		queueName, // queueName
		true,      // durable
		false,     // autoDelete
		false,     // exclusive
		false,     // noWait
		nil,       // args
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}
	return nil
}

func BindQueue(ch *amqp.Channel, queueName string, routingKey string, exchangeName string) error {
	err := ch.QueueBind(
		queueName,    // queueName
		routingKey,   // routingKey
		exchangeName, // exchangeName
		false,        // noWait
		nil,          // args
	)
	if err != nil {
		log.Fatalf("Failed to bind queue: %v", err)
	}
	return nil
}

// GetRabbitMQ 获取 RabbitMQ 连接
func GetRabbitMQ() *RabbitMQConnection {
	if rabbitMQConnection == nil {
		InitConnection()
	}
	return rabbitMQConnection

}

// Close 关闭 RabbitMQ 连接
func (r *RabbitMQConnection) Close() {
	if r.Conn != nil {
		r.Conn.Close()
	}
}

func (rmq *RabbitMQConnection) SendNormalMessage(body string, routingKey string) error {
	err := rmq.Channel.Publish(
		"",         // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		return fmt.Errorf("failed to send normal message: %v", err)
	}

	return nil
}

func (rmq *RabbitMQConnection) SendDelayedMessage(body string, routingKey string, delay time.Duration) error {
	delayMs := int(delay / time.Millisecond)
	err := rmq.Channel.Publish(
		config.DELAYED_EXCHANGE_NAME, // exchange
		routingKey,                   // routing key
		false,                        // mandatory
		false,                        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
			Headers: amqp.Table{
				"x-delay": delayMs,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to send delayed message: %v", err)
	}

	return nil
}

func (rmq *RabbitMQConnection) ConsumeNormalMessage(queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := rmq.Channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume normal message: %v", err)
	}

	return msgs, nil
}
