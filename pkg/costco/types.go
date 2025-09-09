package costco

import "time"

type TokenResponse struct {
	IDToken               string `json:"id_token"`
	TokenType             string `json:"token_type"`
	NotBefore             int64  `json:"not_before"`
	ClientInfo            string `json:"client_info"`
	Scope                 string `json:"scope"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
}

type OnlineOrder struct {
	OrderHeaderID      string          `json:"orderHeaderId"`
	OrderPlacedDate    string          `json:"orderPlacedDate"`
	OrderNumber        string          `json:"orderNumber"`
	OrderTotal         float64         `json:"orderTotal"`
	WarehouseNumber    string          `json:"warehouseNumber"`
	Status             string          `json:"status"`
	EmailAddress       string          `json:"emailAddress"`
	OrderCancelAllowed bool            `json:"orderCancelAllowed"`
	OrderPaymentFailed bool            `json:"orderPaymentFailed"`
	OrderReturnAllowed bool            `json:"orderReturnAllowed"`
	OrderLineItems     []OrderLineItem `json:"orderLineItems"`
}

type OrderLineItem struct {
	OrderLineItemCancelAllowed bool      `json:"orderLineItemCancelAllowed"`
	OrderLineItemID            string    `json:"orderLineItemId"`
	OrderReturnAllowed         bool      `json:"orderReturnAllowed"`
	ItemID                     string    `json:"itemId"`
	ItemNumber                 string    `json:"itemNumber"`
	ItemTypeID                 string    `json:"itemTypeId"`
	LineNumber                 int       `json:"lineNumber"`
	ItemDescription            string    `json:"itemDescription"`
	DeliveryDate               string    `json:"deliveryDate"`
	WarehouseNumber            string    `json:"warehouseNumber"`
	Status                     string    `json:"status"`
	OrderStatus                string    `json:"orderStatus"`
	ParentOrderLineItemID      string    `json:"parentOrderLineItemId"`
	IsFSAEligible              bool      `json:"isFSAEligible"`
	ShippingType               string    `json:"shippingType"`
	ShippingTimeFrame          string    `json:"shippingTimeFrame"`
	IsShipToWarehouse          bool      `json:"isShipToWarehouse"`
	CarrierItemCategory        string    `json:"carrierItemCategory"`
	CarrierContactPhone        string    `json:"carrierContactPhone"`
	ProgramTypeID              string    `json:"programTypeId"`
	IsBuyAgainEligible         bool      `json:"isBuyAgainEligible"`
	ScheduledDeliveryDate      string    `json:"scheduledDeliveryDate"`
	ScheduledDeliveryDateEnd   string    `json:"scheduledDeliveryDateEnd"`
	ConfiguredItemData         string    `json:"configuredItemData"`
	Shipment                   *Shipment `json:"shipment"`
}

type Shipment struct {
	ShipmentID                     string         `json:"shipmentId"`
	OrderHeaderID                  string         `json:"orderHeaderId"`
	OrderShipToID                  string         `json:"orderShipToId"`
	LineNumber                     int            `json:"lineNumber"`
	OrderNumber                    string         `json:"orderNumber"`
	ShippingType                   string         `json:"shippingType"`
	ShippingTimeFrame              string         `json:"shippingTimeFrame"`
	ShippedDate                    string         `json:"shippedDate"`
	PackageNumber                  string         `json:"packageNumber"`
	TrackingNumber                 string         `json:"trackingNumber"`
	TrackingSiteURL                string         `json:"trackingSiteUrl"`
	CarrierName                    string         `json:"carrierName"`
	EstimatedArrivalDate           string         `json:"estimatedArrivalDate"`
	DeliveredDate                  string         `json:"deliveredDate"`
	IsDeliveryDelayed              bool           `json:"isDeliveryDelayed"`
	IsEstimatedArrivalDateEligible bool           `json:"isEstimatedArrivalDateEligible"`
	StatusTypeID                   string         `json:"statusTypeId"`
	Status                         string         `json:"status"`
	PickUpReadyDate                string         `json:"pickUpReadyDate"`
	PickUpCompletedDate            string         `json:"pickUpCompletedDate"`
	ReasonCode                     string         `json:"reasonCode"`
	TrackingEvent                  *TrackingEvent `json:"trackingEvent"`
}

type TrackingEvent struct {
	Event                 string `json:"event"`
	CarrierName           string `json:"carrierName"`
	EventDate             string `json:"eventDate"`
	EstimatedDeliveryDate string `json:"estimatedDeliveryDate"`
	ScheduledDeliveryDate string `json:"scheduledDeliveryDate"`
	TrackingNumber        string `json:"trackingNumber"`
}

type OnlineOrdersResponse struct {
	PageNumber           int           `json:"pageNumber"`
	PageSize             int           `json:"pageSize"`
	TotalNumberOfRecords int           `json:"totalNumberOfRecords"`
	BCOrders             []OnlineOrder `json:"bcOrders"`
}

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

type ReceiptsWithCountsResponse struct {
	InWarehouse   int       `json:"inWarehouse"`
	GasStation    int       `json:"gasStation"`
	CarWash       int       `json:"carWash"`
	GasAndCarWash int       `json:"gasAndCarWash"`
	Receipts      []Receipt `json:"receipts"`
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
	Data   interface{} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type OrdersQueryVariables struct {
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	PageNumber      int    `json:"pageNumber"`
	PageSize        int    `json:"pageSize"`
	WarehouseNumber string `json:"warehouseNumber"`
}

type ReceiptsQueryVariables struct {
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	DocumentType    string `json:"documentType"`
	DocumentSubType string `json:"documentSubType"`
}

type ReceiptDetailQueryVariables struct {
	Barcode      string `json:"barcode"`
	DocumentType string `json:"documentType"`
}

type Config struct {
	Email              string
	Password           string
	WarehouseNumber    string
	TokenRefreshBuffer time.Duration
}
