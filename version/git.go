package version

import (
	"os"
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
		gitHash = getEnvWithDefault("GIT_HASH", "dev")
		gitTag = getEnvWithDefault("GIT_TAG", "")
	})
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}
