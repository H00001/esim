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

func defaultDecoder(b []byte) interface{} {
	return string(b)
}

func defaultEncoder(i interface{}) []byte {
	return []byte(i.(string))
}
