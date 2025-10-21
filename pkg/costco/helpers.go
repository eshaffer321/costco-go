package costco

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"
)

// Analytics helper methods for the Costco client

// GetAllTransactionItems fetches all receipts in a date range and retrieves full item details for each.
// This method combines GetReceipts and GetReceiptDetail to provide complete transaction data
// including all line items for each receipt.
//
// The startDate and endDate should be in YYYY-MM-DD format.
// Returns a slice of TransactionWithItems, each containing full receipt details and all items.
//
// Example:
//
//	transactions, err := client.GetAllTransactionItems(ctx, "2025-01-01", "2025-01-31")
//	for _, tx := range transactions {
//	    fmt.Printf("Transaction on %s: $%.2f (%d items)\n",
//	        tx.TransactionDate.Format("2006-01-02"), tx.Total, len(tx.Items))
//	}
func (c *Client) GetAllTransactionItems(ctx context.Context, startDate, endDate string) ([]TransactionWithItems, error) {
	c.getLogger().Info("fetching all transaction items",
		slog.String("start_date", startDate),
		slog.String("end_date", endDate))

	// First get all receipts
	receipts, err := c.GetReceipts(ctx, startDate, endDate, "all", "all")
	if err != nil {
		return nil, fmt.Errorf("getting receipts: %w", err)
	}

	var transactions []TransactionWithItems

	// For each receipt, get the full details
	for _, receipt := range receipts.Receipts {
		// Skip if no barcode
		if receipt.TransactionBarcode == "" {
			continue
		}

		// Determine document type based on receipt type
		documentType := "warehouse"
		if receipt.ReceiptType == "Gas Station" || receipt.DocumentType == "fuel" {
			documentType = "fuel"
		}

		// Get full receipt details including all items
		detail, err := c.GetReceiptDetail(ctx, receipt.TransactionBarcode, documentType)
		if err != nil {
			c.getLogger().Warn("failed to get receipt details",
				slog.String("barcode", receipt.TransactionBarcode),
				slog.String("document_type", documentType),
				slog.String("error", err.Error()))
			continue
		}

		// Parse the transaction date
		txDate, _ := time.Parse("2006-01-02T15:04:05", detail.TransactionDateTime)

		transaction := TransactionWithItems{
			TransactionBarcode: detail.TransactionBarcode,
			TransactionDate:    txDate,
			WarehouseName:      detail.WarehouseName,
			Total:              detail.Total,
			Items:              detail.ItemArray,
			MembershipNumber:   detail.MembershipNumber,
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetItemHistory retrieves the complete purchase history for a specific item number
// within the given date range. Returns a chronological list of all transactions
// where the item was purchased, including date, quantity, price, and receipt barcode.
//
// The startDate and endDate should be in YYYY-MM-DD format.
// The itemNumber is the Costco item identifier.
//
// Example:
//
//	history, err := client.GetItemHistory(ctx, "12345", "2025-01-01", "2025-12-31")
//	for _, purchase := range history {
//	    fmt.Printf("Bought %d units on %s for $%.2f\n",
//	        purchase.Quantity, purchase.Date, purchase.Price)
//	}
func (c *Client) GetItemHistory(ctx context.Context, itemNumber, startDate, endDate string) ([]ItemPurchase, error) {
	transactions, err := c.GetAllTransactionItems(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var history []ItemPurchase

	for _, tx := range transactions {
		for _, item := range tx.Items {
			if item.ItemNumber == itemNumber {
				history = append(history, ItemPurchase{
					Date:     tx.TransactionDate.Format("2006-01-02"),
					Quantity: item.Unit,
					Price:    item.Amount,
					Barcode:  tx.TransactionBarcode,
				})
			}
		}
	}

	return history, nil
}

// GetSpendingSummary calculates total spending and item counts by department.
// Returns a map keyed by department number, with spending statistics for each department.
//
// The startDate and endDate should be in YYYY-MM-DD format.
//
// Example:
//
//	summary, err := client.GetSpendingSummary(ctx, "2025-01-01", "2025-12-31")
//	for deptNum, stats := range summary {
//	    fmt.Printf("%s: $%.2f across %d items\n",
//	        stats.Department, stats.Total, stats.ItemCount)
//	}
func (c *Client) GetSpendingSummary(ctx context.Context, startDate, endDate string) (map[int]SpendingByDepartment, error) {
	transactions, err := c.GetAllTransactionItems(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	summary := make(map[int]SpendingByDepartment)

	for _, tx := range transactions {
		for _, item := range tx.Items {
			dept := item.ItemDepartmentNumber
			current := summary[dept]
			current.Department = fmt.Sprintf("Department %d", dept)
			current.Total += item.Amount
			current.ItemCount += item.Unit
			summary[dept] = current
		}
	}

	return summary, nil
}

// GetFrequentItems returns the most frequently purchased items within a date range,
// sorted by purchase frequency. Useful for identifying shopping patterns and favorite products.
//
// The startDate and endDate should be in YYYY-MM-DD format.
// The limit parameter controls the maximum number of items returned (0 = return all).
//
// Example:
//
//	// Get top 10 most frequently purchased items
//	items, err := client.GetFrequentItems(ctx, "2025-01-01", "2025-12-31", 10)
//	for i, item := range items {
//	    fmt.Printf("#%d: %s (bought %d times, spent $%.2f)\n",
//	        i+1, item.ItemDescription, item.PurchaseCount, item.TotalSpent)
//	}
func (c *Client) GetFrequentItems(ctx context.Context, startDate, endDate string, limit int) ([]FrequentItem, error) {
	transactions, err := c.GetAllTransactionItems(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	itemMap := make(map[string]*FrequentItem)

	for _, tx := range transactions {
		for _, item := range tx.Items {
			if stats, exists := itemMap[item.ItemNumber]; exists {
				stats.TotalQuantity += item.Unit
				stats.TotalSpent += item.Amount
				stats.PurchaseCount++
			} else {
				itemMap[item.ItemNumber] = &FrequentItem{
					ItemNumber:      item.ItemNumber,
					ItemDescription: item.ItemDescription01,
					TotalQuantity:   item.Unit,
					TotalSpent:      item.Amount,
					PurchaseCount:   1,
				}
			}
		}
	}

	// Convert map to slice for sorting
	items := make([]FrequentItem, 0, len(itemMap))
	for _, stats := range itemMap {
		items = append(items, *stats)
	}

	// Sort by purchase count (descending) using stdlib sort
	sort.Slice(items, func(i, j int) bool {
		return items[i].PurchaseCount > items[j].PurchaseCount
	})

	// Return only the requested limit
	if limit > 0 && limit < len(items) {
		return items[:limit], nil
	}

	return items, nil
}
