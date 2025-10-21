package costco

import "context"

// CostcoClient defines the interface for interacting with Costco's API.
// This interface can be used for mocking in tests or creating alternative implementations.
//
// Example usage with mocking:
//
//	type MockClient struct {
//	    mock.Mock
//	}
//
//	func (m *MockClient) GetOnlineOrders(ctx context.Context, startDate, endDate string, pageNumber, pageSize int) (*OnlineOrdersResponse, error) {
//	    args := m.Called(ctx, startDate, endDate, pageNumber, pageSize)
//	    return args.Get(0).(*OnlineOrdersResponse), args.Error(1)
//	}
//
//	// Use in tests
//	mockClient := new(MockClient)
//	mockClient.On("GetOnlineOrders", ...).Return(&OnlineOrdersResponse{...}, nil)
type CostcoClient interface {
	// GetOnlineOrders retrieves online orders from Costco.com within the specified date range.
	// Supports pagination via pageNumber and pageSize parameters.
	GetOnlineOrders(ctx context.Context, startDate, endDate string, pageNumber, pageSize int) (*OnlineOrdersResponse, error)

	// GetReceipts retrieves warehouse receipts within the specified date range.
	// Can filter by documentType ("all", "warehouse", "fuel") and documentSubType.
	GetReceipts(ctx context.Context, startDate, endDate, documentType, documentSubType string) (*ReceiptsWithCountsResponse, error)

	// GetReceiptDetail retrieves full details for a specific receipt identified by barcode.
	// documentType should be "warehouse" or "fuel" depending on the receipt type.
	GetReceiptDetail(ctx context.Context, barcode, documentType string) (*Receipt, error)

	// GetAllTransactionItems fetches all receipts in a date range and retrieves full item details for each.
	// This is a convenience method that combines GetReceipts and GetReceiptDetail.
	GetAllTransactionItems(ctx context.Context, startDate, endDate string) ([]TransactionWithItems, error)

	// GetItemHistory retrieves the purchase history for a specific item number.
	// Returns a list of all transactions where the item was purchased.
	GetItemHistory(ctx context.Context, itemNumber, startDate, endDate string) ([]ItemPurchase, error)

	// GetSpendingSummary calculates total spending and item counts by department.
	// Returns a map keyed by department number.
	GetSpendingSummary(ctx context.Context, startDate, endDate string) (map[int]SpendingByDepartment, error)

	// GetFrequentItems returns the most frequently purchased items, sorted by purchase frequency.
	// The limit parameter controls how many items to return (0 = return all).
	GetFrequentItems(ctx context.Context, startDate, endDate string, limit int) ([]FrequentItem, error)
}
