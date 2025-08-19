// Package storage provides an abstract interface for file storage operations.
// This package is currently not used by the main application but is kept
// for future storage needs (e.g., persistent file storage, cloud storage).
package storage

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

// StorageProvider defines the interface for different storage backends
type StorageProvider interface {
	// Store saves a file to storage and returns a unique identifier
	Store(ctx context.Context, filename string, data io.Reader) (string, error)

	// Retrieve retrieves a file from storage by ID
	Retrieve(ctx context.Context, id string) (io.ReadCloser, error)

	// Delete removes a file from storage by ID
	Delete(ctx context.Context, id string) error

	// GetURL returns a URL for accessing the file (if supported)
	GetURL(ctx context.Context, id string) (string, error)

	// Cleanup performs maintenance operations (e.g., TTL cleanup)
	Cleanup(ctx context.Context) error

	// Close performs cleanup when the storage provider is no longer needed
	// This is optional and may not be implemented by all providers
	Close() error
}

// FileInfo represents metadata about a stored file
type FileInfo struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	ContentType string    `json:"content_type,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for FileInfo
func (fi *FileInfo) MarshalJSON() ([]byte, error) {
	type Alias FileInfo
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
		ExpiresAt string `json:"expires_at,omitempty"`
	}{
		Alias:     (*Alias)(fi),
		CreatedAt: fi.CreatedAt.Format(time.RFC3339),
		ExpiresAt: fi.ExpiresAt.Format(time.RFC3339),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for FileInfo
func (fi *FileInfo) UnmarshalJSON(data []byte) error {
	type Alias FileInfo
	aux := &struct {
		*Alias
		CreatedAt string `json:"created_at"`
		ExpiresAt string `json:"expires_at,omitempty"`
	}{
		Alias: (*Alias)(fi),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse CreatedAt
	if aux.CreatedAt != "" {
		createdAt, err := time.Parse(time.RFC3339, aux.CreatedAt)
		if err != nil {
			return err
		}
		fi.CreatedAt = createdAt
	}

	// Parse ExpiresAt if present
	if aux.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, aux.ExpiresAt)
		if err != nil {
			return err
		}
		fi.ExpiresAt = expiresAt
	}

	return nil
}

// generateUniqueID creates a unique identifier for stored files
func generateUniqueID() string {
	return randomString(16)
}

// randomString generates a random string of the specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
