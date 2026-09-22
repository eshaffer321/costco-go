package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/eshaffer321/costco-go/pkg/costco"
)

func printImportInstructions(out io.Writer) {
	fmt.Fprintln(out, "  How to get it:")
	fmt.Fprintln(out, "  1. Log in to costco.com in your browser")
	fmt.Fprintln(out, "  2. Open DevTools → Network → filter Fetch/XHR")
	fmt.Fprintln(out, "  3. Search for 'token' and select the token endpoint request")
	fmt.Fprintln(out, "  4. Copy the full Response body (JSON)")
	fmt.Fprintln(out, "  5. Paste it here")
	fmt.Fprintln(out)
}

func processTokenJSON(data []byte, out io.Writer) error {
	var resp costco.TokenResponse
	if err := json.Unmarshal(data, &resp); err != nil {
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

// importTokens reads token JSON from in (e.g. piped stdin) and imports it.
func importTokens(in io.Reader, out io.Writer) error {
	fmt.Fprintln(out, "Paste the JSON response from the Costco token endpoint, then press Ctrl+D:")
	fmt.Fprintln(out)
	printImportInstructions(out)

	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	return processTokenJSON(data, out)
}

// isInteractive reports whether f is attached to a terminal.
func isInteractive(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// editTokenJSON opens $EDITOR (falling back to vi) on a temp file so the
// user can paste the token JSON there, then returns the saved contents.
// Pasting large JSON directly into the terminal followed by Ctrl+D is
// unreliable across terminal emulators, so an editor is used instead.
func editTokenJSON(runEditor func(path string) error) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "costco-token-*.json")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	path := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(path)

	if runErr := runEditor(path); runErr != nil {
		return nil, runErr
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading edited file: %w", err)
	}
	return data, nil
}

func runEditorCmd(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running editor %q: %w", editor, err)
	}
	return nil
}

// runImportTokensInteractive prints instructions, waits for the user to
// press Enter (so a full-screen editor doesn't wipe them out immediately),
// then opens the editor and imports whatever was saved.
func runImportTokensInteractive(out io.Writer, waitIn io.Reader, edit func(func(string) error) ([]byte, error), runEditor func(string) error) error {
	fmt.Fprintln(out, "This will open your editor to paste the token JSON (set $EDITOR to choose one; defaults to vi).")
	fmt.Fprintln(out)
	printImportInstructions(out)
	fmt.Fprint(out, "Press Enter when you're ready to open the editor: ")
	bufio.NewReader(waitIn).ReadString('\n')

	data, err := edit(runEditor)
	if err != nil {
		return err
	}

	return processTokenJSON(data, out)
}

func runImportTokens() error {
	// If a file path was given, or stdin is piped (not a terminal), read
	// straight from stdin - this keeps scripted/non-interactive usage working.
	if !isInteractive(os.Stdin) {
		return importTokens(os.Stdin, os.Stdout)
	}

	return runImportTokensInteractive(os.Stdout, os.Stdin, editTokenJSON, runEditorCmd)
}

func runImportTokensFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading token file: %w", err)
	}
	return processTokenJSON(data, os.Stdout)
}
