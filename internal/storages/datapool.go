package storages

import (
	"context"
	"io"
	"sync"
	"time"
)

const (
	readerChunkSz       int = 8 * 1024
	readerChunkNumLimit int = 32 // * readerChunkSz = 256Kb
	dataChunkSz         int = 1024 * 1024
)

var (
	_ DataReader  = (*DataPool)(nil)
	_ Maintenance = (*DataPool)(nil)
)

type DataPool struct {
	sync.Mutex

	readerPool memoryPool
	valuePool  memoryPool

	current      allocation
	queueToClean queueToClean
}

func NewDataPool() (DataPool, error) {
	valuePool := newMemoryPool(dataChunkSz)
	buf, err := valuePool.Get()
	if err != nil {
		return DataPool{}, nil
	}

	return DataPool{
		readerPool: newMemoryPool(readerChunkSz),
		valuePool:  valuePool,
		current:    newAllocation(buf),
	}, nil
}

func (o *DataPool) ReadValue(data io.Reader, expiration time.Time) ([]byte, error) {
	dataBuf, index, sz, err := o.read(data)
	if err != nil {
		return nil, err
	}

	valueBuf, err := o.allocate(sz, expiration)
	if err != nil {
		return nil, err
	}

	o.copy(valueBuf, dataBuf, index)

	return valueBuf, nil
}

func (o *DataPool) read(data io.Reader) ([readerChunkNumLimit][]byte, int, int, error) {
	var (
		dataBuf [readerChunkNumLimit][]byte
		sz      int
		index   int
	)
	defer func() {
		for i := 0; i < index; i++ {
			o.readerPool.Put(dataBuf[i])
		}
	}()

	for index < readerChunkNumLimit {
		buf, err := o.readerPool.Get()
		if err != nil {
			return dataBuf, 0, 0, err
		}
		read, err := data.Read(buf)
		if err != nil {
			return dataBuf, 0, 0, err
		}

		sz += read
		dataBuf[index] = buf

		if read < len(buf) {
			break
		}

		index++
	}

	if index >= readerChunkNumLimit {
		return dataBuf, 0, 0, ErrOutOfLimit
	}

	return dataBuf, index, sz, nil
}

func (o *DataPool) allocate(sz int, expiration time.Time) ([]byte, error) {
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

func (o *DataPool) copy(buf []byte, read [readerChunkNumLimit][]byte, index int) {
	valueIndex := 0
	for i := 0; i < index; i++ {
		copy(buf[valueIndex:], read[i])
		valueIndex += len(read[i])
	}
}

func (o *DataPool) Clean(ctx context.Context) error {
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

	return nil
}
