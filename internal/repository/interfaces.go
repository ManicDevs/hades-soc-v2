package repository

import (
	"context"
	"time"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditLogFilter provides filtering options for audit log queries
type AuditLogFilter struct {
	UserID    int
	Action    string
	Resource  string
	IPAddress string
	Limit     int
	Offset    int
	OrderBy   string
	OrderDesc bool
	DateFrom  time.Time
	DateTo    time.Time
}

// AuditLogRepository defines the contract for audit log data access
// Uses pure Go database drivers (no CGO)
type AuditLogRepository interface {
	// Create inserts a new audit log entry
	Create(ctx context.Context, log *AuditLog) error

	// GetByID retrieves an audit log by ID
	GetByID(ctx context.Context, id int) (*AuditLog, error)

	// List retrieves audit logs with optional filtering
	List(ctx context.Context, filter AuditLogFilter) ([]*AuditLog, error)

	// Count returns the total number of audit logs matching the filter
	Count(ctx context.Context, filter AuditLogFilter) (int, error)

	// GetByUserID retrieves audit logs for a specific user
	GetByUserID(ctx context.Context, userID int, limit int) ([]*AuditLog, error)

	// GetByAction retrieves audit logs by action type
	GetByAction(ctx context.Context, action string, limit int) ([]*AuditLog, error)

	// DeleteOlder deletes audit logs older than the specified date
	DeleteOlder(ctx context.Context, date time.Time) (int64, error)

	// Ping tests the database connection
	Ping(ctx context.Context) error

	// Close closes the repository and releases resources
	Close() error
}
