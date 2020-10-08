package storages

import (
	"context"
	"sync"
	"time"
)

type DataPool interface {
	Copy(data []byte, expiration time.Time) (Buffer, error)
	Clean(ctx context.Context) error
}

type dataPool struct {
	sync.Mutex

	valuePool MemoryPool

	current      *preAllocatedBuffer
	queueToClean queueAllocations
}

func NewDataPool(memPool MemoryPool) (DataPool, error) {
	buf, err := memPool.Get()
	if err != nil {
		return nil, err
	}

	return &dataPool{
		valuePool: memPool,
		current:   newPreAllocatedBuffer(buf),
	}, nil
}

func (o *dataPool) Copy(data []byte, expiration time.Time) (Buffer, error) {
	valueBuf, err := o.allocate(len(data), expiration)
	if err != nil {
		return Buffer{}, err
	}

	valueBuf.Copy(data)

	return valueBuf, nil
}

func (o *dataPool) allocate(sz int, expiration time.Time) (Buffer, error) {
	o.Lock()
	defer o.Unlock()

	buf, ok := o.current.allocate(sz, expiration)
	if ok {
		return buf, nil
	}

	bufP, err := o.valuePool.Get()
	if err != nil {
		return Buffer{}, err
	}

	nodeToClean := o.current
	o.current = newPreAllocatedBuffer(bufP)

	go o.queueToClean.push(nodeToClean)

	buf, ok = o.current.allocate(sz, expiration)
	if ok {
		return buf, nil
	}

	return Buffer{}, ErrOutOfLimit
}

func (o *dataPool) Clean(ctx context.Context) error {
	now := time.Now()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		node, ok := o.queueToClean.pop(now)
		if !ok {
			return nil
		}

		if node.buf != nil {
			o.valuePool.Put(node.buf)
			node.buf = nil
			node = nil
		}
	}
}
