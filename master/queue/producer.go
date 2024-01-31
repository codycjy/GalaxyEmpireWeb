package queue

import "go.uber.org/zap"

type RabbitMQProducer struct {
	mq          *RabbitMQConnection
	MessageChan chan DelayedMessage
}

type DelayedMessage struct {
	RoutingKey string
	Body       []byte
	Delay      int64
}

func NewRabbitMQProducer(mq *RabbitMQConnection, messageChan chan DelayedMessage) *RabbitMQProducer {
	return &RabbitMQProducer{
		mq:          mq,
		MessageChan: messageChan,
	}
}

func (producer *RabbitMQProducer) StartPublishing(exchange string) {
	go func() {
		for msg := range producer.MessageChan {
			// 在这里将消息发布到消息队列
			err := producer.mq.PublishWithDelay(
				msg.RoutingKey,
				msg.Body,
				msg.Delay,
			)
			if err != nil {
				log.Error("[queue] RabbitMQ Producer - fail to publish message",
					zap.Error(err),
				)
			}
		}
	}()
}
