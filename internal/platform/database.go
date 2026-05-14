package platform

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hades-v2/internal/database"
	"log"
	"sync"
	"time"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type              string        `json:"type"`                // "postgresql", "mysql", or "sqlite"
	Host              string        `json:"host"`                // PostgreSQL host
	Port              int           `json:"port"`                // PostgreSQL port
	Database          string        `json:"database"`            // PostgreSQL database name
	Username          string        `json:"username"`            // PostgreSQL username
	Password          string        `json:"password"`            // PostgreSQL password
	SSLMode           string        `json:"ssl_mode"`            // PostgreSQL SSL mode
	Path              string        `json:"path"`                // SQLite path (for SQLite)
	MaxConnections    int           `json:"max_connections"`     // Max connections
	ConnTimeout       time.Duration `json:"conn_timeout"`        // Connection timeout
	EnableWAL         bool          `json:"enable_wal"`          // SQLite WAL
	EnableForeignKeys bool          `json:"enable_foreign_keys"` // SQLite foreign keys
}

// DefaultDatabaseConfig returns sensible database defaults (PostgreSQL preferred)
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:              "postgresql",
		Host:              "localhost",
		Port:              5432,
		Database:          "hades",
		Username:          "hades",
		Password:          "hades_password",
		SSLMode:           "disable",
		MaxConnections:    25,
		ConnTimeout:       30 * time.Second,
		EnableWAL:         true,
		EnableForeignKeys: true,
	}
}

// Database provides enterprise database functionality using unified DatabaseManager
type Database struct {
	config *DatabaseConfig
	db     *sql.DB
	mu     sync.RWMutex
}

// NewDatabase creates a new database instance using unified DatabaseManager
func NewDatabase(config *DatabaseConfig) (*Database, error) {
	if config == nil {
		config = DefaultDatabaseConfig()
	}

	mgr := database.GetManager()
	if mgr == nil {
		return nil, fmt.Errorf("hades.platform.database: database manager not initialized (set HADES_DB_ENCRYPTION_KEY or HADES_ALLOW_INSECURE_DEV_DB_KEY=true)")
	}

	managerConfig := &database.ManagerConfig{
		PrimaryDSN: fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode),
		Database:     config.Database,
		MaxOpenConns: config.MaxConnections,
		MaxIdleConns: 5,
		ConnLifetime: 5 * time.Minute,
		UseSQLite:    config.Type == "sqlite",
		SQLitePath:   config.Path,
	}

	mgr.SetConfig(managerConfig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := mgr.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("hades.platform.database: failed to initialize database manager: %w", err)
	}

	db, err := mgr.GetConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.database: failed to get connection: %w", err)
	}

	platformDB := &Database{
		config: config,
		db:     db,
	}

	if err := platformDB.initialize(); err != nil {
		return nil, fmt.Errorf("hades.platform.database: %w", err)
	}

	return platformDB, nil
}

// initialize sets up the database schema and configuration
func (d *Database) initialize() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Configure SQLite
	if d.config.EnableWAL {
		if _, err := d.db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			return fmt.Errorf("hades.platform.database: failed to enable WAL: %w", err)
		}
	}

	if d.config.EnableForeignKeys {
		if _, err := d.db.Exec("PRAGMA foreign_keys=ON"); err != nil {
			return fmt.Errorf("hades.platform.database: failed to enable foreign keys: %w", err)
		}
	}

	// Create tables
	if err := d.createTables(); err != nil {
		return fmt.Errorf("hades.platform.database: failed to create tables: %w", err)
	}

	return nil
}

// createTables creates the database schema
func (d *Database) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			role TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			is_active BOOLEAN DEFAULT 1,
			last_login DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS scan_results (
			id TEXT PRIMARY KEY,
			module_name TEXT NOT NULL,
			target TEXT NOT NULL,
			status TEXT NOT NULL,
			result_data TEXT,
			error_message TEXT,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			action TEXT NOT NULL,
			resource TEXT,
			details TEXT,
			ip_address TEXT,
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS configurations (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			description TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("hades.platform.database: failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.db.Close()
}

// StoreUser saves a user to the database
func (d *Database) StoreUser(ctx context.Context, user *User) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	query := `INSERT OR REPLACE INTO users 
		(id, username, email, role, password_hash, is_active, last_login, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	_, err := d.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.Role,
		user.PasswordHash, user.IsActive, user.LastLogin, user.CreatedAt)

	if err != nil {
		return fmt.Errorf("hades.platform.database: failed to store user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by username
func (d *Database) GetUser(ctx context.Context, username string) (*User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `SELECT id, username, email, role, password_hash, is_active, last_login, created_at 
		FROM users WHERE username = ?`

	row := d.db.QueryRowContext(ctx, query, username)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Role,
		&user.PasswordHash, &user.IsActive, &user.LastLogin, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("hades.platform.database: user not found")
		}
		return nil, fmt.Errorf("hades.platform.database: failed to get user: %w", err)
	}

	return &user, nil
}

// StoreSession saves a session to the database
func (d *Database) StoreSession(ctx context.Context, session *Session) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	query := `INSERT OR REPLACE INTO sessions 
		(id, user_id, token, expires_at, ip_address, user_agent, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := d.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.Token, session.ExpiresAt,
		session.IPAddress, session.UserAgent, session.CreatedAt)

	if err != nil {
		return fmt.Errorf("hades.platform.database: failed to store session: %w", err)
	}

	return nil
}

// GetSession retrieves a session by token
func (d *Database) GetSession(ctx context.Context, token string) (*Session, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `SELECT id, user_id, token, expires_at, ip_address, user_agent, created_at 
		FROM sessions WHERE token = ? AND expires_at > CURRENT_TIMESTAMP`

	row := d.db.QueryRowContext(ctx, query, token)

	var session Session
	err := row.Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt,
		&session.IPAddress, &session.UserAgent, &session.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("hades.platform.database: session not found or expired")
		}
		return nil, fmt.Errorf("hades.platform.database: failed to get session: %w", err)
	}

	return &session, nil
}

// DeleteSession removes a session from the database
func (d *Database) DeleteSession(ctx context.Context, token string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	query := `DELETE FROM sessions WHERE token = ?`

	result, err := d.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("hades.platform.database: failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("hades.platform.database: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("hades.platform.database: session not found")
	}

	return nil
}

// StoreScanResult saves scan results to the database
func (d *Database) StoreScanResult(ctx context.Context, moduleName, target, status string, resultData interface{}, errorMsg string, startTime, endTime time.Time) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var resultJSON string
	if resultData != nil {
		data, err := json.Marshal(resultData)
		if err != nil {
			return fmt.Errorf("hades.platform.database: failed to marshal result data: %w", err)
		}
		resultJSON = string(data)
	}

	query := `INSERT INTO scan_results 
		(id, module_name, target, status, result_data, error_message, start_time, end_time) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	id := fmt.Sprintf("%d", time.Now().UnixNano())

	_, err := d.db.ExecContext(ctx, query,
		id, moduleName, target, status, resultJSON, errorMsg, startTime, endTime)

	if err != nil {
		return fmt.Errorf("hades.platform.database: failed to store scan result: %w", err)
	}

	return nil
}

// GetScanResults retrieves scan results with optional filters
func (d *Database) GetScanResults(ctx context.Context, moduleName, target string, limit int) ([]map[string]interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `SELECT id, module_name, target, status, result_data, error_message, start_time, end_time, created_at 
		FROM scan_results WHERE 1=1`

	args := []interface{}{}

	if moduleName != "" {
		query += " AND module_name = ?"
		args = append(args, moduleName)
	}

	if target != "" {
		query += " AND target = ?"
		args = append(args, target)
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.database: failed to query scan results: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var results []map[string]interface{}
	for rows.Next() {
		var id, moduleName, target, status, resultData, errorMsg string
		var startTime, endTime, createdAt time.Time

		err := rows.Scan(&id, &moduleName, &target, &status, &resultData, &errorMsg, &startTime, &endTime, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("hades.platform.database: failed to scan row: %w", err)
		}

		result := map[string]interface{}{
			"id":            id,
			"module_name":   moduleName,
			"target":        target,
			"status":        status,
			"result_data":   resultData,
			"error_message": errorMsg,
			"start_time":    startTime,
			"end_time":      endTime,
			"created_at":    createdAt,
		}

		results = append(results, result)
	}

	return results, nil
}

// LogAudit records an audit event
func (d *Database) LogAudit(ctx context.Context, userID, action, resource, details, ipAddress, userAgent string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	query := `INSERT INTO audit_logs 
		(id, user_id, action, resource, details, ip_address, user_agent, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	id := fmt.Sprintf("%d", time.Now().UnixNano())

	_, err := d.db.ExecContext(ctx, query,
		id, userID, action, resource, details, ipAddress, userAgent)

	if err != nil {
		return fmt.Errorf("hades.platform.database: failed to log audit: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs with optional filters
func (d *Database) GetAuditLogs(ctx context.Context, userID, action string, limit int) ([]map[string]interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `SELECT id, user_id, action, resource, details, ip_address, user_agent, created_at 
		FROM audit_logs WHERE 1=1`

	args := []interface{}{}

	if userID != "" {
		query += " AND user_id = ?"
		args = append(args, userID)
	}

	if action != "" {
		query += " AND action = ?"
		args = append(args, action)
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.database: failed to query audit logs: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var logs []map[string]interface{}
	for rows.Next() {
		var id, userID, action, resource, details, ipAddress, userAgent string
		var createdAt time.Time

		err := rows.Scan(&id, &userID, &action, &resource, &details, &ipAddress, &userAgent, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("hades.platform.database: failed to scan audit row: %w", err)
		}

		log := map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"action":     action,
			"resource":   resource,
			"details":    details,
			"ip_address": ipAddress,
			"user_agent": userAgent,
			"created_at": createdAt,
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// StoreConfiguration saves a configuration value
func (d *Database) StoreConfiguration(ctx context.Context, key, value, description string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	query := `INSERT OR REPLACE INTO configurations (key, value, description, updated_at) 
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)`

	_, err := d.db.ExecContext(ctx, query, key, value, description)
	if err != nil {
		return fmt.Errorf("hades.platform.database: failed to store configuration: %w", err)
	}

	return nil
}

// GetConfiguration retrieves a configuration value
func (d *Database) GetConfiguration(ctx context.Context, key string) (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `SELECT value FROM configurations WHERE key = ?`

	var value string
	err := d.db.QueryRowContext(ctx, query, key).Scan(&value)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("hades.platform.database: configuration not found")
		}
		return "", fmt.Errorf("hades.platform.database: failed to get configuration: %w", err)
	}

	return value, nil
}

// CleanupExpiredData removes expired sessions and old data
func (d *Database) CleanupExpiredData(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Delete expired sessions
	_, err := d.db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP")
	if err != nil {
		return fmt.Errorf("hades.platform.database: failed to cleanup expired sessions: %w", err)
	}

	// Delete old audit logs (older than 90 days)
	_, err = d.db.ExecContext(ctx, "DELETE FROM audit_logs WHERE created_at < date('now', '-90 days')")
	if err != nil {
		return fmt.Errorf("hades.platform.database: failed to cleanup old audit logs: %w", err)
	}

	return nil
}
