package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/jukylin/esim/log"
	"strings"
	"time"
)

type Client struct {
	client *KClient

	transports []func() interface{}

	logger log.Logger
}

type Option func(c *Client)

type ClientOptions struct{}

func NewClient(options ...Option) *Client {
	Client := &Client{
		transports: make([]func() interface{}, 0),
	}

	client := new(KClient)
	Client.client = client

	for _, option := range options {
		option(Client)
	}

	if Client.logger == nil {
		Client.logger = log.NewLogger()
	}

	return Client
}

func (ClientOptions) WithBokerList(b string) Option {
	return func(hc *Client) {
		brokerList := strings.Split(b, ",")
		asyncProducer, err := sarama.NewAsyncProducer(brokerList, hc.client.config)
		if err != nil {
			log.Log.Errorf("kafka connect error:%v", err.Error())
		}
		hc.client.asyncProducer = asyncProducer

		syncProducer, err := sarama.NewSyncProducer(brokerList, hc.client.config)
		if err != nil {
			log.Log.Errorf("kafka connect error:%v", err.Error())
		}
		hc.client.syncProcuder = syncProducer

		consumer, err := sarama.NewConsumer(brokerList, hc.client.config)
		if err != nil {
			log.Log.Errorf("kafka connect error:%v", err.Error())
		}
		hc.client.consumer = consumer
	}
}

func (ClientOptions) WithReturnSuccess() Option {
	return func(hc *Client) {
		hc.client.config.Producer.Return.Successes = true
	}
}

func (ClientOptions) WithLogger(logger log.Logger) Option {
	return func(hc *Client) {
		hc.logger = logger
	}
}

func (ClientOptions) WithConfig(c *sarama.Config) Option {
	return func(hc *Client) {
		if c == nil {
			c = new(sarama.Config)
		}
		c.Admin.Timeout = 20
		c.Producer.MaxMessageBytes = 1024
		c.Producer.Timeout = defaultTimeOut
		c.Producer.Partitioner = sarama.NewHashPartitioner
		c.Consumer.Fetch.Min = 10
		c.Consumer.Fetch.Default = defaultMinFetch
		c.Consumer.MaxWaitTime = 20 * time.Minute
		c.Consumer.MaxProcessingTime = 20
		c.Consumer.Offsets.CommitInterval = 10
		c.Consumer.Offsets.Initial = sarama.OffsetOldest
		c.Consumer.Group.Session.Timeout = 20 * time.Minute
		c.Consumer.Group.Heartbeat.Interval = 1 * time.Minute
		c.Consumer.Group.Rebalance.Strategy = &balanceStrategy{}
		c.Consumer.Group.Rebalance.Timeout = 1 * time.Minute
		hc.client.config = c

	}
}

func (ClientOptions) WithErrors() Option {
	return func(hc *Client) {
		hc.client.config.Consumer.Return.Errors = true
	}
}

func (ClientOptions) WithMaxOpenRequests(m int) Option {
	return func(hc *Client) {
		if m <= 0 {
			m = 1
		}
		hc.client.config.Net.MaxOpenRequests = m
	}
}

func (ClientOptions) WithDialTimeOut(t time.Duration) Option {
	return func(hc *Client) {
		hc.client.config.Net.DialTimeout = t
	}
}

func (ClientOptions) WithReadTimeOut(t time.Duration) Option {
	return func(hc *Client) {
		hc.client.config.Net.ReadTimeout = t
	}
}

func (ClientOptions) WithWriteTimeOut(t time.Duration) Option {
	return func(hc *Client) {
		hc.client.config.Net.WriteTimeout = t
	}
}

func (ClientOptions) WithConsumer(f func(b []byte)) Option {
	return func(hc *Client) {
		hc.client.config.Net.WriteTimeout = 110
	}
}
