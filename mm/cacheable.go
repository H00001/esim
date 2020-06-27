package buffer

type Cached interface {
	Save(key, interface{})
	Get(index key) interface{}
}

type cacheNode struct {
	k     key
	data  interface{}
	prev  *cacheNode
	next  *cacheNode
	times int
}

type baseCaching struct {
	m    map[key]*cacheNode
	head *cacheNode
	tail *cacheNode
}

type cacheing struct {
	base baseCaching
	max  int
	now  int
	as   abandonStrategy
	ss   saveStrategy
	rs   revisitStrategy
}

func (c *cacheing) Save(index key, i interface{}) {
	n := cacheNode{data: i, k: index}
	c.ss(&c.base, &n)

	if c.now > c.max {
		c.as(&c.base)
	}
	c.now++
}

func (c *baseCaching) rm(n *cacheNode) {
	if n.prev == nil {
		c.head = n.next
	} else {
		n.prev.next = n.next
	}
	if n.next == nil {
		c.tail = n.prev
	} else {
		n.next.prev = n.prev
	}
}

func (c *baseCaching) inputHead(n *cacheNode) {
	if c.head == nil && c.tail == nil {
		c.tail = n
		c.head = n
		return
	}
	c.head.prev = n
	n.next = c.head
	c.head = n
}

func (c *cacheing) Init(i uint64) error {
	c.base.m = make(map[key]*cacheNode, i)
	c.max = int(i)
	c.as = func(ca *baseCaching) {
		ca.rm(ca.tail)
	}
	c.rs = func(bc *baseCaching, node *cacheNode) {
		bc.rm(node)
		bc.inputHead(node)
		node.times++
	}
	c.ss = func(b *baseCaching, n *cacheNode) {
		b.inputHead(n)
		b.m[n.k] = n
	}
	return nil
}

func (c *cacheing) Destroy() error {
	return nil
}

func (c *cacheing) Get(index key) *CachOptional {
	w, ok := c.base.m[index]
	if !ok {
		return new(CachOptional)
	}
	c.rs(&c.base, w)

	return &CachOptional{i: w.data}
}
