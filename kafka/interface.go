package kafka

import "github.com/Shopify/sarama"

type IkClient interface {
	AsyncSend(topic string, msg sarama.Encoder) (int32, int64, error)
	SyncSend(topic string, msg sarama.Encoder) (int32, int64, error)
	ReceiveOnceMessage(topic string) ([]byte, error)
	SetConsumer(topic string, consumer func(interface{}))
}
