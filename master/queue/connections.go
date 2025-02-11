package queue

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/logger"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// RabbitMQConnection 管理 RabbitMQ 的连接
type RabbitMQConnection struct {
	Conn            *amqp.Connection
	Channel         *amqp.Channel
	closed          bool
	mutex           sync.Mutex
	reconnecting    bool
	reconnectingMux sync.Mutex
}

const (
	reconnectDelay = 5 * time.Second
	maxRetries     = 100
)

var log = logger.GetLogger()

var rabbitMQConnection *RabbitMQConnection

// NewRabbitMQConnection 创建一个新的 RabbitMQ 连接
func NewRabbitMQConnection(cfg *config.RabbitMQConfig) *RabbitMQConnection {
	var connStr string
	connStr = fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
		cfg.RabbitMQ.Vhost,
	)
	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: %v", zap.Error(err))
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to create channel: %v", zap.Error(err))
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
	log.Info("InitDeclare")
	log.Info("DeclareDelayedExchange")

	DeclareDelayedExchange(rabbitMQConnection.Channel)
	log.Info(fmt.Sprintf("DeclareQueue %s", config.TASK_QUEUE_NAME))
	DeclareQueue(rabbitMQConnection.Channel, config.TASK_QUEUE_NAME)
	log.Info(fmt.Sprintf("DeclareQueue %s", config.RESULT_QUEUE_NAME))
	DeclareQueue(rabbitMQConnection.Channel, config.RESULT_QUEUE_NAME)
	log.Info(fmt.Sprintf("DeclareQueue %s", config.INSTANT_QUEUE_NAME))
	DeclareQueue(rabbitMQConnection.Channel, config.INSTANT_QUEUE_NAME)
	log.Info("BindQueue")
	log.Info(fmt.Sprintf("BindQueue %s %s %s", config.TASK_QUEUE_NAME, config.TASK_QUEUE_NAME, config.DELAYED_EXCHANGE_NAME))
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
		log.Fatal("Failed to declare delayed exchange: %v", zap.Error(err))
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
		log.Fatal("Failed to declare queue: %v", zap.Error(err))
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
		log.Fatal("Failed to bind queue: %v", zap.Error(err))
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

func (rmq *RabbitMQConnection) SendNormalMessage(body string, routingKey string) error {
	rmq.mutex.Lock()
	defer rmq.mutex.Unlock()

	for i := 0; i < maxRetries; i++ {
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
			log.Info("Failed to send normal message: %v, retry: %d", zap.Error(err), zap.Int("retry", i+1))
			rmq.reconnect()
			continue
		}
		return nil
	}
	return fmt.Errorf("failed to send normal message after %d attempts", maxRetries)
}

// SendDelayedMessage 添加重连机制
func (rmq *RabbitMQConnection) SendDelayedMessage(body string, routingKey string, delay time.Duration) error {
	rmq.mutex.Lock()
	defer rmq.mutex.Unlock()

	delayMs := delay.Milliseconds()
	for i := 0; i < maxRetries; i++ {
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
				DeliveryMode: amqp.Persistent,
			})
		if err != nil {
			log.Info("Failed to send delayed message: %v, retry: %d", zap.Error(err), zap.Int("retry", i+1))
			rmq.reconnect()
			continue
		}
		return nil
	}
	return fmt.Errorf("failed to send delayed message after %d attempts", maxRetries)
}

// 添加连接健康检查方法
func (rmq *RabbitMQConnection) ensureConnected() error {
	if rmq.closed {
		return fmt.Errorf("connection is closed")
	}

	// 检查channel是否正常
	if rmq.Channel == nil {
		log.Info("Channel is nil, attempting to reconnect...")
		rmq.reconnect()
	}

	// 检查connection是否正常
	if rmq.Conn == nil || rmq.Conn.IsClosed() {
		log.Info("Connection is nil or closed, attempting to reconnect...")
		rmq.reconnect()
	}

	return nil
}

// ConsumeNormalMessage 添加重连机制
// 修改消费者的通知处理
func (rmq *RabbitMQConnection) ConsumeNormalMessage(queueName string) (<-chan amqp.Delivery, error) {
	deliveries := make(chan amqp.Delivery)

	go func() {
		for {
			if rmq.closed {
				return
			}

			// 检查是否正在重连
			rmq.reconnectingMux.Lock()
			if rmq.reconnecting {
				rmq.reconnectingMux.Unlock()
				time.Sleep(reconnectDelay)
				continue
			}
			rmq.reconnectingMux.Unlock()

			msgs, err := rmq.Channel.Consume(
				queueName,
				"",
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				log.Info("Failed to consume message", zap.Error(err))
				rmq.reconnect()
				continue
			}

			chanClose := rmq.Channel.NotifyClose(make(chan *amqp.Error))
			connClose := rmq.Conn.NotifyClose(make(chan *amqp.Error))

			for {
				select {
				case d, ok := <-msgs:
					if !ok {
						break
					}
					deliveries <- d
				case <-chanClose:
					log.Info("Channel closed, reconnecting...")
					rmq.reconnect()
					goto RECONNECT
				case <-connClose:
					log.Info("Connection closed, reconnecting...")
					rmq.reconnect()
					goto RECONNECT
				}
			}

		RECONNECT:
			continue
		}
	}()

	return deliveries, nil
}

// reconnect 重连方法
func (rmq *RabbitMQConnection) safeClose(timeout time.Duration) {
	if rmq.Channel != nil {
		done := make(chan bool)
		go func() {
			rmq.Channel.Close()
			done <- true
		}()

		select {
		case <-done:
			log.Info("Channel closed successfully")
		case <-time.After(timeout):
			log.Info("Channel close timeout")
		}
	}

	if rmq.Conn != nil {
		done := make(chan bool)
		go func() {
			rmq.Conn.Close()
			done <- true
		}()

		select {
		case <-done:
			log.Info("Connection closed successfully")
		case <-time.After(timeout):
			log.Info("Connection close timeout")
		}
	}
}

func (rmq *RabbitMQConnection) reconnect() {
	// 首先尝试获取重连锁
	rmq.reconnectingMux.Lock()
	if rmq.reconnecting {
		rmq.reconnectingMux.Unlock()
		log.Info("Reconnection already in progress, skipping...")
		return
	}
	rmq.reconnecting = true
	rmq.reconnectingMux.Unlock()

	// 完成后释放重连状态
	defer func() {
		rmq.reconnectingMux.Lock()
		rmq.reconnecting = false
		rmq.reconnectingMux.Unlock()
	}()

	rmq.mutex.Lock()
	defer rmq.mutex.Unlock()

	for i := 0; i < maxRetries; i++ {
		log.Info(fmt.Sprintf("Reconnecting to RabbitMQ, attempt: %d", i+1))
		if rmq.closed {
			return
		}

		// 安全关闭现有连接
		rmq.safeClose(3 * time.Second)
		log.Info("Closed existing connection")

		// 重新连接
		var connStr string
		cfg := config.GetRabbitMQConfig()
		connStr = fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
			cfg.RabbitMQ.User,
			cfg.RabbitMQ.Password,
			cfg.RabbitMQ.Host,
			cfg.RabbitMQ.Port,
			cfg.RabbitMQ.Vhost,
		)

		// 设置连接超时
		dialConfig := amqp.Config{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, 5*time.Second)
			},
		}
		conn, err := amqp.DialConfig(connStr, dialConfig)
		if err != nil {
			log.Info(fmt.Sprintf("Failed to reconnect to RabbitMQ: %v, retry: %d", err, i+1))
			time.Sleep(reconnectDelay)
			continue
		}
		log.Info("Connected to RabbitMQ")

		ch, err := conn.Channel()
		if err != nil {
			log.Info(fmt.Sprintf("Failed to create channel: %v, retry: %d", err, i+1))
			conn.Close()
			time.Sleep(reconnectDelay)
			continue
		}
		log.Info("Created channel")

		rmq.Conn = conn
		rmq.Channel = ch

		return
	}
	log.Info(fmt.Sprintf("Failed to reconnect to RabbitMQ after %d attempts", maxRetries))
}

// Close 优化关闭方法
func (r *RabbitMQConnection) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.closed = true
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}
