package costco

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadConfig(t *testing.T) {
	// Use a temporary test directory
	tempDir := t.TempDir()
	os.Setenv("COSTCO_TEST_CONFIG_PATH", tempDir)
	defer os.Unsetenv("COSTCO_TEST_CONFIG_PATH")

	// Test SaveConfig
	config := &StoredConfig{
		Email:           "test@example.com",
		WarehouseNumber: "847",
	}

	err := SaveConfig(config)
	require.NoError(t, err)

	// Verify file was created with correct permissions
	configPath := filepath.Join(tempDir, configFile)
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Test LoadConfig
	loadedConfig, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, loadedConfig)
	assert.Equal(t, "test@example.com", loadedConfig.Email)
	assert.Equal(t, "847", loadedConfig.WarehouseNumber)
}

func TestLoadConfig_NotExists(t *testing.T) {
	// Use a temporary test directory
	tempDir := t.TempDir()
	os.Setenv("COSTCO_TEST_CONFIG_PATH", tempDir)
	defer os.Unsetenv("COSTCO_TEST_CONFIG_PATH")

	// Load config when file doesn't exist
	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Nil(t, config)
}

func TestSaveAndLoadTokens(t *testing.T) {
	// Use a temporary test directory
	tempDir := t.TempDir()
	os.Setenv("COSTCO_TEST_CONFIG_PATH", tempDir)
	defer os.Unsetenv("COSTCO_TEST_CONFIG_PATH")

	// Test SaveTokens
	expiry := time.Now().Add(1 * time.Hour)
	refreshExpiry := time.Now().Add(30 * 24 * time.Hour)
	tokens := &StoredTokens{
		IDToken:                "test-id-token",
		RefreshToken:           "test-refresh-token",
		TokenExpiry:            expiry,
		RefreshTokenExpiresAt:  refreshExpiry,
	}

	err := SaveTokens(tokens)
	require.NoError(t, err)

	// Verify file was created with correct permissions
	tokenPath := filepath.Join(tempDir, tokenFile)
	info, err := os.Stat(tokenPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify UpdatedAt was set
	assert.False(t, tokens.UpdatedAt.IsZero())

	// Test LoadTokens
	loadedTokens, err := LoadTokens()
	require.NoError(t, err)
	require.NotNil(t, loadedTokens)
	assert.Equal(t, "test-id-token", loadedTokens.IDToken)
	assert.Equal(t, "test-refresh-token", loadedTokens.RefreshToken)
	assert.WithinDuration(t, expiry, loadedTokens.TokenExpiry, time.Second)
	assert.WithinDuration(t, refreshExpiry, loadedTokens.RefreshTokenExpiresAt, time.Second)
	assert.False(t, loadedTokens.UpdatedAt.IsZero())
}

func TestLoadTokens_NotExists(t *testing.T) {
	// Use a temporary test directory
	tempDir := t.TempDir()
	os.Setenv("COSTCO_TEST_CONFIG_PATH", tempDir)
	defer os.Unsetenv("COSTCO_TEST_CONFIG_PATH")

	// Load tokens when file doesn't exist
	tokens, err := LoadTokens()
	require.NoError(t, err)
	assert.Nil(t, tokens)
}

func TestClearTokens(t *testing.T) {
	// Use a temporary test directory
	tempDir := t.TempDir()
	os.Setenv("COSTCO_TEST_CONFIG_PATH", tempDir)
	defer os.Unsetenv("COSTCO_TEST_CONFIG_PATH")

	// First save some tokens
	tokens := &StoredTokens{
		IDToken:               "test-token",
		RefreshToken:          "test-refresh",
		TokenExpiry:           time.Now().Add(1 * time.Hour),
		RefreshTokenExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	err := SaveTokens(tokens)
	require.NoError(t, err)

	// Verify file exists
	tokenPath := filepath.Join(tempDir, tokenFile)
	_, err = os.Stat(tokenPath)
	require.NoError(t, err)

	// Clear tokens
	err = ClearTokens()
	require.NoError(t, err)

	// Verify file is deleted
	_, err = os.Stat(tokenPath)
	assert.True(t, os.IsNotExist(err))

	// Clearing again should not error
	err = ClearTokens()
	require.NoError(t, err)
}

func TestGetConfigInfo(t *testing.T) {
	// Use a temporary test directory
	tempDir := t.TempDir()
	os.Setenv("COSTCO_TEST_CONFIG_PATH", tempDir)
	defer os.Unsetenv("COSTCO_TEST_CONFIG_PATH")

	// Test with no files
	info := GetConfigInfo()
	assert.Contains(t, info, tempDir)
	assert.Contains(t, info, "not found")

	// Create config file
	config := &StoredConfig{
		Email:           "test@example.com",
		WarehouseNumber: "847",
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	// Test with config file only
	info = GetConfigInfo()
	assert.Contains(t, info, tempDir)
	assert.Contains(t, info, "config.json (exists)")
	assert.Contains(t, info, "tokens.json (not found)")

	// Create token file with valid token
	tokens := &StoredTokens{
		IDToken:               "test-token",
		RefreshToken:          "test-refresh",
		TokenExpiry:           time.Now().Add(1 * time.Hour),
		RefreshTokenExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	err = SaveTokens(tokens)
	require.NoError(t, err)

	// Test with both files
	info = GetConfigInfo()
	assert.Contains(t, info, tempDir)
	assert.Contains(t, info, "config.json (exists)")
	assert.Contains(t, info, "tokens.json (exists)")
	assert.Contains(t, info, "Token valid until:")
	assert.Contains(t, info, "Last updated:")

	// Create expired token
	expiredTokens := &StoredTokens{
		IDToken:               "expired-token",
		RefreshToken:          "test-refresh",
		TokenExpiry:           time.Now().Add(-1 * time.Hour), // Expired
		RefreshTokenExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	err = SaveTokens(expiredTokens)
	require.NoError(t, err)

	// Test with expired token
	info = GetConfigInfo()
	assert.Contains(t, info, "Token expired, will refresh")
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Use a temporary test directory
	tempDir := t.TempDir()
	os.Setenv("COSTCO_TEST_CONFIG_PATH", tempDir)
	defer os.Unsetenv("COSTCO_TEST_CONFIG_PATH")

	// Create config dir and write invalid JSON
	err := ensureConfigDir()
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, configFile)
	err = os.WriteFile(configPath, []byte("invalid json"), 0600)
	require.NoError(t, err)

	// Try to load - should error
	config, err := LoadConfig()
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestLoadTokens_InvalidJSON(t *testing.T) {
	// Use a temporary test directory
	tempDir := t.TempDir()
	os.Setenv("COSTCO_TEST_CONFIG_PATH", tempDir)
	defer os.Unsetenv("COSTCO_TEST_CONFIG_PATH")

	// Create config dir and write invalid JSON
	err := ensureConfigDir()
	require.NoError(t, err)

	tokenPath := filepath.Join(tempDir, tokenFile)
	err = os.WriteFile(tokenPath, []byte("invalid json"), 0600)
	require.NoError(t, err)

	// Try to load - should error
	tokens, err := LoadTokens()
	assert.Error(t, err)
	assert.Nil(t, tokens)
}
