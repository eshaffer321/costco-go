# Codebase Audit Report - costco-go

**Date:** 2025-10-20
**Version Audited:** v0.1.1
**Auditor:** Claude (Sonnet 4.5)
**Scope:** Full codebase architecture, organization, and quality review

---

## Executive Summary

The costco-go library is functionally solid with good fundamentals (thread safety, error handling, structured logging), but suffers from **poor architectural organization**. The codebase is structured like a prototype with all logic dumped into a flat package structure, making it difficult to navigate, test, and maintain.

**Recommendation:** Aggressive refactor to establish proper architectural boundaries before the codebase grows larger.

---

## Metrics

### Current State (v0.1.1)

| Metric | Rating | Notes |
|--------|--------|-------|
| **Separation of Concerns** | ⭐⭐☆☆☆ | Everything mixed together in client.go |
| **Discoverability** | ⭐⭐☆☆☆ | 35+ exports in flat namespace |
| **Testability** | ⭐⭐⭐☆☆ | Tests exist but hard to isolate |
| **Maintainability** | ⭐⭐☆☆☆ | 492-line client.go doing too much |
| **Professional Polish** | ⭐⭐⭐☆☆ | Works but lacks organization |
| **Documentation** | ⭐⭐⭐⭐☆ | Good README/CLAUDE.md, missing godoc |

### Code Statistics

```
Total Lines of Code:     2,932
Implementation Code:     1,494
Test Code:              1,102
Files:                      13 (9 implementation, 4 test)
Exported Symbols:          35+
Dependencies:                2 (minimal - good!)
Largest File:          client.go (492 lines)
```

---

## Critical Issues

### 🚨 Issue #1: No Architectural Boundaries (SEVERITY: HIGH)

**File:** `pkg/costco/client.go` (492 lines)

**Problem:** Single file contains ALL business logic:
- OAuth2 authentication flow
- Token lifecycle management
- HTTP transport layer
- GraphQL execution
- Business operations (orders, receipts)
- Utility functions (UUID generation)
- Logging wrappers

**Impact:**
- Hard to understand code flow
- Difficult to test components in isolation
- Changes in one area risk breaking others
- No clear ownership of responsibilities

**Example:**
```go
// Lines 74-151: Authentication
func (c *Client) authenticate() error { ... }

// Lines 153-166: Token parsing
func (c *Client) calculateTokenExpiry(tokenString string) time.Time { ... }

// Lines 168-191: Token refresh logic
func (c *Client) refreshTokenIfNeeded() error { ... }

// Lines 193-271: Refresh implementation
func (c *Client) refreshToken() error { ... }

// Lines 273-351: HTTP transport
func (c *Client) executeGraphQL(...) error { ... }

// Lines 353-492: Business operations
func (c *Client) GetOnlineOrders(...) { ... }
func (c *Client) GetReceipts(...) { ... }
func (c *Client) GetReceiptDetail(...) { ... }

// All in ONE file!
```

### 🚨 Issue #2: Types.go Dumping Ground (SEVERITY: MEDIUM-HIGH)

**File:** `pkg/costco/types.go` (258 lines, 30+ types)

**Problem:** All type definitions mixed together with no logical grouping:
- Domain types (orders, receipts, shipments)
- API request/response types
- GraphQL types
- Configuration types
- Storage types

**Impact:**
- Hard to find specific types
- No clear domain boundaries
- Type pollution in single namespace

**Current Structure:**
```go
// All in one file:
type TokenResponse struct { ... }        // Auth type
type OnlineOrder struct { ... }          // Order domain
type Receipt struct { ... }              // Receipt domain
type GraphQLRequest struct { ... }       // Transport type
type Config struct { ... }               // Config type
type OrdersQueryVariables struct { ... } // Query type
// ... 24 more types
```

### 🚨 Issue #3: Anonymous Return Types (SEVERITY: MEDIUM)

**File:** `pkg/costco/helpers.go`

**Problem:** Functions return anonymous structs that users cannot reference:

```go
// Line 76-81: Anonymous struct return type
func (c *Client) GetItemHistory(ctx context.Context, itemNumber, startDate, endDate string) ([]struct {
    Date     string
    Quantity int
    Price    float64
    Barcode  string
}, error)

// Line 116-120: Another anonymous struct
func (c *Client) GetSpendingSummary(ctx context.Context, startDate, endDate string) (map[int]struct {
    Department string
    Total      float64
    ItemCount  int
}, error)

// Line 147-153: Yet another anonymous struct
func (c *Client) GetFrequentItems(ctx context.Context, startDate, endDate string, limit int) ([]struct {
    ItemNumber      string
    ItemDescription string
    TotalQuantity   int
    TotalSpent      float64
    PurchaseCount   int
}, error)
```

**Impact:**
- Users can't reference these types elsewhere in their code
- No godoc documentation for return types
- Poor IDE autocomplete experience
- Can't easily create test fixtures
- Can't marshal/unmarshal directly

### 🚨 Issue #4: No Interface Definition (SEVERITY: MEDIUM)

**File:** `pkg/costco/client.go`

**Problem:** No interface defined for the client, only concrete type:

```go
type Client struct {
    httpClient  *http.Client
    config      Config
    token       *TokenResponse
    tokenExpiry time.Time
    mu          sync.RWMutex
    logger      *slog.Logger
}
```

**Impact:**
- Users cannot mock the Costco API for testing
- No contract definition
- Tight coupling to concrete implementation
- Hard to create alternative implementations

### 🚨 Issue #5: Code Duplication (SEVERITY: LOW-MEDIUM)

**Files:** `pkg/costco/client.go`

**Problem:** HTTP headers duplicated in multiple auth functions:

```go
// Lines 98-105 in authenticate()
req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
req.Header.Set("Accept", "*/*")
req.Header.Set("Accept-Language", "en-US,en;q=0.9")
req.Header.Set("Cache-Control", "no-cache")
req.Header.Set("Origin", "https://www.costco.com")
req.Header.Set("Pragma", "no-cache")
req.Header.Set("Referer", "https://www.costco.com/")
req.Header.Set("User-Agent", "Mozilla/5.0 ...")

// Lines 219-226 in refreshToken() - EXACT SAME 8 LINES
req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
req.Header.Set("Accept", "*/*")
// ... exact duplicates
```

**Impact:**
- Changes must be made in multiple places
- Risk of inconsistency
- Violates DRY principle

### 🚨 Issue #6: Bubble Sort in Production (SEVERITY: LOW)

**File:** `pkg/costco/helpers.go:214-220`

**Problem:** Using O(n²) bubble sort instead of stdlib:

```go
// Sort by purchase count (you could also sort by TotalQuantity or TotalSpent)
// Simple bubble sort for demonstration
for i := 0; i < len(items)-1; i++ {
    for j := 0; j < len(items)-i-1; j++ {
        if items[j].PurchaseCount < items[j+1].PurchaseCount {
            items[j], items[j+1] = items[j+1], items[j]
        }
    }
}
```

**Impact:**
- Inefficient for large datasets
- Comment says "for demonstration" but this is production code
- Should use `sort.Slice` from stdlib

### 🚨 Issue #7: Missing Godoc Comments (SEVERITY: LOW)

**Files:** Multiple

**Problem:** Many exported functions lack proper godoc comments:

```go
// ❌ No godoc
func SaveConfig(config *StoredConfig) error { ... }
func LoadConfig() (*StoredConfig, error) { ... }
func SaveTokens(tokens *StoredTokens) error { ... }
func LoadTokens() (*StoredTokens, error) { ... }
func ClearTokens() error { ... }
```

**Impact:**
- Poor IDE autocomplete/IntelliSense
- No documentation on pkg.go.dev
- Users must read source to understand usage

### 🚨 Issue #8: Flat Namespace Pollution (SEVERITY: MEDIUM)

**Package:** `pkg/costco`

**Problem:** 35+ exports in single flat namespace:

```go
import "github.com/costco-go/pkg/costco"

// Users get ALL of these in one namespace:
costco.Client
costco.Config
costco.StoredConfig
costco.StoredTokens
costco.TokenResponse
costco.OnlineOrder
costco.OrderLineItem
costco.Shipment
costco.TrackingEvent
costco.OnlineOrdersResponse
costco.Receipt
costco.ReceiptItem
costco.Tender
costco.SubTaxes
costco.ReceiptsWithCountsResponse
costco.TransactionWithItems
costco.GraphQLRequest
costco.GraphQLResponse
costco.OrdersQueryVariables
costco.ReceiptsQueryVariables
costco.ReceiptDetailQueryVariables
costco.SaveConfig
costco.LoadConfig
costco.SaveTokens
costco.LoadTokens
costco.ClearTokens
costco.GetConfigInfo
// ... and more
```

**Impact:**
- Overwhelming for new users
- No logical grouping
- Hard to discover relevant types/functions
- Namespace collision risk

---

## Strengths (Keep These!)

### ✅ Thread-Safe Token Management

**Location:** `pkg/costco/client.go:26, 129-132, 169-173, 249-252, 295-297`

Proper use of `sync.RWMutex` for concurrent token access:

```go
type Client struct {
    mu          sync.RWMutex
    token       *TokenResponse
    tokenExpiry time.Time
}

c.mu.Lock()
c.token = &tokenResp
c.tokenExpiry = c.calculateTokenExpiry(tokenResp.IDToken)
c.mu.Unlock()

c.mu.RLock()
token := c.token.IDToken
c.mu.RUnlock()
```

### ✅ Structured Logging with Context

**Location:** `pkg/costco/client.go:30-37, 48-51`

Good use of `log/slog` with default silent mode:

```go
func (c *Client) getLogger() *slog.Logger {
    if c.logger != nil {
        return c.logger
    }
    return slog.New(slog.NewTextHandler(io.Discard, nil))
}

logger = logger.With(slog.String("client", "costco"))
```

### ✅ Error Wrapping with Context

**Location:** Throughout codebase

Proper error chains using `%w`:

```go
if err := c.refreshTokenIfNeeded(); err != nil {
    return fmt.Errorf("token refresh failed: %w", err)
}
```

### ✅ Minimal Dependencies

**Location:** `go.mod`

Only 2 dependencies (excluding dev):
- `github.com/golang-jwt/jwt/v5` - JWT parsing
- `github.com/stretchr/testify` - Testing only

### ✅ Good Test Coverage

**Location:** `*_test.go` files

- Critical paths tested
- HTTP mocking with `httptest.Server`
- JWT generation helpers
- Proper cleanup and isolation

### ✅ Excellent Documentation

**Files:** `README.md`, `CLAUDE.md`

- Comprehensive usage examples
- Clear versioning process
- Good contributing guidelines

---

## Recommendations

### Priority 1: Create Internal Package Structure

Reorganize code into logical internal packages:

```
pkg/costco/
├── internal/
│   ├── auth/          # Authentication & token management
│   ├── transport/     # HTTP & GraphQL client
│   ├── api/           # API operations
│   ├── analytics/     # Analytics functions
│   └── persistence/   # Config & token storage
```

### Priority 2: Define Client Interface

Create clear interface contract:

```go
type CostcoClient interface {
    GetOnlineOrders(...) (*OnlineOrdersResponse, error)
    GetReceipts(...) (*ReceiptsWithCountsResponse, error)
    // ... other methods
}
```

### Priority 3: Split Types into Domain Files

Organize types by domain:
- `orders.go` - Order-related types
- `receipts.go` - Receipt-related types
- `analytics.go` - Analytics types
- `options.go` - Config types

### Priority 4: Create Named Types

Replace anonymous structs with proper types:

```go
type ItemPurchase struct {
    Date     string
    Quantity int
    Price    float64
    Barcode  string
}

func GetItemHistory(...) ([]ItemPurchase, error)
```

### Priority 5: Quick Wins

- Replace bubble sort with `sort.Slice`
- Extract HTTP header builder functions
- Add godoc to all exported symbols
- Consolidate duplicated code

---

## Test Results

```bash
$ go test ./pkg/costco -v
=== RUN   TestNewClient
--- PASS: TestNewClient
=== RUN   TestAuthenticate
--- PASS: TestAuthenticate
=== RUN   TestRefreshToken
--- PASS: TestRefreshToken
# ... all tests passing

ok      github.com/costco-go/pkg/costco    0.234s
```

All existing tests pass. Good foundation for refactoring.

---

## Conclusion

The library has **solid fundamentals** but **poor organization**. It's structured like a prototype, not a production library. The good news: refactoring is purely organizational - no algorithm changes or complex logic rewrites needed.

**Recommendation:** Execute aggressive refactor now while codebase is small (2,932 lines). Establishing proper architecture early will prevent technical debt as features grow.

---

## Appendix: File Breakdown

| File | Lines | Purpose | Issues |
|------|-------|---------|--------|
| `client.go` | 492 | Everything | Too many responsibilities |
| `types.go` | 258 | All types | No organization |
| `helpers.go` | 228 | Analytics | Anonymous return types |
| `queries.go` | 229 | GraphQL queries | Just constants (fine) |
| `config.go` | 186 | Persistence | Missing godoc |
| `constants.go` | 58 | API constants | Good |
| `test_helpers.go` | 43 | Test utils | Good |

---

**Next Step:** See `MIGRATION_v0.2.0.md` for upgrade guide.
