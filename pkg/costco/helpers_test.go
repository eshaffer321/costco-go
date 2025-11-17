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
							"receipts": []map[string]interface{}{
								{
									"warehouseName":       "TEST",
									"transactionDateTime": "2025-01-01T10:00:00",
									"transactionBarcode":  "123",
									"total":               100.00,
									"totalItemCount":      3,
								},
								{
									"warehouseName":       "TEST",
									"transactionDateTime": "2025-01-02T10:00:00",
									"transactionBarcode":  "456",
									"total":               50.00,
									"totalItemCount":      2,
								},
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			} else if req.Query == ReceiptDetailQuery {
				barcode := req.Variables["barcode"].(string)
				var items []map[string]interface{}

				if barcode == "123" {
					items = []map[string]interface{}{
						{
							"itemNumber":           "ITEM1",
							"itemDescription01":    "Item One",
							"unit":                 2,
							"amount":               10.00,
							"itemDepartmentNumber": 1,
						},
						{
							"itemNumber":           "ITEM2",
							"itemDescription01":    "Item Two",
							"unit":                 1,
							"amount":               5.00,
							"itemDepartmentNumber": 2,
						},
					}
				} else {
					items = []map[string]interface{}{
						{
							"itemNumber":           "ITEM1",
							"itemDescription01":    "Item One",
							"unit":                 3,
							"amount":               15.00,
							"itemDepartmentNumber": 1,
						},
						{
							"itemNumber":           "ITEM3",
							"itemDescription01":    "Item Three",
							"unit":                 1,
							"amount":               8.00,
							"itemDepartmentNumber": 3,
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
									"total":               100.00,
									"membershipNumber":    "111222333",
									"itemArray":           items,
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

	// Test with no limit (return all)
	items, err := client.GetFrequentItems(context.Background(), "2025-01-01", "2025-01-31", 0)
	require.NoError(t, err)

	assert.Len(t, items, 3)
	// ITEM1 should be most frequent (appears in 2 transactions)
	assert.Equal(t, "ITEM1", items[0].ItemNumber)
	assert.Equal(t, "Item One", items[0].ItemDescription)
	assert.Equal(t, 2, items[0].PurchaseCount)
	assert.Equal(t, 5, items[0].TotalQuantity)
	assert.Equal(t, 25.00, items[0].TotalSpent)

	// Test with limit
	limitedItems, err := client.GetFrequentItems(context.Background(), "2025-01-01", "2025-01-31", 2)
	require.NoError(t, err)
	assert.Len(t, limitedItems, 2)
}

func TestGetSpendingSummary(t *testing.T) {
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
							"receipts": []map[string]interface{}{
								{
									"warehouseName":       "TEST",
									"transactionDateTime": "2025-01-01T10:00:00",
									"transactionBarcode":  "123",
									"total":               30.00,
									"totalItemCount":      2,
								},
								{
									"warehouseName":       "TEST",
									"transactionDateTime": "2025-01-02T10:00:00",
									"transactionBarcode":  "456",
									"total":               45.00,
									"totalItemCount":      2,
								},
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			} else if req.Query == ReceiptDetailQuery {
				barcode := req.Variables["barcode"].(string)
				var items []map[string]interface{}

				if barcode == "123" {
					items = []map[string]interface{}{
						{
							"itemNumber":           "ITEM1",
							"itemDescription01":    "Item One",
							"unit":                 2,
							"amount":               10.00,
							"itemDepartmentNumber": 1,
						},
						{
							"itemNumber":           "ITEM2",
							"itemDescription01":    "Item Two",
							"unit":                 1,
							"amount":               20.00,
							"itemDepartmentNumber": 2,
						},
					}
				} else {
					items = []map[string]interface{}{
						{
							"itemNumber":           "ITEM3",
							"itemDescription01":    "Item Three",
							"unit":                 1,
							"amount":               15.00,
							"itemDepartmentNumber": 1,
						},
						{
							"itemNumber":           "ITEM4",
							"itemDescription01":    "Item Four",
							"unit":                 2,
							"amount":               30.00,
							"itemDepartmentNumber": 3,
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
									"total":               100.00,
									"membershipNumber":    "111222333",
									"itemArray":           items,
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

	summary, err := client.GetSpendingSummary(context.Background(), "2025-01-01", "2025-01-31")
	require.NoError(t, err)

	assert.Len(t, summary, 3)
	assert.Equal(t, "Department 1", summary[1].Department)
	assert.Equal(t, 25.00, summary[1].Total)
	assert.Equal(t, 3, summary[1].ItemCount)

	assert.Equal(t, "Department 2", summary[2].Department)
	assert.Equal(t, 20.00, summary[2].Total)
	assert.Equal(t, 1, summary[2].ItemCount)

	assert.Equal(t, "Department 3", summary[3].Department)
	assert.Equal(t, 30.00, summary[3].Total)
	assert.Equal(t, 2, summary[3].ItemCount)
}

func TestGetItemHistory(t *testing.T) {
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
							"receipts": []map[string]interface{}{
								{
									"warehouseName":       "TEST",
									"transactionDateTime": "2025-01-01T10:00:00",
									"transactionBarcode":  "123",
									"total":               30.00,
									"totalItemCount":      2,
								},
								{
									"warehouseName":       "TEST",
									"transactionDateTime": "2025-01-15T14:30:00",
									"transactionBarcode":  "456",
									"total":               45.00,
									"totalItemCount":      2,
								},
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			} else if req.Query == ReceiptDetailQuery {
				barcode := req.Variables["barcode"].(string)
				var items []map[string]interface{}

				if barcode == "123" {
					items = []map[string]interface{}{
						{
							"itemNumber":           "ITEM1",
							"itemDescription01":    "Organic Milk",
							"unit":                 2,
							"amount":               10.00,
							"itemDepartmentNumber": 1,
						},
						{
							"itemNumber":           "ITEM2",
							"itemDescription01":    "Bread",
							"unit":                 1,
							"amount":               5.00,
							"itemDepartmentNumber": 2,
						},
					}
				} else {
					items = []map[string]interface{}{
						{
							"itemNumber":           "ITEM1",
							"itemDescription01":    "Organic Milk",
							"unit":                 3,
							"amount":               15.00,
							"itemDepartmentNumber": 1,
						},
						{
							"itemNumber":           "ITEM3",
							"itemDescription01":    "Eggs",
							"unit":                 2,
							"amount":               8.00,
							"itemDepartmentNumber": 1,
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
									"total":               100.00,
									"membershipNumber":    "111222333",
									"itemArray":           items,
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

	// Get history for ITEM1 which appears in both transactions
	history, err := client.GetItemHistory(context.Background(), "ITEM1", "2025-01-01", "2025-01-31")
	require.NoError(t, err)

	assert.Len(t, history, 2)
	assert.Equal(t, "2025-01-01", history[0].Date)
	assert.Equal(t, 2, history[0].Quantity)
	assert.Equal(t, 10.00, history[0].Price)
	assert.Equal(t, "123", history[0].Barcode)

	assert.Equal(t, "2025-01-01", history[1].Date) // Mock returns same date for all receipts
	assert.Equal(t, 3, history[1].Quantity)
	assert.Equal(t, 15.00, history[1].Price)
	assert.Equal(t, "456", history[1].Barcode)

	// Get history for item that doesn't exist
	emptyHistory, err := client.GetItemHistory(context.Background(), "NONEXISTENT", "2025-01-01", "2025-01-31")
	require.NoError(t, err)
	assert.Empty(t, emptyHistory)
}
