package mocks_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/ch374n/file-downloader/internal/mocks"
)

func TestMockCache_GetSet(t *testing.T) {
	cache := mocks.NewMockCache()
	ctx := context.Background()

	// Initially empty
	data, found, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if found {
		t.Error("Expected not found")
	}
	if data != nil {
		t.Error("Expected nil data")
	}

	// Set data
	testData := []byte("test value")
	err = cache.Set(ctx, "key1", testData)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get data
	data, found, err = cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !found {
		t.Error("Expected found")
	}
	if !bytes.Equal(data, testData) {
		t.Errorf("Expected '%s', got '%s'", testData, data)
	}

	// Check calls were recorded
	if len(cache.GetCalls) != 2 {
		t.Errorf("Expected 2 GetCalls, got %d", len(cache.GetCalls))
	}
	if len(cache.SetCalls) != 1 {
		t.Errorf("Expected 1 SetCalls, got %d", len(cache.SetCalls))
	}
}

func TestMockCache_Errors(t *testing.T) {
	cache := mocks.NewMockCache()
	ctx := context.Background()

	cache.GetError = mocks.ErrCacheUnavailable
	_, _, err := cache.Get(ctx, "key")
	if err != mocks.ErrCacheUnavailable {
		t.Errorf("Expected ErrCacheUnavailable, got %v", err)
	}

	cache.SetError = mocks.ErrCacheTimeout
	err = cache.Set(ctx, "key", []byte("value"))
	if err != mocks.ErrCacheTimeout {
		t.Errorf("Expected ErrCacheTimeout, got %v", err)
	}

	cache.PingError = mocks.ErrCacheUnavailable
	err = cache.Ping(ctx)
	if err != mocks.ErrCacheUnavailable {
		t.Errorf("Expected ErrCacheUnavailable, got %v", err)
	}
}

func TestMockCache_Reset(t *testing.T) {
	cache := mocks.NewMockCache()
	ctx := context.Background()

	cache.Set(ctx, "key", []byte("value"))
	cache.Get(ctx, "key")
	cache.Ping(ctx)
	cache.GetError = mocks.ErrCacheUnavailable

	cache.Reset()

	if len(cache.GetCalls) != 0 {
		t.Error("GetCalls not reset")
	}
	if len(cache.SetCalls) != 0 {
		t.Error("SetCalls not reset")
	}
	if cache.PingCalls != 0 {
		t.Error("PingCalls not reset")
	}
	if cache.GetError != nil {
		t.Error("GetError not reset")
	}

	// Data should be cleared
	_, found, _ := cache.Get(ctx, "key")
	if found {
		t.Error("Data not cleared on reset")
	}
}

func TestMockCache_SetData(t *testing.T) {
	cache := mocks.NewMockCache()
	ctx := context.Background()

	// Pre-populate using SetData
	cache.SetData("preloaded", []byte("preloaded value"))

	data, found, err := cache.Get(ctx, "preloaded")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !found {
		t.Error("Expected to find preloaded key")
	}
	if string(data) != "preloaded value" {
		t.Errorf("Expected 'preloaded value', got '%s'", data)
	}
}

func TestMockCache_ClearData(t *testing.T) {
	cache := mocks.NewMockCache()
	ctx := context.Background()

	cache.SetData("key1", []byte("value1"))
	cache.SetData("key2", []byte("value2"))

	cache.ClearData()

	_, found, _ := cache.Get(ctx, "key1")
	if found {
		t.Error("key1 should be cleared")
	}
	_, found, _ = cache.Get(ctx, "key2")
	if found {
		t.Error("key2 should be cleared")
	}
}

func TestMockStorage_GetSetObject(t *testing.T) {
	storage := mocks.NewMockStorage()
	ctx := context.Background()

	// Initially empty - should return not found error
	_, err := storage.GetObject(ctx, "key1")
	if err != mocks.ErrObjectNotFound {
		t.Fatalf("Expected ErrObjectNotFound, got %v", err)
	}

	// Set object using PutObject
	testData := []byte("test content")
	err = storage.PutObject(ctx, "key1", bytes.NewReader(testData), "text/plain")
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	// Get object
	data, err := storage.GetObject(ctx, "key1")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if !bytes.Equal(data, testData) {
		t.Errorf("Expected '%s', got '%s'", testData, data)
	}

	// Check calls
	if len(storage.GetCalls) != 2 {
		t.Errorf("Expected 2 GetCalls, got %d", len(storage.GetCalls))
	}
	if len(storage.PutCalls) != 1 {
		t.Errorf("Expected 1 PutCalls, got %d", len(storage.PutCalls))
	}
}

func TestMockStorage_DeleteObject(t *testing.T) {
	storage := mocks.NewMockStorage()
	ctx := context.Background()

	storage.SetObject("key1", []byte("content"))

	exists, _ := storage.ObjectExists(ctx, "key1")
	if !exists {
		t.Error("Expected object to exist")
	}

	err := storage.DeleteObject(ctx, "key1")
	if err != nil {
		t.Fatalf("DeleteObject failed: %v", err)
	}

	exists, _ = storage.ObjectExists(ctx, "key1")
	if exists {
		t.Error("Expected object to be deleted")
	}
}

func TestMockStorage_ObjectExists(t *testing.T) {
	storage := mocks.NewMockStorage()
	ctx := context.Background()

	exists, err := storage.ObjectExists(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("ObjectExists failed: %v", err)
	}
	if exists {
		t.Error("Expected not exists")
	}

	storage.SetObject("key1", []byte("content"))

	exists, err = storage.ObjectExists(ctx, "key1")
	if err != nil {
		t.Fatalf("ObjectExists failed: %v", err)
	}
	if !exists {
		t.Error("Expected exists")
	}
}

func TestMockStorage_HealthCheck(t *testing.T) {
	storage := mocks.NewMockStorage()
	ctx := context.Background()

	err := storage.HealthCheck(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if storage.HealthCheckCalls != 1 {
		t.Errorf("Expected 1 HealthCheckCalls, got %d", storage.HealthCheckCalls)
	}

	storage.HealthCheckError = mocks.ErrBucketNotFound
	err = storage.HealthCheck(ctx)
	if err != mocks.ErrBucketNotFound {
		t.Errorf("Expected ErrBucketNotFound, got %v", err)
	}
}

func TestMockStorage_Errors(t *testing.T) {
	storage := mocks.NewMockStorage()
	ctx := context.Background()

	storage.GetError = mocks.ErrStorageError
	_, err := storage.GetObject(ctx, "key")
	if err != mocks.ErrStorageError {
		t.Errorf("Expected ErrStorageError, got %v", err)
	}

	storage.PutError = mocks.ErrStorageTimeout
	err = storage.PutObject(ctx, "key", bytes.NewReader([]byte("data")), "text/plain")
	if err != mocks.ErrStorageTimeout {
		t.Errorf("Expected ErrStorageTimeout, got %v", err)
	}

	storage.DeleteError = mocks.ErrStorageError
	err = storage.DeleteObject(ctx, "key")
	if err != mocks.ErrStorageError {
		t.Errorf("Expected ErrStorageError, got %v", err)
	}
}

func TestMockStorage_Reset(t *testing.T) {
	storage := mocks.NewMockStorage()
	ctx := context.Background()

	storage.SetObject("key1", []byte("content"))
	storage.GetObject(ctx, "key1")
	storage.HealthCheck(ctx)
	storage.GetError = mocks.ErrStorageError

	storage.Reset()

	if len(storage.GetCalls) != 0 {
		t.Error("GetCalls not reset")
	}
	if storage.HealthCheckCalls != 0 {
		t.Error("HealthCheckCalls not reset")
	}
	if storage.GetError != nil {
		t.Error("GetError not reset")
	}

	// Data should be cleared
	_, err := storage.GetObject(ctx, "key1")
	if err != mocks.ErrObjectNotFound {
		t.Error("Objects not cleared on reset")
	}
}

func TestMockStorage_SetObject(t *testing.T) {
	storage := mocks.NewMockStorage()
	ctx := context.Background()

	// Pre-populate using SetObject
	storage.SetObject("preloaded", []byte("preloaded content"))

	data, err := storage.GetObject(ctx, "preloaded")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(data) != "preloaded content" {
		t.Errorf("Expected 'preloaded content', got '%s'", data)
	}
}

func TestMockStorage_ClearObjects(t *testing.T) {
	storage := mocks.NewMockStorage()
	ctx := context.Background()

	storage.SetObject("key1", []byte("value1"))
	storage.SetObject("key2", []byte("value2"))

	storage.ClearObjects()

	_, err := storage.GetObject(ctx, "key1")
	if err != mocks.ErrObjectNotFound {
		t.Error("key1 should be cleared")
	}
	_, err = storage.GetObject(ctx, "key2")
	if err != mocks.ErrObjectNotFound {
		t.Error("key2 should be cleared")
	}
}
