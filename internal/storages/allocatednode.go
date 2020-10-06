package storages

import (
	"sync"
	"sync/atomic"
	"time"
)

type Buffer struct {
	allocation *allocation
	data       []byte
}

func (o *Buffer) inUse() {
	o.allocation.bufferInUse()
}

func (o *Buffer) Free() {
	o.allocation.bufferFree()

	o.Reset()
}

func (o *Buffer) Copy(data []byte) {
	copy(o.data, data)
}

func (o *Buffer) Bytes() []byte {
	return o.data
}

func (o *Buffer) Reset() {
	o.allocation = nil
	o.data = nil
}

type allocation struct {
	expiration time.Time
	index      int
	allocated  int64
	buf        []byte
}

func newAllocation(buf []byte) *allocation {
	return &allocation{
		buf: buf,
	}
}

func (o *allocation) allocate(sz int, expiration time.Time) (Buffer, bool) {
	if o.index+sz >= len(o.buf) {
		return Buffer{}, false
	}

	index := o.index
	o.index += sz

	if o.expiration.Before(expiration) {
		o.expiration = expiration
	}

	return Buffer{
		allocation: o,
		data:       o.buf[index:o.index],
	}, true
}

func (o *allocation) bufferInUse() {
	atomic.AddInt64(&o.allocated, 1)
}

func (o *allocation) bufferFree() {
	atomic.AddInt64(&o.allocated, -1)
}

func (o *allocation) isExpired(now time.Time) bool {
	return atomic.LoadInt64(&o.allocated) == 0 && !o.expiration.After(now)
}

type allocateInUse struct {
	allocation *allocation

	next *allocateInUse
}

func newAllocationInUse(node *allocation) *allocateInUse {
	return &allocateInUse{
		allocation: node,
	}
}

func (o *allocateInUse) IsExpired(now time.Time) bool {
	return o.allocation.isExpired(now)
}

func (o *allocateInUse) Reset() (*allocation, *allocateInUse) {
	allocation := o.allocation
	next := o.next

	o.allocation = nil
	o.next = nil

	return allocation, next
}

type queueToClean struct {
	sync.Mutex

	root *allocateInUse
	last *allocateInUse
}

func (o *queueToClean) push(node *allocation) {
	n := newAllocationInUse(node)

	o.Lock()
	defer o.Unlock()

	if o.root == nil {
		o.root = n
		o.last = o.root

		return
	}

	o.last.next = n
	o.last = o.last.next
}

func (o *queueToClean) pop(now time.Time) (*allocation, bool) {
	o.Lock()
	defer o.Unlock()

	if o.root == nil {
		return nil, false
	}

	if !o.root.IsExpired(now) {
		return nil, false
	}

	allocation, next := o.root.Reset()

	o.root = next
	if o.root == nil {
		o.last = nil
	}

	return allocation, true
}
