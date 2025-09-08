package constants

// HTTP Status Messages
const (
	StatusSuccess = "success"
	StatusError   = "error"
	StatusFailed  = "failed"
)

// Default Values
const (
	DefaultPageSize     = 20
	MaxPageSize         = 100
	DefaultPage         = 1
	DefaultTaxRate      = 8.25  // 8.25%
	DefaultCurrency     = "USD"
	DefaultCountry      = "US"
	DefaultTimeZone     = "UTC"
)

// Order Constants
const (
	OrderNumberPrefix = "ORD"
	MinOrderTotal     = 0.01
	MaxOrderTotal     = 999999.99
)

// Payment Constants
const (
	TransactionIDPrefix = "TXN"
	MinPaymentAmount    = 0.01
	MaxPaymentAmount    = 999999.99
)

// Stock Constants
const (
	MinStockLevel     = 0
	DefaultMinStock   = 5
	DefaultMaxStock   = 1000
	StockUnit         = "pcs"
)

// User Constants
const (
	MinPasswordLength = 6
	MaxPasswordLength = 100
	MinUsernameLength = 3
	MaxUsernameLength = 50
)

// Product Constants
const (
	MaxProductNameLength    = 200
	MaxDescriptionLength    = 1000
	MaxSKULength           = 50
	MaxBarcodeLength       = 50
)

// Customer Constants
const (
	MaxCustomerNameLength = 100
	MaxAddressLength      = 500
	MaxPhoneLength        = 20
)

// API Response Messages
const (
	MsgSuccess           = "Operation completed successfully"
	MsgCreated           = "Resource created successfully"
	MsgUpdated           = "Resource updated successfully"
	MsgDeleted           = "Resource deleted successfully"
	MsgNotFound          = "Resource not found"
	MsgValidationError   = "Validation error"
	MsgUnauthorized      = "Unauthorized access"
	MsgForbidden         = "Access forbidden"
	MsgInternalError     = "Internal server error"
	MsgBadRequest        = "Bad request"
	MsgConflict          = "Resource already exists"
)

// Database Table Names
const (
	TableUsers      = "users"
	TableCustomers  = "customers"
	TableCategories = "categories"
	TableProducts   = "products"
	TableOrders     = "orders"
	TableOrderItems = "order_items"
	TablePayments   = "payments"
)

// Cache Keys
const (
	CacheKeyUser     = "user:"
	CacheKeyProduct  = "product:"
	CacheKeyOrder    = "order:"
	CacheKeyCustomer = "customer:"
	CacheKeyCategory = "category:"
)

// File Upload Constants
const (
	MaxFileSize      = 10 << 20 // 10 MB
	AllowedImageExt  = "jpg,jpeg,png,gif,webp"
	UploadPath       = "uploads/"
	ProductImagePath = "uploads/products/"
)

// JWT Constants
const (
	JWTIssuer           = "t-pos"
	JWTAudience         = "t-pos-api"
	DefaultTokenExpiry  = 24 * 60 * 60 // 24 hours in seconds
	RefreshTokenExpiry  = 7 * 24 * 60 * 60 // 7 days in seconds
)

// Rate Limiting
const (
	RateLimitPerMinute = 60
	RateLimitPerHour   = 1000
	RateLimitPerDay    = 10000
)

// Date Formats
const (
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = "2006-01-02 15:04:05"
	APIDateFormat  = "2006-01-02T15:04:05Z"
)

// Validation Patterns
const (
	EmailPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	PhonePattern    = `^\+?[1-9]\d{1,14}$`
	SKUPattern      = `^[A-Z0-9-_]{3,50}$`
	BarcodePattern  = `^[0-9]{8,13}$`
)

// Error Codes
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeForbidden     = "FORBIDDEN"
	ErrCodeConflict      = "CONFLICT"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeBadRequest    = "BAD_REQUEST"
	ErrCodeTimeout       = "TIMEOUT"
	ErrCodeRateLimit     = "RATE_LIMIT_EXCEEDED"
)