package storages

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func toBytes(i uint64) []byte {
	return []byte(strconv.FormatInt(int64(i), 10))
}

func TestPartitionDictionary_Add(t *testing.T) {
	dictCount := uint64(4)
	dataDicts := make([]*mockDataDictionary, 0, dictCount)
	expiration := time.Now()

	fabric := func() (DataDictionary, error) {
		key := uint64(len(dataDicts))

		d := &mockDataDictionary{}
		d.On("Add", key, toBytes(key), expiration).Return(nil)

		dataDicts = append(dataDicts, d)

		return d, nil
	}

	dict, err := NewPartitionedDictionary(dictCount, 0x3, fabric)
	require.NoError(t, err)

	for i := uint64(0); i < dictCount; i++ {
		err = dict.Add(i, toBytes(i), expiration)
		require.NoError(t, err)
	}

	for _, d := range dataDicts {
		d.AssertExpectations(t)
	}
}

func TestPartitionDictionary_Get(t *testing.T) {
	dictCount := uint64(4)
	dataDicts := make([]*mockDataDictionary, 0, dictCount)
	expiration := time.Now().Add(1 * time.Minute)

	fabric := func() (DataDictionary, error) {
		key := uint64(len(dataDicts))
		value := toBytes(key)

		d := &mockDataDictionary{}
		d.On("Add", key, value, expiration).Return(nil)
		d.On("Get", key).Return(newBuffer(&mockRefCounter{}, value), nil)

		dataDicts = append(dataDicts, d)

		return d, nil
	}

	dict, err := NewPartitionedDictionary(dictCount, 0x3, fabric)
	require.NoError(t, err)

	for i := uint64(0); i < dictCount; i++ {
		err = dict.Add(i, toBytes(i), expiration)
		require.NoError(t, err)
	}

	for i := uint64(0); i < dictCount; i++ {
		value, err := dict.Get(i)
		require.NoError(t, err)
		assert.Equal(t, toBytes(i), value.Bytes())
	}

	for _, d := range dataDicts {
		d.AssertExpectations(t)
	}
}

func TestPartitionDictionary_Clean(t *testing.T) {
	dictCount := uint64(4)
	dataDicts := make([]*mockDataDictionary, 0, dictCount)
	ctx := context.Background()

	fabric := func() (DataDictionary, error) {
		d := &mockDataDictionary{}
		d.On("Clean", ctx).Return(nil)

		dataDicts = append(dataDicts, d)

		return d, nil
	}

	dict, err := NewPartitionedDictionary(dictCount, 0x3, fabric)
	require.NoError(t, err)

	err = dict.Clean(ctx)
	require.NoError(t, err)

	for _, d := range dataDicts {
		d.AssertExpectations(t)
	}
}
