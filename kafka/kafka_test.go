package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/jukylin/esim/log"
	"os"
	"testing"
)

const (
	host1 = "127.0.0.1"
)

var logger log.Logger

func TestMain(m *testing.M) {
	loggerOptions := log.LoggerOptions{}
	logger = log.NewLogger(loggerOptions.WithDebug(true))

	code := m.Run()

	os.Exit(code)
}

func TestMulLevelRoundTrip(t *testing.T) {
	clientOptions := ClientOptions{}
	httpClient := NewClient(
		clientOptions.WithConfig(new(sarama.Config)),
		clientOptions.WithDialTimeOut(10),
		clientOptions.WithReadTimeOut(10),
		clientOptions.WithWriteTimeOut(10),
		clientOptions.WithMaxOpenRequests(10),
		clientOptions.WithLogger(logger),
		clientOptions.WithBokerList("123,45"),
	)
	httpClient.SyncSend("123", "456")

}
