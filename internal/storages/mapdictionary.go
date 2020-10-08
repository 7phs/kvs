package storages

import (
	"context"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	preAllocatedCap    = 1024
	clearedPortionSize = 100
	maxClearedAttempt  = 5
)

type expiredList struct {
	sync.Mutex

	list               []uint64
	clearedIndex       int
	clearedPartitionSz int
}

func newExpiredList(cap, clearedPartitionSz int) expiredList {
	return expiredList{
		list:               make([]uint64, 0, cap),
		clearedPartitionSz: clearedPartitionSz,
	}
}

func (o *expiredList) push(index ...uint64) {
	o.Lock()
	defer o.Unlock()

	o.list = append(o.list, index...)
}

func (o *expiredList) Clean(fn func(keys []uint64) bool) {
	o.Lock()
	defer o.Unlock()

	for i := 0; i < maxClearedAttempt; i++ {
		start := o.clearedIndex
		limit := start + o.clearedPartitionSz

		if limit > len(o.list) {
			limit = len(o.list)
		}

		o.clearedIndex += o.clearedPartitionSz

		if !fn(o.list[start:limit]) {
			return
		}

		if o.clearedIndex >= len(o.list) {
			o.clearedIndex = 0
			o.list = o.list[:0]

			break
		}
	}
}

type MapDictionary struct {
	sync.RWMutex

	pool    DataPool
	data    map[uint64]record
	expired expiredList
}

func NewMapDictionary(pool DataPool) DataDictionary {
	return &MapDictionary{
		pool:    pool,
		data:    make(map[uint64]record, preAllocatedCap),
		expired: newExpiredList(preAllocatedCap, clearedPortionSize),
	}
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

	if !rec.isExpired() {
		return rec.get(), nil
	}

	go o.expired.push(key)

	return Buffer{}, ErrKeyExpired
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

	o.expired.Clean(func(keys []uint64) bool {
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
