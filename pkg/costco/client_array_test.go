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

// TestGetOnlineOrdersArrayResponse tests that the client correctly handles
// the GraphQL response returning an array instead of an object
func TestGetOnlineOrdersArrayResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/graphql" {
			// Simulate the array response from Costco's GraphQL API
			response := GraphQLResponse{
				Data: json.RawMessage(`{
					"getOnlineOrders": [
						{
							"pageNumber": 1,
							"pageSize": 10,
							"totalNumberOfRecords": 2,
							"bcOrders": [
								{
									"orderHeaderId": "123",
									"orderPlacedDate": "2025-01-01",
									"orderNumber": "ORD-001",
									"orderTotal": 99.99,
									"status": "Delivered",
									"warehouseNumber": "847"
								},
								{
									"orderHeaderId": "124",
									"orderPlacedDate": "2025-01-02",
									"orderNumber": "ORD-002",
									"orderTotal": 149.99,
									"status": "Processing",
									"warehouseNumber": "847"
								}
							]
						}
					]
				}`),
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
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
			Email:           "test@example.com",
			WarehouseNumber: "847",
		},
		token: &TokenResponse{
			IDToken: generateTestJWT(time.Now().Add(1 * time.Hour).Unix()),
		},
		tokenExpiry: time.Now().Add(1 * time.Hour),
	}

	orders, err := client.GetOnlineOrders(context.Background(), "2025-01-01", "2025-01-31", 1, 10)
	require.NoError(t, err)
	assert.NotNil(t, orders)
	assert.Equal(t, 2, orders.TotalNumberOfRecords)
	assert.Equal(t, 1, orders.PageNumber)
	assert.Len(t, orders.BCOrders, 2)
	assert.Equal(t, "ORD-001", orders.BCOrders[0].OrderNumber)
	assert.Equal(t, "ORD-002", orders.BCOrders[1].OrderNumber)
}

// TestGetReceiptsArrayResponse tests array response handling for receipts
func TestGetReceiptsArrayResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/graphql" {
			// First call returns array (new format)
			response := GraphQLResponse{
				Data: json.RawMessage(`{
					"receiptsWithCounts": [
						{
							"inWarehouse": 5,
							"gasStation": 3,
							"carWash": 1,
							"receipts": [
								{
									"warehouseName": "Test Warehouse",
									"receiptType": "In-Warehouse",
									"transactionBarcode": "123456789",
									"total": 123.45
								}
							]
						}
					]
				}`),
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
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
			Email:           "test@example.com",
			WarehouseNumber: "847",
		},
		token: &TokenResponse{
			IDToken: generateTestJWT(time.Now().Add(1 * time.Hour).Unix()),
		},
		tokenExpiry: time.Now().Add(1 * time.Hour),
	}

	receipts, err := client.GetReceipts(context.Background(), "1/01/2025", "1/31/2025", "all", "all")
	require.NoError(t, err)
	assert.NotNil(t, receipts)
	assert.Equal(t, 5, receipts.InWarehouse)
	assert.Equal(t, 3, receipts.GasStation)
	assert.Equal(t, 1, receipts.CarWash)
	assert.Len(t, receipts.Receipts, 1)
	assert.Equal(t, "123456789", receipts.Receipts[0].TransactionBarcode)
}

// TestScopeConfiguration tests that the scope is correctly configured
func TestScopeConfiguration(t *testing.T) {
	// Verify the scope matches what the Costco web application uses
	// The scope includes 'profile' which allows tokens from both web and CLI to work interchangeably
	assert.Equal(t, "openid profile offline_access", Scope)
	assert.Contains(t, Scope, "profile", "Scope should contain 'profile' to match Costco web application")
}

// TestEmptyArrayResponse tests handling of empty array responses
func TestEmptyArrayResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/graphql" {
			response := GraphQLResponse{
				Data: json.RawMessage(`{
					"getOnlineOrders": []
				}`),
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
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
			Email:           "test@example.com",
			WarehouseNumber: "847",
		},
		token: &TokenResponse{
			IDToken: generateTestJWT(time.Now().Add(1 * time.Hour).Unix()),
		},
		tokenExpiry: time.Now().Add(1 * time.Hour),
	}

	orders, err := client.GetOnlineOrders(context.Background(), "2025-01-01", "2025-01-31", 1, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no order data returned")
	assert.Nil(t, orders)
}
