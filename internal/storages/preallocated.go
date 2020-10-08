package storages

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	inc = 1
)

type preAllocatedBuffer struct {
	expiration time.Time
	index      int
	allocated  int64
	buf        []byte
}

func newPreAllocatedBuffer(buf []byte) *preAllocatedBuffer {
	return &preAllocatedBuffer{
		buf: buf,
	}
}

func (o *preAllocatedBuffer) allocate(sz int, expiration time.Time) (Buffer, bool) {
	if o.index+sz >= len(o.buf) {
		return Buffer{}, false
	}

	index := o.index
	o.index += sz

	if o.expiration.Before(expiration) {
		o.expiration = expiration
	}

	return newBuffer(o, o.buf[index:o.index]), true
}

func (o *preAllocatedBuffer) BufferInUse() {
	atomic.AddInt64(&o.allocated, 1)
}

func (o *preAllocatedBuffer) BufferFree() {
	atomic.AddInt64(&o.allocated, -1)
}

func (o *preAllocatedBuffer) isExpired(now time.Time) bool {
	return atomic.LoadInt64(&o.allocated) == 0 && !o.expiration.After(now)
}

type allocationInUse struct {
	allocation *preAllocatedBuffer

	next *allocationInUse
}

func newAllocationInUse(node *preAllocatedBuffer) *allocationInUse {
	return &allocationInUse{
		allocation: node,
	}
}

func (o *allocationInUse) IsExpired(now time.Time) bool {
	return o.allocation.isExpired(now)
}

func (o *allocationInUse) Reset() (*preAllocatedBuffer, *allocationInUse) {
	allocation := o.allocation
	next := o.next

	o.allocation = nil
	o.next = nil

	return allocation, next
}

type queueAllocations struct {
	sync.Mutex

	root *allocationInUse
	last *allocationInUse
}

func (o *queueAllocations) push(node *preAllocatedBuffer) {
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

func (o *queueAllocations) pop(now time.Time) (*preAllocatedBuffer, bool) {
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

func (o *queueAllocations) len() int {
	count := 0

	for cursor := o.root; cursor != nil; cursor, count = cursor.next, count+inc {
	}

	return count
}
