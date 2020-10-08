package storages

import (
	"sync"
)

var (
	_ MemoryPool = (*memoryPool)(nil)
)

type MemoryPool interface {
	Get() ([]byte, error)
	Put(d []byte)
}

type memoryPool struct {
	sync.Pool
}

func NewMemoryPool(sz int) MemoryPool {
	return &memoryPool{
		Pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, sz)
			},
		},
	}
}

func (o *memoryPool) Get() ([]byte, error) {
	var (
		buf []byte
		err error
	)

	func() {
		defer func() {
			if r := recover(); r != nil {
				err = ErrOutOfLimit
			}
		}()

		buf, _ = o.Pool.Get().([]byte)
	}()

	return buf, err
}

//nolint:staticcheck
func (o *memoryPool) Put(d []byte) {
	o.Pool.Put(d)
}
