package costco

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	configDir  = ".costco"
	configFile = "config.json"
	tokenFile  = "tokens.json"
)

type StoredConfig struct {
	Email           string `json:"email"`
	WarehouseNumber string `json:"warehouse_number"`
}

type StoredTokens struct {
	IDToken               string    `json:"id_token"`
	RefreshToken          string    `json:"refresh_token"`
	TokenExpiry           time.Time `json:"token_expiry"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

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
