package costco

import (
	"log/slog"
	"time"
)

// Configuration and options for the Costco client

// Config holds the configuration for creating a new Costco client.
// Email and Password are required for authentication.
// WarehouseNumber defaults to "847" if not provided.
// TokenRefreshBuffer controls how early tokens are refreshed (default: 5 minutes before expiry).
// Logger is optional - if nil, all logs are silently discarded.
type Config struct {
	Email              string          // Costco account email (required)
	Password           string          // Costco account password (required)
	WarehouseNumber    string          // Default warehouse number (default: "847")
	TokenRefreshBuffer time.Duration   // How early to refresh tokens before expiry (default: 5min)
	Logger             *slog.Logger    // Optional structured logger (nil = silent)
}

// StoredConfig represents user configuration persisted to disk.
// This is saved to ~/.costco/config.json and contains non-sensitive settings.
type StoredConfig struct {
	Email           string `json:"email"`
	WarehouseNumber string `json:"warehouse_number"`
}

// StoredTokens represents authentication tokens persisted to disk.
// This is saved to ~/.costco/tokens.json with 0600 permissions (user read/write only).
// Tokens are automatically loaded on client creation and refreshed as needed.
type StoredTokens struct {
	IDToken               string    `json:"id_token"`
	RefreshToken          string    `json:"refresh_token"`
	TokenExpiry           time.Time `json:"token_expiry"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// TokenResponse represents the OAuth2 token response from Costco's authentication endpoint.
// This is returned during initial authentication and token refresh operations.
type TokenResponse struct {
	IDToken               string `json:"id_token"`
	TokenType             string `json:"token_type"`
	NotBefore             int64  `json:"not_before"`
	ClientInfo            string `json:"client_info"`
	Scope                 string `json:"scope"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
}
