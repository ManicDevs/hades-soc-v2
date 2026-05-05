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
	if dm.config.UseSQLite {
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
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = ?
		)`
		err := dm.primary.QueryRowContext(ctx, query, tableName).Scan(&exists)
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

	// Create table
	query := `
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

	if dm.config.UseSQLite {
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
	} else {
		// PostgreSQL version
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
	// Verification guard: ensure table exists
	if _, err := dm.primary.ExecContext(ctx, `
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
		)
	`); err != nil {
		return fmt.Errorf("failed to ensure governor_actions table exists: %w", err)
	}

	var query string

	if dm.config.UseSQLite {
		query = `
		INSERT OR REPLACE INTO governor_actions (
			action_id, action_name, target, reasoning, requester, status,
			requires_approval, approved, requires_manual_ack, block_reason,
			execution_time, metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
	} else {
		query = `
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
			updated_at = CURRENT_TIMESTAMP
		`
	}

	_, err := dm.primary.ExecContext(ctx, query,
		action.ActionID, action.ActionName, action.Target, action.Reasoning,
		action.Requester, action.Status, action.RequiresApproval, action.Approved,
		action.RequiresManualAck, action.BlockReason, action.ExecutionTime,
		action.Metadata, action.CreatedAt, action.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to record governor action: %w", err)
	}

	return nil
}

// GetBlockCountInLastHour returns the number of approved actions in the last hour
func (dm *DatabaseManager) GetBlockCountInLastHour(ctx context.Context) (int, error) {
	// Verification guard: ensure table exists
	if _, err := dm.primary.ExecContext(ctx, `
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
		)
	`); err != nil {
		return 0, fmt.Errorf("failed to ensure governor_actions table exists: %w", err)
	}

	var query string

	if dm.config.UseSQLite {
		query = `
		SELECT COUNT(*) FROM governor_actions 
		WHERE status = ? AND approved = ? AND created_at >= ?
		`
	} else {
		query = `
		SELECT COUNT(*) FROM governor_actions 
		WHERE status = ? AND approved = ? AND created_at >= ?
		`
	}

	oneHourAgo := time.Now().Add(-1 * time.Hour)
	var count int

	err := dm.primary.QueryRowContext(ctx, query,
		GovernorActionStatusApproved, true, oneHourAgo,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get block count in last hour: %w", err)
	}

	return count, nil
}

// GetRecentActions returns recent governor actions within the specified time window
func (dm *DatabaseManager) GetRecentActions(ctx context.Context, since time.Time, limit int) ([]*GovernorAction, error) {
	var query string

	if dm.config.UseSQLite {
		query = `
		SELECT id, action_id, action_name, target, reasoning, requester, status,
			requires_approval, approved, requires_manual_ack, block_reason,
			execution_time, metadata, created_at, updated_at
		FROM governor_actions 
		WHERE created_at >= ?
		ORDER BY created_at DESC
		LIMIT ?
		`
	} else {
		query = `
		SELECT id, action_id, action_name, target, reasoning, requester, status,
			requires_approval, approved, requires_manual_ack, block_reason,
			execution_time, metadata, created_at, updated_at
		FROM governor_actions 
		WHERE created_at >= ?
		ORDER BY created_at DESC
		LIMIT ?
		`
	}

	rows, err := dm.primary.QueryContext(ctx, query, since, limit)
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
		actions = append(actions, action)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating governor actions: %w", err)
	}

	return actions, nil
}

// GetGovernorStats returns statistics for the safety governor
func (dm *DatabaseManager) GetGovernorStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get approved actions in last hour
	approvedLastHour, err := dm.GetBlockCountInLastHour(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get approved actions count: %w", err)
	}
	stats["approved_last_hour"] = approvedLastHour

	// Get total actions in last 24 hours
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	var query string

	if dm.config.UseSQLite {
		query = `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN approved = 1 THEN 1 END) as approved,
			COUNT(CASE WHEN requires_manual_ack = 1 THEN 1 END) as manual_ack,
			COUNT(CASE WHEN status = ? THEN 1 END) as blocked
		FROM governor_actions 
		WHERE created_at >= ?
		`
	} else {
		query = `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN approved = true THEN 1 END) as approved,
			COUNT(CASE WHEN requires_manual_ack = true THEN 1 END) as manual_ack,
			COUNT(CASE WHEN status = ? THEN 1 END) as blocked
		FROM governor_actions 
		WHERE created_at >= ?
		`
	}

	var total, approved, manualAck, blocked int
	err = dm.primary.QueryRowContext(ctx, query, GovernorActionStatusBlocked, twentyFourHoursAgo).Scan(
		&total, &approved, &manualAck, &blocked,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get governor stats: %w", err)
	}

	stats["total_last_24h"] = total
	stats["approved_last_24h"] = approved
	stats["manual_ack_last_24h"] = manualAck
	stats["blocked_last_24h"] = blocked

	// Get time until next reset (when oldest approved action is 1 hour old)
	var resetQuery string

	if dm.config.UseSQLite {
		resetQuery = `
		SELECT MIN(created_at) FROM governor_actions 
		WHERE approved = ? AND created_at >= ?
		`
	} else {
		resetQuery = `
		SELECT MIN(created_at) FROM governor_actions 
		WHERE approved = ? AND created_at >= ?
		`
	}

	oneHourAgo := time.Now().Add(-1 * time.Hour)
	var oldestApprovedStr string
	err = dm.primary.QueryRowContext(ctx, resetQuery, true, oneHourAgo).Scan(&oldestApprovedStr)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get oldest approved action: %w", err)
	}

	if oldestApprovedStr != "" {
		oldestApproved, err := time.Parse("2006-01-02 15:04:05", oldestApprovedStr)
		if err != nil {
			// Try parsing with different formats
			layouts := []string{
				"2006-01-02T15:04:05Z",
				"2006-01-02T15:04:05-07:00",
				"2006-01-02 15:04:05.999999999-07:00",
			}
			for _, layout := range layouts {
				if oldestApproved, err = time.Parse(layout, oldestApprovedStr); err == nil {
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

// CleanupOldGovernorActions removes actions older than the specified duration
func (dm *DatabaseManager) CleanupOldGovernorActions(ctx context.Context, olderThan time.Duration) error {
	var query string

	if dm.config.UseSQLite {
		query = `DELETE FROM governor_actions WHERE created_at < ?`
	} else {
		query = `DELETE FROM governor_actions WHERE created_at < ?`
	}

	cutoff := time.Now().Add(-olderThan)
	result, err := dm.primary.ExecContext(ctx, query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup old governor actions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("DatabaseManager: Cleaned up %d old governor actions", rowsAffected)

	return nil
}
