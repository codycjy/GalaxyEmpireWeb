package queue

import (
	"GalaxyEmpireWeb/config"
	"fmt"
	"log"

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
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
		cfg.RabbitMQ.Vhost,
	)

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
func InitConnection() {
	rabbitMQConnection = NewRabbitMQConnection(config.GetRabbitMQConfig())

}
func GetRabbitMQ() *RabbitMQConnection {
	if rabbitMQConnection==nil{
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
