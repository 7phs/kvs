package storages

import (
	"context"
	"time"

	"github.com/7phs/kvs/internal/config"
	"github.com/minio/highwayhash"
)

var (
	_ Storages = (*InMemStorages)(nil)
)

type Storages interface {
	ID() string
	Add(key, body []byte) error
	Get(key []byte) (Buffer, error)
	Clean(ctx context.Context) error
}

type DataDictionary interface {
	Add(key uint64, data []byte, expiration time.Time) error
	Get(key uint64) (Buffer, error)
	Clean(ctx context.Context) error
}

type InMemStorages struct {
	dataDict DataDictionary

	nonce      [32]byte
	expired    time.Duration
	timeSource config.TimeSource
}

func NewInMemStorages(
	config config.Config,
	dataDict DataDictionary,
) (Storages, error) {
	return &InMemStorages{
		dataDict:   dataDict,
		expired:    config.Expiration(),
		timeSource: config.TimeSource(),
	}, nil
}

func (o *InMemStorages) ID() string {
	return "in-memory-storages"
}

func (o *InMemStorages) Add(key, body []byte) error {
	expiration := o.timeSource.Now().Add(o.expired)

	return o.dataDict.Add(o.hash(key), body, expiration)
}

func (o *InMemStorages) Get(key []byte) (Buffer, error) {
	return o.dataDict.Get(o.hash(key))
}

func (o *InMemStorages) Clean(ctx context.Context) error {
	return o.dataDict.Clean(ctx)
}

func (o *InMemStorages) hash(key []byte) uint64 {
	return highwayhash.Sum64(key, o.nonce[:])
}
