package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/eshaffer321/costco-go/pkg/costco"
)

func importTokens() error {
	fmt.Println("Paste the JSON response from the Costco token endpoint, then press Ctrl+D:")
	fmt.Println()
	fmt.Println("  How to get it:")
	fmt.Println("  1. Log in to costco.com in your browser")
	fmt.Println("  2. Open DevTools → Network → filter Fetch/XHR")
	fmt.Println("  3. Search for 'token' and select the token endpoint request")
	fmt.Println("  4. Copy the full Response body (JSON)")
	fmt.Println("  5. Paste it here")
	fmt.Println()

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	var resp costco.TokenResponse
	if err = json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("parsing JSON: %w\n\nMake sure you copied the Response body (not the Headers)", err)
	}

	tokens, err := costco.ImportTokenResponse(&resp)
	if err != nil {
		return err
	}

	if err = costco.SaveTokens(tokens); err != nil {
		return fmt.Errorf("saving tokens: %w", err)
	}

	fmt.Println("✓ Tokens saved to ~/.costco/tokens.json")
	fmt.Printf("  ID token valid until:      %s\n", tokens.TokenExpiry.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("  Refresh token valid until: %s\n", tokens.RefreshTokenExpiresAt.Format("2006-01-02 15:04:05 MST"))
	return nil
}
