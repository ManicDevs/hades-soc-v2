package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	defaultManager *DatabaseManager
	managerOnce    sync.Once
)

type DatabaseManager struct {
	mu          sync.RWMutex
	primary     *sql.DB
	config      *ManagerConfig
	initialized bool
	connections map[string]*sql.DB
}

type ManagerConfig struct {
	PrimaryDSN   string
	Database     string
	MaxOpenConns int
	MaxIdleConns int
	ConnLifetime time.Duration
	UseSQLite    bool
	SQLitePath   string
}

func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		Database:     "hades",
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		ConnLifetime: 5 * time.Minute,
		UseSQLite:    false,
		SQLitePath:   "hades.db",
	}
}

func GetManager() *DatabaseManager {
	managerOnce.Do(func() {
		defaultManager = NewDatabaseManager(DefaultManagerConfig())
	})
	return defaultManager
}

func NewDatabaseManager(config *ManagerConfig) *DatabaseManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	return &DatabaseManager{
		config:      config,
		connections: make(map[string]*sql.DB),
	}
}

func (dm *DatabaseManager) Initialize(ctx context.Context) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.initialized {
		return nil
	}

	var db *sql.DB
	var err error

	if dm.config.UseSQLite {
		db, err = dm.connectSQLite()
	} else {
		db, err = dm.connectPostgreSQL()
	}

	if err != nil {
		log.Printf("Warning: Primary connection failed, falling back to SQLite: %v", err)
		db, err = dm.connectSQLite()
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
	}

	dm.primary = db
	dm.connections["primary"] = db
	dm.initialized = true

	log.Printf("DatabaseManager: Initialized with primary connection to %s",
		dm.config.Database)

	return nil
}

func (dm *DatabaseManager) connectPostgreSQL() (*sql.DB, error) {
	dsn := dm.config.PrimaryDSN
	if dsn == "" {
		dsn = "host=localhost port=5432 user=hades dbname=hades sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	dm.configurePool(db)
	log.Println("DatabaseManager: Connected to PostgreSQL")

	return db, nil
}

func (dm *DatabaseManager) connectSQLite() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dm.config.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Printf("Warning: Failed to enable WAL: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		log.Printf("Warning: Failed to enable foreign keys: %v", err)
	}

	dm.configurePool(db)
	log.Println("DatabaseManager: Connected to SQLite")

	return db, nil
}

func (dm *DatabaseManager) configurePool(db *sql.DB) {
	db.SetMaxOpenConns(dm.config.MaxOpenConns)
	db.SetMaxIdleConns(dm.config.MaxIdleConns)
	db.SetConnMaxLifetime(dm.config.ConnLifetime)
}

func (dm *DatabaseManager) GetConnection(ctx context.Context) (*sql.DB, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if !dm.initialized {
		return nil, fmt.Errorf("database manager not initialized")
	}

	return dm.primary, nil
}

func (dm *DatabaseManager) GetPrimary() *sql.DB {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.primary
}

func (dm *DatabaseManager) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.primary == nil {
		return nil, fmt.Errorf("no primary connection")
	}

	return dm.primary.QueryContext(ctx, query, args...)
}

func (dm *DatabaseManager) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.primary == nil {
		return nil
	}

	return dm.primary.QueryRowContext(ctx, query, args...)
}

func (dm *DatabaseManager) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.primary == nil {
		return nil, fmt.Errorf("no primary connection")
	}

	return dm.primary.ExecContext(ctx, query, args...)
}

func (dm *DatabaseManager) Begin(ctx context.Context) (*sql.Tx, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.primary == nil {
		return nil, fmt.Errorf("no primary connection")
	}

	return dm.primary.BeginTx(ctx, nil)
}

func (dm *DatabaseManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for name, conn := range dm.connections {
		if err := conn.Close(); err != nil {
			log.Printf("Warning: Failed to close connection %s: %v", name, err)
		}
	}

	dm.connections = make(map[string]*sql.DB)
	dm.primary = nil
	dm.initialized = false

	log.Println("DatabaseManager: Closed all connections")
	return nil
}

func (dm *DatabaseManager) Ping(ctx context.Context) error {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.primary == nil {
		return fmt.Errorf("no primary connection")
	}

	return dm.primary.PingContext(ctx)
}

func (dm *DatabaseManager) IsInitialized() bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.initialized
}

func (dm *DatabaseManager) GetStats() map[string]interface{} {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	stats := make(map[string]interface{})
	if dm.primary != nil {
		stats["primary"] = map[string]interface{}{
			"open_conns": dm.primary.Stats().OpenConnections,
			"idle_conns": dm.primary.Stats().Idle,
			"in_use":     dm.primary.Stats().InUse,
			"wait_count": dm.primary.Stats().WaitCount,
			"wait_time":  dm.primary.Stats().WaitDuration.String(),
		}
	}

	stats["initialized"] = dm.initialized
	stats["connections"] = len(dm.connections)

	return stats
}

func (dm *DatabaseManager) SetConfig(config *ManagerConfig) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.config = config
}

func (dm *DatabaseManager) GetConfig() *ManagerConfig {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.config
}
