package costco

import (
	"context"
	"fmt"
	"time"
)

// TransactionWithItems represents a receipt with its full item details
type TransactionWithItems struct {
	TransactionBarcode string
	TransactionDate    time.Time
	WarehouseName      string
	Total              float64
	Items              []ReceiptItem
	MembershipNumber   string
}

// GetAllTransactionItems fetches all receipts in a date range and retrieves full item details for each
func (c *Client) GetAllTransactionItems(ctx context.Context, startDate, endDate string) ([]TransactionWithItems, error) {
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
			// Log but continue with other receipts
			fmt.Printf("Warning: Could not get details for %s (type: %s): %v\n", receipt.TransactionBarcode, documentType, err)
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

// GetItemHistory returns all purchases of a specific item number
func (c *Client) GetItemHistory(ctx context.Context, itemNumber, startDate, endDate string) ([]struct {
	Date     string
	Quantity int
	Price    float64
	Barcode  string
}, error) {
	transactions, err := c.GetAllTransactionItems(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var history []struct {
		Date     string
		Quantity int
		Price    float64
		Barcode  string
	}

	for _, tx := range transactions {
		for _, item := range tx.Items {
			if item.ItemNumber == itemNumber {
				history = append(history, struct {
					Date     string
					Quantity int
					Price    float64
					Barcode  string
				}{
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

// GetSpendingSummary calculates total spending by category
func (c *Client) GetSpendingSummary(ctx context.Context, startDate, endDate string) (map[int]struct {
	Department string
	Total      float64
	ItemCount  int
}, error) {
	transactions, err := c.GetAllTransactionItems(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	summary := make(map[int]struct {
		Department string
		Total      float64
		ItemCount  int
	})

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

// GetFrequentItems returns the most frequently purchased items
func (c *Client) GetFrequentItems(ctx context.Context, startDate, endDate string, limit int) ([]struct {
	ItemNumber      string
	ItemDescription string
	TotalQuantity   int
	TotalSpent      float64
	PurchaseCount   int
}, error) {
	transactions, err := c.GetAllTransactionItems(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	type itemStats struct {
		ItemNumber      string
		ItemDescription string
		TotalQuantity   int
		TotalSpent      float64
		PurchaseCount   int
	}

	itemMap := make(map[string]*itemStats)

	for _, tx := range transactions {
		for _, item := range tx.Items {
			if stats, exists := itemMap[item.ItemNumber]; exists {
				stats.TotalQuantity += item.Unit
				stats.TotalSpent += item.Amount
				stats.PurchaseCount++
			} else {
				itemMap[item.ItemNumber] = &itemStats{
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
	var items []struct {
		ItemNumber      string
		ItemDescription string
		TotalQuantity   int
		TotalSpent      float64
		PurchaseCount   int
	}

	for _, stats := range itemMap {
		items = append(items, struct {
			ItemNumber      string
			ItemDescription string
			TotalQuantity   int
			TotalSpent      float64
			PurchaseCount   int
		}{
			ItemNumber:      stats.ItemNumber,
			ItemDescription: stats.ItemDescription,
			TotalQuantity:   stats.TotalQuantity,
			TotalSpent:      stats.TotalSpent,
			PurchaseCount:   stats.PurchaseCount,
		})
	}

	// Sort by purchase count (you could also sort by TotalQuantity or TotalSpent)
	// Simple bubble sort for demonstration
	for i := 0; i < len(items)-1; i++ {
		for j := 0; j < len(items)-i-1; j++ {
			if items[j].PurchaseCount < items[j+1].PurchaseCount {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}

	// Return only the requested limit
	if limit > 0 && limit < len(items) {
		return items[:limit], nil
	}

	return items, nil
}
