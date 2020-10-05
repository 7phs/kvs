package storages

import (
	"context"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	partitionNum      uint64 = 16
	maxUint64                = ^uint64(0)
	divider                  = maxUint64 / partitionNum
	preAllocatedCap          = 1024
	clearedPortion           = 100
	maxClearedAttempt        = 5
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

type record struct {
	value      []byte
	expiration time.Time
}

type partition struct {
	sync.RWMutex

	pool    *dataPool
	data    map[uint64]record
	expired expiredList
}

func newPartition() (partition, error) {
	pool, err := newDataPool()
	if err != nil {
		return partition{}, err
	}

	return partition{
		pool:    pool,
		data:    make(map[uint64]record, preAllocatedCap),
		expired: newExpiredList(preAllocatedCap),
	}, nil
}

func (o *partition) add(key uint64, value []byte, expiration time.Time) error {
	value, err := o.pool.Copy(value, expiration)
	if err != nil {
		return err
	}

	o.Lock()
	defer o.Unlock()

	o.data[key] = record{
		value:      value,
		expiration: expiration,
	}

	return nil
}

func (o *partition) get(key uint64) ([]byte, error) {
	o.RLock()
	rec, ok := o.data[key]
	o.RUnlock()

	if !ok {
		return nil, ErrKeyNotFound
	}

	if rec.expiration.After(time.Now()) {
		return rec.value, nil
	}

	go o.expired.pushToExpire(key)

	return nil, ErrKeyNotFound
}

func (o *partition) clean(ctx context.Context) error {
	var (
		wg  errgroup.Group
		now = time.Now()
	)

	wg.Go(func() error {
		return o.pool.Clean(ctx)
	})

	wg.Go(func() error {
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
	})

	return wg.Wait()
}

type PartitionedDictionary struct {
	partitions [partitionNum]partition
}

func NewPartitionedDictionary() (*PartitionedDictionary, error) {
	var (
		dict = PartitionedDictionary{}
		err  error
	)

	for i := range dict.partitions {
		dict.partitions[i], err = newPartition()
		if err != nil {
			return nil, err
		}
	}

	return &dict, nil
}

func (o *PartitionedDictionary) Add(key uint64, value []byte, expiration time.Time) error {
	return o.partitions[chunkKey(key)].add(key, value, expiration)
}

func (o *PartitionedDictionary) Get(key uint64) ([]byte, error) {
	return o.partitions[chunkKey(key)].get(key)
}

func chunkKey(key uint64) int {
	return int(key / divider)
}

func (o *PartitionedDictionary) Clean(ctx context.Context) error {
	for i := 0; i < len(o.partitions); i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		err := o.partitions[i].clean(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
