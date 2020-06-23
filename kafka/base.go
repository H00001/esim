package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/prometheus/common/log"
	"sync"
)

type KClient struct {
	asyncProducer sarama.AsyncProducer
	syncProcuder  sarama.SyncProducer
	consumer      sarama.Consumer
	consumerGroup sarama.ConsumerGroup
	config        *sarama.Config
}

func (kc *KClient) asyncSend(topic string, msg string) (int32, int64, error) {
	p := kc.asyncProducer
	p.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(msg),
	}
	select {
	case res := <-p.Successes():
		return res.Partition, res.Offset, nil
	case err := <-p.Errors():
		log.Errorln("Produced message failure: ", err)
		return 0, 0, err
	}
}

func (kc *KClient) syncSend(topic string, msg string) (int32, int64, error) {
	partition, offset, err := kc.syncProcuder.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(msg),
	})
	return partition, offset, err
}

func (kc *KClient) recvMsg(topic string) {
	partitions, err := kc.consumer.Partitions(topic)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	log.Infof("kafka receving msg from topic:%s,partitions:%v", topic, partitions)
	var wg sync.WaitGroup
	for _, partition := range partitions {
		partitionConsumer, err := kc.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Error("partitionConsumer err:", err)
			continue
		}
		wg.Add(1)
		go func(partitionConsumer sarama.PartitionConsumer) {
			for {
				select {
				case res := <-partitionConsumer.Messages():
					log.Error(sarama.StringEncoder(res.Value))
				case err := <-partitionConsumer.Errors():
					log.Error(err.Error())
				}
			}
			wg.Done()
		}(partitionConsumer)
	}
	wg.Wait()
}
