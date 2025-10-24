package costco

import "strings"

// Receipt-related types for Costco warehouse and online receipts

// Receipt represents a single receipt from a Costco transaction
type Receipt struct {
	WarehouseName       string        `json:"warehouseName"`
	ReceiptType         string        `json:"receiptType"`
	DocumentType        string        `json:"documentType"`
	TransactionDateTime string        `json:"transactionDateTime"`
	TransactionDate     string        `json:"transactionDate"`
	CompanyNumber       int           `json:"companyNumber"`
	WarehouseNumber     int           `json:"warehouseNumber"`
	OperatorNumber      int           `json:"operatorNumber"`
	WarehouseShortName  string        `json:"warehouseShortName"`
	RegisterNumber      int           `json:"registerNumber"`
	TransactionNumber   int           `json:"transactionNumber"`
	TransactionType     string        `json:"transactionType"`
	TransactionBarcode  string        `json:"transactionBarcode"`
	Total               float64       `json:"total"`
	WarehouseAddress1   string        `json:"warehouseAddress1"`
	WarehouseAddress2   string        `json:"warehouseAddress2"`
	WarehouseCity       string        `json:"warehouseCity"`
	WarehouseState      string        `json:"warehouseState"`
	WarehouseCountry    string        `json:"warehouseCountry"`
	WarehousePostalCode string        `json:"warehousePostalCode"`
	TotalItemCount      int           `json:"totalItemCount"`
	SubTotal            float64       `json:"subTotal"`
	Taxes               float64       `json:"taxes"`
	InvoiceNumber       interface{}   `json:"invoiceNumber"`  // Can be string or number for fuel receipts
	SequenceNumber      interface{}   `json:"sequenceNumber"` // Can be string or number for fuel receipts
	ItemArray           []ReceiptItem `json:"itemArray"`
	TenderArray         []Tender      `json:"tenderArray"`
	SubTaxes            *SubTaxes     `json:"subTaxes"`
	InstantSavings      float64       `json:"instantSavings"`
	MembershipNumber    string        `json:"membershipNumber"`
}

// ReceiptItem represents a single line item on a receipt
type ReceiptItem struct {
	ItemNumber             string  `json:"itemNumber"`
	ItemDescription01      string  `json:"itemDescription01"`
	FrenchItemDescription1 string  `json:"frenchItemDescription1"`
	ItemDescription02      string  `json:"itemDescription02"`
	FrenchItemDescription2 string  `json:"frenchItemDescription2"`
	ItemIdentifier         string  `json:"itemIdentifier"`
	ItemDepartmentNumber   int     `json:"itemDepartmentNumber"`
	Unit                   int     `json:"unit"`
	Amount                 float64 `json:"amount"`
	TaxFlag                string  `json:"taxFlag"`
	MerchantID             string  `json:"merchantID"`
	EntryMethod            string  `json:"entryMethod"`
	TransDepartmentNumber  int     `json:"transDepartmentNumber"`
	FuelUnitQuantity       float64 `json:"fuelUnitQuantity"`
	FuelGradeCode          string  `json:"fuelGradeCode"`
	ItemUnitPriceAmount    float64 `json:"itemUnitPriceAmount"`
	FuelUomCode            string  `json:"fuelUomCode"`
	FuelUomDescription     string  `json:"fuelUomDescription"`
	FuelUomDescriptionFr   string  `json:"fuelUomDescriptionFr"`
	FuelGradeDescription   string  `json:"fuelGradeDescription"`
	FuelGradeDescriptionFr string  `json:"fuelGradeDescriptionFr"`
}

// IsDiscount returns true if this line item represents a discount applied to another item.
// Discount items have:
//   - Negative amount and negative unit
//   - Description starting with "/" followed by the parent item number (e.g., "/1553261")
//
// Note: Returns (refunds) also have negative amounts but have normal descriptions
// and appear in receipts with TransactionType: "Refund". This method will return
// false for return items since they don't have the "/" prefix.
//
// Example:
//
//	for _, item := range receipt.ItemArray {
//	    if item.IsDiscount() {
//	        fmt.Printf("Discount of $%.2f on item %s\n",
//	            math.Abs(item.Amount),
//	            item.GetParentItemNumber())
//	        continue
//	    }
//	    // Process regular items...
//	}
func (item *ReceiptItem) IsDiscount() bool {
	return item.Amount < 0 &&
		item.Unit < 0 &&
		strings.HasPrefix(item.ItemDescription01, "/")
}

// GetParentItemNumber returns the item number this discount applies to.
// For discount items with descriptions like "/1553261" or "/ 1857091" (with spaces),
// this returns "1553261" or "1857091" respectively.
// Returns empty string if this is not a discount item.
//
// Example:
//
//	if item.IsDiscount() {
//	    parentItemNum := item.GetParentItemNumber()
//	    // Use parentItemNum to find the item this discount applies to
//	}
func (item *ReceiptItem) GetParentItemNumber() string {
	if !item.IsDiscount() {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(item.ItemDescription01, "/"))
}

// Tender represents payment information on a receipt
type Tender struct {
	TenderTypeCode               string  `json:"tenderTypeCode"`
	TenderSubTypeCode            string  `json:"tenderSubTypeCode"`
	TenderDescription            string  `json:"tenderDescription"`
	AmountTender                 float64 `json:"amountTender"`
	DisplayAccountNumber         string  `json:"displayAccountNumber"`
	SequenceNumber               string  `json:"sequenceNumber"`
	ApprovalNumber               string  `json:"approvalNumber"`
	ResponseCode                 string  `json:"responseCode"`
	TenderTypeName               string  `json:"tenderTypeName"`
	TransactionID                string  `json:"transactionID"`
	MerchantID                   string  `json:"merchantID"`
	EntryMethod                  string  `json:"entryMethod"`
	TenderAcctTxnNumber          string  `json:"tenderAcctTxnNumber"`
	TenderAuthorizationCode      string  `json:"tenderAuthorizationCode"`
	TenderTypeNameFr             string  `json:"tenderTypeNameFr"`
	TenderEntryMethodDescription string  `json:"tenderEntryMethodDescription"`
	WalletType                   string  `json:"walletType"`
	WalletID                     string  `json:"walletId"`
	StoredValueBucket            string  `json:"storedValueBucket"`
}

// SubTaxes represents detailed tax breakdown on a receipt
type SubTaxes struct {
	Tax1               float64 `json:"tax1"`
	Tax2               float64 `json:"tax2"`
	Tax3               float64 `json:"tax3"`
	Tax4               float64 `json:"tax4"`
	ATaxPercent        float64 `json:"aTaxPercent"`
	ATaxLegend         string  `json:"aTaxLegend"`
	ATaxAmount         float64 `json:"aTaxAmount"`
	ATaxPrintCode      string  `json:"aTaxPrintCode"`
	ATaxPrintCodeFR    string  `json:"aTaxPrintCodeFR"`
	ATaxIdentifierCode string  `json:"aTaxIdentifierCode"`
	BTaxPercent        float64 `json:"bTaxPercent"`
	BTaxLegend         string  `json:"bTaxLegend"`
	BTaxAmount         float64 `json:"bTaxAmount"`
	BTaxPrintCode      string  `json:"bTaxPrintCode"`
	BTaxPrintCodeFR    string  `json:"bTaxPrintCodeFR"`
	BTaxIdentifierCode string  `json:"bTaxIdentifierCode"`
	CTaxPercent        float64 `json:"cTaxPercent"`
	CTaxLegend         string  `json:"cTaxLegend"`
	CTaxAmount         float64 `json:"cTaxAmount"`
	CTaxIdentifierCode string  `json:"cTaxIdentifierCode"`
	DTaxPercent        float64 `json:"dTaxPercent"`
	DTaxLegend         string  `json:"dTaxLegend"`
	DTaxAmount         float64 `json:"dTaxAmount"`
	DTaxPrintCode      string  `json:"dTaxPrintCode"`
	DTaxPrintCodeFR    string  `json:"dTaxPrintCodeFR"`
	DTaxIdentifierCode string  `json:"dTaxIdentifierCode"`
	UTaxLegend         string  `json:"uTaxLegend"`
	UTaxAmount         float64 `json:"uTaxAmount"`
	UTaxableAmount     float64 `json:"uTaxableAmount"`
}

// ReceiptsWithCountsResponse represents the response from the receipts query
type ReceiptsWithCountsResponse struct {
	InWarehouse   int       `json:"inWarehouse"`
	GasStation    int       `json:"gasStation"`
	CarWash       int       `json:"carWash"`
	GasAndCarWash int       `json:"gasAndCarWash"`
	Receipts      []Receipt `json:"receipts"`
}
