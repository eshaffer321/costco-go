package costco

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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
	logger      *slog.Logger
}

// getLogger returns the client's logger or a no-op logger if none is set
func (c *Client) getLogger() *slog.Logger {
	if c.logger != nil {
		return c.logger
	}
	// Return a no-op logger that discards all output
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func NewClient(config Config) *Client {
	if config.TokenRefreshBuffer == 0 {
		config.TokenRefreshBuffer = 5 * time.Minute
	}

	// Initialize logger with no-op logger if not provided
	logger := config.Logger
	if logger == nil {
		// Use a no-op logger that discards all output
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	// Add "client=costco" attribute to all log messages
	logger = logger.With(slog.String("client", "costco"))

	client := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
		logger: logger,
	}

	// Try to load existing tokens
	if tokens, err := LoadTokens(); err == nil && tokens != nil {
		client.token = &TokenResponse{
			IDToken:      tokens.IDToken,
			RefreshToken: tokens.RefreshToken,
		}
		client.tokenExpiry = tokens.TokenExpiry
		logger.Info("token initialized from disk", slog.Time("token_expiry", client.tokenExpiry))
	}

	return client
}

func (c *Client) authenticate() error {
	c.getLogger().Debug("authenticating with costco", slog.String("email", c.config.Email))

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
		c.getLogger().Error("failed to create auth request", slog.String("error", err.Error()))
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

	c.getLogger().Debug("sending auth request", slog.String("endpoint", TokenEndpoint), slog.String("method", "POST"))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.getLogger().Error("auth request failed", slog.String("error", err.Error()))
		return fmt.Errorf("executing auth request: %w", err)
	}
	defer resp.Body.Close()

	c.getLogger().Debug("auth response received", slog.Int("status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.getLogger().Error("authentication failed", slog.Int("status_code", resp.StatusCode))
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		c.getLogger().Error("failed to decode token response", slog.String("error", err.Error()))
		return fmt.Errorf("decoding token response: %w", err)
	}

	c.mu.Lock()
	c.token = &tokenResp
	c.tokenExpiry = c.calculateTokenExpiry(tokenResp.IDToken)
	c.mu.Unlock()

	c.getLogger().Info("authenticated", slog.Time("token_expiry", c.tokenExpiry))

	// Save tokens to disk
	storedTokens := &StoredTokens{
		IDToken:               tokenResp.IDToken,
		RefreshToken:          tokenResp.RefreshToken,
		TokenExpiry:           c.tokenExpiry,
		RefreshTokenExpiresAt: time.Now().Add(time.Duration(tokenResp.RefreshTokenExpiresIn) * time.Second),
	}
	c.getLogger().Debug("saving tokens to disk")
	if err := SaveTokens(storedTokens); err != nil {
		c.getLogger().Warn("failed to save tokens", slog.String("error", err.Error()))
	} else {
		c.getLogger().Info("tokens saved successfully")
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
	tokenExpiry := c.tokenExpiry
	c.mu.RUnlock()

	if !needsRefresh {
		// Check if token is expiring soon
		timeUntilExpiry := time.Until(tokenExpiry)
		if timeUntilExpiry > 0 && timeUntilExpiry < 5*time.Minute {
			c.getLogger().Warn("token expiring soon", slog.Duration("time_until_expiry", timeUntilExpiry))
		}
		return nil
	}

	c.getLogger().Debug("token refresh needed", slog.Bool("has_refresh_token", hasRefreshToken))

	if hasRefreshToken {
		return c.refreshToken()
	}

	return c.authenticate()
}

func (c *Client) refreshToken() error {
	c.getLogger().Debug("refreshing token")

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
		c.getLogger().Error("failed to create refresh request", slog.String("error", err.Error()))
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

	c.getLogger().Debug("sending refresh request", slog.String("endpoint", TokenEndpoint), slog.String("method", "POST"))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.getLogger().Error("refresh request failed", slog.String("error", err.Error()))
		return fmt.Errorf("executing refresh request: %w", err)
	}
	defer resp.Body.Close()

	c.getLogger().Debug("refresh response received", slog.Int("status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		c.getLogger().Warn("token refresh failed, falling back to authentication", slog.Int("status_code", resp.StatusCode))
		return c.authenticate()
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		c.getLogger().Error("failed to decode refresh response", slog.String("error", err.Error()))
		return fmt.Errorf("decoding refresh response: %w", err)
	}

	c.mu.Lock()
	c.token = &tokenResp
	c.tokenExpiry = c.calculateTokenExpiry(tokenResp.IDToken)
	c.mu.Unlock()

	c.getLogger().Info("token refreshed", slog.Time("token_expiry", c.tokenExpiry))

	// Save refreshed tokens to disk
	storedTokens := &StoredTokens{
		IDToken:               tokenResp.IDToken,
		RefreshToken:          tokenResp.RefreshToken,
		TokenExpiry:           c.tokenExpiry,
		RefreshTokenExpiresAt: time.Now().Add(time.Duration(tokenResp.RefreshTokenExpiresIn) * time.Second),
	}
	c.getLogger().Debug("saving refreshed tokens to disk")
	if err := SaveTokens(storedTokens); err != nil {
		c.getLogger().Warn("failed to save refreshed tokens", slog.String("error", err.Error()))
	} else {
		c.getLogger().Info("refreshed tokens saved successfully")
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
		c.getLogger().Error("failed to marshal graphql request", slog.String("error", err.Error()))
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", GraphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		c.getLogger().Error("failed to create graphql request", slog.String("error", err.Error()))
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

	c.getLogger().Debug("sending graphql request", slog.String("endpoint", GraphQLEndpoint), slog.String("method", "POST"))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.getLogger().Error("graphql request failed", slog.String("error", err.Error()))
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	c.getLogger().Debug("graphql response received", slog.Int("status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.getLogger().Error("graphql request failed", slog.Int("status_code", resp.StatusCode))
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var graphQLResp GraphQLResponse
	graphQLResp.Data = result

	if err := json.NewDecoder(resp.Body).Decode(&graphQLResp); err != nil {
		c.getLogger().Debug("failed to decode graphql response", slog.String("error", err.Error()))
		return fmt.Errorf("decoding response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		c.getLogger().Warn("graphql errors in response", slog.Int("error_count", len(graphQLResp.Errors)))
		return fmt.Errorf("GraphQL errors: %v", graphQLResp.Errors)
	}

	return nil
}

func (c *Client) GetOnlineOrders(ctx context.Context, startDate, endDate string, pageNumber, pageSize int) (*OnlineOrdersResponse, error) {
	c.getLogger().Info("fetching online orders",
		slog.String("start_date", startDate),
		slog.String("end_date", endDate),
		slog.Int("page_number", pageNumber),
		slog.Int("page_size", pageSize))

	variables := map[string]interface{}{
		"startDate":       startDate,
		"endDate":         endDate,
		"pageNumber":      pageNumber,
		"pageSize":        pageSize,
		"warehouseNumber": c.config.WarehouseNumber,
	}

	c.getLogger().Debug("executing graphql query", slog.String("operation", "getOnlineOrders"))

	var result struct {
		GetOnlineOrders []OnlineOrdersResponse `json:"getOnlineOrders"`
	}

	if err := c.executeGraphQL(ctx, OnlineOrdersQuery, variables, &result); err != nil {
		return nil, err
	}

	if len(result.GetOnlineOrders) == 0 {
		return nil, fmt.Errorf("no order data returned")
	}

	orderCount := len(result.GetOnlineOrders[0].BCOrders)
	c.getLogger().Info("fetched online orders",
		slog.Int("order_count", orderCount),
		slog.String("date_range", startDate+" to "+endDate))

	return &result.GetOnlineOrders[0], nil
}

func (c *Client) GetReceipts(ctx context.Context, startDate, endDate, documentType, documentSubType string) (*ReceiptsWithCountsResponse, error) {
	c.getLogger().Info("fetching receipts",
		slog.String("start_date", startDate),
		slog.String("end_date", endDate),
		slog.String("document_type", documentType))

	variables := map[string]interface{}{
		"startDate":       startDate,
		"endDate":         endDate,
		"documentType":    documentType,
		"documentSubType": documentSubType,
	}

	c.getLogger().Debug("executing graphql query", slog.String("operation", "receiptsWithCounts"))

	// Try object format first (this is what Costco's API currently returns)
	var resultObject struct {
		ReceiptsWithCounts ReceiptsWithCountsResponse `json:"receiptsWithCounts"`
	}

	if err := c.executeGraphQL(ctx, ReceiptsQuery, variables, &resultObject); err != nil {
		// TODO: If this fallback is never hit over time, we can remove the array format code entirely.
		// The array format may have been from API changes or incorrect assumptions during initial development.
		// Monitor logs for the "🚨 ARRAY FALLBACK" message - if it never appears, delete this fallback code.
		c.getLogger().Warn("🚨🚨🚨 OBJECT FORMAT FAILED - attempting ARRAY format fallback 🚨🚨🚨",
			slog.String("object_error", err.Error()),
			slog.String("document_type", documentType))

		var resultArray struct {
			ReceiptsWithCounts []ReceiptsWithCountsResponse `json:"receiptsWithCounts"`
		}
		if err2 := c.executeGraphQL(ctx, ReceiptsQuery, variables, &resultArray); err2 != nil {
			return nil, fmt.Errorf("failed to decode as object: %v, and as array: %v", err, err2)
		}

		if len(resultArray.ReceiptsWithCounts) == 0 {
			return nil, fmt.Errorf("no receipt data returned")
		}

		receiptCount := len(resultArray.ReceiptsWithCounts[0].Receipts)
		c.getLogger().Warn("✅✅✅ ARRAY FALLBACK SUCCEEDED! Array format worked! (DO NOT DELETE THIS CODE) ✅✅✅",
			slog.Int("receipt_count", receiptCount),
			slog.String("document_type", documentType))
		return &resultArray.ReceiptsWithCounts[0], nil
	}

	receiptCount := len(resultObject.ReceiptsWithCounts.Receipts)
	c.getLogger().Info("fetched receipts",
		slog.Int("receipt_count", receiptCount),
		slog.String("document_type", documentType))

	return &resultObject.ReceiptsWithCounts, nil
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
	c.getLogger().Info("fetching receipt detail",
		slog.String("barcode", barcode),
		slog.String("document_type", documentType))

	variables := map[string]interface{}{
		"barcode":      barcode,
		"documentType": documentType,
	}

	c.getLogger().Debug("executing graphql query", slog.String("operation", "getReceiptDetail"))

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

	receipt := &result.ReceiptsWithCounts.Receipts[0]
	c.getLogger().Info("fetched receipt detail",
		slog.String("barcode", barcode),
		slog.String("document_type", documentType),
		slog.Int("item_count", len(receipt.ItemArray)),
		slog.Float64("total", receipt.Total))

	return receipt, nil
}
