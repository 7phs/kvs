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

func (o *MapDictionary) Add(key uint64, value []byte, expiration time.Time) error {
	value, err := o.pool.Copy(value, expiration)
	if err != nil {
		return err
	}

	o.Store(key, record{
		value:      value,
		expiration: expiration,
	})

	return nil
}

func (o *MapDictionary) Get(key uint64) ([]byte, error) {
	v, ok := o.Load(key)
	if !ok {
		return nil, ErrKeyNotFound
	}

	rec, ok := v.(record)
	if !ok {
		return nil, ErrKeyNotFound
	}

	if rec.expiration.Before(time.Now()) {
		return nil, ErrKeyNotFound
	}

	return rec.value, nil
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

		rec.value = nil
	}

	return nil
}
