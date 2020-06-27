package buffer

type Develop interface {
	Init(uint64) error
	Destroy() error
}

type ByteBufferDevelop interface {
	Init(uint64, Allocator)
	Destroy() error
}
type IOer interface {
	Writer
	Reader
}

type Writer interface {
	Write([]byte) error
}

type Reader interface {
	Read(uint64) ([]byte, error)
	ReadAll() ([]byte, error)
}

type key int
type abandonStrategy func(*baseCaching)
type saveStrategy func(*baseCaching, *cacheNode)
type revisitStrategy func(*baseCaching, *cacheNode)

type CachOptional struct {
	i   interface{}
	def interface{}
	k   key
}

func (c *CachOptional) ifPresent(fn func(interface{})) {
	if c.i != nil {
		fn(c.i)
	}
}

func (c *CachOptional) Get() interface{} {
	if c.i != nil {
		return c.i
	}
	return c.def
}

func (c *CachOptional) ofNullable(i interface{}) {
	c.def = i
}
