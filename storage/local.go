// Package storage provides local filesystem storage implementation.
// This package is currently not used by the main application but is kept
// for future storage needs.
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LocalStorage implements StorageProvider for local filesystem
type LocalStorage struct {
	baseDir   string
	retention time.Duration
	mu        sync.RWMutex
	metadata  map[string]*FileInfo
}

// NewLocalStorage creates a new local storage provider
func NewLocalStorage(baseDir string, retention time.Duration) (*LocalStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	ls := &LocalStorage{
		baseDir:   baseDir,
		retention: retention,
		metadata:  make(map[string]*FileInfo),
	}

	// Load existing metadata
	if err := ls.loadMetadata(); err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	// Start background cleanup
	go ls.backgroundCleanup()

	return ls, nil
}

// Store saves a file to local storage
func (ls *LocalStorage) Store(ctx context.Context, filename string, data io.Reader) (string, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Generate unique ID
	id := generateUniqueID()
	filePath := filepath.Join(ls.baseDir, id)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	size, err := io.Copy(file, data)
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return "", fmt.Errorf("failed to copy data: %w", err)
	}

	// Create metadata
	now := time.Now()
	fileInfo := &FileInfo{
		ID:          id,
		Filename:    filename,
		Size:        size,
		CreatedAt:   now,
		ExpiresAt:   now.Add(ls.retention),
		ContentType: "application/octet-stream",
	}

	// Store metadata
	ls.metadata[id] = fileInfo
	if err := ls.saveMetadata(); err != nil {
		os.Remove(filePath) // Clean up on error
		return "", fmt.Errorf("failed to save metadata: %w", err)
	}

	return id, nil
}

// Retrieve retrieves a file from local storage
func (ls *LocalStorage) Retrieve(ctx context.Context, id string) (io.ReadCloser, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	if _, exists := ls.metadata[id]; !exists {
		return nil, fmt.Errorf("file not found: %s", id)
	}

	filePath := filepath.Join(ls.baseDir, id)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete removes a file from local storage
func (ls *LocalStorage) Delete(ctx context.Context, id string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	filePath := filepath.Join(ls.baseDir, id)

	// Remove file
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	// Remove metadata
	delete(ls.metadata, id)
	if err := ls.saveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// GetURL returns a local file path (for local storage, this is just the path)
func (ls *LocalStorage) GetURL(ctx context.Context, id string) (string, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	if _, exists := ls.metadata[id]; !exists {
		return "", fmt.Errorf("file not found: %s", id)
	}

	return filepath.Join(ls.baseDir, id), nil
}

// Cleanup removes expired files
func (ls *LocalStorage) Cleanup(ctx context.Context) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	now := time.Now()
	var expiredIDs []string

	// Find expired files
	for id, fileInfo := range ls.metadata {
		if now.After(fileInfo.ExpiresAt) {
			expiredIDs = append(expiredIDs, id)
		}
	}

	// Remove expired files
	for _, id := range expiredIDs {
		filePath := filepath.Join(ls.baseDir, id)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Failed to remove expired file %s: %v\n", id, err)
		}
		delete(ls.metadata, id)
	}

	// Save updated metadata
	if len(expiredIDs) > 0 {
		if err := ls.saveMetadata(); err != nil {
			return fmt.Errorf("failed to save metadata: %w", err)
		}
	}

	return nil
}

// Close performs cleanup when the storage provider is no longer needed
func (ls *LocalStorage) Close() error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Clean up all files
	for id := range ls.metadata {
		filePath := filepath.Join(ls.baseDir, id)
		os.Remove(filePath)
	}
	ls.metadata = make(map[string]*FileInfo)
	return nil
}

// Helper methods for metadata management
func (ls *LocalStorage) loadMetadata() error {
	metadataPath := filepath.Join(ls.baseDir, "metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No metadata file yet
		}
		return err
	}

	return json.Unmarshal(data, &ls.metadata)
}

func (ls *LocalStorage) SaveMetadata() error {
	metadataPath := filepath.Join(ls.baseDir, "metadata.json")

	data, err := json.Marshal(ls.metadata)
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

func (ls *LocalStorage) saveMetadata() error {
	return ls.SaveMetadata()
}

func (ls *LocalStorage) backgroundCleanup() {
	ticker := time.NewTicker(ls.retention / 2)
	defer ticker.Stop()

	for range ticker.C {
		if err := ls.Cleanup(context.Background()); err != nil {
			fmt.Printf("Background cleanup failed: %v\n", err)
		}
	}
}
