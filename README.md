# Costco Go Client

[![Version](https://img.shields.io/badge/version-0.3.1-blue.svg)](https://github.com/eshaffer321/costco-go/releases/tag/v0.3.1)

A Go client library and CLI for accessing Costco order history and receipt data via their GraphQL API.

## Features

- OAuth2 authentication with automatic token refresh
- Get online order history
- Get warehouse receipts
- Get detailed receipt information with line items
- Command-line interface
- JSON output support
- Test-driven development with comprehensive test coverage

## Installation

Install the latest version:

```bash
go get github.com/costco-go/pkg/costco
```

Or install a specific version:

```bash
go get github.com/costco-go/pkg/costco@v0.1.0
```

## Library Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/costco-go/pkg/costco"
)

func main() {
    config := costco.Config{
        Email:              "your-email@example.com",
        Password:           "your-password",
        WarehouseNumber:    "847", // Your local warehouse number
        TokenRefreshBuffer: 5 * time.Minute,
    }

    client := costco.NewClient(config)
    ctx := context.Background()

    // Get online orders
    orders, err := client.GetOnlineOrders(ctx, "2025-01-01", "2025-01-31", 1, 10)
    if err != nil {
        log.Fatal(err)
    }

    for _, order := range orders.BCOrders {
        fmt.Printf("Order %s: $%.2f\n", order.OrderNumber, order.OrderTotal)
    }

    // Get receipts
    receipts, err := client.GetReceipts(ctx, "1/01/2025", "1/31/2025", "all", "all")
    if err != nil {
        log.Fatal(err)
    }

    for _, receipt := range receipts.Receipts {
        fmt.Printf("Receipt from %s: $%.2f\n", receipt.TransactionDateTime, receipt.Total)
    }

    // Get detailed receipt
    receipt, err := client.GetReceiptDetail(ctx, "21134300501862509051323", "warehouse")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Receipt total: $%.2f with %d items\n", receipt.Total, receipt.TotalItemCount)
}
```

## Logging

The client supports optional logger injection using Go's standard `log/slog` package. By default, if no logger is provided, all logs are silently discarded.

### Basic Usage (Silent Mode)

```go
// Logs are silently discarded (default behavior)
config := costco.Config{
    Email:           "your-email@example.com",
    Password:        "your-password",
    WarehouseNumber: "847",
}
client := costco.NewClient(config)
```

### With Custom Logger

```go
import (
    "log/slog"
    "os"

    "github.com/costco-go/pkg/costco"
)

// Create a text logger that outputs to stdout
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

config := costco.Config{
    Email:           "your-email@example.com",
    Password:        "your-password",
    WarehouseNumber: "847",
    Logger:          logger,
}

client := costco.NewClient(config)
```

### JSON Logging

For structured JSON logs, use `slog.NewJSONHandler`:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug, // Include debug logs
}))

config := costco.Config{
    Email:           "your-email@example.com",
    Password:        "your-password",
    WarehouseNumber: "847",
    Logger:          logger,
}

client := costco.NewClient(config)
```

### Log Levels

The client uses the following log levels:

- `Info`: High-level operations (fetching orders, receipts, authentication success)
- `Debug`: Detailed debugging information (API requests, token refresh)
- `Warn`: Non-critical issues (token expiring soon, fallback behavior)
- `Error`: Error conditions (authentication failures, API errors)

### Structured Logging

All logs use structured key-value pairs for easy parsing and filtering:

```json
{
  "time": "2025-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "fetching online orders",
  "client": "costco",
  "start_date": "2025-01-01",
  "end_date": "2025-01-31",
  "page_number": 1,
  "page_size": 10
}
```

Every log message includes a `client=costco` attribute for easy identification in multi-client applications.

## CLI Usage

### Build the CLI

```bash
go build -o costco-cli cmd/costco-cli/main.go
```

### Set credentials via environment variables

```bash
export COSTCO_EMAIL="your-email@example.com"
export COSTCO_PASSWORD="your-password"
export COSTCO_WAREHOUSE="847"  # Optional, defaults to 847
```

### Get online orders

```bash
# Get orders from last 3 months (default)
./costco-cli -cmd orders

# Get orders for specific date range
./costco-cli -cmd orders -start 2025-01-01 -end 2025-01-31

# Get orders with pagination
./costco-cli -cmd orders -page 2 -size 20

# Output as JSON
./costco-cli -cmd orders -json
```

### Get receipts

```bash
# Get all receipts from last 3 months
./costco-cli -cmd receipts

# Get receipts for specific date range
./costco-cli -cmd receipts -start 2025-01-01 -end 2025-01-31

# Output as JSON
./costco-cli -cmd receipts -json
```

### Get receipt details

```bash
# Get detailed receipt with all line items
./costco-cli -cmd receipt-detail -barcode 21134300501862509051323

# Output as JSON
./costco-cli -cmd receipt-detail -barcode 21134300501862509051323 -json
```

### CLI Flags

- `-email`: Costco account email (overrides COSTCO_EMAIL env var)
- `-password`: Costco account password (overrides COSTCO_PASSWORD env var)
- `-warehouse`: Warehouse number (overrides COSTCO_WAREHOUSE env var, default: 847)
- `-cmd`: Command to run: `orders`, `receipts`, or `receipt-detail`
- `-start`: Start date in YYYY-MM-DD format
- `-end`: End date in YYYY-MM-DD format
- `-barcode`: Receipt barcode (required for receipt-detail command)
- `-page`: Page number for orders (default: 1)
- `-size`: Page size for orders (default: 10)
- `-json`: Output results as JSON

## Running Tests

```bash
go test ./pkg/costco -v
```

## API Details

The client uses Costco's OAuth2 authentication flow and GraphQL API:

- **Auth endpoint**: `https://signin.costco.com/.../oauth2/v2.0/token`
- **GraphQL endpoint**: `https://ecom-api.costco.com/ebusiness/order/v1/orders/graphql`
- **Auth header**: `costco-x-authorization: Bearer {id_token}`

The client handles:
- Initial authentication with email/password
- Automatic token refresh before expiry
- Thread-safe token management
- GraphQL query construction and response parsing

## Data Structures

### Online Orders
- Order header information (date, total, status)
- Line items with shipping details
- Shipment tracking information

### Receipts
- Transaction details (date, warehouse, total)
- Complete line item details with prices
- Tax breakdown
- Payment information
- Membership number

## Handling Discount Line Items

Costco's API returns discounts as separate line items in receipts. These discount items have special characteristics that allow you to identify and process them differently from regular items.

### Discount Item Characteristics

Discount line items have:
- **Negative amount** (e.g., `-4.00`)
- **Negative unit** (e.g., `-1`)
- **Description starting with "/"** followed by the parent item number (e.g., `"/1553261"`)

**Important:** The discount amount is already factored into the receipt's `SubTotal`. You should not double-count discounts when calculating totals.

### Distinguishing Discounts from Returns

Return items also have negative amounts, but they differ from discounts:
- Returns have **normal descriptions** (e.g., "RED GRAPE")
- Returns appear in receipts with **`TransactionType: "Refund"`**
- Returns do **NOT** have the "/" prefix in their description

### Helper Methods

The library provides two helper methods to identify and process discount items:

#### IsDiscount()

Returns `true` if a line item is a discount:

```go
for _, item := range receipt.ItemArray {
    if item.IsDiscount() {
        fmt.Printf("Found discount: $%.2f\n", math.Abs(item.Amount))
        continue
    }
    // Process regular items...
}
```

#### GetParentItemNumber()

Returns the item number that the discount applies to:

```go
for _, item := range receipt.ItemArray {
    if item.IsDiscount() {
        parentItemNum := item.GetParentItemNumber()
        fmt.Printf("Discount of $%.2f applies to item %s\n",
            math.Abs(item.Amount),
            parentItemNum)
    }
}
```

### Example: Calculating Net Item Amounts

Here's how to build a map of items with discounts applied:

```go
// Build net amounts map
itemAmounts := make(map[string]float64)
itemDescs := make(map[string]string)

for _, item := range receipt.ItemArray {
    if item.IsDiscount() {
        // Apply discount to parent item
        parentNum := item.GetParentItemNumber()
        itemAmounts[parentNum] += item.Amount
    } else {
        // Add regular item
        itemAmounts[item.ItemNumber] += item.Amount
        itemDescs[item.ItemNumber] = item.ItemDescription01
    }
}

// Now process items with net amounts
for itemNum, netAmount := range itemAmounts {
    fmt.Printf("%s: $%.2f\n", itemDescs[itemNum], netAmount)
}
```

### Real-World Example

Given this receipt data:

```json
{
  "itemArray": [
    {
      "itemNumber": "1553261",
      "itemDescription01": "GUAC BOWL",
      "amount": 13.99,
      "unit": 1
    },
    {
      "itemNumber": "363064",
      "itemDescription01": "/1553261",
      "amount": -4.00,
      "unit": -1
    }
  ],
  "subTotal": 9.99,
  "instantSavings": 4.00
}
```

Processing with helpers:

```go
// Item 1: Regular item
item1.IsDiscount() // Returns: false

// Item 2: Discount item
item2.IsDiscount()           // Returns: true
item2.GetParentItemNumber()  // Returns: "1553261"

// Net amount: 13.99 + (-4.00) = 9.99 (matches subTotal)
```

### Use Cases

**Budgeting Applications:** Calculate net amounts per item to accurately categorize spending.

**Price Tracking:** Track both original and discounted prices to analyze savings over time.

**Receipt Processing:** Filter out discount line items to avoid confusion when presenting items to users.

**Analytics:** Aggregate `instantSavings` data across receipts to measure total savings.

## License

MIT