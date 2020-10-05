package storages

import (
	"context"
	"sync"
	"time"
)

const (
	dataChunkSz int = 1024 * 1024
)

type dataPool struct {
	sync.Mutex

	valuePool *memoryPool

	current      allocation
	queueToClean queueToClean
}

func newDataPool() (*dataPool, error) {
	valuePool := newMemoryPool(dataChunkSz)

	buf, err := valuePool.Get()
	if err != nil {
		return nil, err
	}

	return &dataPool{
		valuePool: valuePool,
		current:   newAllocation(buf),
	}, nil
}

func (o *dataPool) Copy(data []byte, expiration time.Time) ([]byte, error) {
	valueBuf, err := o.allocate(len(data), expiration)
	if err != nil {
		return nil, err
	}

	copy(valueBuf, data)

	return valueBuf, nil
}

func (o *dataPool) allocate(sz int, expiration time.Time) ([]byte, error) {
	o.Lock()
	defer o.Unlock()

	buf, ok := o.current.allocate(sz, expiration)
	if ok {
		return buf, nil
	}

	buf, err := o.valuePool.Get()
	if err != nil {
		return nil, err
	}

	nodeToClean := o.current
	o.current = newAllocation(buf)

	go o.queueToClean.push(nodeToClean)

	buf, ok = o.current.allocate(sz, expiration)
	if ok {
		return buf, nil
	}

	return nil, ErrOutOfLimit
}

func (o *dataPool) Clean(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		node, ok := o.queueToClean.pop()
		if !ok {
			return nil
		}

		if len(node.buf) > 0 {
			o.valuePool.Put(node.buf)
			node.buf = nil
		}
	}
}
