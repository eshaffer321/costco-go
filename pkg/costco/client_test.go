package costco

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	config := Config{
		Email:           "test@example.com",
		Password:        "password",
		WarehouseNumber: "847",
	}

	client := NewClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, config.Email, client.config.Email)
	assert.Equal(t, config.Password, client.config.Password)
	assert.Equal(t, config.WarehouseNumber, client.config.WarehouseNumber)
	assert.Equal(t, 5*time.Minute, client.config.TokenRefreshBuffer)
}

func TestAuthenticate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/oauth2/v2.0/token", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, HeaderContentTypeForm, r.Header.Get("Content-Type"))

		err := r.ParseForm()
		require.NoError(t, err)

		assert.Equal(t, ClientID, r.Form.Get("client_id"))
		assert.Equal(t, "test@example.com", r.Form.Get("username"))
		assert.Equal(t, "password123", r.Form.Get("password"))
		assert.Equal(t, GrantType, r.Form.Get("grant_type"))

		resp := TokenResponse{
			IDToken:               generateTestJWT(time.Now().Add(1 * time.Hour).Unix()),
			TokenType:             "Bearer",
			RefreshToken:          "test-refresh-token",
			RefreshTokenExpiresIn: 7776000,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
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

	err := client.authenticate()
	require.NoError(t, err)

	assert.NotNil(t, client.token)
	assert.NotEmpty(t, client.token.IDToken)
	assert.Equal(t, "test-refresh-token", client.token.RefreshToken)
	assert.True(t, client.tokenExpiry.After(time.Now()))
}

func TestRefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		require.NoError(t, err)

		if r.Form.Get("grant_type") == "refresh_token" {
			assert.Equal(t, "old-refresh-token", r.Form.Get("refresh_token"))

			resp := TokenResponse{
				IDToken:               generateTestJWT(time.Now().Add(2 * time.Hour).Unix()),
				TokenType:             "Bearer",
				RefreshToken:          "new-refresh-token",
				RefreshTokenExpiresIn: 7776000,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
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
		token: &TokenResponse{
			IDToken:      generateTestJWT(time.Now().Add(-1 * time.Hour).Unix()),
			RefreshToken: "old-refresh-token",
		},
		tokenExpiry: time.Now().Add(-1 * time.Hour),
	}

	err := client.refreshToken()
	require.NoError(t, err)

	assert.NotNil(t, client.token)
	assert.Equal(t, "new-refresh-token", client.token.RefreshToken)
	assert.True(t, client.tokenExpiry.After(time.Now()))
}

func TestGetOnlineOrders(t *testing.T) {
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
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, HeaderContentTypeJSON, r.Header.Get("Content-Type"))
			assert.Contains(t, r.Header.Get("costco-x-authorization"), "Bearer ")

			var req GraphQLRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			assert.Contains(t, req.Query, "getOnlineOrders")
			assert.Equal(t, "2025-01-01", req.Variables["startDate"])
			assert.Equal(t, "2025-01-31", req.Variables["endDate"])

			resp := map[string]interface{}{
				"data": map[string]interface{}{
					"getOnlineOrders": map[string]interface{}{
						"pageNumber":           1,
						"pageSize":             10,
						"totalNumberOfRecords": 1,
						"bcOrders": []map[string]interface{}{
							{
								"orderHeaderId":      "12345",
								"orderPlacedDate":    "2025-01-15",
								"orderNumber":        "ORD-001",
								"orderTotal":         99.99,
								"warehouseNumber":    "847",
								"status":             "Delivered",
								"emailAddress":       "test@example.com",
								"orderCancelAllowed": false,
								"orderPaymentFailed": false,
								"orderReturnAllowed": true,
								"orderLineItems":     []interface{}{},
							},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
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

	orders, err := client.GetOnlineOrders(context.Background(), "2025-01-01", "2025-01-31", 1, 10)
	require.NoError(t, err)

	assert.NotNil(t, orders)
	assert.Equal(t, 1, orders.PageNumber)
	assert.Equal(t, 10, orders.PageSize)
	assert.Equal(t, 1, orders.TotalNumberOfRecords)
	assert.Len(t, orders.BCOrders, 1)
	assert.Equal(t, "12345", orders.BCOrders[0].OrderHeaderID)
	assert.Equal(t, "ORD-001", orders.BCOrders[0].OrderNumber)
	assert.Equal(t, 99.99, orders.BCOrders[0].OrderTotal)
}

func TestGetReceiptDetail(t *testing.T) {
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

			assert.Contains(t, req.Query, "receiptsWithCounts")
			assert.Equal(t, "21134300501862509051323", req.Variables["barcode"])

			resp := map[string]interface{}{
				"data": map[string]interface{}{
					"receiptsWithCounts": map[string]interface{}{
						"receipts": []map[string]interface{}{
							{
								"warehouseName":       "MERIDIAN",
								"receiptType":         "In-Warehouse",
								"documentType":        "WarehouseReceiptDetail",
								"transactionDateTime": "2025-09-05T13:23:00",
								"transactionBarcode":  "21134300501862509051323",
								"total":               269.13,
								"totalItemCount":      20,
								"subTotal":            253.9,
								"taxes":               15.23,
								"membershipNumber":    "111869503713",
								"itemArray": []map[string]interface{}{
									{
										"itemNumber":          "1529345",
										"itemDescription01":   "ALM TORTILLA",
										"itemDescription02":   "20CT T8H5 P720 SL45",
										"unit":                1,
										"amount":              11.89,
										"itemUnitPriceAmount": 11.89,
									},
								},
								"tenderArray": []map[string]interface{}{
									{
										"tenderTypeCode":       "064",
										"tenderDescription":    "COSTCO VISA",
										"amountTender":         269.13,
										"displayAccountNumber": "9920",
									},
								},
							},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
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

	receipt, err := client.GetReceiptDetail(context.Background(), "21134300501862509051323", "warehouse")
	require.NoError(t, err)

	assert.NotNil(t, receipt)
	assert.Equal(t, "MERIDIAN", receipt.WarehouseName)
	assert.Equal(t, "21134300501862509051323", receipt.TransactionBarcode)
	assert.Equal(t, 269.13, receipt.Total)
	assert.Equal(t, 20, receipt.TotalItemCount)
	assert.Len(t, receipt.ItemArray, 1)
	assert.Equal(t, "1529345", receipt.ItemArray[0].ItemNumber)
}

func generateTestJWT(exp int64) string {
	// Create a base64-encoded JWT with the correct exp value
	// This is a simplified JWT for testing - not cryptographically valid
	payload := fmt.Sprintf(`{"exp":%d,"iat":1757379753,"email":"test@example.com"}`, exp)
	// In a real JWT these would be base64url encoded, but for testing we can use a simplified version
	return "eyJhbGciOiJSUzI1NiIsImtpZCI6InRlc3QiLCJ0eXAiOiJKV1QifQ." +
		base64Encode(payload) +
		".signature"
}

func base64Encode(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

// testTransport is a custom RoundTripper that redirects requests to our test server
type testTransport struct {
	baseURL string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace the host with our test server
	testURL := t.baseURL
	if req.URL.Path == "/e0714dd4-784d-46d6-a278-3e29553483eb/b2c_1a_sso_wcs_signup_signin_157/oauth2/v2.0/token" {
		testURL += "/oauth2/v2.0/token"
	} else if req.URL.Path == "/ebusiness/order/v1/orders/graphql" {
		testURL += "/graphql"
	} else {
		testURL += req.URL.Path
	}

	newReq, err := http.NewRequest(req.Method, testURL, req.Body)
	if err != nil {
		return nil, err
	}

	newReq.Header = req.Header
	return http.DefaultTransport.RoundTrip(newReq)
}
