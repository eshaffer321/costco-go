package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/eshaffer321/costco-go/pkg/costco"
)

func setupCredentials() error {
	reader := bufio.NewReader(os.Stdin)

	// Load existing config if any
	existingConfig, _ := costco.LoadConfig()

	fmt.Println("Costco CLI Setup")
	fmt.Println("================")
	fmt.Println("Your credentials will be stored in ~/.costco/")
	fmt.Println()

	// Get email
	defaultEmail := ""
	if existingConfig != nil && existingConfig.Email != "" {
		defaultEmail = existingConfig.Email
		fmt.Printf("Email [%s]: ", defaultEmail)
	} else {
		fmt.Print("Email: ")
	}

	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	if email == "" && defaultEmail != "" {
		email = defaultEmail
	}

	// Get warehouse
	defaultWarehouse := "847"
	if existingConfig != nil && existingConfig.WarehouseNumber != "" {
		defaultWarehouse = existingConfig.WarehouseNumber
		fmt.Printf("Warehouse Number [%s]: ", defaultWarehouse)
	} else {
		fmt.Printf("Warehouse Number [%s]: ", defaultWarehouse)
	}

	warehouse, _ := reader.ReadString('\n')
	warehouse = strings.TrimSpace(warehouse)
	if warehouse == "" {
		warehouse = defaultWarehouse
	}

	// Save config
	config := &costco.StoredConfig{
		Email:           email,
		WarehouseNumber: warehouse,
	}

	if err := costco.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("\n✓ Configuration saved to ~/.costco/config.json")

	// Ask if they want to authenticate now
	fmt.Print("\nDo you want to authenticate now? (y/n): ")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "y" || answer == "yes" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println()

		password := string(passwordBytes)

		// Create client and authenticate
		client := costco.NewClient(costco.Config{
			Email:           email,
			Password:        password,
			WarehouseNumber: warehouse,
		})

		fmt.Print("Authenticating...")

		// Force authentication by making a simple request
		ctx := context.Background()
		_, err = client.GetReceipts(ctx, "1/01/2025", "1/09/2025", "all", "all")
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		fmt.Println(" ✓")
		fmt.Println("✓ Authentication successful! Tokens saved to ~/.costco/tokens.json")
	}

	fmt.Println("\nSetup complete! You can now use the CLI commands.")
	fmt.Println("\nExample commands:")
	fmt.Println("  costco-cli orders      - Get recent orders")
	fmt.Println("  costco-cli receipts    - Get recent receipts")
	fmt.Println("  costco-cli info        - Show config info")

	return nil
}
