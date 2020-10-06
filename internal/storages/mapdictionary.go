package storages

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type MapDictionary struct {
	sync.Map

	pool *dataPool
}

func NewMapDictionary() (*MapDictionary, error) {
	pool, err := newDataPool()
	if err != nil {
		return nil, err
	}

	return &MapDictionary{
		pool: pool,
	}, nil
}

func (o *MapDictionary) Add(key uint64, data []byte, expiration time.Time) error {
	buf, err := o.pool.Copy(data, expiration)
	if err != nil {
		return err
	}

	o.Store(key, newRecord(buf, expiration))

	return nil
}

func (o *MapDictionary) Get(key uint64) (Buffer, error) {
	v, ok := o.Load(key)
	if !ok {
		return Buffer{}, ErrKeyNotFound
	}

	rec, ok := v.(record)
	if !ok {
		return Buffer{}, ErrKeyNotFound
	}

	if rec.isExpired() {
		return Buffer{}, ErrKeyNotFound
	}

	return rec.get(), nil
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
	v := [100]uint64{}
	index := -1
	now := time.Now()

	o.Range(func(key, value interface{}) bool {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		rec, ok := value.(record)
		if !ok {
			return true
		}

		if rec.expiration.After(now) {
			return true
		}

		index++
		v[index], ok = key.(uint64)
		if !ok {
			return true
		}

		if index == len(v)-1 {
			return false
		}

		return true
	})

	for _, key := range v {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		v, ok := o.LoadAndDelete(key)
		if !ok {
			continue
		}

		rec, ok := v.(record)
		if !ok {
			continue
		}

		rec.reset()
	}

	return nil
}
