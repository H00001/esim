package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/jukylin/esim/log"
	"strings"
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

func (ClientOptions) WithErrors() Option {
	return func(hc *Client) {
		hc.client.config.Consumer.Return.Errors = true
	}
}