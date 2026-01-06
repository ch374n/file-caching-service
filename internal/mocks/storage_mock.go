package mocks

import (
	"context"
	"errors"
	"io"
	"sync"
)

// MockStorage is a mock implementation of storage.Storage for testing
type MockStorage struct {
	mu      sync.RWMutex
	objects map[string][]byte

	// Control behavior
	GetError         error
	PutError         error
	DeleteError      error
	ExistsError      error
	HealthCheckError error

	// Track calls
	GetCalls         []string
	PutCalls         []PutCall
	DeleteCalls      []string
	ExistsCalls      []string
	HealthCheckCalls int
}

type PutCall struct {
	Key         string
	ContentType string
	Data        []byte
}

// NewMockStorage creates a new mock storage
func NewMockStorage() *MockStorage {
	return &MockStorage{
		objects:     make(map[string][]byte),
		GetCalls:    make([]string, 0),
		PutCalls:    make([]PutCall, 0),
		DeleteCalls: make([]string, 0),
		ExistsCalls: make([]string, 0),
	}
}

// GetObject retrieves an object from mock storage
func (m *MockStorage) GetObject(ctx context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetCalls = append(m.GetCalls, key)

	if m.GetError != nil {
		return nil, m.GetError
	}

	data, found := m.objects[key]
	if !found {
		return nil, ErrObjectNotFound
	}

	return data, nil
}

// PutObject stores an object in mock storage
func (m *MockStorage) PutObject(ctx context.Context, key string, data io.Reader, contentType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	content, err := io.ReadAll(data)
	if err != nil {
		return err
	}

	m.PutCalls = append(m.PutCalls, PutCall{
		Key:         key,
		ContentType: contentType,
		Data:        content,
	})

	if m.PutError != nil {
		return m.PutError
	}

	m.objects[key] = content
	return nil
}

// DeleteObject deletes an object from mock storage
func (m *MockStorage) DeleteObject(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DeleteCalls = append(m.DeleteCalls, key)

	if m.DeleteError != nil {
		return m.DeleteError
	}

	delete(m.objects, key)
	return nil
}

// ObjectExists checks if an object exists in mock storage
func (m *MockStorage) ObjectExists(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ExistsCalls = append(m.ExistsCalls, key)

	if m.ExistsError != nil {
		return false, m.ExistsError
	}

	_, found := m.objects[key]
	return found, nil
}

// HealthCheck checks mock storage health
func (m *MockStorage) HealthCheck(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.HealthCheckCalls++
	return m.HealthCheckError
}

// SetObject pre-populates storage data for testing
func (m *MockStorage) SetObject(key string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[key] = data
}

// ClearObjects clears all stored objects
func (m *MockStorage) ClearObjects() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects = make(map[string][]byte)
}

// Reset resets all mock state
func (m *MockStorage) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.objects = make(map[string][]byte)
	m.GetCalls = make([]string, 0)
	m.PutCalls = make([]PutCall, 0)
	m.DeleteCalls = make([]string, 0)
	m.ExistsCalls = make([]string, 0)
	m.HealthCheckCalls = 0
	m.GetError = nil
	m.PutError = nil
	m.DeleteError = nil
	m.ExistsError = nil
	m.HealthCheckError = nil
}

// Common errors for testing
var (
	ErrObjectNotFound = errors.New("NoSuchKey: The specified key does not exist")
	ErrStorageTimeout = errors.New("storage timeout")
	ErrStorageError   = errors.New("storage error")
	ErrBucketNotFound = errors.New("bucket not found")
)
