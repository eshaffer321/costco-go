package costco

// GraphQL-related types for API communication

// GraphQLRequest represents a GraphQL request sent to the Costco API
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// GraphQLResponse represents a GraphQL response from the Costco API
type GraphQLResponse struct {
	Data   interface{} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// OrdersQueryVariables represents the variables for the online orders GraphQL query
type OrdersQueryVariables struct {
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	PageNumber      int    `json:"pageNumber"`
	PageSize        int    `json:"pageSize"`
	WarehouseNumber string `json:"warehouseNumber"`
}

// ReceiptsQueryVariables represents the variables for the receipts GraphQL query
type ReceiptsQueryVariables struct {
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	DocumentType    string `json:"documentType"`
	DocumentSubType string `json:"documentSubType"`
}

// ReceiptDetailQueryVariables represents the variables for the receipt detail GraphQL query
type ReceiptDetailQueryVariables struct {
	Barcode      string `json:"barcode"`
	DocumentType string `json:"documentType"`
}
