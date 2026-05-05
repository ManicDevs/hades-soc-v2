package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// =============================================================================
// MySQL Audit Log Repository Implementation (Pure Go)
// =============================================================================

// mysqlAuditLogRepository implements AuditLogRepository for MySQL
type mysqlAuditLogRepository struct {
	db *sql.DB
}

// NewMySQLAuditLogRepository creates a new MySQL audit log repository
func NewMySQLAuditLogRepository(db *sql.DB) *mysqlAuditLogRepository {
	return &mysqlAuditLogRepository{db: db}
}

// Create inserts a new audit log entry
func (r *mysqlAuditLogRepository) Create(ctx context.Context, log *AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, action, resource, details, ip_address, user_agent, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	result, err := r.db.ExecContext(ctx, query,
		log.UserID, log.Action, log.Resource, log.Details,
		log.IPAddress, log.UserAgent, log.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	log.ID = int(id)
	return nil
}

// GetByID retrieves an audit log by ID
func (r *mysqlAuditLogRepository) GetByID(ctx context.Context, id int) (*AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource, details, ip_address, user_agent, timestamp
		FROM audit_logs WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	var log AuditLog
	err := row.Scan(&log.ID, &log.UserID, &log.Action, &log.Resource,
		&log.Details, &log.IPAddress, &log.UserAgent, &log.Timestamp)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audit log not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	return &log, nil
}

// List retrieves audit logs with optional filtering
func (r *mysqlAuditLogRepository) List(ctx context.Context, filter AuditLogFilter) ([]*AuditLog, error) {
	query, args := r.buildSelectQuery(filter)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var logs []*AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(&log.ID, &log.UserID, &log.Action, &log.Resource,
			&log.Details, &log.IPAddress, &log.UserAgent, &log.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, nil
}

// Count returns the total number of audit logs matching the filter
func (r *mysqlAuditLogRepository) Count(ctx context.Context, filter AuditLogFilter) (int, error) {
	query, args := r.buildCountQuery(filter)

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// GetByUserID retrieves audit logs for a specific user
func (r *mysqlAuditLogRepository) GetByUserID(ctx context.Context, userID int, limit int) ([]*AuditLog, error) {
	return r.List(ctx, AuditLogFilter{UserID: userID, Limit: limit})
}

// GetByAction retrieves audit logs by action type
func (r *mysqlAuditLogRepository) GetByAction(ctx context.Context, action string, limit int) ([]*AuditLog, error) {
	return r.List(ctx, AuditLogFilter{Action: action, Limit: limit})
}

// DeleteOlder deletes audit logs older than the specified date
func (r *mysqlAuditLogRepository) DeleteOlder(ctx context.Context, date time.Time) (int64, error) {
	query := `DELETE FROM audit_logs WHERE timestamp < ?`

	result, err := r.db.ExecContext(ctx, query, date)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	return result.RowsAffected()
}

// Ping tests the database connection
func (r *mysqlAuditLogRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Close closes the repository and releases resources
func (r *mysqlAuditLogRepository) Close() error {
	return r.db.Close()
}

// buildSelectQuery builds the SELECT query with filters (MySQL uses ? placeholders)
func (r *mysqlAuditLogRepository) buildSelectQuery(filter AuditLogFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if filter.UserID > 0 {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}
	if filter.Resource != "" {
		conditions = append(conditions, "resource = ?")
		args = append(args, filter.Resource)
	}
	if filter.IPAddress != "" {
		conditions = append(conditions, "ip_address = ?")
		args = append(args, filter.IPAddress)
	}
	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, filter.DateFrom)
	}
	if !filter.DateTo.IsZero() {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, filter.DateTo)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderBy := "ORDER BY timestamp DESC"
	if filter.OrderBy != "" {
		orderDir := "ASC"
		if filter.OrderDesc {
			orderDir = "DESC"
		}
		orderBy = fmt.Sprintf("ORDER BY %s %s", filter.OrderBy, orderDir)
	}

	limitClause := ""
	if filter.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			limitClause += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	return fmt.Sprintf(`
		SELECT id, user_id, action, resource, details, ip_address, user_agent, timestamp
		FROM audit_logs
		%s %s %s`, whereClause, orderBy, limitClause), args
}

// buildCountQuery builds the COUNT query with filters
func (r *mysqlAuditLogRepository) buildCountQuery(filter AuditLogFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if filter.UserID > 0 {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}
	if filter.Resource != "" {
		conditions = append(conditions, "resource = ?")
		args = append(args, filter.Resource)
	}
	if filter.IPAddress != "" {
		conditions = append(conditions, "ip_address = ?")
		args = append(args, filter.IPAddress)
	}
	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, filter.DateFrom)
	}
	if !filter.DateTo.IsZero() {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, filter.DateTo)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause), args
}
