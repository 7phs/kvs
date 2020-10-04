package storages

import (
	"io"
	"time"

	"github.com/7phs/kvs/internal/config"
	"github.com/minio/highwayhash"
)

var (
	_ Storages = (*InMemStorages)(nil)
)

type Storages interface {
	Add(key string, data io.Reader) error
	Get(key string) ([]byte, error)
}

type DataReader interface {
	ReadValue(data io.Reader, expiration time.Time) ([]byte, error)
}

type DataDictionary interface {
	Add(key uint64, data []byte, expiration time.Time) error
	Get(key uint64) ([]byte, error)
}

type InMemStorages struct {
	dataReader DataReader
	dataDict   DataDictionary

	nonce   [32]byte
	expired time.Duration
}

func NewInMemStorages(
	config config.Config,
	dataReader DataReader,
	dataDict DataDictionary,
) (Storages, error) {
	return &InMemStorages{
		dataReader: dataReader,
		dataDict:   dataDict,
		expired:    config.Expiration(),
	}, nil
}

func (o *InMemStorages) Add(key string, data io.Reader) error {
	expiration := time.Now().Add(o.expired)

	value, err := o.dataReader.ReadValue(data, expiration)
	if err != nil {
		return err
	}

	return o.dataDict.Add(o.hash(key), value, expiration)
}

func (o *InMemStorages) Get(key string) ([]byte, error) {
	return o.dataDict.Get(o.hash(key))
}

func (o *InMemStorages) hash(key string) uint64 {
	return highwayhash.Sum64([]byte(key), o.nonce[:])
}
