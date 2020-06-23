package kafka

type IkClient interface {
	SyncSend(topic string, msg string) (int32, int64, error)
	RecvMsg(topic string) (string, error)
	AsyncSend(topic string, msg string) (int32, int64, error)
}
