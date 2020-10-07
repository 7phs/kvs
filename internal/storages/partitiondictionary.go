package storages

import (
	"context"
	"time"
)

const (
	DefaultPartitionNum  uint64 = 16
	DefaultPartitionMask uint64 = 0xF
)

type PartitionedDictionary struct {
	partitions    []DataDictionary
	partitionMask uint64
}

func NewPartitionedDictionary(
	partitionNum uint64,
	partitionMask uint64,
	partitionFabric func() (DataDictionary, error),
) (DataDictionary, error) {
	var (
		dict = PartitionedDictionary{
			partitions:    make([]DataDictionary, partitionNum),
			partitionMask: partitionMask,
		}
		err error
	)

	for i := range dict.partitions {
		dict.partitions[i], err = partitionFabric()
		if err != nil {
			return nil, err
		}
	}

	return &dict, nil
}

func (o *PartitionedDictionary) Add(key uint64, value []byte, expiration time.Time) error {
	return o.partitions[o.chunkKey(key)].Add(key, value, expiration)
}

func (o *PartitionedDictionary) Get(key uint64) (Buffer, error) {
	return o.partitions[o.chunkKey(key)].Get(key)
}

func (o *PartitionedDictionary) chunkKey(key uint64) int {
	return int(key & o.partitionMask)
}

func (o *PartitionedDictionary) Clean(ctx context.Context) error {
	for i := 0; i < len(o.partitions); i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		err := o.partitions[i].Clean(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
