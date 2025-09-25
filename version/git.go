package version

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	gitHash string
	gitTag  string
	once    sync.Once
)

// GitHash returns the git hash, loading from environment if needed
func GitHash() string {
	ensureLoaded()
	return gitHash
}

// GitTag returns the git tag, loading from environment if needed
func GitTag() string {
	ensureLoaded()
	return gitTag
}

func ensureLoaded() {
	once.Do(func() {
		// Try to load .env from executable directory if not already loaded
		loadEnvFile()
		gitHash = getEnvWithDefault("GIT_HASH", "dev")
		gitTag = getEnvWithDefault("GIT_TAG", "")
	})
}

func loadEnvFile() {
	// If environment variables are not set, try loading from executable directory
	if os.Getenv("GIT_HASH") == "" || os.Getenv("GIT_TAG") == "" {
		if execPath, err := os.Executable(); err == nil {
			execDir := filepath.Dir(execPath)
			envFile := filepath.Join(execDir, ".env")
			if data, err := os.ReadFile(envFile); err == nil {
				lines := strings.Split(string(data), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}
					if idx := strings.Index(line, "="); idx > 0 {
						key := strings.TrimSpace(line[:idx])
						value := strings.TrimSpace(line[idx+1:])
						os.Setenv(key, value)
					}
				}
			}
		}
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}
