package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/prometheus/common/log"
	"sync"
)

type KClient struct {
	asyncProducer sarama.AsyncProducer
	syncProducer  sarama.SyncProducer
	consumer      sarama.Consumer
	consumerGroup sarama.ConsumerGroup
	config        *sarama.Config
	consumers     map[string]Inbound
	tran          struct {
		encoder sarama.Encoder
	}
}

func (c *Client) AsyncSend(topic string, msg sarama.Encoder) (int32, int64, error) {
	p := c.client.asyncProducer
	p.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Value: c.client.tran.encoder,
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
	partition, offset, err := kc.client.syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: msg,
	})
	return partition, offset, err
}

func (kc *Client) ReceiveOnceMessage(topic string) ([]byte, error) {
	var err error
	var data []byte
	partitions, err := kc.client.consumer.Partitions(topic)
	if err != nil {
		return nil, err
	}
	kc.consume(partitions, topic, false, true, func(b []byte, e error) bool {
		data = b
		err = e
		return false
	})
	return data, err
}

func (kc *Client) consume(partitions []int32, topic string, parallel bool, s bool, fn func(b []byte, e error) bool) *sync.WaitGroup {
	wg := sync.WaitGroup{}
	for _, partition := range partitions {
		partitionConsumer, err := kc.client.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Error("partition consumer err:", err)
			continue
		}
		if parallel {
			go kc.doConsumer(partitionConsumer, fn)(&wg)
		} else {
			kc.doConsumer(partitionConsumer, fn)(nil)
		}
	}
	if s {
		wg.Wait()
		return nil
	}
	return &wg
}

func (kc *Client) doConsumer(p sarama.PartitionConsumer, fn func(b []byte, e error) bool) func(wg *sync.WaitGroup) {
	return func(wg *sync.WaitGroup) {
		isComplete := false
		if wg != nil {
			wg.Add(1)
			defer wg.Done()
		}
		for !isComplete {
			select {
			case res := <-p.Messages():
				isComplete = !fn(res.Value, nil)
			case err := <-p.Errors():
				isComplete = !fn(nil, err)
			}
		}
	}
}

type ConsumerHandle struct {
	wg *sync.WaitGroup
}

func (c *ConsumerHandle) join() {
	c.wg.Wait()
}

func (kc *Client) loopConsume(topic string) (*ConsumerHandle, error) {
	partitions, _ := kc.client.consumer.Partitions(topic)
	c := kc.client.consumers[topic]
	wg := kc.consume(partitions, topic, true, false, func(b []byte, e error) bool {
		v := c.decoder(b)
		c.consumer(v)
		return true
	})
	ch := ConsumerHandle{wg: wg}
	return &ch, nil
}

func (kc *Client) SetConsumer(topic string, consumer func(interface{})) {
	b := Inbound{consumer: consumer, topic: topic, decoder: defaultDecoder}
	kc.client.consumers[topic] = b
}
