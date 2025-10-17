package costco

import (
	"os"
	"path/filepath"
	"testing"
)

// SetupTestConfig creates a temporary directory for test configuration files
// and sets the COSTCO_TEST_CONFIG_PATH environment variable.
// It returns a cleanup function that should be deferred.
func SetupTestConfig(t *testing.T) func() {
	t.Helper()
	
	// Create a temporary directory for test config
	tempDir, err := os.MkdirTemp("", "costco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	// Create the .costco subdirectory
	configPath := filepath.Join(tempDir, ".costco")
	if err := os.MkdirAll(configPath, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	// Set the environment variable
	originalPath := os.Getenv("COSTCO_TEST_CONFIG_PATH")
	os.Setenv("COSTCO_TEST_CONFIG_PATH", configPath)
	
	// Return cleanup function
	return func() {
		// Restore original environment variable
		if originalPath == "" {
			os.Unsetenv("COSTCO_TEST_CONFIG_PATH")
		} else {
			os.Setenv("COSTCO_TEST_CONFIG_PATH", originalPath)
		}
		
		// Remove temp directory
		os.RemoveAll(tempDir)
	}
}