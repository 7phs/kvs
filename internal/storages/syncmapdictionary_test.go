package storages

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncMapDictionary_Add(t *testing.T) {
	hashedKey := uint64(0x8208d73d0fcfef26)
	expiration := time.Now().Add(1 * time.Minute)
	value := []byte("test-value")

	dataPool := &mockDataPool{}
	dataPool.On("Copy", value, expiration).Return(newBuffer(dataPool, value), nil)
	dataPool.On("BufferInUse")

	dict := NewSyncMapDictionary(dataPool)

	err := dict.Add(hashedKey, value, expiration)
	require.NoError(t, err)

	storedValue, err := dict.Get(hashedKey)
	require.NoError(t, err)
	assert.Equal(t, value, storedValue.Bytes())

	dataPool.AssertExpectations(t)
}

func TestSyncMapDictionary_GetNotFound(t *testing.T) {
	hashedKey := uint64(0x8208d73d0fcfef26)

	dataPool := &mockDataPool{}

	dict := NewSyncMapDictionary(dataPool)

	_, err := dict.Get(hashedKey)
	require.Error(t, err)
	assert.EqualError(t, err, ErrKeyNotFound.Error())

	dataPool.AssertExpectations(t)
}

func TestSyncMapDictionary_GetExpiration(t *testing.T) {
	hashedKey := uint64(0x8208d73d0fcfef26)
	expiration := time.Now()
	value := []byte("test-value")

	dataPool := &mockDataPool{}
	dataPool.On("Copy", value, expiration).Return(newBuffer(dataPool, value), nil)

	dict := NewSyncMapDictionary(dataPool)

	err := dict.Add(hashedKey, value, expiration)
	require.NoError(t, err)

	_, err = dict.Get(hashedKey)
	require.Error(t, err)
	assert.EqualError(t, err, ErrKeyExpired.Error())

	dataPool.AssertExpectations(t)
}

func TestSyncMapDictionary_Clean(t *testing.T) {
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

	dict := NewSyncMapDictionary(dataPool)

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
