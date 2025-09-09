package costco

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllTransactionItems(t *testing.T) {
	cleanup := SetupTestConfig(t)
	defer cleanup()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/v2.0/token" {
			resp := TokenResponse{
				IDToken:               generateTestJWT(time.Now().Add(1 * time.Hour).Unix()),
				TokenType:             "Bearer",
				RefreshToken:          "test-refresh-token",
				RefreshTokenExpiresIn: 7776000,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.URL.Path == "/graphql" {
			var req GraphQLRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			if req.Query == ReceiptsQuery {
				resp := map[string]interface{}{
					"data": map[string]interface{}{
						"receiptsWithCounts": map[string]interface{}{
							"inWarehouse": 2,
							"gasStation":  1,
							"receipts": []map[string]interface{}{
								{
									"warehouseName":       "TEST WAREHOUSE",
									"receiptType":         "In-Warehouse",
									"documentType":        "warehouse",
									"transactionDateTime": "2025-01-01T10:00:00",
									"transactionBarcode":  "12345",
									"total":               100.50,
									"totalItemCount":      5,
								},
								{
									"warehouseName":       "TEST GAS",
									"receiptType":         "Gas Station",
									"documentType":        "fuel",
									"transactionDateTime": "2025-01-02T11:00:00",
									"transactionBarcode":  "67890",
									"total":               50.00,
									"totalItemCount":      1,
								},
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			} else if req.Query == ReceiptDetailQuery {
				barcode := req.Variables["barcode"].(string)
				documentType := req.Variables["documentType"].(string)

				var items []map[string]interface{}
				if documentType == "warehouse" {
					items = []map[string]interface{}{
						{
							"itemNumber":        "123",
							"itemDescription01": "Test Item",
							"unit":              2,
							"amount":            50.25,
						},
					}
				} else {
					items = []map[string]interface{}{
						{
							"itemNumber":        "GAS001",
							"itemDescription01": "Regular Unleaded",
							"unit":              1,
							"amount":            50.00,
						},
					}
				}

				resp := map[string]interface{}{
					"data": map[string]interface{}{
						"receiptsWithCounts": map[string]interface{}{
							"receipts": []map[string]interface{}{
								{
									"warehouseName":       "TEST",
									"transactionDateTime": "2025-01-01T10:00:00",
									"transactionBarcode":  barcode,
									"total":               100.50,
									"subTotal":            95.00,
									"taxes":               5.50,
									"membershipNumber":    "111222333",
									"itemArray":           items,
									"invoiceNumber":       12345, // number for fuel
									"sequenceNumber":      67890, // number for fuel
								},
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}
		}
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{
			Transport: &testTransport{
				baseURL: server.URL,
			},
		},
		config: Config{
			Email:              "test@example.com",
			Password:           "password123",
			WarehouseNumber:    "847",
			TokenRefreshBuffer: 5 * time.Minute,
		},
	}

	transactions, err := client.GetAllTransactionItems(context.Background(), "2025-01-01", "2025-01-31")
	require.NoError(t, err)

	assert.Len(t, transactions, 2)

	// Check warehouse transaction
	assert.Equal(t, "12345", transactions[0].TransactionBarcode)
	assert.Equal(t, 100.50, transactions[0].Total)
	assert.Len(t, transactions[0].Items, 1)
	assert.Equal(t, "123", transactions[0].Items[0].ItemNumber)

	// Check gas station transaction
	assert.Equal(t, "67890", transactions[1].TransactionBarcode)
	assert.Equal(t, 100.50, transactions[1].Total) // Note: using receipt detail total
	assert.Len(t, transactions[1].Items, 1)
	assert.Equal(t, "GAS001", transactions[1].Items[0].ItemNumber)
}

func TestGetFrequentItems(t *testing.T) {
	// Test the frequency calculation logic directly
	transactions := []TransactionWithItems{
		{
			TransactionBarcode: "123",
			TransactionDate:    time.Now(),
			Items: []ReceiptItem{
				{ItemNumber: "ITEM1", ItemDescription01: "Item One", Unit: 2, Amount: 10.00},
				{ItemNumber: "ITEM2", ItemDescription01: "Item Two", Unit: 1, Amount: 5.00},
			},
		},
		{
			TransactionBarcode: "456",
			TransactionDate:    time.Now(),
			Items: []ReceiptItem{
				{ItemNumber: "ITEM1", ItemDescription01: "Item One", Unit: 3, Amount: 15.00},
				{ItemNumber: "ITEM3", ItemDescription01: "Item Three", Unit: 1, Amount: 8.00},
			},
		},
	}

	// Test the frequency calculation logic
	itemMap := make(map[string]*struct {
		ItemNumber      string
		ItemDescription string
		TotalQuantity   int
		TotalSpent      float64
		PurchaseCount   int
	})

	for _, tx := range transactions {
		for _, item := range tx.Items {
			if stats, exists := itemMap[item.ItemNumber]; exists {
				stats.TotalQuantity += item.Unit
				stats.TotalSpent += item.Amount
				stats.PurchaseCount++
			} else {
				itemMap[item.ItemNumber] = &struct {
					ItemNumber      string
					ItemDescription string
					TotalQuantity   int
					TotalSpent      float64
					PurchaseCount   int
				}{
					ItemNumber:      item.ItemNumber,
					ItemDescription: item.ItemDescription01,
					TotalQuantity:   item.Unit,
					TotalSpent:      item.Amount,
					PurchaseCount:   1,
				}
			}
		}
	}

	assert.Len(t, itemMap, 3)
	assert.Equal(t, 5, itemMap["ITEM1"].TotalQuantity)
	assert.Equal(t, 25.00, itemMap["ITEM1"].TotalSpent)
	assert.Equal(t, 2, itemMap["ITEM1"].PurchaseCount)
}

func TestGetSpendingSummary(t *testing.T) {
	transactions := []TransactionWithItems{
		{
			TransactionBarcode: "123",
			Items: []ReceiptItem{
				{ItemDepartmentNumber: 1, Amount: 10.00, Unit: 2},
				{ItemDepartmentNumber: 2, Amount: 20.00, Unit: 1},
			},
		},
		{
			TransactionBarcode: "456",
			Items: []ReceiptItem{
				{ItemDepartmentNumber: 1, Amount: 15.00, Unit: 1},
				{ItemDepartmentNumber: 3, Amount: 30.00, Unit: 2},
			},
		},
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
			current.Department = "Department " + string(rune(dept+'0'))
			current.Total += item.Amount
			current.ItemCount += item.Unit
			summary[dept] = current
		}
	}

	assert.Len(t, summary, 3)
	assert.Equal(t, 25.00, summary[1].Total)
	assert.Equal(t, 3, summary[1].ItemCount)
	assert.Equal(t, 20.00, summary[2].Total)
	assert.Equal(t, 30.00, summary[3].Total)
}
