package queue

// RabbitMQConsumer 负责从RabbitMQ消费消息
type RabbitMQConsumer struct {
	mq          *RabbitMQConnection
	MessageChan chan []byte
}

// NewRabbitMQConsumer 创建一个新的RabbitMQ消费者实例
func NewRabbitMQConsumer(mq *RabbitMQConnection, messageChan chan []byte) *RabbitMQConsumer {
	return &RabbitMQConsumer{
		mq:          mq,
		MessageChan: messageChan,
	}
}

// StartConsuming 开始消费消息
func (consumer *RabbitMQConsumer) StartConsuming(queueName string) error {
	msgs, err := consumer.mq.Channel.Consume(
		queueName,
		"",
		true, // Auto ack later change it?
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			consumer.MessageChan <- msg.Body
		}
	}()
	return nil
}
