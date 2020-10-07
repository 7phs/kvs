package storages

import (
	"context"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	preAllocatedCap   = 1024
	clearedPortion    = 100
	maxClearedAttempt = 5
)

type expiredList struct {
	sync.Mutex

	list         []uint64
	clearedIndex int
}

func newExpiredList(cap int) expiredList {
	return expiredList{
		list: make([]uint64, 0, cap),
	}
}

func (o *expiredList) pushToExpire(index ...uint64) {
	o.Lock()
	defer o.Unlock()

	o.list = append(o.list, index...)
}

func (o *expiredList) Range(fn func(keys []uint64) bool) {
	o.Lock()
	defer o.Unlock()

	for i := 0; i < maxClearedAttempt; i++ {
		limit := o.clearedIndex + clearedPortion
		if limit > len(o.list) {
			limit = len(o.list)
		}

		if !fn(o.list[o.clearedIndex:limit]) {
			return
		}

		o.clearedIndex += clearedPortion

		if o.clearedIndex > len(o.list) {
			o.clearedIndex = 0
			o.list = o.list[:0]

			break
		}
	}
}

type MapDictionary struct {
	sync.RWMutex

	pool    *dataPool
	data    map[uint64]record
	expired expiredList
}

func NewMapDictionary() (DataDictionary, error) {
	pool, err := newDataPool()
	if err != nil {
		return nil, err
	}

	return &MapDictionary{
		pool:    pool,
		data:    make(map[uint64]record, preAllocatedCap),
		expired: newExpiredList(preAllocatedCap),
	}, nil
}

func (o *MapDictionary) Add(key uint64, data []byte, expiration time.Time) error {
	buf, err := o.pool.Copy(data, expiration)
	if err != nil {
		return err
	}

	o.Lock()
	defer o.Unlock()

	o.data[key] = newRecord(buf, expiration)

	return nil
}

func (o *MapDictionary) Get(key uint64) (Buffer, error) {
	o.RLock()
	rec, ok := o.data[key]
	o.RUnlock()

	if !ok {
		return Buffer{}, ErrKeyNotFound
	}

	if rec.expiration.After(time.Now()) {
		return rec.get(), nil
	}

	go o.expired.pushToExpire(key)

	return Buffer{}, ErrKeyNotFound
}

func (o *MapDictionary) Clean(ctx context.Context) error {
	var (
		wg errgroup.Group
	)

	wg.Go(func() error {
		return o.pool.Clean(ctx)
	})

	wg.Go(func() error {
		return o.cleanDictionary(ctx)
	})

	return wg.Wait()
}

func (o *MapDictionary) cleanDictionary(ctx context.Context) error {
	now := time.Now()

	o.expired.Range(func(keys []uint64) bool {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		prevK := uint64(0)

		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		o.Lock()
		defer o.Unlock()

		for _, k := range keys {
			if k != prevK {
				if r, ok := o.data[k]; ok && r.expiration.Before(now) {
					delete(o.data, k)
				}
			}

			prevK = k
		}

		return true
	})

	return nil
}
