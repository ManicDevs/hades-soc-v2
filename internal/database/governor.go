package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// GovernorActionStatus represents the status of a governor action
type GovernorActionStatus string

const (
	GovernorActionStatusApproved          GovernorActionStatus = "approved"
	GovernorActionStatusBlocked           GovernorActionStatus = "blocked"
	GovernorActionStatusManualAckRequired GovernorActionStatus = "manual_ack_required"
	GovernorActionStatusPending           GovernorActionStatus = "pending"
)

// EnsureTableExists checks if a table exists in the database
func (dm *DatabaseManager) EnsureTableExists(ctx context.Context, tableName string) (bool, error) {
	if dm.IsSQLite() {
		var name string
		query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?`
		err := dm.primary.QueryRowContext(ctx, query, tableName).Scan(&name)
		if err != nil {
			if err == sql.ErrNoRows {
				return false, nil // Table doesn't exist
			}
			return false, err
		}
		return name == tableName, nil
	} else {
		// PostgreSQL and MySQL use information_schema
		var exists bool
		query := fmt.Sprintf(`SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = %s AND table_name = %s
		)`, dm.GetPlaceholderForIndex(1), dm.GetPlaceholderForIndex(2))

		schema := "public"
		if dm.IsMySQL() {
			schema = dm.config.Database // MySQL uses database name as schema
		}

		err := dm.primary.QueryRowContext(ctx, query, schema, tableName).Scan(&exists)
		return exists, err
	}
}

// CreateGovernorActionTable creates governor_actions table if it doesn't exist
func (dm *DatabaseManager) CreateGovernorActionTable(ctx context.Context) error {
	// Check if table already exists
	exists, err := dm.EnsureTableExists(ctx, "governor_actions")
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if exists {
		log.Println("DatabaseManager: governor_actions table already exists")
		return nil
	}

	var query string

	if dm.IsSQLite() {
		query = `
		CREATE TABLE IF NOT EXISTS governor_actions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			action_id TEXT UNIQUE NOT NULL,
			action_name TEXT NOT NULL,
			target TEXT NOT NULL,
			reasoning TEXT,
			requester TEXT,
			status TEXT NOT NULL,
			requires_approval BOOLEAN DEFAULT FALSE,
			approved BOOLEAN DEFAULT FALSE,
			requires_manual_ack BOOLEAN DEFAULT FALSE,
			block_reason TEXT,
			execution_time INTEGER DEFAULT 0,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_governor_actions_created_at ON governor_actions(created_at);
		CREATE INDEX IF NOT EXISTS idx_governor_actions_status ON governor_actions(status);
		CREATE INDEX IF NOT EXISTS idx_governor_actions_action_id ON governor_actions(action_id);
		`
	} else if dm.IsPostgreSQL() {
		query = `
		CREATE TABLE IF NOT EXISTS governor_actions (
			id SERIAL PRIMARY KEY,
			action_id TEXT UNIQUE NOT NULL,
			action_name TEXT NOT NULL,
			target TEXT NOT NULL,
			reasoning TEXT,
			requester TEXT,
			status TEXT NOT NULL,
			requires_approval BOOLEAN DEFAULT FALSE,
			approved BOOLEAN DEFAULT FALSE,
			requires_manual_ack BOOLEAN DEFAULT FALSE,
			block_reason TEXT,
			execution_time INTEGER DEFAULT 0,
			metadata TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_governor_actions_created_at ON governor_actions(created_at);
		CREATE INDEX IF NOT EXISTS idx_governor_actions_status ON governor_actions(status);
		CREATE INDEX IF NOT EXISTS idx_governor_actions_action_id ON governor_actions(action_id);
		`
	} else if dm.IsMySQL() {
		query = `
		CREATE TABLE IF NOT EXISTS governor_actions (
			id INT AUTO_INCREMENT PRIMARY KEY,
			action_id VARCHAR(255) UNIQUE NOT NULL,
			action_name VARCHAR(255) NOT NULL,
			target VARCHAR(255) NOT NULL,
			reasoning TEXT,
			requester VARCHAR(255),
			status VARCHAR(50) NOT NULL,
			requires_approval BOOLEAN DEFAULT FALSE,
			approved BOOLEAN DEFAULT FALSE,
			requires_manual_ack BOOLEAN DEFAULT FALSE,
			block_reason TEXT,
			execution_time BIGINT DEFAULT 0,
			metadata TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);
		
		CREATE INDEX idx_governor_actions_created_at ON governor_actions(created_at);
		CREATE INDEX idx_governor_actions_status ON governor_actions(status);
		CREATE INDEX idx_governor_actions_action_id ON governor_actions(action_id);
		`
	}

	_, err = dm.primary.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create governor_actions table: %w", err)
	}

	log.Println("DatabaseManager: governor_actions table created/verified")
	return nil
}

// RecordGovernorAction records a governor action in the database
func (dm *DatabaseManager) RecordGovernorAction(ctx context.Context, action *GovernorAction) error {
	// Note: Table should be created via CreateGovernorActionTable before calling this method

	// Use Go-calculated timestamp instead of database functions
	now := time.Now().UTC() // Use UTC to prevent timezone drift

	// Update the action timestamps to ensure consistency
	// Only set CreatedAt if it's zero (not explicitly set)
	if action.CreatedAt.IsZero() {
		action.CreatedAt = now
	}
	// Always update UpdatedAt to current time
	action.UpdatedAt = now

	// Encrypt sensitive fields before storage using helper functions
	encryptedTarget := dm.encryptField(action.Target)
	encryptedReasoning := dm.encryptField(action.Reasoning)
	encryptedBlockReason := dm.encryptField(action.BlockReason)

	var queryTemplate string
	var args []interface{}

	if dm.IsSQLite() {
		queryTemplate = `
		INSERT OR REPLACE INTO governor_actions (
			action_id, action_name, target, reasoning, requester, status,
			requires_approval, approved, requires_manual_ack, block_reason,
			execution_time, metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		args = []interface{}{
			action.ActionID, action.ActionName, encryptedTarget, encryptedReasoning,
			action.Requester, action.Status, action.RequiresApproval, action.Approved,
			action.RequiresManualAck, encryptedBlockReason, action.ExecutionTime,
			action.Metadata, action.CreatedAt, action.UpdatedAt,
		}
	} else if dm.IsPostgreSQL() {
		queryTemplate = `
		INSERT INTO governor_actions (
			action_id, action_name, target, reasoning, requester, status,
			requires_approval, approved, requires_manual_ack, block_reason,
			execution_time, metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(action_id) DO UPDATE SET
			status = EXCLUDED.status,
			approved = EXCLUDED.approved,
			requires_manual_ack = EXCLUDED.requires_manual_ack,
			block_reason = EXCLUDED.block_reason,
			execution_time = EXCLUDED.execution_time,
			updated_at = ?
		`
		args = []interface{}{
			action.ActionID, action.ActionName, encryptedTarget, encryptedReasoning,
			action.Requester, action.Status, action.RequiresApproval, action.Approved,
			action.RequiresManualAck, encryptedBlockReason, action.ExecutionTime,
			action.Metadata, action.CreatedAt, action.UpdatedAt,
			action.UpdatedAt, // Additional parameter for the UPDATE clause
		}
	} else if dm.IsMySQL() {
		queryTemplate = `
		INSERT INTO governor_actions (
			action_id, action_name, target, reasoning, requester, status,
			requires_approval, approved, requires_manual_ack, block_reason,
			execution_time, metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			status = VALUES(status),
			approved = VALUES(approved),
			requires_manual_ack = VALUES(requires_manual_ack),
			block_reason = VALUES(block_reason),
			execution_time = VALUES(execution_time),
			updated_at = ?
		`
		args = []interface{}{
			action.ActionID, action.ActionName, encryptedTarget, encryptedReasoning,
			action.Requester, action.Status, action.RequiresApproval, action.Approved,
			action.RequiresManualAck, encryptedBlockReason, action.ExecutionTime,
			action.Metadata, action.CreatedAt, action.UpdatedAt,
			action.UpdatedAt, // Additional parameter for the UPDATE clause
		}
	} else {
		return fmt.Errorf("unsupported database type: %s", dm.config.DBType)
	}

	// Build driver-agnostic query
	query, finalArgs := dm.BuildQuery(queryTemplate, args...)

	_, err := dm.primary.ExecContext(ctx, query, finalArgs...)

	if err != nil {
		return fmt.Errorf("failed to record governor action: %w", err)
	}

	return nil
}

// GetBlockCountInLastHour returns the number of approved actions in the last hour
func (dm *DatabaseManager) GetBlockCountInLastHour(ctx context.Context) (int, error) {
	// Note: Table should be created via CreateGovernorActionTable before calling this method

	// Use Go to calculate the timestamp with UTC to prevent timezone drift
	oneHourAgo := time.Now().UTC().Add(-1 * time.Hour)

	// Use driver-agnostic query building
	queryTemplate := `
	SELECT COUNT(*) FROM governor_actions 
	WHERE status = ? AND approved = ? AND created_at >= ?
	`

	args := []interface{}{
		GovernorActionStatusApproved, true, oneHourAgo,
	}

	// Build driver-agnostic query
	query, finalArgs := dm.BuildQuery(queryTemplate, args...)

	var count int
	err := dm.primary.QueryRowContext(ctx, query, finalArgs...).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get block count in last hour: %w", err)
	}

	return count, nil
}

// GetRecentActions returns recent governor actions within the specified time window
func (dm *DatabaseManager) GetRecentActions(ctx context.Context, since time.Time, limit int) ([]*GovernorAction, error) {
	// Ensure since time is in UTC for consistency
	sinceUTC := since.UTC()

	queryTemplate := `
	SELECT id, action_id, action_name, target, reasoning, requester, status,
		requires_approval, approved, requires_manual_ack, block_reason,
		execution_time, metadata, created_at, updated_at
	FROM governor_actions 
	WHERE created_at >= ?
	ORDER BY created_at DESC
	LIMIT ?
	`
	if dm.IsSQLite() {
		queryTemplate = `
		SELECT id, action_id, action_name, target, reasoning, requester, status,
			requires_approval, approved, requires_manual_ack, block_reason,
			execution_time, metadata, created_at, updated_at
		FROM governor_actions
		WHERE datetime(created_at) >= datetime(?)
		ORDER BY datetime(created_at) DESC
		LIMIT ?
		`
	}

	args := []interface{}{sinceUTC, limit}

	// Build driver-agnostic query
	query, finalArgs := dm.BuildQuery(queryTemplate, args...)

	rows, err := dm.primary.QueryContext(ctx, query, finalArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent actions: %w", err)
	}
	defer rows.Close()

	var actions []*GovernorAction
	for rows.Next() {
		action := &GovernorAction{}
		err := rows.Scan(
			&action.ID, &action.ActionID, &action.ActionName, &action.Target,
			&action.Reasoning, &action.Requester, &action.Status,
			&action.RequiresApproval, &action.Approved, &action.RequiresManualAck,
			&action.BlockReason, &action.ExecutionTime, &action.Metadata,
			&action.CreatedAt, &action.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan governor action: %w", err)
		}

		// Decrypt sensitive fields after retrieval using helper functions
		action.Target = dm.decryptField(action.Target)
		action.Reasoning = dm.decryptField(action.Reasoning)
		action.BlockReason = dm.decryptField(action.BlockReason)

		actions = append(actions, action)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating governor actions: %w", err)
	}

	return actions, nil
}

// GetGovernorStats returns statistics about governor actions
func (dm *DatabaseManager) GetGovernorStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get approved actions in last hour
	approvedLastHour, err := dm.GetBlockCountInLastHour(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get approved actions count: %w", err)
	}
	stats["approved_last_hour"] = approvedLastHour

	// Get total actions in last 24 hours - use Go to calculate timestamp with UTC
	twentyFourHoursAgo := time.Now().UTC().Add(-24 * time.Hour)

	// Use driver-agnostic query building
	queryTemplate := `
	SELECT 
		COUNT(*) as total,
		COUNT(CASE WHEN approved = ? THEN 1 END) as approved,
		COUNT(CASE WHEN requires_manual_ack = ? THEN 1 END) as manual_ack,
		COUNT(CASE WHEN status = ? THEN 1 END) as blocked
	FROM governor_actions 
	WHERE created_at >= ?
	`

	// Use appropriate boolean values for the database type
	var approvedValue interface{} = true
	if dm.IsSQLite() {
		approvedValue = 1 // SQLite uses integer for boolean
	}

	args := []interface{}{
		approvedValue, approvedValue, GovernorActionStatusBlocked, twentyFourHoursAgo,
	}

	// Build driver-agnostic query
	query, finalArgs := dm.BuildQuery(queryTemplate, args...)

	var total, approved, manualAck, blocked int
	err = dm.primary.QueryRowContext(ctx, query, finalArgs...).Scan(
		&total, &approved, &manualAck, &blocked,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get governor stats: %w", err)
	}

	stats["total_last_24h"] = total
	stats["approved_last_24h"] = approved
	stats["manual_ack_last_24h"] = manualAck
	stats["blocked_last_24h"] = blocked

	// Get time until next reset - use Go to calculate timestamps with UTC
	oneHourAgo := time.Now().UTC().Add(-1 * time.Hour)
	resetQueryTemplate := `
	SELECT MIN(created_at) FROM governor_actions 
	WHERE approved = ? AND created_at >= ?
	`

	// Use appropriate boolean value for the database type
	var resetApprovedValue interface{} = true
	if dm.IsSQLite() {
		resetApprovedValue = 1
	}

	resetArgs := []interface{}{resetApprovedValue, oneHourAgo}
	resetQuery, resetFinalArgs := dm.BuildQuery(resetQueryTemplate, resetArgs...)

	var oldestApprovedStr sql.NullString
	err = dm.primary.QueryRowContext(ctx, resetQuery, resetFinalArgs...).Scan(&oldestApprovedStr)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get oldest approved action: %w", err)
	}

	if oldestApprovedStr.Valid && oldestApprovedStr.String != "" {
		oldestApproved, err := time.Parse("2006-01-02 15:04:05", oldestApprovedStr.String)
		if err != nil {
			// Try parsing with different formats
			layouts := []string{
				"2006-01-02T15:04:05Z",
				"2006-01-02T15:04:05-07:00",
				"2006-01-02 15:04:05.999999999-07:00",
			}
			for _, layout := range layouts {
				if oldestApproved, err = time.Parse(layout, oldestApprovedStr.String); err == nil {
					break
				}
			}
			if err != nil {
				// If parsing fails, skip time calculation
				stats["time_until_reset"] = "0s"
				stats["time_until_reset_seconds"] = 0
			}
		}

		if err == nil {
			timeUntilReset := time.Until(oldestApproved.Add(1 * time.Hour))
			if timeUntilReset < 0 {
				timeUntilReset = 0
			}
			stats["time_until_reset"] = timeUntilReset.String()
			stats["time_until_reset_seconds"] = int(timeUntilReset.Seconds())
		}
	} else {
		stats["time_until_reset"] = "0s"
		stats["time_until_reset_seconds"] = 0
	}

	return stats, nil
}

// GetGovernorActionByID retrieves a specific governor action by its action_id
func (dm *DatabaseManager) GetGovernorActionByID(ctx context.Context, actionID string) (*GovernorAction, error) {
	queryTemplate := `
	SELECT id, action_id, action_name, target, reasoning, requester, status,
		requires_approval, approved, requires_manual_ack, block_reason,
		execution_time, metadata, created_at, updated_at
	FROM governor_actions 
	WHERE action_id = ?
	`

	args := []interface{}{actionID}

	// Build driver-agnostic query
	query, finalArgs := dm.BuildQuery(queryTemplate, args...)

	action := &GovernorAction{}
	err := dm.primary.QueryRowContext(ctx, query, finalArgs...).Scan(
		&action.ID, &action.ActionID, &action.ActionName, &action.Target,
		&action.Reasoning, &action.Requester, &action.Status,
		&action.RequiresApproval, &action.Approved, &action.RequiresManualAck,
		&action.BlockReason, &action.ExecutionTime, &action.Metadata,
		&action.CreatedAt, &action.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Action not found
		}
		return nil, fmt.Errorf("failed to get governor action by ID: %w", err)
	}

	// Decrypt sensitive fields after retrieval using helper functions
	action.Target = dm.decryptField(action.Target)
	action.Reasoning = dm.decryptField(action.Reasoning)
	action.BlockReason = dm.decryptField(action.BlockReason)

	return action, nil
}

// UpdateGovernorAction updates an existing governor action
func (dm *DatabaseManager) UpdateGovernorAction(ctx context.Context, action *GovernorAction) error {
	// Use Go-calculated timestamp with UTC to prevent timezone drift
	action.UpdatedAt = time.Now().UTC()

	queryTemplate := `
	UPDATE governor_actions 
	SET action_name = ?, target = ?, reasoning = ?, requester = ?, status = ?,
		requires_approval = ?, approved = ?, requires_manual_ack = ?, block_reason = ?,
		execution_time = ?, metadata = ?, updated_at = ?
	WHERE id = ?
	`

	// Encrypt sensitive fields before update using helper functions
	encryptedTarget := dm.encryptField(action.Target)
	encryptedReasoning := dm.encryptField(action.Reasoning)
	encryptedBlockReason := dm.encryptField(action.BlockReason)

	args := []interface{}{
		action.ActionName, encryptedTarget, encryptedReasoning, action.Requester, action.Status,
		action.RequiresApproval, action.Approved, action.RequiresManualAck, encryptedBlockReason,
		action.ExecutionTime, action.Metadata, action.UpdatedAt, action.ID,
	}

	// Build driver-agnostic query
	query, finalArgs := dm.BuildQuery(queryTemplate, args...)

	_, err := dm.primary.ExecContext(ctx, query, finalArgs...)

	if err != nil {
		return fmt.Errorf("failed to update governor action: %w", err)
	}

	return nil
}

// GetPendingActions returns actions that require manual approval
func (dm *DatabaseManager) GetPendingActions(ctx context.Context) ([]*GovernorAction, error) {
	queryTemplate := `
	SELECT id, action_id, action_name, target, reasoning, requester, status,
		requires_approval, approved, requires_manual_ack, block_reason,
		execution_time, metadata, created_at, updated_at
	FROM governor_actions 
	WHERE status = ? OR status = ?
	ORDER BY created_at DESC
	`

	args := []interface{}{
		GovernorActionStatusManualAckRequired, GovernorActionStatusPending,
	}

	// Build driver-agnostic query
	query, finalArgs := dm.BuildQuery(queryTemplate, args...)

	rows, err := dm.primary.QueryContext(ctx, query, finalArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending actions: %w", err)
	}
	defer rows.Close()

	var actions []*GovernorAction
	for rows.Next() {
		action := &GovernorAction{}
		err := rows.Scan(
			&action.ID, &action.ActionID, &action.ActionName, &action.Target,
			&action.Reasoning, &action.Requester, &action.Status,
			&action.RequiresApproval, &action.Approved, &action.RequiresManualAck,
			&action.BlockReason, &action.ExecutionTime, &action.Metadata,
			&action.CreatedAt, &action.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending action: %w", err)
		}

		// Decrypt sensitive fields after retrieval using helper functions
		action.Target = dm.decryptField(action.Target)
		action.Reasoning = dm.decryptField(action.Reasoning)
		action.BlockReason = dm.decryptField(action.BlockReason)

		actions = append(actions, action)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pending actions: %w", err)
	}

	return actions, nil
}

// CleanupOldGovernorActions removes actions older than the specified duration
func (dm *DatabaseManager) CleanupOldGovernorActions(ctx context.Context, olderThan time.Duration) error {
	queryTemplate := `DELETE FROM governor_actions WHERE created_at < ?`
	if dm.IsSQLite() {
		queryTemplate = `DELETE FROM governor_actions WHERE datetime(created_at) < datetime(?)`
	}

	// Use Go-calculated timestamp with UTC to prevent timezone drift
	cutoff := time.Now().UTC().Add(-olderThan)
	args := []interface{}{cutoff}

	// Build driver-agnostic query
	query, finalArgs := dm.BuildQuery(queryTemplate, args...)

	result, err := dm.primary.ExecContext(ctx, query, finalArgs...)
	if err != nil {
		return fmt.Errorf("failed to cleanup old governor actions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("DatabaseManager: Cleaned up %d old governor actions", rowsAffected)

	return nil
}
