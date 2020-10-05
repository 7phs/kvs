package storages

import (
	"sync"
	"time"
)

type allocation struct {
	expiration time.Time
	index      int
	buf        []byte
}

func newAllocation(buf []byte) allocation {
	return allocation{
		buf: buf,
	}
}

func (o *allocation) allocate(sz int, expiration time.Time) ([]byte, bool) {
	if o.index+sz >= len(o.buf) {
		return nil, false
	}

	index := o.index
	o.index += sz

	if o.expiration.Before(expiration) {
		o.expiration = expiration
	}

	return o.buf[index:o.index], true
}

type allocateInUse struct {
	allocation

	next *allocateInUse
}

func newAllocationInUse(node allocation) *allocateInUse {
	return &allocateInUse{
		allocation: node,
	}
}

type queueToClean struct {
	sync.Mutex

	root *allocateInUse
	last *allocateInUse
}

func (o *queueToClean) push(node allocation) {
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

func (o *queueToClean) pop(now time.Time) (allocation, bool) {
	o.Lock()
	defer o.Unlock()

	if o.root == nil {
		return allocation{}, false
	}

	if o.root.expiration.After(now) {
		return allocation{}, false
	}

	n := o.root
	o.root = o.root.next

	if o.root == nil {
		o.last = nil
	}

	return n.allocation, true
}
