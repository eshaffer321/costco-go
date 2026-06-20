package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
	fmt.Println("\nSetup complete! Next, run:")
	fmt.Println("  costco-cli -cmd import-token")
	fmt.Println("\nThen log in to costco.com in your browser and paste the OAuth token response.")

	return nil
}
