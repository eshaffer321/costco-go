package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/eshaffer321/costco-go/pkg/costco"
)

func importTokens(in io.Reader, out io.Writer) error {
	fmt.Fprintln(out, "Paste the JSON response from the Costco token endpoint, then press Ctrl+D:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  How to get it:")
	fmt.Fprintln(out, "  1. Log in to costco.com in your browser")
	fmt.Fprintln(out, "  2. Open DevTools → Network → filter Fetch/XHR")
	fmt.Fprintln(out, "  3. Search for 'token' and select the token endpoint request")
	fmt.Fprintln(out, "  4. Copy the full Response body (JSON)")
	fmt.Fprintln(out, "  5. Paste it here")
	fmt.Fprintln(out)

	data, err := io.ReadAll(in)
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

	fmt.Fprintln(out, "✓ Tokens saved to ~/.costco/tokens.json")
	fmt.Fprintf(out, "  ID token valid until:      %s\n", tokens.TokenExpiry.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(out, "  Refresh token valid until: %s\n", tokens.RefreshTokenExpiresAt.Format("2006-01-02 15:04:05 MST"))
	return nil
}

func runImportTokens() error {
	return importTokens(os.Stdin, os.Stdout)
}
