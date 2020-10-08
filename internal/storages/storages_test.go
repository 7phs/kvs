package storages

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemStorages_Add(t *testing.T) {
	now := time.Now()

	conf := &mockConfig{
		Exp:   1 * time.Second,
		TimeS: constantTime(now),
	}

	key := []byte("0123456789")
	hashedKey := uint64(0x8208d73d0fcfef26)
	expiration := now.Add(conf.Exp)
	value := []byte("test-value")

	mockDict := &mockDataDictionary{}
	mockDict.On("Add", hashedKey, value, expiration).Return(nil)

	storage, err := NewInMemStorages(conf, mockDict)
	require.NoError(t, err)

	err = storage.Add(key, value)
	require.NoError(t, err)

	mockDict.AssertExpectations(t)
}

func TestInMemStorages_Get(t *testing.T) {
	now := time.Now()

	conf := &mockConfig{
		Exp:   1 * time.Second,
		TimeS: constantTime(now),
	}

	key := []byte("0123456789")
	hashedKey := uint64(0x8208d73d0fcfef26)
	value := []byte("test-value")

	mockDict := &mockDataDictionary{}
	mockDict.On("Get", hashedKey).Return(newBuffer(&mockRefCounter{}, value), nil)

	storage, err := NewInMemStorages(conf, mockDict)
	require.NoError(t, err)

	storedValue, err := storage.Get(key)
	require.NoError(t, err)
	assert.Equal(t, value, storedValue.Bytes())
	assert.Equal(t, value, storedValue.Bytes())

	mockDict.AssertExpectations(t)
}

func TestInMemStorages_GetNotFound(t *testing.T) {
	now := time.Now()

	conf := &mockConfig{
		Exp:   1 * time.Second,
		TimeS: constantTime(now),
	}

	key := []byte("0123456789")
	hashedKey := uint64(0x8208d73d0fcfef26)

	mockDict := &mockDataDictionary{}
	mockDict.On("Get", hashedKey).Return(newBuffer(nil, nil), ErrKeyNotFound)

	storage, err := NewInMemStorages(conf, mockDict)
	require.NoError(t, err)

	_, err = storage.Get(key)
	require.Error(t, err)
	assert.EqualError(t, err, ErrKeyNotFound.Error())

	mockDict.AssertExpectations(t)
}

func TestInMemStorages_Clean(t *testing.T) {
	conf := &mockConfig{}

	ctx := context.Background()

	mockDict := &mockDataDictionary{}
	mockDict.On("Clean", ctx).Return(nil)

	storage, err := NewInMemStorages(conf, mockDict)
	require.NoError(t, err)

	err = storage.Clean(ctx)
	require.NoError(t, err)

	mockDict.AssertExpectations(t)
}

func TestInMemStorages_CleanFailed(t *testing.T) {
	conf := &mockConfig{}

	ctx := context.Background()

	mockDict := &mockDataDictionary{}
	mockDict.On("Clean", ctx).Return(ErrOutOfLimit)

	storage, err := NewInMemStorages(conf, mockDict)
	require.NoError(t, err)

	err = storage.Clean(ctx)
	require.Error(t, err)
	assert.EqualError(t, err, ErrOutOfLimit.Error())

	mockDict.AssertExpectations(t)
}
