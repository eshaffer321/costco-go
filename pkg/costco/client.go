package costco

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Constants moved to constants.go for better organization

type Client struct {
	httpClient  *http.Client
	config      Config
	token       *TokenResponse
	tokenExpiry time.Time
	mu          sync.RWMutex
}

func NewClient(config Config) *Client {
	if config.TokenRefreshBuffer == 0 {
		config.TokenRefreshBuffer = 5 * time.Minute
	}

	client := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}

	// Try to load existing tokens
	if tokens, err := LoadTokens(); err == nil && tokens != nil {
		client.token = &TokenResponse{
			IDToken:      tokens.IDToken,
			RefreshToken: tokens.RefreshToken,
		}
		client.tokenExpiry = tokens.TokenExpiry
	}

	return client
}

func (c *Client) authenticate() error {
	data := url.Values{}
	data.Set("client_id", ClientID)
	data.Set("scope", Scope)
	data.Set("grant_type", GrantType)
	data.Set("username", c.config.Email)
	data.Set("password", c.config.Password)
	data.Set("response_type", ResponseType)
	data.Set("client_info", "1")
	data.Set("x-client-SKU", "msal.js.browser")
	data.Set("x-client-VER", "2.32.1")
	data.Set("x-ms-lib-capability", "retry-after, h429")
	data.Set("x-client-current-telemetry", "5|61,0,,,|@azure/msal-react,1.5.1")
	data.Set("x-client-last-telemetry", "5|0|||0,0")
	data.Set("client-request-id", generateUUID())

	req, err := http.NewRequest("POST", TokenEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Origin", "https://www.costco.com")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://www.costco.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decoding token response: %w", err)
	}

	c.mu.Lock()
	c.token = &tokenResp
	c.tokenExpiry = c.calculateTokenExpiry(tokenResp.IDToken)
	c.mu.Unlock()

	// Save tokens to disk
	storedTokens := &StoredTokens{
		IDToken:               tokenResp.IDToken,
		RefreshToken:          tokenResp.RefreshToken,
		TokenExpiry:           c.tokenExpiry,
		RefreshTokenExpiresAt: time.Now().Add(time.Duration(tokenResp.RefreshTokenExpiresIn) * time.Second),
	}
	if err := SaveTokens(storedTokens); err != nil {
		// Log error but don't fail the auth
		fmt.Printf("Warning: Could not save tokens: %v\n", err)
	}

	return nil
}

func (c *Client) calculateTokenExpiry(tokenString string) time.Time {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return time.Now().Add(50 * time.Minute)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			return time.Unix(int64(exp), 0).Add(-c.config.TokenRefreshBuffer)
		}
	}

	return time.Now().Add(50 * time.Minute)
}

func (c *Client) refreshTokenIfNeeded() error {
	c.mu.RLock()
	needsRefresh := c.token == nil || time.Now().After(c.tokenExpiry)
	hasRefreshToken := c.token != nil && c.token.RefreshToken != ""
	c.mu.RUnlock()

	if !needsRefresh {
		return nil
	}

	if hasRefreshToken {
		return c.refreshToken()
	}

	return c.authenticate()
}

func (c *Client) refreshToken() error {
	c.mu.RLock()
	refreshToken := c.token.RefreshToken
	c.mu.RUnlock()

	data := url.Values{}
	data.Set("client_id", ClientID)
	data.Set("scope", Scope)
	data.Set("grant_type", RefreshGrantType)
	data.Set("client_info", "1")
	data.Set("x-client-SKU", "msal.js.browser")
	data.Set("x-client-VER", "2.32.1")
	data.Set("x-ms-lib-capability", "retry-after, h429")
	data.Set("x-client-current-telemetry", "5|61,0,,,|@azure/msal-react,1.5.1")
	data.Set("x-client-last-telemetry", "5|0|||0,0")
	data.Set("client-request-id", generateUUID())
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", TokenEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Origin", "https://www.costco.com")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://www.costco.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.authenticate()
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decoding refresh response: %w", err)
	}

	c.mu.Lock()
	c.token = &tokenResp
	c.tokenExpiry = c.calculateTokenExpiry(tokenResp.IDToken)
	c.mu.Unlock()

	// Save refreshed tokens to disk
	storedTokens := &StoredTokens{
		IDToken:               tokenResp.IDToken,
		RefreshToken:          tokenResp.RefreshToken,
		TokenExpiry:           c.tokenExpiry,
		RefreshTokenExpiresAt: time.Now().Add(time.Duration(tokenResp.RefreshTokenExpiresIn) * time.Second),
	}
	if err := SaveTokens(storedTokens); err != nil {
		// Log error but don't fail the refresh
		fmt.Printf("Warning: Could not save refreshed tokens: %v\n", err)
	}

	return nil
}

func (c *Client) executeGraphQL(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	if err := c.refreshTokenIfNeeded(); err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", GraphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	c.mu.RLock()
	token := c.token.IDToken
	c.mu.RUnlock()

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json-patch+json")
	req.Header.Set("DNT", "1")
	req.Header.Set("Origin", "https://www.costco.com")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://www.costco.com/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Set(HeaderClientIdentifier, ClientIdentifier)
	req.Header.Set(HeaderAuthorization, "Bearer "+token)
	req.Header.Set(HeaderWCSClientID, WCSClientID)
	req.Header.Set(HeaderCostcoEnv, CostcoEnvironment)
	req.Header.Set(HeaderCostcoService, CostcoService)
	req.Header.Set("sec-ch-ua", `"Chromium";v="139", "Not;A=Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var graphQLResp GraphQLResponse
	graphQLResp.Data = result

	if err := json.NewDecoder(resp.Body).Decode(&graphQLResp); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %v", graphQLResp.Errors)
	}

	return nil
}

func (c *Client) GetOnlineOrders(ctx context.Context, startDate, endDate string, pageNumber, pageSize int) (*OnlineOrdersResponse, error) {
	variables := map[string]interface{}{
		"startDate":       startDate,
		"endDate":         endDate,
		"pageNumber":      pageNumber,
		"pageSize":        pageSize,
		"warehouseNumber": c.config.WarehouseNumber,
	}

	var result struct {
		GetOnlineOrders []OnlineOrdersResponse `json:"getOnlineOrders"`
	}

	if err := c.executeGraphQL(ctx, OnlineOrdersQuery, variables, &result); err != nil {
		return nil, err
	}

	if len(result.GetOnlineOrders) == 0 {
		return nil, fmt.Errorf("no order data returned")
	}

	return &result.GetOnlineOrders[0], nil
}

func (c *Client) GetReceipts(ctx context.Context, startDate, endDate, documentType, documentSubType string) (*ReceiptsWithCountsResponse, error) {
	variables := map[string]interface{}{
		"startDate":       startDate,
		"endDate":         endDate,
		"documentType":    documentType,
		"documentSubType": documentSubType,
	}

	// Try as array first, similar to orders
	var resultArray struct {
		ReceiptsWithCounts []ReceiptsWithCountsResponse `json:"receiptsWithCounts"`
	}

	if err := c.executeGraphQL(ctx, ReceiptsQuery, variables, &resultArray); err != nil {
		// If array fails, try as object (backward compatibility)
		var resultObject struct {
			ReceiptsWithCounts ReceiptsWithCountsResponse `json:"receiptsWithCounts"`
		}
		if err2 := c.executeGraphQL(ctx, ReceiptsQuery, variables, &resultObject); err2 != nil {
			return nil, fmt.Errorf("failed to decode as array: %v, and as object: %v", err, err2)
		}
		return &resultObject.ReceiptsWithCounts, nil
	}

	if len(resultArray.ReceiptsWithCounts) == 0 {
		return nil, fmt.Errorf("no receipt data returned")
	}

	return &resultArray.ReceiptsWithCounts[0], nil
}

func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based UUID if random fails
		return fmt.Sprintf("%d-%d-%d-%d-%d",
			time.Now().Unix(),
			time.Now().UnixNano()%1000000,
			time.Now().UnixNano()%100000,
			time.Now().UnixNano()%10000,
			time.Now().UnixNano()%1000)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (c *Client) GetReceiptDetail(ctx context.Context, barcode, documentType string) (*Receipt, error) {
	variables := map[string]interface{}{
		"barcode":      barcode,
		"documentType": documentType,
	}

	var result struct {
		ReceiptsWithCounts struct {
			Receipts []Receipt `json:"receipts"`
		} `json:"receiptsWithCounts"`
	}

	if err := c.executeGraphQL(ctx, ReceiptDetailQuery, variables, &result); err != nil {
		return nil, err
	}

	if len(result.ReceiptsWithCounts.Receipts) == 0 {
		return nil, fmt.Errorf("no receipt found for barcode %s", barcode)
	}

	return &result.ReceiptsWithCounts.Receipts[0], nil
}
