package storage

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateUniqueID(t *testing.T) {
	// Test that IDs are generated
	id1 := generateUniqueID()
	id2 := generateUniqueID()

	if id1 == "" {
		t.Error("Generated ID should not be empty")
	}

	if id2 == "" {
		t.Error("Generated ID should not be empty")
	}

	// IDs should be different (very unlikely to be the same)
	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}

	// IDs should be 16 characters long
	if len(id1) != 16 {
		t.Errorf("Expected ID length 16, got %d", len(id1))
	}

	if len(id2) != 16 {
		t.Errorf("Expected ID length 16, got %d", len(id2))
	}

	// IDs should only contain valid characters
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, char := range id1 {
		if !strings.ContainsRune(validChars, char) {
			t.Errorf("ID contains invalid character: %c", char)
		}
	}
}

func TestRandomString(t *testing.T) {
	// Test different lengths
	testCases := []int{1, 5, 10, 20, 100}

	for _, length := range testCases {
		result := randomString(length)
		if len(result) != length {
			t.Errorf("Expected length %d, got %d", length, len(result))
		}

		// Check that all characters are valid
		validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		for _, char := range result {
			if !strings.ContainsRune(validChars, char) {
				t.Errorf("String contains invalid character: %c", char)
			}
		}
	}

	// Test that consecutive calls produce different results
	// Note: This test might occasionally fail due to timing, but it's very unlikely
	str1 := randomString(10)
	str2 := randomString(10)

	// Add a small delay to ensure different timestamps
	time.Sleep(1 * time.Microsecond)
	str3 := randomString(10)

	// At least one of the strings should be different
	if str1 == str2 && str2 == str3 {
		t.Error("Random strings should be different (this test might occasionally fail due to timing)")
	}
}

func TestFileInfo(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(5 * time.Minute)

	fileInfo := &FileInfo{
		ID:          "test-id-123",
		Filename:    "test.gpx",
		Size:        1024,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		ContentType: "application/gpx+xml",
	}

	// Test basic fields
	if fileInfo.ID != "test-id-123" {
		t.Errorf("Expected ID 'test-id-123', got '%s'", fileInfo.ID)
	}

	if fileInfo.Filename != "test.gpx" {
		t.Errorf("Expected filename 'test.gpx', got '%s'", fileInfo.Filename)
	}

	if fileInfo.Size != 1024 {
		t.Errorf("Expected size 1024, got %d", fileInfo.Size)
	}

	if fileInfo.ContentType != "application/gpx+xml" {
		t.Errorf("Expected content type 'application/gpx+xml', got '%s'", fileInfo.ContentType)
	}

	// Test time fields
	if !fileInfo.CreatedAt.Equal(now) {
		t.Error("CreatedAt should equal the provided time")
	}

	if !fileInfo.ExpiresAt.Equal(expiresAt) {
		t.Error("ExpiresAt should equal the provided time")
	}

	// Test that ExpiresAt is after CreatedAt
	if !fileInfo.ExpiresAt.After(fileInfo.CreatedAt) {
		t.Error("ExpiresAt should be after CreatedAt")
	}
}

func TestFileInfoJSON(t *testing.T) {
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	expiresAt := now.Add(1 * time.Hour)

	fileInfo := &FileInfo{
		ID:          "test-123",
		Filename:    "example.gpx",
		Size:        2048,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		ContentType: "application/gpx+xml",
	}

	// Test JSON marshaling
	data, err := fileInfo.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal FileInfo to JSON: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled FileInfo
	err = unmarshaled.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal FileInfo from JSON: %v", err)
	}

	// Verify all fields are preserved
	if unmarshaled.ID != fileInfo.ID {
		t.Errorf("ID mismatch: expected '%s', got '%s'", fileInfo.ID, unmarshaled.ID)
	}

	if unmarshaled.Filename != fileInfo.Filename {
		t.Errorf("Filename mismatch: expected '%s', got '%s'", fileInfo.Filename, unmarshaled.Filename)
	}

	if unmarshaled.Size != fileInfo.Size {
		t.Errorf("Size mismatch: expected %d, got %d", fileInfo.Size, unmarshaled.Size)
	}

	if unmarshaled.ContentType != fileInfo.ContentType {
		t.Errorf("ContentType mismatch: expected '%s', got '%s'", fileInfo.ContentType, unmarshaled.ContentType)
	}

	// Time fields should be approximately equal (allowing for timezone differences)
	// Note: JSON marshaling/unmarshaling might lose sub-second precision
	if !unmarshaled.CreatedAt.Equal(fileInfo.CreatedAt) {
		// Check if they're within 1 second (JSON precision limitation)
		diff := unmarshaled.CreatedAt.Sub(fileInfo.CreatedAt)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("CreatedAt mismatch: expected %v, got %v (diff: %v)", fileInfo.CreatedAt, unmarshaled.CreatedAt, diff)
		}
	}

	if !unmarshaled.ExpiresAt.Equal(fileInfo.ExpiresAt) {
		// Check if they're within 1 second (JSON precision limitation)
		diff := unmarshaled.ExpiresAt.Sub(fileInfo.ExpiresAt)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("ExpiresAt mismatch: expected %v, got %v (diff: %v)", fileInfo.ExpiresAt, unmarshaled.ExpiresAt, diff)
		}
	}
}

func TestFileInfoWithoutOptionalFields(t *testing.T) {
	now := time.Now()

	fileInfo := &FileInfo{
		ID:        "minimal-id",
		Filename:  "minimal.gpx",
		Size:      512,
		CreatedAt: now,
		// ExpiresAt and ContentType are omitted
	}

	// Test JSON marshaling without optional fields
	data, err := fileInfo.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal minimal FileInfo to JSON: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled FileInfo
	err = unmarshaled.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal minimal FileInfo from JSON: %v", err)
	}

	// Verify required fields are preserved
	if unmarshaled.ID != fileInfo.ID {
		t.Errorf("ID mismatch: expected '%s', got '%s'", fileInfo.ID, unmarshaled.ID)
	}

	if unmarshaled.Filename != fileInfo.Filename {
		t.Errorf("Filename mismatch: expected '%s', got '%s'", fileInfo.Filename, unmarshaled.Filename)
	}

	if unmarshaled.Size != fileInfo.Size {
		t.Errorf("Size mismatch: expected %d, got %d", fileInfo.Size, unmarshaled.Size)
	}

	// Check if they're within 1 second (JSON precision limitation)
	diff := unmarshaled.CreatedAt.Sub(fileInfo.CreatedAt)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("CreatedAt mismatch: expected %v, got %v (diff: %v)", fileInfo.CreatedAt, unmarshaled.CreatedAt, diff)
	}

	// Optional fields should be zero values
	if !unmarshaled.ExpiresAt.IsZero() {
		t.Errorf("ExpiresAt should be zero value, got %v", unmarshaled.ExpiresAt)
	}

	if unmarshaled.ContentType != "" {
		t.Errorf("ContentType should be empty string, got '%s'", unmarshaled.ContentType)
	}
}
