# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2025-10-23

### Added
- **Discount Line Item Helpers**: Added helper methods to identify and process discount line items in receipts
  - `ReceiptItem.IsDiscount()` - Returns true if the item is a discount/adjustment line item
  - `ReceiptItem.GetParentItemNumber()` - Returns the item number the discount applies to
- **Comprehensive Documentation**: Added "Handling Discount Line Items" section to README with:
  - Explanation of discount item characteristics
  - Distinction between discounts and returns
  - Real-world examples and use cases
  - Code examples for calculating net item amounts
- **Test Coverage**: Added comprehensive tests for discount helpers including:
  - Edge case testing (returns, empty descriptions, slash in middle of string)
  - Real-world data from actual Costco receipts
  - Workflow tests demonstrating practical usage patterns

### Fixed
- Clarified that discount amounts are already factored into receipt subtotals to prevent double-counting in consuming applications

[0.3.0]: https://github.com/eshaffer321/costco-go/compare/v0.2.0...v0.3.0

## [0.2.0] - 2025-10-20

### Added
- **New Named Types**: Replaced anonymous struct return types with proper named types for better usability:
  - `ItemPurchase` - for `GetItemHistory()` results
  - `SpendingByDepartment` - for `GetSpendingSummary()` results
  - `FrequentItem` - for `GetFrequentItems()` results
- **Client Interface**: Added `CostcoClient` interface for better testability and mocking support
- **Comprehensive Godoc**: Added detailed documentation to all exported types and functions with examples
- **Domain-Specific Files**: Organized types into logical domain files:
  - `orders.go` - Order-related types
  - `receipts.go` - Receipt-related types
  - `analytics.go` - Analytics types
  - `options.go` - Configuration types
  - `graphql.go` - GraphQL types
  - `interface.go` - Client interface definition
- **Migration Guide**: Added `MIGRATION_v0.2.0.md` with comprehensive upgrade instructions for AI agents
- **Audit Documentation**: Added `AUDIT.md` with detailed codebase analysis and recommendations

### Changed
- **Breaking**: `GetItemHistory()` now returns `[]ItemPurchase` instead of anonymous struct
- **Breaking**: `GetSpendingSummary()` now returns `map[int]SpendingByDepartment` instead of anonymous struct
- **Breaking**: `GetFrequentItems()` now returns `[]FrequentItem` instead of anonymous struct
- **File Organization**: Split `types.go` into domain-specific files for better code organization
- **Performance**: Replaced O(n²) bubble sort with O(n log n) stdlib `sort.Slice` in `GetFrequentItems()`
- **Code Quality**: Removed duplicate type definitions between files

### Removed
- **Breaking**: Removed `types.go` (types moved to domain-specific files)
- Removed duplicate `TransactionWithItems` definition from `helpers.go` (now only in `analytics.go`)

### Improved
- Better code organization with clear separation of concerns
- Enhanced discoverability with logical type grouping
- Improved testability with interface-based design
- Better IDE autocomplete and documentation support
- Cleaner, more maintainable codebase structure

### Migration Notes
See `MIGRATION_v0.2.0.md` for detailed upgrade instructions. Key changes:
- Update code using `GetItemHistory()`, `GetSpendingSummary()`, or `GetFrequentItems()` to use new named types
- No changes needed for core API methods (`GetOnlineOrders`, `GetReceipts`, `GetReceiptDetail`)
- Import path remains the same: `github.com/costco-go/pkg/costco`

[0.2.0]: https://github.com/eshaffer321/costco-go/compare/v0.1.1...v0.2.0

## [0.1.1] - 2025-10-19

### Fixed
- Improved receipt fetching logging to eliminate misleading ERROR logs during normal operation
- Optimized `GetReceipts()` fallback logic to try object format first (the format Costco API currently returns), reducing unnecessary failed attempts
- Added clear diagnostic logging with emojis to identify if array fallback is ever needed
- Reduced log noise by changing generic GraphQL decode errors to DEBUG level

[0.1.1]: https://github.com/eshaffer321/costco-go/compare/v0.1.0...v0.1.1

## [0.1.0] - 2025-10-19

### Added
- Initial release of Costco Go client library
- Authentication using Azure AD B2C OAuth2/OIDC flow
- Order history retrieval via GraphQL API
- Support for refresh tokens
- Structured logging with slog (silent by default using io.Discard)
- CLI tool for command-line usage
- Configurable warehouse selection
- Pagination support for order history

### Fixed
- Azure AD B2C authentication flow
- GraphQL array response handling
- Linting and formatting issues
- Test failures with structured logging

[0.1.0]: https://github.com/costco-go/compare/v0.1.0
