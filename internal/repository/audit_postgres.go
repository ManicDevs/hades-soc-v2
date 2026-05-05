package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =============================================================================
// PostgreSQL Audit Log Repository Implementation (Pure Go)
// =============================================================================

// postgresAuditLogRepository implements AuditLogRepository for PostgreSQL
type postgresAuditLogRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAuditLogRepository creates a new PostgreSQL audit log repository
func NewPostgresAuditLogRepository(pool *pgxpool.Pool) *postgresAuditLogRepository {
	return &postgresAuditLogRepository{pool: pool}
}

// Create inserts a new audit log entry
func (r *postgresAuditLogRepository) Create(ctx context.Context, log *AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, action, resource, details, ip_address, user_agent, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	err := r.pool.QueryRow(ctx, query,
		log.UserID, log.Action, log.Resource, log.Details,
		log.IPAddress, log.UserAgent, log.Timestamp,
	).Scan(&log.ID)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetByID retrieves an audit log by ID
func (r *postgresAuditLogRepository) GetByID(ctx context.Context, id int) (*AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource, details, ip_address, user_agent, timestamp
		FROM audit_logs WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)

	var log AuditLog
	err := row.Scan(&log.ID, &log.UserID, &log.Action, &log.Resource,
		&log.Details, &log.IPAddress, &log.UserAgent, &log.Timestamp)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("audit log not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	return &log, nil
}

// List retrieves audit logs with optional filtering
func (r *postgresAuditLogRepository) List(ctx context.Context, filter AuditLogFilter) ([]*AuditLog, error) {
	query, args := r.buildSelectQuery(filter)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

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
func (r *postgresAuditLogRepository) Count(ctx context.Context, filter AuditLogFilter) (int, error) {
	query, args := r.buildCountQuery(filter)

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// GetByUserID retrieves audit logs for a specific user
func (r *postgresAuditLogRepository) GetByUserID(ctx context.Context, userID int, limit int) ([]*AuditLog, error) {
	return r.List(ctx, AuditLogFilter{UserID: userID, Limit: limit})
}

// GetByAction retrieves audit logs by action type
func (r *postgresAuditLogRepository) GetByAction(ctx context.Context, action string, limit int) ([]*AuditLog, error) {
	return r.List(ctx, AuditLogFilter{Action: action, Limit: limit})
}

// DeleteOlder deletes audit logs older than the specified date
func (r *postgresAuditLogRepository) DeleteOlder(ctx context.Context, date time.Time) (int64, error) {
	query := `DELETE FROM audit_logs WHERE timestamp < $1`

	result, err := r.pool.Exec(ctx, query, date)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	return result.RowsAffected(), nil
}

// Ping tests the database connection
func (r *postgresAuditLogRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// Close closes the repository and releases resources
func (r *postgresAuditLogRepository) Close() error {
	r.pool.Close()
	return nil
}

// buildSelectQuery builds the SELECT query with filters
func (r *postgresAuditLogRepository) buildSelectQuery(filter AuditLogFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.UserID > 0 {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argNum))
		args = append(args, filter.UserID)
		argNum++
	}
	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argNum))
		args = append(args, filter.Action)
		argNum++
	}
	if filter.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argNum))
		args = append(args, filter.Resource)
		argNum++
	}
	if filter.IPAddress != "" {
		conditions = append(conditions, fmt.Sprintf("ip_address = $%d", argNum))
		args = append(args, filter.IPAddress)
		argNum++
	}
	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argNum))
		args = append(args, filter.DateFrom)
		argNum++
	}
	if !filter.DateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argNum))
		args = append(args, filter.DateTo)
		argNum++
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
func (r *postgresAuditLogRepository) buildCountQuery(filter AuditLogFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.UserID > 0 {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argNum))
		args = append(args, filter.UserID)
		argNum++
	}
	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argNum))
		args = append(args, filter.Action)
		argNum++
	}
	if filter.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argNum))
		args = append(args, filter.Resource)
		argNum++
	}
	if filter.IPAddress != "" {
		conditions = append(conditions, fmt.Sprintf("ip_address = $%d", argNum))
		args = append(args, filter.IPAddress)
		argNum++
	}
	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argNum))
		args = append(args, filter.DateFrom)
		argNum++
	}
	if !filter.DateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argNum))
		args = append(args, filter.DateTo)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause), args
}
