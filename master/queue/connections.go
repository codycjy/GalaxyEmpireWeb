package queue

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/logger"
	"fmt"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// RabbitMQConnection 管理 RabbitMQ 的连接
type RabbitMQConnection struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

var log = logger.GetLogger()

var rabbitMQConnection *RabbitMQConnection

// NewRabbitMQConnection 创建一个新的 RabbitMQ 连接
func NewRabbitMQConnection(cfg *config.RabbitMQConfig) *RabbitMQConnection {
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
		cfg.RabbitMQ.Vhost,
	)

	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: %v",
			zap.Error(err),
		)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel: %v", zap.Error(err))
	}

	return &RabbitMQConnection{
		Conn:    conn,
		Channel: ch,
	}
}

// InitConnection 初始化 RabbitMQ 连接
func InitConnection() {
	rabbitMQConnection = NewRabbitMQConnection(config.GetRabbitMQConfig())

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
func (rmq *RabbitMQConnection) publish(exchange string, routingKey string, body []byte, delay int64) error {
	headers := make(amqp.Table)

	log.Debug("publishing to %q %q",
		zap.String("exchange", exchange),
		zap.String("routingKey", routingKey),
		zap.ByteString("body", body),
	)

	if delay != 0 {
		headers["x-delay"] = delay
	}

	return rmq.Channel.Publish(exchange, routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         body,
		Headers:      headers,
	})
}

// PublishWithDelay
// delay millisecond
func (rmq *RabbitMQConnection) PublishWithDelay(routingKey string, body []byte, delay int64) error {
	return rmq.publish("delayed", routingKey, body, delay)
}
