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

type allocationInUse struct {
	allocation *allocation

	next *allocationInUse
}

func newAllocationInUse(node *allocation) *allocationInUse {
	return &allocationInUse{
		allocation: node,
	}
}

func (o *allocationInUse) IsExpired(now time.Time) bool {
	return o.allocation.isExpired(now)
}

func (o *allocationInUse) Reset() (*allocation, *allocationInUse) {
	allocation := o.allocation
	next := o.next

	o.allocation = nil
	o.next = nil

	return allocation, next
}

type queueToClean struct {
	sync.Mutex

	root *allocationInUse
	last *allocationInUse
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

	var (
		prev   *allocationInUse
		cursor = o.root
	)

	for cursor != nil {
		if cursor.IsExpired(now) {
			allocation, next := cursor.Reset()

			if prev != nil {
				prev.next = next
			} else {
				o.root = next
			}

			if o.root == nil {
				o.last = nil
			}

			return allocation, true
		}

		prev = cursor
		cursor = cursor.next
	}

	prev = nil

	return nil, false
}
