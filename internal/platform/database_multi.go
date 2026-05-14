package platform

import (
	"context"
	"database/sql"
	"fmt"
	"hades-v2/internal/database"
	"log"
	"sync"
	"time"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	DatabaseSQLite   DatabaseType = "sqlite"
	DatabasePostgres DatabaseType = "postgres"
	DatabaseMySQL    DatabaseType = "mysql"
)

// DatabaseConfigMulti holds multi-database configuration
type DatabaseConfigMulti struct {
	Type              DatabaseType  `json:"type"`
	Host              string        `json:"host"`
	Port              int           `json:"port"`
	Database          string        `json:"database"`
	Username          string        `json:"username"`
	Password          string        `json:"password"`
	SSLMode           string        `json:"ssl_mode"`
	MaxConnections    int           `json:"max_connections"`
	ConnTimeout       time.Duration `json:"conn_timeout"`
	EnableWAL         bool          `json:"enable_wal"`
	EnableForeignKeys bool          `json:"enable_foreign_keys"`
	Charset           string        `json:"charset"`
}

// DefaultDatabaseConfigMulti returns sensible defaults for each database type
func DefaultDatabaseConfigMulti(dbType DatabaseType) *DatabaseConfigMulti {
	switch dbType {
	case DatabaseSQLite:
		return &DatabaseConfigMulti{
			Type:              DatabaseSQLite,
			Database:          "hades.db",
			MaxConnections:    10,
			ConnTimeout:       30 * time.Second,
			EnableWAL:         true,
			EnableForeignKeys: true,
		}
	case DatabasePostgres:
		return &DatabaseConfigMulti{
			Type:              DatabasePostgres,
			Host:              "localhost",
			Port:              5432,
			Database:          "hades",
			Username:          "postgres",
			SSLMode:           "disable",
			MaxConnections:    20,
			ConnTimeout:       30 * time.Second,
			EnableForeignKeys: true,
		}
	case DatabaseMySQL:
		return &DatabaseConfigMulti{
			Type:              DatabaseMySQL,
			Host:              "localhost",
			Port:              3306,
			Database:          "hades",
			Username:          "root",
			Charset:           "utf8mb4",
			MaxConnections:    20,
			ConnTimeout:       30 * time.Second,
			EnableForeignKeys: true,
		}
	default:
		return DefaultDatabaseConfigMulti(DatabaseSQLite)
	}
}

// DatabaseMulti provides multi-database support with connection pooling
type DatabaseMulti struct {
	config *DatabaseConfigMulti
	db     *sql.DB
	mu     sync.RWMutex
}

// NewDatabaseMulti creates a new multi-database instance using DatabaseManager
func NewDatabaseMulti(config *DatabaseConfigMulti) (*DatabaseMulti, error) {
	if config == nil {
		config = DefaultDatabaseConfigMulti(DatabaseSQLite)
	}

	// Map platform.DatabaseType to database.DatabaseType
	var dbType database.DatabaseType
	switch config.Type {
	case DatabaseSQLite:
		dbType = database.SQLite
	case DatabasePostgres:
		dbType = database.PostgreSQL
	case DatabaseMySQL:
		dbType = database.MySQL
	default:
		dbType = database.SQLite
	}

	managerConfig := &database.ManagerConfig{
		Database:     config.Database,
		MaxOpenConns: config.MaxConnections,
		MaxIdleConns: 5,
		ConnLifetime: 5 * time.Minute,
		UseSQLite:    config.Type == DatabaseSQLite || config.Type == "",
		SQLitePath:   config.Database,
		DBType:       dbType,
	}

	if config.Type == DatabasePostgres {
		managerConfig.PrimaryDSN = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)
	}

	// Create a fresh database manager instead of using singleton
	mgr := database.NewDatabaseManager(managerConfig)
	if mgr == nil {
		return nil, fmt.Errorf("hades.platform.database: failed to create database manager")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := mgr.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("hades.platform.database: failed to initialize database manager: %w", err)
	}

	db, err := mgr.GetConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.database: failed to get connection: %w", err)
	}

	multiDB := &DatabaseMulti{
		config: config,
		db:     db,
	}

	if err := multiDB.initialize(); err != nil {
		return nil, fmt.Errorf("hades.platform.database: %w", err)
	}

	return multiDB, nil
}

// initialize sets up the database with proper configuration
func (dm *DatabaseMulti) initialize() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Configure connection pool
	dm.db.SetMaxOpenConns(dm.config.MaxConnections)
	dm.db.SetMaxIdleConns(dm.config.MaxConnections / 2)
	dm.db.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), dm.config.ConnTimeout)
	defer cancel()

	if err := dm.db.PingContext(ctx); err != nil {
		return fmt.Errorf("hades.platform.database: failed to ping database: %w", err)
	}

	// Configure database-specific settings
	if err := dm.configureDatabase(); err != nil {
		return fmt.Errorf("hades.platform.database: failed to configure database: %w", err)
	}

	// Create tables
	if err := dm.createTables(); err != nil {
		return fmt.Errorf("hades.platform.database: failed to create tables: %w", err)
	}

	return nil
}

// configureDatabase applies database-specific configurations
func (dm *DatabaseMulti) configureDatabase() error {
	switch dm.config.Type {
	case DatabaseSQLite:
		if dm.config.EnableWAL {
			if _, err := dm.db.Exec("PRAGMA journal_mode=WAL"); err != nil {
				return fmt.Errorf("hades.platform.database: failed to enable WAL: %w", err)
			}
		}
		if dm.config.EnableForeignKeys {
			if _, err := dm.db.Exec("PRAGMA foreign_keys=ON"); err != nil {
				return fmt.Errorf("hades.platform.database: failed to enable foreign keys: %w", err)
			}
		}
	case DatabasePostgres:
		// PostgreSQL-specific configurations
		if _, err := dm.db.Exec("SET timezone TO 'UTC'"); err != nil {
			return fmt.Errorf("hades.platform.database: failed to set timezone: %w", err)
		}
	case DatabaseMySQL:
		// MySQL-specific configurations
		if _, err := dm.db.Exec("SET sql_mode = 'STRICT_TRANS_TABLES'"); err != nil {
			return fmt.Errorf("hades.platform.database: failed to set SQL mode: %w", err)
		}
	}
	return nil
}

// createTables creates the database schema with proper data types for each database
func (dm *DatabaseMulti) createTables() error {
	var queries []string

	log.Printf("DatabaseMulti.createTables: Creating tables for type %v", dm.config.Type)

	switch dm.config.Type {
	case DatabaseSQLite:
		queries = dm.getSQLiteSchema()
	case DatabasePostgres:
		queries = dm.getPostgresSchema()
	case DatabaseMySQL:
		queries = dm.getMySQLSchema()
	default:
		queries = dm.getSQLiteSchema()
	}

	for i, query := range queries {
		log.Printf("DatabaseMulti.createTables: Executing query %d/%d", i+1, len(queries))
		if _, err := dm.db.Exec(query); err != nil {
			return fmt.Errorf("hades.platform.database: failed to execute query: %s, error: %w", query, err)
		}
	}

	log.Printf("DatabaseMulti.createTables: All %d queries executed successfully", len(queries))
	return nil
}

// getSQLiteSchema returns SQLite-specific schema
func (dm *DatabaseMulti) getSQLiteSchema() []string {
	return []string{
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
}

// getPostgresSchema returns PostgreSQL-specific schema
func (dm *DatabaseMulti) getPostgresSchema() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			role VARCHAR(50) NOT NULL,
			password_hash TEXT NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			last_login TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			ip_address INET,
			user_agent TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS scan_results (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			module_name VARCHAR(255) NOT NULL,
			target TEXT NOT NULL,
			status VARCHAR(50) NOT NULL,
			result_data JSONB,
			error_message TEXT,
			start_time TIMESTAMP WITH TIME ZONE NOT NULL,
			end_time TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID,
			action VARCHAR(255) NOT NULL,
			resource TEXT,
			details JSONB,
			ip_address INET,
			user_agent TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS configurations (
			key VARCHAR(255) PRIMARY KEY,
			value JSONB NOT NULL,
			description TEXT,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
	}
}

// getMySQLSchema returns MySQL-specific schema
func (dm *DatabaseMulti) getMySQLSchema() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			role VARCHAR(50) NOT NULL,
			password_hash TEXT NOT NULL,
			is_active TINYINT(1) DEFAULT 1,
			last_login TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
			user_id VARCHAR(36) NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			ip_address VARCHAR(45),
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS scan_results (
			id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
			module_name VARCHAR(255) NOT NULL,
			target TEXT NOT NULL,
			status VARCHAR(50) NOT NULL,
			result_data JSON,
			error_message TEXT,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
			user_id VARCHAR(36),
			action VARCHAR(255) NOT NULL,
			resource TEXT,
			details JSON,
			ip_address VARCHAR(45),
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS configurations (
			key VARCHAR(255) PRIMARY KEY,
			value JSON NOT NULL,
			description TEXT,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
}

// Close closes the database connection
func (dm *DatabaseMulti) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	return dm.db.Close()
}

// GetDB returns the underlying database connection
func (dm *DatabaseMulti) GetDB() *sql.DB {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.db
}

// Migration represents a database migration
type Migration struct {
	Version     int    `json:"version"`
	Description string `json:"description"`
	SQL         string `json:"sql"`
}

// MigrationManager handles database migrations
type MigrationManager struct {
	db *DatabaseMulti
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *DatabaseMulti) *MigrationManager {
	return &MigrationManager{db: db}
}

// CreateMigrationsTable creates the migrations tracking table
func (mm *MigrationManager) CreateMigrationsTable() error {
	var query string
	switch mm.db.config.Type {
	case DatabasePostgres:
		query = `CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`
	case DatabaseMySQL:
		query = `CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`
	default:
		query = `CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	}

	_, err := mm.db.db.Exec(query)
	return err
}

// GetAppliedMigrations returns list of applied migration versions
func (mm *MigrationManager) GetAppliedMigrations() ([]int, error) {
	rows, err := mm.db.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// ApplyMigration applies a single migration
func (mm *MigrationManager) ApplyMigration(migration Migration) error {
	tx, err := mm.db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			fmt.Printf("Warning: failed to rollback transaction: %v\n", err)
		}
	}()

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("hades.platform.database: migration SQL failed: %w", err)
	}

	// Record migration
	_, err = tx.Exec("INSERT INTO schema_migrations (version, description) VALUES (?, ?)",
		migration.Version, migration.Description)
	if err != nil {
		return fmt.Errorf("hades.platform.database: failed to record migration: %w", err)
	}

	return tx.Commit()
}

// Migrate applies all pending migrations
func (mm *MigrationManager) Migrate(migrations []Migration) error {
	if err := mm.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("hades.platform.database: failed to create migrations table: %w", err)
	}

	applied, err := mm.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("hades.platform.database: failed to get applied migrations: %w", err)
	}

	appliedSet := make(map[int]bool)
	for _, v := range applied {
		appliedSet[v] = true
	}

	for _, migration := range migrations {
		if !appliedSet[migration.Version] {
			if err := mm.ApplyMigration(migration); err != nil {
				return fmt.Errorf("hades.platform.database: migration %d failed: %w", migration.Version, err)
			}
		}
	}

	return nil
}
