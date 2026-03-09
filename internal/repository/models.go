package repository

import (
	"time"

	"github.com/shopspring/decimal"
)

// UploadStatus represents the status of an upload
type UploadStatus string

const (
	UploadStatusReceived   UploadStatus = "RECEIVED"
	UploadStatusProcessing UploadStatus = "PROCESSING"
	UploadStatusDone       UploadStatus = "DONE"
	UploadStatusFailed     UploadStatus = "FAILED"
)

// Upload represents a file upload record
type Upload struct {
	ID                   int64
	TgUserID             int64
	TgChatID             int64
	TelegramMessageID    *int64
	TelegramFileID       *string
	TelegramFileUniqueID *string
	OriginalFilename     *string
	MimeType             *string
	FileSize             *int64
	Status               UploadStatus
	ErrorMessage         *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// RawFile represents stored file metadata
type RawFile struct {
	ID          int64
	UploadID    int64
	SHA256      string
	StorageKind string
	StorageKey  string
	StoredAt    time.Time
}

// ValuationSnapshot represents portfolio total value for a period
type ValuationSnapshot struct {
	ID              int64
	AccountID       int64
	Period          string // "2025-10"
	TotalValue      decimal.Decimal
	CashBalance     decimal.Decimal
	SecuritiesValue decimal.Decimal
	Currency        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// PositionSnapshot represents a single position in portfolio for a period
type PositionSnapshot struct {
	ID           int64
	AccountID    int64
	Period       string // "2025-10"
	ISIN         string
	SecurityName string
	Quantity     decimal.Decimal
	Price        decimal.Decimal
	MarketValue  decimal.Decimal
	Currency     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TgUser represents a Telegram user
type TgUser struct {
	ID             int64
	TelegramUserID int64
	CreatedAt      time.Time
}

// TgChat represents a Telegram chat
type TgChat struct {
	ID             int64
	TelegramChatID int64
	CreatedAt      time.Time
}

// Account represents a broker account
type Account struct {
	ID          int64
	Name        string
	Institution string
	Type        string
	Currency    string
	CreatedAt   time.Time
}

// Portfolio represents a portfolio (real or virtual)
type Portfolio struct {
	ID        int64
	UserID    int64
	Name      string
	Kind      string // "real" or "virtual"
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PortfolioAccount represents the mapping between portfolio and accounts
type PortfolioAccount struct {
	ID          int64
	PortfolioID int64
	AccountID   int64
	CreatedAt   time.Time
}

// CashFlowOperationType represents the type of cash flow operation
type CashFlowOperationType string

const (
	CashFlowOpTypeDeposit          CashFlowOperationType = "DEPOSIT"
	CashFlowOpTypeWithdrawal       CashFlowOperationType = "WITHDRAWAL"
	CashFlowOpTypeSecurityPurchase CashFlowOperationType = "SECURITY_PURCHASE"
	CashFlowOpTypeSecuritySale     CashFlowOperationType = "SECURITY_SALE"
	CashFlowOpTypeDividend         CashFlowOperationType = "DIVIDEND"
	CashFlowOpTypeTax              CashFlowOperationType = "TAX"
	CashFlowOpTypeFee              CashFlowOperationType = "FEE"
	CashFlowOpTypeOther            CashFlowOperationType = "OTHER"
)

// CashFlowOperation represents a cash flow operation in the database
type CashFlowOperation struct {
	ID            int64
	AccountID     int64
	Period        string // "2025-10"
	OperationDate time.Time
	Amount        decimal.Decimal
	Currency      string
	OperationType CashFlowOperationType
	Comment       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
