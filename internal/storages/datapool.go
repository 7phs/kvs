package storages

import (
	"context"
	"sync"
	"time"
)

const (
	dataChunkSz int = 1 * 1024 * 1024
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

	bufP, err := o.valuePool.Get()
	if err != nil {
		return nil, err
	}

	nodeToClean := o.current
	o.current = newAllocation(bufP)

	go o.queueToClean.push(nodeToClean)

	buf, ok = o.current.allocate(sz, expiration)
	if ok {
		return buf, nil
	}

	return nil, ErrOutOfLimit
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
		}
	}
}
