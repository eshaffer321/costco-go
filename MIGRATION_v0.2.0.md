# Migration Guide: v0.1.1 → v0.2.0

**Target Audience:** AI agents and developers consuming costco-go library
**Migration Difficulty:** Low (mostly type renames)
**Breaking Changes:** Yes (type names and locations changed)
**Estimated Migration Time:** 5-10 minutes

---

## Overview

Version 0.2.0 is a **major architectural refactor** focused on code organization and maintainability. While the public API surface remains largely the same, types have been reorganized into logical domain files.

**Key Changes:**
- Types reorganized into domain-specific files (`orders.go`, `receipts.go`, `analytics.go`)
- Anonymous struct return types replaced with named types
- Client interface added for better testability
- Internal packages created for better code organization
- Bubble sort replaced with stdlib `sort.Slice`
- All exports now have proper godoc comments

---

## Quick Migration Checklist

- [ ] Update import statements (import path unchanged)
- [ ] Replace anonymous struct handling with new named types
- [ ] Update code using spending summary return type (now `SpendingByDepartment`)
- [ ] Update code using frequent items return type (now `FrequentItem`)
- [ ] Update code using item history return type (now `ItemPurchase`)
- [ ] Optional: Use new `CostcoClient` interface for mocking
- [ ] Run tests to verify everything works

---

## What Stayed the Same (No Changes Needed)

### ✅ Import Path
```go
// Still the same
import "github.com/costco-go/pkg/costco"
```

### ✅ Client Creation
```go
// Still the same
config := costco.Config{
    Email:              "user@example.com",
    Password:           "password",
    WarehouseNumber:    "847",
    TokenRefreshBuffer: 5 * time.Minute,
    Logger:             logger,
}
client := costco.NewClient(config)
```

### ✅ Core API Methods
```go
// All of these are UNCHANGED
orders, err := client.GetOnlineOrders(ctx, "2025-01-01", "2025-01-31", 1, 10)
receipts, err := client.GetReceipts(ctx, "1/01/2025", "1/31/2025", "all", "all")
receipt, err := client.GetReceiptDetail(ctx, "barcode", "warehouse")
transactions, err := client.GetAllTransactionItems(ctx, "2025-01-01", "2025-01-31")
```

### ✅ Main Domain Types
```go
// All of these are UNCHANGED
costco.OnlineOrder
costco.OnlineOrdersResponse
costco.Receipt
costco.ReceiptsWithCountsResponse
costco.ReceiptItem
costco.OrderLineItem
costco.Shipment
costco.Tender
costco.SubTaxes
costco.TransactionWithItems
costco.Config
costco.StoredConfig
costco.StoredTokens
```

---

## What Changed (Action Required)

### 🔄 Change #1: Item Purchase History

**Old (v0.1.1):**
```go
// Anonymous struct return type
history, err := client.GetItemHistory(ctx, "12345", "2025-01-01", "2025-01-31")
// Type: []struct { Date string; Quantity int; Price float64; Barcode string }

for _, purchase := range history {
    fmt.Printf("Date: %s, Qty: %d, Price: %.2f\n",
        purchase.Date, purchase.Quantity, purchase.Price)
}
```

**New (v0.2.0):**
```go
// Named type: ItemPurchase
history, err := client.GetItemHistory(ctx, "12345", "2025-01-01", "2025-01-31")
// Type: []costco.ItemPurchase

for _, purchase := range history {
    fmt.Printf("Date: %s, Qty: %d, Price: %.2f\n",
        purchase.Date, purchase.Quantity, purchase.Price)
}

// Now you can reference the type elsewhere:
func processHistory(purchases []costco.ItemPurchase) { ... }
```

**Type Definition:**
```go
type ItemPurchase struct {
    Date     string  // Purchase date in YYYY-MM-DD format
    Quantity int     // Number of units purchased
    Price    float64 // Total price for this purchase
    Barcode  string  // Receipt barcode for this transaction
}
```

### 🔄 Change #2: Spending Summary

**Old (v0.1.1):**
```go
// Anonymous struct in map value
summary, err := client.GetSpendingSummary(ctx, "2025-01-01", "2025-01-31")
// Type: map[int]struct { Department string; Total float64; ItemCount int }

for deptNum, info := range summary {
    fmt.Printf("Dept %d: $%.2f (%d items)\n",
        deptNum, info.Total, info.ItemCount)
}
```

**New (v0.2.0):**
```go
// Named type: SpendingByDepartment
summary, err := client.GetSpendingSummary(ctx, "2025-01-01", "2025-01-31")
// Type: map[int]costco.SpendingByDepartment

for deptNum, info := range summary {
    fmt.Printf("Dept %d (%s): $%.2f (%d items)\n",
        deptNum, info.Department, info.Total, info.ItemCount)
}

// Now you can reference the type elsewhere:
func analyzeDepartment(dept costco.SpendingByDepartment) { ... }
```

**Type Definition:**
```go
type SpendingByDepartment struct {
    Department string  // Department name (e.g., "Department 42")
    Total      float64 // Total spending in this department
    ItemCount  int     // Total number of items purchased
}
```

### 🔄 Change #3: Frequent Items

**Old (v0.1.1):**
```go
// Anonymous struct return type
items, err := client.GetFrequentItems(ctx, "2025-01-01", "2025-01-31", 10)
// Type: []struct { ItemNumber string; ItemDescription string; TotalQuantity int; TotalSpent float64; PurchaseCount int }

for _, item := range items {
    fmt.Printf("%s: bought %d times\n",
        item.ItemDescription, item.PurchaseCount)
}
```

**New (v0.2.0):**
```go
// Named type: FrequentItem
items, err := client.GetFrequentItems(ctx, "2025-01-01", "2025-01-31", 10)
// Type: []costco.FrequentItem

for _, item := range items {
    fmt.Printf("%s: bought %d times, total: $%.2f\n",
        item.ItemDescription, item.PurchaseCount, item.TotalSpent)
}

// Now you can reference the type elsewhere:
func rankItems(items []costco.FrequentItem) { ... }
```

**Type Definition:**
```go
type FrequentItem struct {
    ItemNumber      string  // Costco item number
    ItemDescription string  // Item name/description
    TotalQuantity   int     // Total units purchased across all transactions
    TotalSpent      float64 // Total amount spent on this item
    PurchaseCount   int     // Number of times this item was purchased
}
```

---

## New Features in v0.2.0

### 🎉 Feature #1: Client Interface

You can now use the `CostcoClient` interface for mocking and testing:

```go
// Define the interface
type CostcoClient interface {
    GetOnlineOrders(ctx context.Context, startDate, endDate string, pageNumber, pageSize int) (*OnlineOrdersResponse, error)
    GetReceipts(ctx context.Context, startDate, endDate, documentType, documentSubType string) (*ReceiptsWithCountsResponse, error)
    GetReceiptDetail(ctx context.Context, barcode, documentType string) (*Receipt, error)
    GetAllTransactionItems(ctx context.Context, startDate, endDate string) ([]TransactionWithItems, error)
    GetItemHistory(ctx context.Context, itemNumber, startDate, endDate string) ([]ItemPurchase, error)
    GetSpendingSummary(ctx context.Context, startDate, endDate string) (map[int]SpendingByDepartment, error)
    GetFrequentItems(ctx context.Context, startDate, endDate string, limit int) ([]FrequentItem, error)
}

// Use it in your code
type MyService struct {
    costco costco.CostcoClient // Now mockable!
}

// Mock it for testing
type MockCostcoClient struct {
    mock.Mock
}

func (m *MockCostcoClient) GetOnlineOrders(ctx context.Context, startDate, endDate string, pageNumber, pageSize int) (*costco.OnlineOrdersResponse, error) {
    args := m.Called(ctx, startDate, endDate, pageNumber, pageSize)
    return args.Get(0).(*costco.OnlineOrdersResponse), args.Error(1)
}
```

### 🎉 Feature #2: Better Godoc

All exported types and functions now have comprehensive godoc comments:

```go
// GetItemHistory retrieves the complete purchase history for a specific item number
// within the given date range. Returns a chronological list of all transactions
// where the item was purchased, including date, quantity, price, and receipt barcode.
//
// The startDate and endDate should be in YYYY-MM-DD format.
// The itemNumber is the Costco item identifier.
//
// Example:
//   history, err := client.GetItemHistory(ctx, "12345", "2025-01-01", "2025-12-31")
//   for _, purchase := range history {
//       fmt.Printf("Bought %d units on %s for $%.2f\n",
//           purchase.Quantity, purchase.Date, purchase.Price)
//   }
func (c *Client) GetItemHistory(ctx context.Context, itemNumber, startDate, endDate string) ([]ItemPurchase, error)
```

---

## Migration Examples

### Example 1: AI Agent Processing Item History

**Before (v0.1.1):**
```go
package main

import (
    "context"
    "fmt"
    "github.com/costco-go/pkg/costco"
)

func analyzeItemPurchases(client *costco.Client, itemNumber string) error {
    history, err := client.GetItemHistory(context.Background(),
        itemNumber, "2025-01-01", "2025-12-31")
    if err != nil {
        return err
    }

    // Had to work with anonymous struct
    for _, purchase := range history {
        fmt.Printf("%s: %d @ $%.2f\n",
            purchase.Date, purchase.Quantity, purchase.Price)
    }

    return nil
}
```

**After (v0.2.0):**
```go
package main

import (
    "context"
    "fmt"
    "github.com/costco-go/pkg/costco"
)

func analyzeItemPurchases(client costco.CostcoClient, itemNumber string) error {
    history, err := client.GetItemHistory(context.Background(),
        itemNumber, "2025-01-01", "2025-12-31")
    if err != nil {
        return err
    }

    // Can now pass to helper functions
    printPurchases(history)
    calculateStats(history)

    return nil
}

// NEW: Can create helper functions with typed parameters
func printPurchases(purchases []costco.ItemPurchase) {
    for _, p := range purchases {
        fmt.Printf("%s: %d @ $%.2f\n", p.Date, p.Quantity, p.Price)
    }
}

func calculateStats(purchases []costco.ItemPurchase) {
    total := 0.0
    for _, p := range purchases {
        total += p.Price
    }
    fmt.Printf("Total spent: $%.2f\n", total)
}
```

### Example 2: AI Agent Analyzing Spending

**Before (v0.1.1):**
```go
func analyzeDepartmentSpending(client *costco.Client) error {
    summary, err := client.GetSpendingSummary(context.Background(),
        "2025-01-01", "2025-12-31")
    if err != nil {
        return err
    }

    // Could not extract to helper function easily
    maxSpend := 0.0
    maxDept := 0
    for dept, info := range summary {
        if info.Total > maxSpend {
            maxSpend = info.Total
            maxDept = dept
        }
    }

    fmt.Printf("Biggest spending: Dept %d with $%.2f\n", maxDept, maxSpend)
    return nil
}
```

**After (v0.2.0):**
```go
func analyzeDepartmentSpending(client costco.CostcoClient) error {
    summary, err := client.GetSpendingSummary(context.Background(),
        "2025-01-01", "2025-12-31")
    if err != nil {
        return err
    }

    // Can now pass to helper functions
    topDept := findTopDepartment(summary)
    fmt.Printf("Biggest spending: %s with $%.2f\n",
        topDept.Department, topDept.Total)

    return nil
}

// NEW: Can create typed helper functions
func findTopDepartment(summary map[int]costco.SpendingByDepartment) costco.SpendingByDepartment {
    var top costco.SpendingByDepartment
    maxSpend := 0.0

    for _, dept := range summary {
        if dept.Total > maxSpend {
            maxSpend = dept.Total
            top = dept
        }
    }

    return top
}
```

### Example 3: AI Agent Creating Test Mocks

**Before (v0.1.1):**
```go
// Could not easily mock - had to use real client
func TestMyService(t *testing.T) {
    config := costco.Config{
        Email:    "test@example.com",
        Password: "test",
    }
    client := costco.NewClient(config) // Real client!

    service := NewMyService(client)
    // ... test with real API calls (not ideal)
}
```

**After (v0.2.0):**
```go
import "github.com/stretchr/testify/mock"

// Create a mock client
type MockCostcoClient struct {
    mock.Mock
}

func (m *MockCostcoClient) GetOnlineOrders(ctx context.Context, startDate, endDate string, pageNumber, pageSize int) (*costco.OnlineOrdersResponse, error) {
    args := m.Called(ctx, startDate, endDate, pageNumber, pageSize)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*costco.OnlineOrdersResponse), args.Error(1)
}

// Now can test with mocks!
func TestMyService(t *testing.T) {
    mockClient := new(MockCostcoClient)
    mockClient.On("GetOnlineOrders",
        mock.Anything, "2025-01-01", "2025-01-31", 1, 10,
    ).Return(&costco.OnlineOrdersResponse{
        BCOrders: []costco.OnlineOrder{
            {OrderNumber: "12345", OrderTotal: 99.99},
        },
    }, nil)

    service := NewMyService(mockClient)
    // ... test with mocked responses

    mockClient.AssertExpectations(t)
}
```

---

## File Organization Changes

For AI agents that reference specific files:

| Old Location (v0.1.1) | New Location (v0.2.0) | Notes |
|------------------------|----------------------|-------|
| `pkg/costco/types.go` (all types) | `pkg/costco/orders.go` | Order-related types |
| `pkg/costco/types.go` (all types) | `pkg/costco/receipts.go` | Receipt-related types |
| `pkg/costco/types.go` (all types) | `pkg/costco/analytics.go` | Analytics types |
| `pkg/costco/types.go` (all types) | `pkg/costco/options.go` | Config types |
| `pkg/costco/types.go` (all types) | `pkg/costco/graphql.go` | GraphQL types |
| `pkg/costco/client.go` (auth) | `pkg/costco/internal/auth/` | Auth logic (internal) |
| `pkg/costco/client.go` (transport) | `pkg/costco/internal/transport/` | HTTP/GraphQL (internal) |
| `pkg/costco/helpers.go` | `pkg/costco/internal/analytics/` | Analytics impl (internal) |
| `pkg/costco/config.go` | `pkg/costco/internal/persistence/` | Storage (internal) |

**Note:** `internal/` packages are not importable by external users - only used internally.

---

## Troubleshooting

### Issue: "undefined: costco.ItemPurchase"

**Cause:** Using old code that expected anonymous struct.

**Fix:** Update your type references:
```go
// Old
var history []struct { Date string; Quantity int; Price float64; Barcode string }

// New
var history []costco.ItemPurchase
```

### Issue: "cannot use client (type *costco.Client) as type costco.CostcoClient"

**Cause:** The concrete `*costco.Client` implements the interface, but you need to explicitly type it.

**Fix:**
```go
// This works:
var client costco.CostcoClient = costco.NewClient(config)

// This also works:
func doSomething(client costco.CostcoClient) { ... }
doSomething(costco.NewClient(config)) // Client implements interface
```

### Issue: Tests failing after upgrade

**Cause:** Likely using anonymous struct types in test fixtures.

**Fix:** Update test fixtures to use new named types:
```go
// Old
expected := []struct { Date string; Quantity int; Price float64; Barcode string }{
    {"2025-01-15", 2, 19.99, "12345"},
}

// New
expected := []costco.ItemPurchase{
    {Date: "2025-01-15", Quantity: 2, Price: 19.99, Barcode: "12345"},
}
```

---

## Performance & Behavior Changes

### Improved Sorting Performance

The `GetFrequentItems()` function now uses `sort.Slice` instead of bubble sort:

- **Old:** O(n²) bubble sort
- **New:** O(n log n) quicksort (stdlib)

**Impact:** Significant performance improvement for large result sets (100+ items).

### No Other Behavior Changes

All other functionality remains **identical**:
- Authentication flow unchanged
- Token refresh logic unchanged
- API request/response format unchanged
- Error handling unchanged
- Logging unchanged

---

## Testing Your Migration

Run this checklist after updating:

```bash
# 1. Update dependency
go get github.com/costco-go/pkg/costco@v0.2.0

# 2. Run tests
go test ./...

# 3. Build your application
go build ./...

# 4. Verify runtime behavior
# (Run your application and check logs for any issues)
```

---

## Getting Help

If you encounter issues during migration:

1. Check this migration guide
2. Review the [AUDIT.md](./AUDIT.md) for detailed changes
3. Check the [CHANGELOG.md](./CHANGELOG.md)
4. Open an issue at https://github.com/eshaffer321/costco-go/issues

---

## Complete Before/After Reference

### All Type Renames

| Old Type (v0.1.1) | New Type (v0.2.0) | Location |
|-------------------|-------------------|----------|
| Anonymous `[]struct{Date, Quantity, Price, Barcode}` | `[]ItemPurchase` | `analytics.go` |
| Anonymous `map[int]struct{Department, Total, ItemCount}` | `map[int]SpendingByDepartment` | `analytics.go` |
| Anonymous `[]struct{ItemNumber, ItemDescription, ...}` | `[]FrequentItem` | `analytics.go` |

### All API Signatures

```go
// UNCHANGED - no signature changes
GetOnlineOrders(ctx context.Context, startDate, endDate string, pageNumber, pageSize int) (*OnlineOrdersResponse, error)
GetReceipts(ctx context.Context, startDate, endDate, documentType, documentSubType string) (*ReceiptsWithCountsResponse, error)
GetReceiptDetail(ctx context.Context, barcode, documentType string) (*Receipt, error)
GetAllTransactionItems(ctx context.Context, startDate, endDate string) ([]TransactionWithItems, error)

// CHANGED - return type now named instead of anonymous
GetItemHistory(ctx context.Context, itemNumber, startDate, endDate string) ([]ItemPurchase, error)
GetSpendingSummary(ctx context.Context, startDate, endDate string) (map[int]SpendingByDepartment, error)
GetFrequentItems(ctx context.Context, startDate, endDate string, limit int) ([]FrequentItem, error)
```

---

**Migration complete!** Welcome to v0.2.0 🎉
