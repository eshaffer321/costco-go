package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildImportTestJWT(exp int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"exp":%d}`, exp)))
	return header + "." + payload + ".fakesignature"
}

func tokenJSON(t *testing.T, exp int64) string {
	t.Helper()
	return fmt.Sprintf(`{"id_token":%q,"refresh_token":"refresh-abc","refresh_token_expires_in":7776000}`,
		buildImportTestJWT(exp))
}

func withTempConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("COSTCO_TEST_CONFIG_PATH", filepath.Join(dir, ".costco"))
}

func TestImportTokens_Success(t *testing.T) {
	withTempConfig(t)

	exp := time.Now().Add(15 * time.Minute).Unix()
	in := strings.NewReader(tokenJSON(t, exp))
	var out bytes.Buffer

	err := importTokens(in, &out)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "✓ Tokens saved")
	assert.Contains(t, out.String(), "ID token valid until")
	assert.Contains(t, out.String(), "Refresh token valid until")
}

func TestImportTokens_InvalidJSON(t *testing.T) {
	withTempConfig(t)

	in := strings.NewReader("not json at all")
	var out bytes.Buffer

	err := importTokens(in, &out)
	assert.ErrorContains(t, err, "parsing JSON")
}

func TestImportTokens_MissingIDToken(t *testing.T) {
	withTempConfig(t)

	in := strings.NewReader(`{"refresh_token":"abc","refresh_token_expires_in":7776000}`)
	var out bytes.Buffer

	err := importTokens(in, &out)
	assert.ErrorContains(t, err, "id_token")
}

func TestImportTokens_MissingRefreshToken(t *testing.T) {
	withTempConfig(t)

	exp := time.Now().Add(15 * time.Minute).Unix()
	json := fmt.Sprintf(`{"id_token":%q,"refresh_token_expires_in":7776000}`, buildImportTestJWT(exp))
	in := strings.NewReader(json)
	var out bytes.Buffer

	err := importTokens(in, &out)
	assert.ErrorContains(t, err, "refresh_token")
}

func TestImportTokens_WritesToDisk(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".costco")
	t.Setenv("COSTCO_TEST_CONFIG_PATH", configDir)

	exp := time.Now().Add(15 * time.Minute).Unix()
	in := strings.NewReader(tokenJSON(t, exp))
	var out bytes.Buffer

	require.NoError(t, importTokens(in, &out))
	_, err := os.Stat(filepath.Join(configDir, "tokens.json"))
	assert.NoError(t, err, "tokens.json should exist on disk after import")
}

func TestEditTokenJSON_ReturnsFileContentsWrittenByEditor(t *testing.T) {
	exp := time.Now().Add(15 * time.Minute).Unix()
	want := tokenJSON(t, exp)

	fakeEditor := func(path string) error {
		return os.WriteFile(path, []byte(want), 0o600)
	}

	got, err := editTokenJSON(fakeEditor)
	require.NoError(t, err)
	assert.Equal(t, want, string(got))
}

func TestEditTokenJSON_PropagatesEditorError(t *testing.T) {
	fakeEditor := func(path string) error {
		return fmt.Errorf("boom")
	}

	_, err := editTokenJSON(fakeEditor)
	assert.ErrorContains(t, err, "boom")
}

func TestRunImportTokensFromFile_Success(t *testing.T) {
	withTempConfig(t)

	exp := time.Now().Add(15 * time.Minute).Unix()
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	require.NoError(t, os.WriteFile(path, []byte(tokenJSON(t, exp)), 0o600))

	err := runImportTokensFromFile(path)
	require.NoError(t, err)

	_, statErr := os.Stat(filepath.Join(os.Getenv("COSTCO_TEST_CONFIG_PATH"), "tokens.json"))
	assert.NoError(t, statErr)
}

func TestRunImportTokensFromFile_MissingFile(t *testing.T) {
	withTempConfig(t)

	err := runImportTokensFromFile(filepath.Join(t.TempDir(), "does-not-exist.json"))
	assert.ErrorContains(t, err, "reading token file")
}
