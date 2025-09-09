package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/costco-go/pkg/costco"
	"golang.org/x/term"
)

func main() {
	var (
		command         = flag.String("cmd", "", "Command: setup, info, orders, receipts, receipt-detail")
		startDate       = flag.String("start", "", "Start date (YYYY-MM-DD)")
		endDate         = flag.String("end", "", "End date (YYYY-MM-DD)")
		barcode         = flag.String("barcode", "", "Receipt barcode (for receipt-detail)")
		pageNumber      = flag.Int("page", 1, "Page number for orders")
		pageSize        = flag.Int("size", 10, "Page size for orders")
		outputJSON      = flag.Bool("json", false, "Output as JSON")
	)

	flag.Parse()

	// Handle setup and info commands first
	if *command == "setup" {
		if err := setupCredentials(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *command == "info" {
		fmt.Println(costco.GetConfigInfo())
		return
	}

	// Load stored config
	storedConfig, err := costco.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	
	if storedConfig == nil {
		log.Fatal("No configuration found. Run 'costco-cli -cmd setup' first")
	}

	// Check if we have tokens, if not we need password
	tokens, _ := costco.LoadTokens()
	password := ""
	
	if tokens == nil || time.Now().After(tokens.RefreshTokenExpiresAt) {
		// Need to authenticate - get password
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal("Failed to read password")
		}
		fmt.Println()
		password = string(passwordBytes)
	}

	// Default date range if not provided
	if *startDate == "" {
		*startDate = time.Now().AddDate(0, -3, 0).Format("2006-01-02")
	}
	if *endDate == "" {
		*endDate = time.Now().Format("2006-01-02")
	}

	config := costco.Config{
		Email:              storedConfig.Email,
		Password:           password,
		WarehouseNumber:    storedConfig.WarehouseNumber,
		TokenRefreshBuffer: 5 * time.Minute,
	}

	client := costco.NewClient(config)
	ctx := context.Background()

	switch *command {
	case "orders":
		getOrders(ctx, client, *startDate, *endDate, *pageNumber, *pageSize, *outputJSON)
	case "receipts":
		getReceipts(ctx, client, *startDate, *endDate, *outputJSON)
	case "receipt-detail":
		if *barcode == "" {
			log.Fatal("Barcode is required for receipt-detail command")
		}
		getReceiptDetail(ctx, client, *barcode, *outputJSON)
	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func getOrders(ctx context.Context, client *costco.Client, startDate, endDate string, pageNumber, pageSize int, outputJSON bool) {
	orders, err := client.GetOnlineOrders(ctx, startDate, endDate, pageNumber, pageSize)
	if err != nil {
		log.Fatalf("Error getting orders: %v", err)
	}

	if outputJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(orders); err != nil {
			log.Fatalf("Error encoding JSON: %v", err)
		}
		return
	}

	fmt.Printf("Online Orders (%s to %s)\n", startDate, endDate)
	fmt.Printf("Page %d of %d total records\n", pageNumber, orders.TotalNumberOfRecords)
	fmt.Println("=" + string(make([]byte, 80)))

	for _, order := range orders.BCOrders {
		fmt.Printf("\nOrder #%s\n", order.OrderNumber)
		fmt.Printf("  Date: %s\n", order.OrderPlacedDate)
		fmt.Printf("  Status: %s\n", order.Status)
		fmt.Printf("  Total: $%.2f\n", order.OrderTotal)
		fmt.Printf("  Warehouse: %s\n", order.WarehouseNumber)
		
		if len(order.OrderLineItems) > 0 {
			fmt.Printf("  Items: %d\n", len(order.OrderLineItems))
			for i, item := range order.OrderLineItems {
				if i < 3 {
					fmt.Printf("    - %s (Status: %s)\n", item.ItemDescription, item.Status)
				}
			}
			if len(order.OrderLineItems) > 3 {
				fmt.Printf("    ... and %d more items\n", len(order.OrderLineItems)-3)
			}
		}
	}
}

func getReceipts(ctx context.Context, client *costco.Client, startDate, endDate string, outputJSON bool) {
	// Convert date format for receipts API (M/DD/YYYY)
	startTime, _ := time.Parse("2006-01-02", startDate)
	endTime, _ := time.Parse("2006-01-02", endDate)
	startDateFormatted := fmt.Sprintf("%d/%02d/%d", startTime.Month(), startTime.Day(), startTime.Year())
	endDateFormatted := fmt.Sprintf("%d/%02d/%d", endTime.Month(), endTime.Day(), endTime.Year())

	receipts, err := client.GetReceipts(ctx, startDateFormatted, endDateFormatted, "all", "all")
	if err != nil {
		log.Fatalf("Error getting receipts: %v", err)
	}

	if outputJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(receipts); err != nil {
			log.Fatalf("Error encoding JSON: %v", err)
		}
		return
	}

	fmt.Printf("Receipts (%s to %s)\n", startDate, endDate)
	fmt.Printf("In-Warehouse: %d, Gas Station: %d, Car Wash: %d\n", 
		receipts.InWarehouse, receipts.GasStation, receipts.CarWash)
	fmt.Println("=" + string(make([]byte, 80)))

	for _, receipt := range receipts.Receipts {
		fmt.Printf("\n%s - %s\n", receipt.TransactionDateTime, receipt.ReceiptType)
		fmt.Printf("  Warehouse: %s\n", receipt.WarehouseName)
		fmt.Printf("  Barcode: %s\n", receipt.TransactionBarcode)
		fmt.Printf("  Total: $%.2f\n", receipt.Total)
		fmt.Printf("  Items: %d\n", receipt.TotalItemCount)
	}
}

func getReceiptDetail(ctx context.Context, client *costco.Client, barcode string, outputJSON bool) {
	receipt, err := client.GetReceiptDetail(ctx, barcode, "warehouse")
	if err != nil {
		log.Fatalf("Error getting receipt detail: %v", err)
	}

	if outputJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(receipt); err != nil {
			log.Fatalf("Error encoding JSON: %v", err)
		}
		return
	}

	fmt.Printf("Receipt Detail\n")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Printf("Date: %s\n", receipt.TransactionDateTime)
	fmt.Printf("Warehouse: %s (#%d)\n", receipt.WarehouseName, receipt.WarehouseNumber)
	fmt.Printf("Address: %s, %s, %s %s\n", 
		receipt.WarehouseAddress1, receipt.WarehouseCity, 
		receipt.WarehouseState, receipt.WarehousePostalCode)
	fmt.Printf("Barcode: %s\n", receipt.TransactionBarcode)
	fmt.Printf("Member: %s\n", receipt.MembershipNumber)
	fmt.Println()

	fmt.Println("Items:")
	for _, item := range receipt.ItemArray {
		fmt.Printf("  %s - %s %s\n", item.ItemNumber, item.ItemDescription01, item.ItemDescription02)
		if item.Unit > 1 {
			fmt.Printf("    Qty: %d @ $%.2f = $%.2f\n", item.Unit, item.ItemUnitPriceAmount, item.Amount)
		} else {
			fmt.Printf("    $%.2f\n", item.Amount)
		}
	}

	fmt.Println()
	fmt.Printf("Subtotal: $%.2f\n", receipt.SubTotal)
	fmt.Printf("Tax: $%.2f\n", receipt.Taxes)
	fmt.Printf("Total: $%.2f\n", receipt.Total)

	if len(receipt.TenderArray) > 0 {
		fmt.Println("\nPayment:")
		for _, tender := range receipt.TenderArray {
			fmt.Printf("  %s (%s): $%.2f\n", 
				tender.TenderDescription, tender.DisplayAccountNumber, tender.AmountTender)
		}
	}
}