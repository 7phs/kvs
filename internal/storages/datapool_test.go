package storages

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataPool_Copy(t *testing.T) {
	data := []byte("0123456789")
	bufSize := 16
	expiration := time.Now()

	memPool := &mockMemoryPool{}
	memPool.On("Get").Return(make([]byte, bufSize), nil)
	memPool.On("Get").Return(make([]byte, bufSize), nil)

	pool, err := NewDataPool(memPool)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		buf, err := pool.Copy(data, expiration)
		require.NoError(t, err)
		assert.Equal(t, data, buf.Bytes())
	}

	memPool.AssertExpectations(t)
}

func TestDataPool_Clean(t *testing.T) {
	data := []byte("0123456789")
	bufSize := 16
	buf1 := make([]byte, bufSize)
	expiration := time.Now()
	ctx := context.Background()

	memPool := &mockMemoryPool{}
	memPool.On("Get").Return(buf1, nil)
	memPool.On("Get").Return(make([]byte, bufSize), nil)
	memPool.On("Put", buf1)

	pool, err := NewDataPool(memPool)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		buf, err := pool.Copy(data, expiration)
		require.NoError(t, err)
		assert.Equal(t, data, buf.Bytes())
	}

	runtime.Gosched()

	err = pool.Clean(ctx)
	require.NoError(t, err)

	memPool.AssertExpectations(t)
}
