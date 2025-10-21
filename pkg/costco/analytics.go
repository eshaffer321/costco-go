package costco

import "time"

// Analytics-related types for purchase analysis and reporting

// TransactionWithItems represents a receipt with its full item details.
// This is used by GetAllTransactionItems to provide complete transaction data.
type TransactionWithItems struct {
	TransactionBarcode string
	TransactionDate    time.Time
	WarehouseName      string
	Total              float64
	Items              []ReceiptItem
	MembershipNumber   string
}

// ItemPurchase represents a single purchase instance of an item.
// This is returned by GetItemHistory to show when and how an item was bought.
type ItemPurchase struct {
	Date     string  // Purchase date in YYYY-MM-DD format
	Quantity int     // Number of units purchased
	Price    float64 // Total price for this purchase
	Barcode  string  // Receipt barcode for this transaction
}

// SpendingByDepartment represents spending statistics for a single department.
// This is returned by GetSpendingSummary, keyed by department number.
type SpendingByDepartment struct {
	Department string  // Department name (e.g., "Department 42")
	Total      float64 // Total spending in this department
	ItemCount  int     // Total number of items purchased in this department
}

// FrequentItem represents statistics for a frequently purchased item.
// This is returned by GetFrequentItems, sorted by purchase frequency.
type FrequentItem struct {
	ItemNumber      string  // Costco item number
	ItemDescription string  // Item name/description
	TotalQuantity   int     // Total units purchased across all transactions
	TotalSpent      float64 // Total amount spent on this item
	PurchaseCount   int     // Number of times this item was purchased
}
