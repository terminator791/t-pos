package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// SyncRequest represents the complete sync request from mobile client
type SyncRequest struct {
	LastSyncTimestamp   *time.Time                    `json:"last_sync_timestamp"`
	Carts               []entities.Cart               `json:"carts"`
	Categories          []entities.Category           `json:"categories"`
	Expenses            []entities.Expense            `json:"expenses"`
	Histories           []entities.History            `json:"histories"`
	Payments            []entities.Payment            `json:"payments"`
	Products            []entities.Product            `json:"products"`
	Receipts            []entities.Receipt            `json:"receipts"`
	Shops               []entities.Shop               `json:"shops"`
	StockHistories      []entities.StockHistory       `json:"stock_histories"`
	TransactionProducts []entities.TransactionProduct `json:"transaction_products"`
	Transactions        []entities.Transaction        `json:"transactions"`
	Users               []entities.User               `json:"users"`
}

// SyncResponse represents the sync response sent back to mobile client
type SyncResponse struct {
	SyncTimestamp       time.Time                     `json:"sync_timestamp"`
	Carts               []entities.Cart               `json:"carts,omitempty"`
	Categories          []entities.Category           `json:"categories,omitempty"`
	Expenses            []entities.Expense            `json:"expenses,omitempty"`
	Histories           []entities.History            `json:"histories,omitempty"`
	Payments            []entities.Payment            `json:"payments,omitempty"`
	Products            []entities.Product            `json:"products,omitempty"`
	Receipts            []entities.Receipt            `json:"receipts,omitempty"`
	Shops               []entities.Shop               `json:"shops,omitempty"`
	StockHistories      []entities.StockHistory       `json:"stock_histories,omitempty"`
	TransactionProducts []entities.TransactionProduct `json:"transaction_products,omitempty"`
	Transactions        []entities.Transaction        `json:"transactions,omitempty"`
	Users               []entities.User               `json:"users,omitempty"`
	Conflicts           []ConflictInfo                `json:"conflicts,omitempty"`
	Errors              []SyncError                   `json:"errors,omitempty"`
	Stats               SyncStats                     `json:"stats"`
}

// ConflictInfo represents information about a data conflict during sync
type ConflictInfo struct {
	EntityType   string      `json:"entity_type"`
	EntityID     uuid.UUID   `json:"entity_id"`
	ConflictType string      `json:"conflict_type"`
	Resolution   string      `json:"resolution"`
	Details      string      `json:"details"`
	ServerData   interface{} `json:"server_data,omitempty"`
	ClientData   interface{} `json:"client_data,omitempty"`
}

// CleanConflictData creates a clean version of conflict data without nested empty relations
type CleanConflictData struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name,omitempty"`
	ShopID    *uuid.UUID             `json:"shop_id,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Extra     map[string]interface{} `json:",inline"`
}

// SyncError represents an error that occurred during sync for a specific entity
type SyncError struct {
	EntityType string    `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
	ErrorCode  string    `json:"error_code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
}

// SyncStats provides statistics about the sync operation
type SyncStats struct {
	ProcessedEntities map[string]int `json:"processed_entities"`
	CreatedEntities   map[string]int `json:"created_entities"`
	UpdatedEntities   map[string]int `json:"updated_entities"`
	ConflictCount     int            `json:"conflict_count"`
	ErrorCount        int            `json:"error_count"`
	ProcessingTimeMs  int64          `json:"processing_time_ms"`
}

// SyncMetadata stores sync tracking information
type SyncMetadata struct {
	UserID        uuid.UUID  `json:"user_id"`
	LicenseID     uuid.UUID  `json:"license_id"`
	LastSyncAt    *time.Time `json:"last_sync_at"`
	DeviceID      *string    `json:"device_id"`
	AppVersion    *string    `json:"app_version"`
	SyncVersion   string     `json:"sync_version"`
	TotalSyncs    int        `json:"total_syncs"`
	LastSyncStats SyncStats  `json:"last_sync_stats"`
}

// Syncable interface defines common methods for all syncable entities
type Syncable interface {
	GetID() uuid.UUID
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
}

// ConflictResolutionStrategy defines how conflicts should be resolved
type ConflictResolutionStrategy string

const (
	// LastWriteWins resolves conflicts by choosing the entity with latest updated_at
	LastWriteWins ConflictResolutionStrategy = "last_write_wins"
	// ServerWins always chooses the server version
	ServerWins ConflictResolutionStrategy = "server_wins"
	// ClientWins always chooses the client version
	ClientWins ConflictResolutionStrategy = "client_wins"
)

// SyncJob represents an async sync job
type SyncJob struct {
	ID          uuid.UUID     `json:"id"`
	UserID      uuid.UUID     `json:"user_id"`
	LicenseID   uuid.UUID     `json:"license_id"`
	Status      SyncStatus    `json:"status"`
	Request     SyncRequest   `json:"request"`
	Response    *SyncResponse `json:"response,omitempty"`
	Error       *string       `json:"error,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	StartedAt   *time.Time    `json:"started_at,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
}

// SyncStatus represents the status of a sync job
type SyncStatus string

const (
	SyncStatusPending    SyncStatus = "pending"
	SyncStatusProcessing SyncStatus = "processing"
	SyncStatusCompleted  SyncStatus = "completed"
	SyncStatusFailed     SyncStatus = "failed"
)
