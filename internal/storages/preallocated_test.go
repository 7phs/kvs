package storages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreAllocatedBuffer_allocate(t *testing.T) {
	allocateSz := 16
	now := time.Now()
	expiration := now.Add(1 * time.Minute)

	data := make([]byte, 128)
	preAllocated := newPreAllocatedBuffer(data)

	buf, ok := preAllocated.allocate(allocateSz, expiration)
	require.True(t, ok)
	assert.Len(t, buf.Bytes(), allocateSz)
	assert.False(t, preAllocated.isExpired(now))
}

func TestPreAllocatedBuffer_allocateExpired(t *testing.T) {
	allocateSz := 16
	now := time.Now()
	expiration := now

	data := make([]byte, 128)
	preAllocated := newPreAllocatedBuffer(data)

	buf, ok := preAllocated.allocate(allocateSz, expiration)
	require.True(t, ok)
	assert.Len(t, buf.Bytes(), allocateSz)
	assert.True(t, preAllocated.isExpired(now))
}

func TestPreAllocatedBuffer_allocateOutOfLimit(t *testing.T) {
	allocateSz := 16
	now := time.Now()
	expiration := now.Add(1 * time.Minute)

	data := make([]byte, 129)
	preAllocated := newPreAllocatedBuffer(data)

	bufCap := len(data) / allocateSz

	for i := 0; i < bufCap; i++ {
		buf, ok := preAllocated.allocate(allocateSz, expiration)
		require.True(t, ok)
		assert.Len(t, buf.Bytes(), allocateSz)
	}

	_, ok := preAllocated.allocate(allocateSz, expiration)
	require.False(t, ok)

	assert.False(t, preAllocated.isExpired(now))
}

func TestPreAllocatedBuffer_allocateNotExpired(t *testing.T) {
	allocateSz := 16
	now := time.Now()
	expiration := now

	data := make([]byte, 129)
	preAllocated := newPreAllocatedBuffer(data)

	bufCap := len(data) / allocateSz

	for i := 0; i < bufCap-1; i++ {
		buf, ok := preAllocated.allocate(allocateSz, expiration)
		require.True(t, ok)
		assert.Len(t, buf.Bytes(), allocateSz)
	}

	_, ok := preAllocated.allocate(allocateSz, expiration.Add(1*time.Minute))
	require.True(t, ok)

	assert.False(t, preAllocated.isExpired(now))
}

func TestPreAllocatedBuffer_allocateNotExpiredInUse(t *testing.T) {
	allocateSz := 16
	now := time.Now()
	expiration := now

	data := make([]byte, 129)
	preAllocated := newPreAllocatedBuffer(data)

	bufCap := len(data) / allocateSz

	for i := 0; i < bufCap-1; i++ {
		buf, ok := preAllocated.allocate(allocateSz, expiration)
		require.True(t, ok)
		assert.Len(t, buf.Bytes(), allocateSz)
	}

	buf, ok := preAllocated.allocate(allocateSz, expiration)
	require.True(t, ok)

	buf.inUse()

	assert.False(t, preAllocated.isExpired(now))

	buf.Free()

	assert.True(t, preAllocated.isExpired(now))
}

func TestQueue_push(t *testing.T) {
	count := 2

	queue := &queueAllocations{}

	for i := 0; i < count; i++ {
		queue.push(&preAllocatedBuffer{})
	}

	assert.Equal(t, count, queue.len())
}

func TestQueue_pop(t *testing.T) {
	count := 4
	expiredCount := 3
	expired := make([]*preAllocatedBuffer, expiredCount)

	queue := &queueAllocations{}

	// Add expired 1
	expired[0] = newPreAllocatedBuffer(make([]byte, 128))
	_, ok := expired[0].allocate(16, time.Now())
	require.True(t, ok)

	queue.push(expired[0])

	// Add in use 1
	for i := 0; i < count/2; i++ {
		buf := newPreAllocatedBuffer(make([]byte, 128))
		_, ok := buf.allocate(16, time.Now().Add(1*time.Minute))
		require.True(t, ok)

		queue.push(buf)
	}

	// Add expired 2
	expired[1] = newPreAllocatedBuffer(make([]byte, 128))
	_, ok = expired[1].allocate(16, time.Now())
	require.True(t, ok)

	queue.push(expired[1])

	// Add in use 2
	for i := 0; i < count/2; i++ {
		buf := newPreAllocatedBuffer(make([]byte, 128))
		_, ok := buf.allocate(16, time.Now().Add(1*time.Minute))
		require.True(t, ok)

		queue.push(buf)
	}

	// Add expired 3
	expired[2] = newPreAllocatedBuffer(make([]byte, 128))
	_, ok = expired[2].allocate(16, time.Now())
	require.True(t, ok)

	queue.push(expired[2])

	assert.Equal(t, count+expiredCount, queue.len())

	// Pop

	for i := 0; i < expiredCount; i++ {
		popExpired, ok := queue.pop(time.Now())
		require.True(t, ok)
		assert.Equal(t, expired[i], popExpired)
	}

	assert.Equal(t, count, queue.len())

	// Another are in use
	_, ok = queue.pop(time.Now())
	require.False(t, ok)
}
