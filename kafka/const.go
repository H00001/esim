package kafka

const (
	defaultTimeOut  = 20
	defaultFetch    = 40
	defaultMinFetch = 10
)

type DeCoder func([]byte) interface{}
type EnCoder func(interface{}) []byte

type Inbound struct {
	decoder  DeCoder
	topic    string
	consumer func(interface{})
}

func defaultDecorder(b []byte) interface{} {
	return string(b)
}
