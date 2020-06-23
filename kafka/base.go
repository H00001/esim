package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/prometheus/common/log"
)

type KClient struct {
	asyncProducer sarama.AsyncProducer
	syncProcuder  sarama.SyncProducer
	consumer      sarama.Consumer
	consumerGroup sarama.ConsumerGroup
	config        *sarama.Config
	consumers     map[string]Inbound
	tran          struct {
		encoder sarama.Encoder
	}
}

func (kc *Client) AsyncSend(topic string, msg sarama.Encoder) (int32, int64, error) {
	p := kc.client.asyncProducer
	p.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Value: kc.client.tran.encoder,
	}
	select {
	case res := <-p.Successes():
		return res.Partition, res.Offset, nil
	case err := <-p.Errors():
		log.Errorln("Produced message failure: ", err)
		return 0, 0, err
	}
}

func (kc *Client) SyncSend(topic string, msg sarama.Encoder) (int32, int64, error) {
	partition, offset, err := kc.client.syncProcuder.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: msg,
	})
	return partition, offset, err
}

func (kc *Client) RecvMsg(topic string, fn func(b []byte, e error) bool) {
	partitions, err := kc.client.consumer.Partitions(topic)
	if err != nil {
		log.Fatalln(err.Error())
		fn(nil, err)
	}
	log.Infof("kafka receving msg from topic:%s,partitions:%v", topic, partitions)
	for _, partition := range partitions {
		partitionConsumer, err := kc.client.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Error("partition consumer err:", err)
			continue
		}
		select {
		case res := <-partitionConsumer.Messages():
			fn(res.Value, nil)
		case err := <-partitionConsumer.Errors():
			fn(nil, err)
		}
	}
}

func (c *Client) SetConsumer(topic string, consumer func(interface{})) {
	b := Inbound{consumer: consumer, topic: topic, decoder: defaultDecorder}
	c.client.consumers[topic] = b
}
