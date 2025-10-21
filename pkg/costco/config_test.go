package costco

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigPathOverride(t *testing.T) {
	// Test that environment variable overrides the default path
	cleanup := SetupTestConfig(t)
	defer cleanup()

	// Get the config path - should be the temp directory
	configPath, err := getConfigPath()
	require.NoError(t, err)

	// Should not be the user's home directory
	home, _ := os.UserHomeDir()
	assert.NotEqual(t, filepath.Join(home, configDir), configPath)

	// Should be the test directory
	assert.Contains(t, configPath, "costco-test")

	// Test that we can save and load tokens without affecting real config
	testTokens := &StoredTokens{
		IDToken:      "test-token",
		RefreshToken: "test-refresh",
	}

	err = SaveTokens(testTokens)
	require.NoError(t, err)

	// Verify the file was created in the test directory
	tokenPath := filepath.Join(configPath, tokenFile)
	_, err = os.Stat(tokenPath)
	assert.NoError(t, err)

	// Load the tokens back
	loadedTokens, err := LoadTokens()
	require.NoError(t, err)
	assert.Equal(t, "test-token", loadedTokens.IDToken)
	assert.Equal(t, "test-refresh", loadedTokens.RefreshToken)

	// Ensure real home directory wasn't touched by checking if the test file is different
	realTokenPath := filepath.Join(home, configDir, tokenFile)
	if realTokens, err := os.ReadFile(realTokenPath); err == nil {
		// If real token file exists, verify it wasn't modified (it should be different from test tokens)
		testTokens, _ := os.ReadFile(tokenPath)
		assert.NotEqual(t, string(realTokens), string(testTokens), "Real token file should not have been modified by test")
	}
	// If real token file doesn't exist, that's fine too - just means user hasn't run the CLI yet
}

func TestConfigPathDefault(t *testing.T) {
	// Test that without the environment variable, we get the default path
	// Save current env var if it exists
	oldValue := os.Getenv("COSTCO_TEST_CONFIG_PATH")
	os.Unsetenv("COSTCO_TEST_CONFIG_PATH")
	defer func() {
		if oldValue != "" {
			os.Setenv("COSTCO_TEST_CONFIG_PATH", oldValue)
		}
	}()

	configPath, err := getConfigPath()
	require.NoError(t, err)

	home, _ := os.UserHomeDir()
	expectedPath := filepath.Join(home, configDir)
	assert.Equal(t, expectedPath, configPath)
}
