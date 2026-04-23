package costco

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildTestJWT creates a minimal unsigned JWT with the given exp claim.
// ParseUnverified doesn't check signatures so this is sufficient for tests.
func buildTestJWT(exp int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"exp":%d}`, exp)))
	return header + "." + payload + ".fakesignature"
}

func TestImportTokenResponse_SetsTokenExpiry(t *testing.T) {
	exp := time.Now().Add(15 * time.Minute).Unix()
	resp := &TokenResponse{
		IDToken:               buildTestJWT(exp),
		RefreshToken:          "refresh-token-value",
		RefreshTokenExpiresIn: 7776000, // 90 days
	}

	tokens, err := ImportTokenResponse(resp)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Unix(exp, 0), tokens.TokenExpiry, time.Second)
}

func TestImportTokenResponse_SetsRefreshTokenExpiry(t *testing.T) {
	resp := &TokenResponse{
		IDToken:               buildTestJWT(time.Now().Add(15 * time.Minute).Unix()),
		RefreshToken:          "refresh-token-value",
		RefreshTokenExpiresIn: 7776000,
	}

	before := time.Now()
	tokens, err := ImportTokenResponse(resp)
	require.NoError(t, err)

	expectedExpiry := before.Add(7776000 * time.Second)
	assert.WithinDuration(t, expectedExpiry, tokens.RefreshTokenExpiresAt, 2*time.Second)
}

func TestImportTokenResponse_CopiesTokenValues(t *testing.T) {
	resp := &TokenResponse{
		IDToken:               buildTestJWT(time.Now().Add(15 * time.Minute).Unix()),
		RefreshToken:          "my-refresh-token",
		RefreshTokenExpiresIn: 7776000,
	}

	tokens, err := ImportTokenResponse(resp)
	require.NoError(t, err)
	assert.Equal(t, resp.IDToken, tokens.IDToken)
	assert.Equal(t, resp.RefreshToken, tokens.RefreshToken)
}

func TestImportTokenResponse_MissingIDToken(t *testing.T) {
	resp := &TokenResponse{
		RefreshToken:          "my-refresh-token",
		RefreshTokenExpiresIn: 7776000,
	}

	_, err := ImportTokenResponse(resp)
	assert.ErrorContains(t, err, "id_token")
}

func TestImportTokenResponse_MissingRefreshToken(t *testing.T) {
	resp := &TokenResponse{
		IDToken:               buildTestJWT(time.Now().Add(15 * time.Minute).Unix()),
		RefreshTokenExpiresIn: 7776000,
	}

	_, err := ImportTokenResponse(resp)
	assert.ErrorContains(t, err, "refresh_token")
}

func TestImportTokenResponse_MalformedJWTFallsBackToDefault(t *testing.T) {
	resp := &TokenResponse{
		IDToken:               "not.a.jwt",
		RefreshToken:          "my-refresh-token",
		RefreshTokenExpiresIn: 7776000,
	}

	tokens, err := ImportTokenResponse(resp)
	require.NoError(t, err)
	// Should fall back to a short default expiry, not zero time
	assert.True(t, tokens.TokenExpiry.After(time.Now()))
}
