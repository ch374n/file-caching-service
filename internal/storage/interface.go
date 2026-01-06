package storage

import (
	"context"
	"io"
)

// Storage defines the interface for object storage operations
// This allows for easy mocking in tests
type Storage interface {
	GetObject(ctx context.Context, key string) ([]byte, error)
	PutObject(ctx context.Context, key string, data io.Reader, contentType string) error
	DeleteObject(ctx context.Context, key string) error
	ObjectExists(ctx context.Context, key string) (bool, error)
	HealthCheck(ctx context.Context) error
}

// Ensure R2Client implements Storage interface
var _ Storage = (*R2Client)(nil)
