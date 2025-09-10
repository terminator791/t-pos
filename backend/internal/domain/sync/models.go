package sync

import (
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// SyncMetadata represents metadata for sync operations
type SyncMetadata struct {
	LastSyncTimestamp *time.Time `json:"last_sync_timestamp,omitempty"`
	SyncTimestamp     time.Time  `json:"sync_timestamp"`
	ClientID          string     `json:"client_id,omitempty"`
	LicenseID         uuid.UUID  `json:"license_id"`
	UserID            uuid.UUID  `json:"user_id"`
	ShopID            *uuid.UUID `json:"shop_id,omitempty"` // For cashier role restrictions
}

// SyncRequest represents the payload for push sync operations
type SyncRequest struct {
	Metadata SyncMetadata `json:"metadata"`
	Data     SyncData     `json:"data"`
}

// SyncResponse represents the response for sync operations
type SyncResponse struct {
	Success           bool         `json:"success"`
	Message           string       `json:"message"`
	Data              *SyncData    `json:"data,omitempty"`
	SyncTimestamp     time.Time    `json:"sync_timestamp"`
	ConflictsResolved int          `json:"conflicts_resolved,omitempty"`
	RecordsProcessed  int          `json:"records_processed"`
	Errors            []SyncError  `json:"errors,omitempty"`
}

// SyncData contains all entities that can be synchronized
type SyncData struct {
	Carts              []entities.Cart              `json:"carts,omitempty"`
	Categories         []entities.Category          `json:"categories,omitempty"`
	Expenses           []entities.Expense           `json:"expenses,omitempty"`
	Histories          []entities.History           `json:"histories,omitempty"`
	Payments           []entities.Payment           `json:"payments,omitempty"`
	Products           []entities.Product           `json:"products,omitempty"`
	Receipts           []entities.Receipt           `json:"receipts,omitempty"`
	Shops              []entities.Shop              `json:"shops,omitempty"`
	StockHistories     []entities.StockHistory      `json:"stock_histories,omitempty"`
	TransactionProducts []entities.TransactionProduct `json:"transaction_products,omitempty"`
	Transactions       []entities.Transaction       `json:"transactions,omitempty"`
	Users              []entities.User              `json:"users,omitempty"`
}

// SyncError represents errors that occur during sync operations
type SyncError struct {
	EntityType string    `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
	Field      string    `json:"field,omitempty"`
	Message    string    `json:"message"`
	ErrorType  string    `json:"error_type"` // validation, conflict, database, etc.
}

// ConflictResolution represents how conflicts were resolved
type ConflictResolution struct {
	EntityType     string    `json:"entity_type"`
	EntityID       uuid.UUID `json:"entity_id"`
	ConflictType   string    `json:"conflict_type"`   // update_conflict, create_conflict
	Resolution     string    `json:"resolution"`      // server_wins, client_wins, merged
	ServerVersion  time.Time `json:"server_version"`
	ClientVersion  time.Time `json:"client_version"`
	ResolvedAt     time.Time `json:"resolved_at"`
}

// PullSyncRequest represents the query parameters for pull sync operations
type PullSyncRequest struct {
	SinceTimestamp *time.Time `form:"since" json:"since,omitempty"`
	LicenseID      uuid.UUID  `json:"license_id"`
	UserID         uuid.UUID  `json:"user_id"`
	ShopID         *uuid.UUID `json:"shop_id,omitempty"` // For cashier role restrictions
	Limit          int        `form:"limit" json:"limit,omitempty"`          // Pagination
	Offset         int        `form:"offset" json:"offset,omitempty"`        // Pagination
}

// FullSyncRequest combines both push and pull operations
type FullSyncRequest struct {
	PushData SyncRequest `json:"push_data"`
	PullSince *time.Time `json:"pull_since,omitempty"`
}

// FullSyncResponse combines both push and pull responses
type FullSyncResponse struct {
	PushResult SyncResponse `json:"push_result"`
	PullResult SyncResponse `json:"pull_result"`
}

// SyncStats provides statistics about sync operations
type SyncStats struct {
	EntityType       string `json:"entity_type"`
	TotalRecords     int    `json:"total_records"`
	CreatedRecords   int    `json:"created_records"`
	UpdatedRecords   int    `json:"updated_records"`
	ConflictRecords  int    `json:"conflict_records"`
	ErrorRecords     int    `json:"error_records"`
	ProcessingTimeMs int64  `json:"processing_time_ms"`
}

// BatchSyncRequest represents a batch of sync operations with size limits
type BatchSyncRequest struct {
	Metadata  SyncMetadata `json:"metadata"`
	BatchSize int          `json:"batch_size,omitempty"` // Default: 1000
	Data      SyncData     `json:"data"`
}

// GetEntityCount returns the total number of entities in the sync data
func (sd *SyncData) GetEntityCount() int {
	return len(sd.Carts) +
		len(sd.Categories) +
		len(sd.Expenses) +
		len(sd.Histories) +
		len(sd.Payments) +
		len(sd.Products) +
		len(sd.Receipts) +
		len(sd.Shops) +
		len(sd.StockHistories) +
		len(sd.TransactionProducts) +
		len(sd.Transactions) +
		len(sd.Users)
}

// IsEmpty checks if the sync data contains any entities
func (sd *SyncData) IsEmpty() bool {
	return sd.GetEntityCount() == 0
}

// Validate performs basic validation on sync request
func (sr *SyncRequest) Validate() []SyncError {
	var errors []SyncError

	// Validate metadata
	if sr.Metadata.LicenseID == uuid.Nil {
		errors = append(errors, SyncError{
			EntityType: "metadata",
			Message:    "license_id is required",
			ErrorType:  "validation",
		})
	}

	if sr.Metadata.UserID == uuid.Nil {
		errors = append(errors, SyncError{
			EntityType: "metadata",
			Message:    "user_id is required",
			ErrorType:  "validation",
		})
	}

	// Validate that we have some data to sync
	if sr.Data.IsEmpty() {
		errors = append(errors, SyncError{
			EntityType: "data",
			Message:    "no data provided for sync",
			ErrorType:  "validation",
		})
	}

	return errors
}

// NewSyncResponse creates a new sync response with default values
func NewSyncResponse() *SyncResponse {
	return &SyncResponse{
		SyncTimestamp:    time.Now().UTC(),
		Success:          true,
		RecordsProcessed: 0,
		Errors:           make([]SyncError, 0),
	}
}

// AddError adds an error to the sync response
func (sr *SyncResponse) AddError(entityType string, entityID uuid.UUID, field string, message string, errorType string) {
	sr.Errors = append(sr.Errors, SyncError{
		EntityType: entityType,
		EntityID:   entityID,
		Field:      field,
		Message:    message,
		ErrorType:  errorType,
	})
	sr.Success = false
}

// HasErrors checks if the sync response has any errors
func (sr *SyncResponse) HasErrors() bool {
	return len(sr.Errors) > 0
}