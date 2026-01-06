package mocks

import (
	"context"
	"errors"
	"sync"
)

// MockCache is a mock implementation of cache.Cache for testing
type MockCache struct {
	mu   sync.RWMutex
	data map[string][]byte

	// Control behavior
	GetError   error
	SetError   error
	PingError  error
	CloseError error

	// Track calls
	GetCalls   []string
	SetCalls   []SetCall
	PingCalls  int
	CloseCalls int
}

type SetCall struct {
	Key  string
	Data []byte
}

// NewMockCache creates a new mock cache
func NewMockCache() *MockCache {
	return &MockCache{
		data:     make(map[string][]byte),
		GetCalls: make([]string, 0),
		SetCalls: make([]SetCall, 0),
	}
}

// Get retrieves data from mock cache
func (m *MockCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetCalls = append(m.GetCalls, key)

	if m.GetError != nil {
		return nil, false, m.GetError
	}

	data, found := m.data[key]
	return data, found, nil
}

// Set stores data in mock cache
func (m *MockCache) Set(ctx context.Context, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SetCalls = append(m.SetCalls, SetCall{Key: key, Data: data})

	if m.SetError != nil {
		return m.SetError
	}

	m.data[key] = data
	return nil
}

// Ping checks mock cache health
func (m *MockCache) Ping(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PingCalls++
	return m.PingError
}

// Close closes mock cache
func (m *MockCache) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CloseCalls++
	return m.CloseError
}

// SetData pre-populates cache data for testing
func (m *MockCache) SetData(key string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = data
}

// ClearData clears all cached data
func (m *MockCache) ClearData() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string][]byte)
}

// Reset resets all mock state
func (m *MockCache) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string][]byte)
	m.GetCalls = make([]string, 0)
	m.SetCalls = make([]SetCall, 0)
	m.PingCalls = 0
	m.CloseCalls = 0
	m.GetError = nil
	m.SetError = nil
	m.PingError = nil
	m.CloseError = nil
}

// Common errors for testing
var (
	ErrCacheUnavailable = errors.New("cache unavailable")
	ErrCacheTimeout     = errors.New("cache timeout")
)
