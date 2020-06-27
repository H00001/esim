package buffer

import (
	"github.com/jukylin/esim/util"
	"sync"
)

func NewLikedBufferAllocator() Allocator {
	a := linkedAllocator{}
	if err := a.Init(0); err == nil {
		return &a
	}
	return nil
}

type linkedAllocator struct {
	counter util.DynamicCounter
	unUsed  *util.Skiplist
	l       sync.Mutex
}

func (a *linkedAllocator) Init(uint64) error {
	a.unUsed = util.NewSkipList()
	a.counter = util.NewCounter()
	a.counter.Boot()
	return nil
}

func (a *linkedAllocator) PoolSize() uint64 {
	return a.counter.Size()
}

func (a *linkedAllocator) Destroy() error {
	a.unUsed = nil
	a.counter = nil
	//help gc
	return nil
}

func (a *linkedAllocator) Alloc(length uint64) ByteBuffer {
	if length == 0 {
		return nil
	}
	a.counter.Push(length)
	a.l.Lock()
	defer a.l.Unlock()
	if bf := a.findByUnusedList(length); bf != nil {
		return bf
	}
	bf := new(linkedByteBuffer)
	bf.Init(length, a)
	return bf
}

func (a *linkedAllocator) findByUnusedList(i uint64) ByteBuffer {
	k, v := a.unUsed.Search(i)
	if k != 0 {
		a.unUsed.Delete(k)
		return v.(ByteBuffer)
	}
	return nil
}

func (a *linkedAllocator) release(buffer ByteBuffer) {
	a.l.Lock()
	a.counter.Push(-buffer.Size())
	a.unUsed.Insert(buffer.Size(), buffer)
	a.l.Unlock()
	go a.dynamicShrink()
}

func (a *linkedAllocator) dynamicShrink() {

}

type linkedByteBuffer struct {
	BaseByteBuffer
}

func (a *linkedAllocator) OperatorTimes() uint64 {
	return a.counter.Size()
}

func (a *linkedAllocator) AllocSize() uint64 {
	return a.counter.Use()
}
