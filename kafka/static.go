package kafka

import "sync"

const (
	defaultTimeOut  = 20
	defaultFetch    = 40
	defaultMinFetch = 10
)

type control int

const (
	next   control = 1
	cancel         = 1 << 1
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

type ConsumerHandle interface {
	join()
	cancel()
	get() control
}

type consumerHandle struct {
	wg *sync.WaitGroup
	ch chan control
}

func NewConsumerHandle(wg *sync.WaitGroup) ConsumerHandle {
	c := consumerHandle{}
	c.wg = wg
	c.ch = make(chan control, 2)
	return &c
}

func (c *consumerHandle) join() {
	c.wg.Wait()
}

func (c *consumerHandle) cancel() {
	c.ch <- cancel
}

func (c *consumerHandle) get() control {
	select {
	case c := <-c.ch:
		return c
	default:
		return next

	}
}
