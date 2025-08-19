// Package storage provides Google Cloud Storage implementation.
// This package is currently not used by the main application but is kept
// for future cloud storage needs.
package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// GCSStorage implements StorageProvider for Google Cloud Storage
type GCSStorage struct {
	client     *storage.Client
	bucketName string
	projectID  string
	retention  time.Duration
}

// NewGCSStorage creates a new GCS storage provider
func NewGCSStorage(ctx context.Context, bucketName, projectID string, retention time.Duration) (*GCSStorage, error) {
	var client *storage.Client
	var err error

	if projectID != "" {
		client, err = storage.NewClient(ctx, option.WithQuotaProject(projectID))
	} else {
		client, err = storage.NewClient(ctx) // Auto-detect project ID
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	// Verify bucket exists
	bucket := client.Bucket(bucketName)
	if _, err := bucket.Attrs(ctx); err != nil {
		return nil, fmt.Errorf("bucket %s not accessible: %w", bucketName, err)
	}

	// Auto-detect project ID if not provided
	if projectID == "" {
		if attrs, err := bucket.Attrs(ctx); err == nil {
			projectID = fmt.Sprintf("%d", attrs.ProjectNumber)
		}
	}

	return &GCSStorage{
		client:     client,
		bucketName: bucketName,
		projectID:  projectID,
		retention:  retention,
	}, nil
}

// Store saves a file to GCS
func (gcs *GCSStorage) Store(ctx context.Context, filename string, data io.Reader) (string, error) {
	// Generate unique ID
	id := generateUniqueID()
	objectName := fmt.Sprintf("uploads/%s/%s", time.Now().Format("2006-01-02"), id)

	// Create object writer
	bucket := gcs.client.Bucket(gcs.bucketName)
	obj := bucket.Object(objectName)
	writer := obj.NewWriter(ctx)

	// Set metadata
	writer.Metadata = map[string]string{
		"original-filename": filename,
		"upload-time":       time.Now().Format(time.RFC3339),
	}

	// Copy data
	if _, err := io.Copy(writer, data); err != nil {
		writer.Close()
		return "", fmt.Errorf("failed to copy data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	return id, nil
}

// Retrieve retrieves a file from GCS
func (gcs *GCSStorage) Retrieve(ctx context.Context, id string) (io.ReadCloser, error) {
	// Find object by ID (this is simplified - in production you'd need a mapping)
	bucket := gcs.client.Bucket(gcs.bucketName)

	// List objects to find the one with matching ID
	query := &storage.Query{
		Prefix: "uploads/",
	}

	it := bucket.Objects(ctx, query)
	for {
		obj, err := it.Next()
		if err != nil {
			if err.Error() == "iterator done" {
				break
			}
			return nil, fmt.Errorf("failed to iterate objects: %w", err)
		}

		// Check if this object contains our ID
		if obj.Metadata != nil && obj.Metadata["id"] == id {
			reader, err := bucket.Object(obj.Name).NewReader(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to create reader: %w", err)
			}
			return reader, nil
		}
	}

	return nil, fmt.Errorf("file not found: %s", id)
}

// Delete removes a file from GCS
func (gcs *GCSStorage) Delete(ctx context.Context, id string) error {
	// Find and delete object by ID (simplified)
	bucket := gcs.client.Bucket(gcs.bucketName)

	query := &storage.Query{
		Prefix: "uploads/",
	}

	it := bucket.Objects(ctx, query)
	for {
		obj, err := it.Next()
		if err != nil {
			if err.Error() == "iterator done" {
				break
			}
			return fmt.Errorf("failed to iterate objects: %w", err)
		}

		if obj.Metadata != nil && obj.Metadata["id"] == id {
			if err := bucket.Object(obj.Name).Delete(ctx); err != nil {
				return fmt.Errorf("failed to delete object: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("file not found: %s", id)
}

// GetURL returns a GCS URL for the file
func (gcs *GCSStorage) GetURL(ctx context.Context, id string) (string, error) {
	// For now, return a simple URL - in production you might want to generate signed URLs
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", gcs.bucketName, id), nil
}

// Cleanup removes expired files (GCS handles this via lifecycle policies)
func (gcs *GCSStorage) Cleanup(ctx context.Context) error {
	// GCS handles cleanup via bucket lifecycle policies
	// This method is kept for interface compatibility
	return nil
}

// Close closes the GCS client
func (gcs *GCSStorage) Close() error {
	return gcs.client.Close()
}
