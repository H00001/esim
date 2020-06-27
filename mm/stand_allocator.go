// A fast memory allocator support high concurrency
package buffer

import (
	"errors"
	"github.com/jukylin/esim/util"
	"sync"
	"time"
)

const MinBlock = 2

func NewSandBufferAllocator() Allocator {
	alloc := blockAllocator{}
	_ = alloc.Init(20)
	return &alloc
}

func (s *standBlockByteBuffer) Release() {
	s.a.release(s)
}

type divide struct {
	first ByteBuffer
	l     sync.Mutex
	s     uint64
	c     *sync.Cond
	index uint8
	give  uint64
	oper  uint64
	a     Allocator
	// load is a dynamic value
	load uint8
}

func (d *divide) Init(a Allocator, load uint8, index uint8) {
	d.c = sync.NewCond(&d.l)
	d.load = load
	d.index = index
	d.s = MinBlock << index
	d.a = a
}

func (d *divide) Destroy() {
	d.first = nil
	// help gc
}

func (d *divide) alloc() ByteBuffer {
	d.Lock()
	d.oper++
	d.give += d.s
	for d.first == nil {
		d.c.Wait()
	}
	v := d.first.(*standBlockByteBuffer)
	d.first = v.addLast(nil)
	// pre allocator
	if d.first == nil {
		go d.backAlloc()
	}
	d.Unlock()
	return v
}

func (d *divide) release(buffer *standBlockByteBuffer) {
	d.l.Lock()
	defer d.l.Unlock()
	d.oper++
	d.give -= d.s
	buffer.next = d.first
	d.first = buffer
}

func (d *divide) Lock() {
	d.l.Lock()
}

func (d *divide) Unlock() {
	d.l.Unlock()
}

func (d *divide) setLink(sta *standBlockByteBuffer) {
	d.first = sta
}

func (d *divide) backAlloc() {
	d.Lock()
	result := make([]standBlockByteBuffer, d.load)
	for i := 0; i < len(result); i++ {
		result[i].init(d.s, d.a, d.index)
		if i == len(result)-1 {
			result[i].next = nil
		} else {
			result[i].next = &result[i+1]
		}
	}
	d.setLink(&result[0])
	d.Unlock()
	d.c.Broadcast()
}

type standBlockByteBuffer struct {
	BaseByteBuffer
	next    ByteBuffer
	index   uint8
	sumSize uint64
	active  []bool
}

func (s *standBlockByteBuffer) init(size uint64, all Allocator, index uint8) {
	s.BaseByteBuffer.Init(size, all)
	s.sumSize = size
	s.active = make([]bool, 2)
	s.index = index
}

func (s *standBlockByteBuffer) Size() uint64 {
	return s.sumSize
}

func (s *standBlockByteBuffer) addLast(buffer ByteBuffer) ByteBuffer {
	p := s.next
	s.next = buffer
	return p
}

type blockAllocator struct {
	divs  []divide
	min   uint64
	psize uint8
	load  uint8
	r     bool
	max   uint64
	regs  int64
	s     time.Duration
	w     sync.WaitGroup
}

func (s *blockAllocator) Init(i uint64) error {
	// create alloc index
	s.r = true
	s.psize = uint8(i)
	s.min = MinBlock
	s.max = MinBlock<<s.psize - 1
	s.divs = make([]divide, s.psize)
	s.s = 10 * time.Second
	s.w.Add(int(s.psize))
	for i := range s.divs {
		go func(vl *divide, index int) {
			vl.Init(s, s.psize, uint8(index))
			// at init process,it don't need create threads.
			vl.backAlloc()
			s.w.Done()
		}(&s.divs[i], i)
	}
	go s.dynamicRegulate()
	s.w.Wait()
	return nil
}

func (s *blockAllocator) OperatorTimes() uint64 {
	var val uint64 = 0
	for _, v := range s.divs {
		val += v.oper
	}
	return val
}

func (s *blockAllocator) Destroy() error {
	s.r = false
	for i := range s.divs {
		s.divs[i].Destroy()
	}
	return nil
}

func (s *blockAllocator) Alloc(length uint64) ByteBuffer {
	if length > s.max || length == 0 {
		return nil
	}
	return s.doAlloc(length)
}

func (s *blockAllocator) PoolSize() uint64 {
	return util.Int2Uint64(int(s.psize))
}

func (s *blockAllocator) release(b ByteBuffer) {
	release := b
	next := release
	for next != nil {
		release = release.(*standBlockByteBuffer).next
		go func(n *standBlockByteBuffer) { s.divs[n.index].release(n) }(next.(*standBlockByteBuffer))
		next = release
	}
}

func (s *blockAllocator) doAlloc(length uint64) ByteBuffer {
	r := position(length)
	//divide first
	var b = s.divs[r[len(r)-1]].alloc().(*standBlockByteBuffer)
	l := b
	for i := len(r) - 2; i >= 0; i-- {
		l.next = s.divs[r[i]].alloc()
		l = l.next.(*standBlockByteBuffer)
		b.sumSize += l.capital
	}
	return b
}

func position(length uint64) []uint8 {
	return util.IsPow2(length)
}

func (s *blockAllocator) AllocSize() uint64 {
	var val uint64 = 0
	for _, v := range s.divs {
		val += v.give
	}
	return val
}

func (s *blockAllocator) dynamicRegulate() {
	c := util.NewCounter()
	c.Boot()
	for s.r {
		for i := 0; i < len(s.divs); i++ {
			c.Push(s.divs[i].oper)
		}
		time.Sleep(s.s)
		ave := c.Ave()
		for i := 0; i < len(s.divs); i++ {
			if s.divs[i].oper <= ave {
				s.divs[i].load--
			} else {
				s.divs[i].load++
			}
		}
		c.Reset()
	}
}

func (s *standBlockByteBuffer) Read(len uint64) ([]byte, error) {
	if s.sumSize-s.RP < len {
		return nil, errors.New(util.INDEX_OUTOF_BOUND)
	}
	send := make([]byte, len)
	s.Read0(len, 0, send)
	return send, nil
}

func (s *standBlockByteBuffer) ReadAll() ([]byte, error) {
	return nil, nil
}

func (s *standBlockByteBuffer) Read0(len, pos uint64, send []byte) {
	i := pos
	for ; i < len; i++ {
		if s.RP == s.capital {
			break
		}
		send[i] = util.ReadOne(s.s, &s.RP)
	}
	if i < len {
		s.next.(*standBlockByteBuffer).Read0(len, i, send)
	}
}

func (s *standBlockByteBuffer) Write(_b []byte) error {
	if s.sumSize-s.WP < uint64(len(_b)) {
		return errors.New(util.INDEX_OUTOF_BOUND)
	}
	s.Write0(_b, 0)
	return nil
}

func (s *standBlockByteBuffer) Write0(_b []byte, position uint64) {
	i := position
	for ; i < uint64(len(_b)); i++ {
		if s.WP == s.capital {
			break
		}
		util.WriteOne(s.s, _b[i], &s.WP)
	}
	if i != uint64(len(_b)) {
		s.next.(*standBlockByteBuffer).Write0(_b, i)
	}
}

func (s *standBlockByteBuffer) AvailableReadSum() uint64 {
	return s.globalWP(0) - s.globalRP(0) - 1
}

func (s *standBlockByteBuffer) globalWP(now uint64) uint64 {
	if s.WP != s.capital || s.next == nil {
		return now + s.WP
	}
	return s.next.(*standBlockByteBuffer).globalWP(now + s.capital)

}

func (s *standBlockByteBuffer) globalRP(now uint64) uint64 {
	if s.WP != s.capital || s.next == nil {
		return now + s.RP
	}
	return s.next.(*standBlockByteBuffer).globalRP(now + s.capital)
}

func (s *standBlockByteBuffer) FastMoveOut() *[]byte {
	panic("method `FastMoveOut` of struct standBlockByteBuffer do not support!")
	return nil
}

func (s *standBlockByteBuffer) FastMoveIn(_ *[]byte) {
	panic("method `FastMoveIn` of struct standBlockByteBuffer do not support!")
}
