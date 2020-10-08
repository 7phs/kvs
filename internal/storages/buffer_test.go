package storages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferCreate(t *testing.T) {
	mockRefCounter := &mockRefCounter{}
	mockRefCounter.On("BufferInUse").Return()
	mockRefCounter.On("BufferFree").Return()

	data := []byte("test data")

	buf := newBuffer(mockRefCounter, data)
	buf.inUse()
	assert.Equal(t, data, buf.Bytes())

	buf.Free()
	assert.Nil(t, buf.refCounter)
	assert.Nil(t, buf.buf)
	mockRefCounter.AssertExpectations(t)
}

func TestBufferCopy(t *testing.T) {
	mockRefCounter := &mockRefCounter{}

	value := []byte("0123456789")
	data := make([]byte, len(value))

	buf := newBuffer(mockRefCounter, data)
	assert.Len(t, buf.Bytes(), len(data))

	buf.Copy(value)
	assert.Equal(t, value, buf.Bytes())

	mockRefCounter.AssertExpectations(t)
}
