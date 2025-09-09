package costco

// API Endpoints
const (
	TokenEndpoint   = "https://signin.costco.com/e0714dd4-784d-46d6-a278-3e29553483eb/b2c_1a_sso_wcs_signup_signin_157/oauth2/v2.0/token"
	GraphQLEndpoint = "https://ecom-api.costco.com/ebusiness/order/v1/orders/graphql"
)

// OAuth2/OIDC Configuration
const (
	ClientID         = "a3a5186b-7c89-4b4c-93a8-dd604e930757" // Public OAuth2 client ID
	ClientIdentifier = "481b1aec-aa3b-454b-b81b-48187e28f205" // Public API client identifier
	WCSClientID      = "4900eb1f-0c10-4bd9-99c3-c59e6c1ecebf" // Public WCS client ID
	Scope            = "openid offline_access"
	GrantType        = "password"
	RefreshGrantType = "refresh_token"
	ResponseType     = "token id_token"
)

// MSAL Library Configuration (Microsoft Authentication Library)
const (
	MSALClientSKU        = "msal.js.browser"
	MSALClientVersion    = "2.32.1"
	MSALLibCapability    = "retry-after, h429"
	MSALCurrentTelemetry = "5|61,0,,,|@azure/msal-react,1.5.1"
	MSALLastTelemetry    = "5|0|||0,0"
)

// HTTP Headers
const (
	HeaderContentType      = "Content-Type"
	HeaderContentTypeJSON  = "application/json-patch+json"
	HeaderContentTypeForm  = "application/x-www-form-urlencoded;charset=utf-8"
	HeaderAuthorization    = "costco-x-authorization"
	HeaderClientIdentifier = "client-identifier"
	HeaderWCSClientID      = "costco-x-wcs-clientId"
	HeaderCostcoEnv        = "costco.env"
	HeaderCostcoService    = "costco.service"
	HeaderUserAgent        = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"
)

// Costco Service Configuration
const (
	CostcoEnvironment = "ecom"
	CostcoService     = "restOrders"
)

// Default Values
const (
	DefaultWarehouse = "847"
	DefaultPageSize  = 10
	DefaultTimeout   = 30 // seconds
)
