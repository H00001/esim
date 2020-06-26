package kafka

import "sync"

const (
	defaultTimeOut  = 20
	defaultFetch    = 40
	defaultMinFetch = 10
)

type control int

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

type consumerHandle struct {
	wg *sync.WaitGroup
	ch chan bool
}

func NewConsumerHandle(wg *sync.WaitGroup) ConsumerHandle {
	c := consumerHandle{}
	c.wg = wg
	c.ch = make(chan bool, 1)
	return &c
}

func (c *consumerHandle) Join() {
	c.wg.Wait()
}

func (c *consumerHandle) Cancel() {
	c.ch <- false
	close(c.ch)
}

func (c *consumerHandle) get() bool {
	select {
	case b, ok := <-c.ch:
		{
			if ok {
				return b
			} else {
				return false
			}
		}
	default:
		return true
	}
}
