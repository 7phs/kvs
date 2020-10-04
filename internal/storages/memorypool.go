package storages

import (
	"sync"
)

type memoryPool struct {
	sync.Pool
}

func newMemoryPool(sz int) memoryPool {
	return memoryPool{
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

		buf = o.Pool.Get().([]byte)
	}()

	return buf, err
}

func (o *memoryPool) Put(d []byte) {
	o.Pool.Put(d)
}
