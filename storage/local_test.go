package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewLocalStorage(t *testing.T) {
	// Test with valid parameters
	tempDir := t.TempDir()
	retention := 5 * time.Minute

	ls, err := NewLocalStorage(tempDir, retention)
	if err != nil {
		t.Fatalf("Failed to create LocalStorage: %v", err)
	}

	if ls == nil {
		t.Fatal("LocalStorage should not be nil")
	}

	if ls.baseDir != tempDir {
		t.Errorf("Expected baseDir %s, got %s", tempDir, ls.baseDir)
	}

	if ls.retention != retention {
		t.Errorf("Expected retention %v, got %v", retention, ls.retention)
	}

	if ls.metadata == nil {
		t.Error("Metadata map should be initialized")
	}

	// Test with invalid directory (non-writable)
	// Use /dev/null/subdir which will always fail to create
	invalidDir := "/dev/null/invalid_storage_dir"

	_, err = NewLocalStorage(invalidDir, retention)
	if err == nil {
		t.Error("Expected error when creating storage in invalid directory")
	}
}

func TestLocalStorage_Store(t *testing.T) {
	tempDir := t.TempDir()
	retention := 5 * time.Minute
	ls, err := NewLocalStorage(tempDir, retention)
	if err != nil {
		t.Fatalf("Failed to create LocalStorage: %v", err)
	}

	ctx := context.Background()
	filename := "test.gpx"
	content := "<?xml version=\"1.0\"?><gpx></gpx>"
	reader := strings.NewReader(content)

	// Test storing a file
	id, err := ls.Store(ctx, filename, reader)
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}

	if id == "" {
		t.Error("Store should return a non-empty ID")
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, id)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Stored file should exist on disk")
	}

	// Verify metadata was stored
	ls.mu.RLock()
	fileInfo, exists := ls.metadata[id]
	ls.mu.RUnlock()

	if !exists {
		t.Error("File metadata should be stored")
	}

	if fileInfo.Filename != filename {
		t.Errorf("Expected filename %s, got %s", filename, fileInfo.Filename)
	}

	if fileInfo.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), fileInfo.Size)
	}

	if fileInfo.ContentType != "application/octet-stream" {
		t.Errorf("Expected content type 'application/octet-stream', got %s", fileInfo.ContentType)
	}

	// Verify expiration time
	expectedExpiry := fileInfo.CreatedAt.Add(retention)
	if !fileInfo.ExpiresAt.Equal(expectedExpiry) {
		t.Errorf("Expected expiry %v, got %v", expectedExpiry, fileInfo.ExpiresAt)
	}
}

func TestLocalStorage_Retrieve(t *testing.T) {
	tempDir := t.TempDir()
	retention := 5 * time.Minute
	ls, err := NewLocalStorage(tempDir, retention)
	if err != nil {
		t.Fatalf("Failed to create LocalStorage: %v", err)
	}

	ctx := context.Background()
	filename := "test.gpx"
	content := "test content"
	reader := strings.NewReader(content)

	// Store a file first
	id, err := ls.Store(ctx, filename, reader)
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}

	// Test retrieving the file
	retrievedReader, err := ls.Retrieve(ctx, id)
	if err != nil {
		t.Fatalf("Failed to retrieve file: %v", err)
	}
	defer retrievedReader.Close()

	// Read and verify content
	retrievedContent, err := io.ReadAll(retrievedReader)
	if err != nil {
		t.Fatalf("Failed to read retrieved content: %v", err)
	}

	if string(retrievedContent) != content {
		t.Errorf("Expected content '%s', got '%s'", content, string(retrievedContent))
	}

	// Test retrieving non-existent file
	_, err = ls.Retrieve(ctx, "non-existent-id")
	if err == nil {
		t.Error("Expected error when retrieving non-existent file")
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	tempDir := t.TempDir()
	retention := 5 * time.Minute
	ls, err := NewLocalStorage(tempDir, retention)
	if err != nil {
		t.Fatalf("Failed to create LocalStorage: %v", err)
	}

	ctx := context.Background()
	filename := "test.gpx"
	content := "test content"
	reader := strings.NewReader(content)

	// Store a file first
	id, err := ls.Store(ctx, filename, reader)
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tempDir, id)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File should exist before deletion")
	}

	// Test deleting the file
	err = ls.Delete(ctx, id)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify file was removed from disk
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("File should not exist after deletion")
	}

	// Verify metadata was removed
	ls.mu.RLock()
	_, exists := ls.metadata[id]
	ls.mu.RUnlock()

	if exists {
		t.Error("File metadata should be removed after deletion")
	}

	// Test deleting non-existent file (should not error)
	err = ls.Delete(ctx, "non-existent-id")
	if err != nil {
		t.Errorf("Deleting non-existent file should not error: %v", err)
	}
}

func TestLocalStorage_GetURL(t *testing.T) {
	tempDir := t.TempDir()
	retention := 5 * time.Minute
	ls, err := NewLocalStorage(tempDir, retention)
	if err != nil {
		t.Fatalf("Failed to create LocalStorage: %v", err)
	}

	ctx := context.Background()
	filename := "test.gpx"
	content := "test content"
	reader := strings.NewReader(content)

	// Store a file first
	id, err := ls.Store(ctx, filename, reader)
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}

	// Test getting URL
	url, err := ls.GetURL(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}

	expectedURL := filepath.Join(tempDir, id)
	if url != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, url)
	}

	// Test getting URL for non-existent file
	_, err = ls.GetURL(ctx, "non-existent-id")
	if err == nil {
		t.Error("Expected error when getting URL for non-existent file")
	}
}

func TestLocalStorage_Cleanup(t *testing.T) {
	tempDir := t.TempDir()
	retention := 1 * time.Millisecond // Very short retention for testing
	ls, err := NewLocalStorage(tempDir, retention)
	if err != nil {
		t.Fatalf("Failed to create LocalStorage: %v", err)
	}

	ctx := context.Background()
	filename := "test.gpx"
	content := "test content"
	reader := strings.NewReader(content)

	// Store a file
	id, err := ls.Store(ctx, filename, reader)
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}

	// Wait for file to expire
	time.Sleep(2 * time.Millisecond)

	// Run cleanup
	err = ls.Cleanup(ctx)
	if err != nil {
		t.Fatalf("Failed to run cleanup: %v", err)
	}

	// Verify expired file was removed
	ls.mu.RLock()
	_, exists := ls.metadata[id]
	ls.mu.RUnlock()

	if exists {
		t.Error("Expired file metadata should be removed after cleanup")
	}

	// Verify file was removed from disk
	filePath := filepath.Join(tempDir, id)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Expired file should be removed from disk after cleanup")
	}
}

func TestLocalStorage_Close(t *testing.T) {
	tempDir := t.TempDir()
	retention := 5 * time.Minute
	ls, err := NewLocalStorage(tempDir, retention)
	if err != nil {
		t.Fatalf("Failed to create LocalStorage: %v", err)
	}

	ctx := context.Background()
	filename := "test.gpx"
	content := "test content"
	reader := strings.NewReader(content)

	// Store a file
	id, err := ls.Store(ctx, filename, reader)
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tempDir, id)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File should exist before close")
	}

	// Test close
	ls.Close()

	// Verify all files were removed
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("File should be removed after close")
	}

	// Verify metadata was cleared
	ls.mu.RLock()
	if len(ls.metadata) != 0 {
		t.Error("Metadata should be cleared after close")
	}
	ls.mu.RUnlock()
}

func TestLocalStorage_MetadataPersistence(t *testing.T) {
	tempDir := t.TempDir()
	retention := 5 * time.Minute

	// Create first storage instance
	ls1, err := NewLocalStorage(tempDir, retention)
	if err != nil {
		t.Fatalf("Failed to create first LocalStorage: %v", err)
	}

	ctx := context.Background()
	filename := "test.gpx"
	content := "test content"
	reader := strings.NewReader(content)

	// Store a file
	id, err := ls1.Store(ctx, filename, reader)
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}

	// Close first instance
	ls1.Close()

	// Create second storage instance (should load existing metadata)
	ls2, err := NewLocalStorage(tempDir, retention)
	if err != nil {
		t.Fatalf("Failed to create second LocalStorage: %v", err)
	}

	// Verify metadata was loaded
	ls2.mu.RLock()
	fileInfo, exists := ls2.metadata[id]
	ls2.mu.RUnlock()

	if !exists {
		t.Error("Metadata should be loaded from disk")
	}

	if fileInfo.Filename != filename {
		t.Errorf("Expected filename %s, got %s", filename, fileInfo.Filename)
	}

	// Clean up
	ls2.Close()
}

func TestLocalStorage_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	retention := 5 * time.Minute
	ls, err := NewLocalStorage(tempDir, retention)
	if err != nil {
		t.Fatalf("Failed to create LocalStorage: %v", err)
	}

	ctx := context.Background()
	numGoroutines := 10
	results := make(chan error, numGoroutines)

	// Test concurrent store operations
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			filename := fmt.Sprintf("test%d.gpx", index)
			content := fmt.Sprintf("content%d", index)
			reader := strings.NewReader(content)

			_, err := ls.Store(ctx, filename, reader)
			results <- err
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Goroutine %d failed: %v", i, err)
		}
	}

	// Verify all files were stored
	ls.mu.RLock()
	if len(ls.metadata) != numGoroutines {
		t.Errorf("Expected %d files, got %d", numGoroutines, len(ls.metadata))
	}
	ls.mu.RUnlock()

	// Clean up
	ls.Close()
}
