package storages

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpiredList_push(t *testing.T) {
	cap := 10
	clearedSz := 5
	keys := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	list := newExpiredList(cap, clearedSz)

	list.push(keys...)

	index := 0
	list.Clean(func(storedKeys []uint64) bool {
		assert.Equal(t, keys[index:index+len(storedKeys)], storedKeys)
		index += len(storedKeys)
		assert.Equal(t, index, list.clearedIndex)

		return true
	})

	assert.Equal(t, len(keys), index)
}

func TestExpiredList_clean(t *testing.T) {
	cap := 10
	clearedSz := 10
	keys := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	list := newExpiredList(cap, clearedSz)

	list.push(keys...)

	index := 0
	list.Clean(func(storedKeys []uint64) bool {
		assert.Equal(t, keys, storedKeys)
		index += len(storedKeys)
		assert.Equal(t, index, list.clearedIndex)

		return true
	})
	assert.Equal(t, len(keys), index)
	assert.Len(t, list.list, 0)
	assert.Equal(t, 0, list.clearedIndex)
}

func TestExpiredList_cleanNotMultiple(t *testing.T) {
	cap := 10
	clearedSz := 8
	keys := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	list := newExpiredList(cap, clearedSz)

	list.push(keys...)

	index := 0
	list.Clean(func(storedKeys []uint64) bool {
		assert.Equal(t, keys[index:index+len(storedKeys)], storedKeys)
		index += clearedSz
		assert.Equal(t, index, list.clearedIndex)

		return true
	})
	assert.Equal(t, clearedSz*2, index)
	assert.Len(t, list.list, 0)
	assert.Equal(t, 0, list.clearedIndex)
}

func TestExpiredList_cleanBreak(t *testing.T) {
	cap := 10
	clearedSz := 5
	keys := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	list := newExpiredList(cap, clearedSz)

	list.push(keys...)

	index := 0
	list.Clean(func(storedKeys []uint64) bool {
		assert.Equal(t, keys[index:index+len(storedKeys)], storedKeys)
		index += len(storedKeys)
		assert.Equal(t, index, list.clearedIndex)

		return false
	})
	assert.Equal(t, clearedSz, index)
	assert.Len(t, list.list, len(keys))
	assert.Equal(t, clearedSz, list.clearedIndex)
}

func TestExpiredList_cleanMaximum(t *testing.T) {
	cap := 10
	clearedSz := 5
	keys := []uint64{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
	}

	list := newExpiredList(cap, clearedSz)

	list.push(keys...)

	index := 0
	list.Clean(func(storedKeys []uint64) bool {
		assert.Equal(t, keys[index:index+len(storedKeys)], storedKeys)
		index += len(storedKeys)
		assert.Equal(t, index, list.clearedIndex)

		return true
	})
	assert.Equal(t, clearedSz*maxClearedAttempt, index)
	assert.Len(t, list.list, len(keys))
	assert.Equal(t, clearedSz*maxClearedAttempt, list.clearedIndex)
}

func TestMapDictionary_Add(t *testing.T) {
	hashedKey := uint64(0x8208d73d0fcfef26)
	expiration := time.Now().Add(1 * time.Minute)
	value := []byte("test-value")

	dataPool := &mockDataPool{}
	dataPool.On("Copy", value, expiration).Return(newBuffer(dataPool, value), nil)
	dataPool.On("BufferInUse")

	dict := NewMapDictionary(dataPool)

	err := dict.Add(hashedKey, value, expiration)
	require.NoError(t, err)

	storedValue, err := dict.Get(hashedKey)
	require.NoError(t, err)
	assert.Equal(t, value, storedValue.Bytes())

	dataPool.AssertExpectations(t)
}

func TestMapDictionary_GetNotFound(t *testing.T) {
	hashedKey := uint64(0x8208d73d0fcfef26)

	dataPool := &mockDataPool{}

	dict := NewMapDictionary(dataPool)

	_, err := dict.Get(hashedKey)
	require.Error(t, err)
	assert.EqualError(t, err, ErrKeyNotFound.Error())

	dataPool.AssertExpectations(t)
}

func TestMapDictionary_GetExpiration(t *testing.T) {
	hashedKey := uint64(0x8208d73d0fcfef26)
	expiration := time.Now()
	value := []byte("test-value")

	dataPool := &mockDataPool{}
	dataPool.On("Copy", value, expiration).Return(newBuffer(dataPool, value), nil)

	dict := NewMapDictionary(dataPool)

	err := dict.Add(hashedKey, value, expiration)
	require.NoError(t, err)

	_, err = dict.Get(hashedKey)
	require.Error(t, err)
	assert.EqualError(t, err, ErrKeyExpired.Error())

	dataPool.AssertExpectations(t)
}

func TestMapDictionary_Clean(t *testing.T) {
	hashedKey1 := uint64(0x8208d73d0fcfef26)
	value1 := []byte("test-value")
	expiration1 := time.Now()
	hashedKey2 := uint64(0x8208d73d0fcfef27)
	value2 := []byte("test-value2")
	expiration2 := time.Now().Add(1 * time.Minute)

	ctx := context.Background()

	dataPool := &mockDataPool{}
	dataPool.On("Copy", value1, expiration1).Return(newBuffer(dataPool, value1), nil)
	dataPool.On("Copy", value2, expiration2).Return(newBuffer(dataPool, value2), nil)
	dataPool.On("BufferInUse")
	dataPool.On("Clean", ctx).Return(nil)

	dict := NewMapDictionary(dataPool)

	// Add
	err := dict.Add(hashedKey1, value1, expiration1)
	require.NoError(t, err)

	err = dict.Add(hashedKey2, value2, expiration2)
	require.NoError(t, err)

	// Get: 1
	_, err = dict.Get(hashedKey1)
	require.Error(t, err)
	assert.EqualError(t, err, ErrKeyExpired.Error())

	storedValue2, err := dict.Get(hashedKey2)
	require.NoError(t, err)
	assert.Equal(t, value2, storedValue2.Bytes())

	// Clean
	err = dict.Clean(ctx)
	require.NoError(t, err)

	// Get: 2
	_, err = dict.Get(hashedKey1)
	require.Error(t, err)
	assert.EqualError(t, err, ErrKeyNotFound.Error())

	storedValue2, err = dict.Get(hashedKey2)
	require.NoError(t, err)
	assert.Equal(t, value2, storedValue2.Bytes())

	dataPool.AssertExpectations(t)
}
