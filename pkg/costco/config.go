package costco

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Configuration and token persistence

const (
	configDir  = ".costco"
	configFile = "config.json"
	tokenFile  = "tokens.json"
)

func getConfigPath() (string, error) {
	// Allow overriding config path for testing
	if testPath := os.Getenv("COSTCO_TEST_CONFIG_PATH"); testPath != "" {
		return testPath, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir), nil
}

func ensureConfigDir() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	return os.MkdirAll(configPath, 0700) // Only user can read/write
}

// SaveConfig persists user configuration to disk at ~/.costco/config.json.
// The config file stores non-sensitive settings like email and warehouse number.
// The file is created with 0600 permissions (user read/write only).
//
// Example:
//
//	config := &costco.StoredConfig{
//	    Email:           "user@example.com",
//	    WarehouseNumber: "847",
//	}
//	err := costco.SaveConfig(config)
func SaveConfig(config *StoredConfig) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(configPath, configFile)
	return os.WriteFile(filePath, data, 0600) // Only user can read/write
}

// LoadConfig loads user configuration from ~/.costco/config.json.
// Returns nil if the config file doesn't exist (not an error).
// Returns an error only if the file exists but cannot be read or parsed.
//
// Example:
//
//	config, err := costco.LoadConfig()
//	if err != nil {
//	    return err
//	}
//	if config != nil {
//	    fmt.Printf("Email: %s\n", config.Email)
//	}
func LoadConfig() (*StoredConfig, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(configPath, configFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No config file yet
		}
		return nil, err
	}

	var config StoredConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveTokens persists authentication tokens to disk at ~/.costco/tokens.json.
// The token file is created with 0600 permissions (user read/write only) for security.
// This function is automatically called by the client after successful authentication
// or token refresh. The UpdatedAt timestamp is automatically set to the current time.
//
// Example:
//
//	tokens := &costco.StoredTokens{
//	    IDToken:      "eyJhbGc...",
//	    RefreshToken: "1//0g...",
//	    TokenExpiry:  time.Now().Add(1 * time.Hour),
//	}
//	err := costco.SaveTokens(tokens)
func SaveTokens(tokens *StoredTokens) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	tokens.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(configPath, tokenFile)
	return os.WriteFile(filePath, data, 0600) // Only user can read/write
}

// LoadTokens loads authentication tokens from ~/.costco/tokens.json.
// Returns nil if the token file doesn't exist (not an error).
// Returns an error only if the file exists but cannot be read or parsed.
// The client automatically calls this during initialization to restore saved tokens.
//
// Example:
//
//	tokens, err := costco.LoadTokens()
//	if err != nil {
//	    return err
//	}
//	if tokens != nil && time.Now().Before(tokens.TokenExpiry) {
//	    fmt.Println("Valid token found")
//	}
func LoadTokens() (*StoredTokens, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(configPath, tokenFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No tokens file yet
		}
		return nil, err
	}

	var tokens StoredTokens
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, err
	}

	return &tokens, nil
}

// ClearTokens removes the saved token file from ~/.costco/tokens.json.
// This is useful for forcing re-authentication or cleaning up after logout.
// Returns nil if the file doesn't exist (already cleared).
//
// Example:
//
//	err := costco.ClearTokens()
//	if err != nil {
//	    log.Printf("Failed to clear tokens: %v", err)
//	}
func ClearTokens() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	filePath := filepath.Join(configPath, tokenFile)
	err = os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// GetConfigInfo returns a human-readable summary of the current configuration state.
// This includes the config directory path, whether config and token files exist,
// token expiry status, and last update time. Useful for debugging and status checks.
//
// Example:
//
//	info := costco.GetConfigInfo()
//	fmt.Println(info)
//	// Output:
//	// Config directory: /Users/username/.costco
//	// Config file: /Users/username/.costco/config.json (exists)
//	// Token file: /Users/username/.costco/tokens.json (exists)
//	//   - Token valid until: 2025-10-20T15:30:00Z
//	//   - Last updated: 2025-10-20T14:30:00Z
func GetConfigInfo() string {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Sprintf("Error getting config path: %v", err)
	}

	info := fmt.Sprintf("Config directory: %s\n", configPath)

	// Check if config exists
	configFile := filepath.Join(configPath, configFile)
	if _, err := os.Stat(configFile); err == nil {
		info += fmt.Sprintf("Config file: %s (exists)\n", configFile)
	} else {
		info += fmt.Sprintf("Config file: %s (not found)\n", configFile)
	}

	// Check if tokens exist
	tokenFile := filepath.Join(configPath, tokenFile)
	if _, err := os.Stat(tokenFile); err == nil {
		info += fmt.Sprintf("Token file: %s (exists)\n", tokenFile)

		// Try to load and show token status
		if tokens, err := LoadTokens(); err == nil && tokens != nil {
			if time.Now().Before(tokens.TokenExpiry) {
				info += fmt.Sprintf("  - Token valid until: %s\n", tokens.TokenExpiry.Format(time.RFC3339))
			} else {
				info += "  - Token expired, will refresh\n"
			}
			info += fmt.Sprintf("  - Last updated: %s\n", tokens.UpdatedAt.Format(time.RFC3339))
		}
	} else {
		info += fmt.Sprintf("Token file: %s (not found)\n", tokenFile)
	}

	return info
}
