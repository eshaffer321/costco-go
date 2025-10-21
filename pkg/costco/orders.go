package costco

// Order-related types for Costco online orders

// OnlineOrder represents a single online order from Costco.com
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

// OrderLineItem represents a single line item within an online order
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

// Shipment represents shipping information for an order line item
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

// TrackingEvent represents a tracking event for a shipment
type TrackingEvent struct {
	Event                 string `json:"event"`
	CarrierName           string `json:"carrierName"`
	EventDate             string `json:"eventDate"`
	EstimatedDeliveryDate string `json:"estimatedDeliveryDate"`
	ScheduledDeliveryDate string `json:"scheduledDeliveryDate"`
	TrackingNumber        string `json:"trackingNumber"`
}

// OnlineOrdersResponse represents the response from the online orders query
type OnlineOrdersResponse struct {
	PageNumber           int           `json:"pageNumber"`
	PageSize             int           `json:"pageSize"`
	TotalNumberOfRecords int           `json:"totalNumberOfRecords"`
	BCOrders             []OnlineOrder `json:"bcOrders"`
}
