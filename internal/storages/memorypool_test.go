package storages

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryPool_Get(t *testing.T) {
	sz := 1024
	pool := NewMemoryPool(sz)

	buf, err := pool.Get()
	require.NoError(t, err)
	assert.Len(t, buf, sz)
}

func TestMemoryPool_Put(t *testing.T) {
	sz := 1024
	v := []byte("01234567")
	pool := NewMemoryPool(sz)

	buf, err := pool.Get()
	require.NoError(t, err)
	assert.Len(t, buf, sz)

	copy(buf[:len(v)], v)

	pool.Put(buf)

	buf, err = pool.Get()
	require.NoError(t, err)
	assert.Len(t, buf, sz)
	assert.Equal(t, v, buf[:len(v)])
}
