package storages

import (
	"context"
	"time"

	"github.com/7phs/kvs/internal/config"
	"github.com/stretchr/testify/mock"
)

var (
	_ refCounter     = (*mockRefCounter)(nil)
	_ refCounter     = (*mockDataPool)(nil)
	_ DataPool       = (*mockDataPool)(nil)
	_ DataDictionary = (*mockDataDictionary)(nil)
	_ config.Config  = (*mockConfig)(nil)
)

type mockRefCounter struct {
	mock.Mock
}

func (m *mockRefCounter) BufferInUse() {
	m.Called()
}

func (m *mockRefCounter) BufferFree() {
	m.Called()
}

type constantTime time.Time

func (o constantTime) Now() time.Time {
	return time.Time(o)
}

type mockConfig struct {
	Exp      time.Duration
	TimeS    config.TimeSource
	PreAlloc int
}

func (o *mockConfig) Port() int {
	return 0
}

func (o *mockConfig) Expiration() time.Duration {
	return o.Exp
}

func (o *mockConfig) Maintenance() time.Duration {
	return 0
}

func (o *mockConfig) PreAllocated() int {
	return o.PreAlloc
}

func (o *mockConfig) Mode() config.StorageMode {
	return config.StorageModeMap
}

func (o *mockConfig) TimeSource() config.TimeSource {
	return o.TimeS
}

type mockDataDictionary struct {
	mock.Mock
}

func (m *mockDataDictionary) Add(key uint64, data []byte, expiration time.Time) error {
	args := m.Called(key, data, expiration)

	return args.Error(0)
}

func (m *mockDataDictionary) Get(key uint64) (Buffer, error) {
	args := m.Called(key)

	return args.Get(0).(Buffer), args.Error(1)
}

func (m *mockDataDictionary) Clean(ctx context.Context) error {
	args := m.Called(ctx)

	return args.Error(0)
}

type mockDataPool struct {
	mock.Mock
}

func (m *mockDataPool) Copy(data []byte, expiration time.Time) (Buffer, error) {
	args := m.Called(data, expiration)

	return args.Get(0).(Buffer), args.Error(1)
}

func (m *mockDataPool) BufferInUse() {
	m.Called()
}

func (m *mockDataPool) BufferFree() {
	m.Called()
}

func (m *mockDataPool) Clean(ctx context.Context) error {
	args := m.Called(ctx)

	return args.Error(0)
}
