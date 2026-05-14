package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	// PostgreSQL driver
	_ "github.com/lib/pq"
	// MySQL driver
	_ "github.com/go-sql-driver/mysql"
	// SQLite driver (pure Go - no CGO required)
	_ "modernc.org/sqlite"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
	SQLite     DatabaseType = "sqlite"
	// MongoDB and Redis will be added later
)

// Database interface for abstraction
type Database interface {
	Connect(config DatabaseConfig) error
	Close() error
	Ping() error
	GetType() DatabaseType
	GetConnection() interface{}
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     DatabaseType           `json:"type"`
	Host     string                 `json:"host"`
	Port     int                    `json:"port"`
	Database string                 `json:"database"`
	Username string                 `json:"username"`
	Password string                 `json:"password"`
	SSLMode  string                 `json:"ssl_mode"`
	Options  map[string]interface{} `json:"options"`
}

// SQLDatabase extends Database for SQL databases
type SQLDatabase struct {
	db     *sql.DB
	config DatabaseConfig
	dbType DatabaseType
}

// NewDatabase creates a new database instance based on type
func NewDatabase(dbType DatabaseType) Database {
	switch dbType {
	case PostgreSQL:
		return &SQLDatabase{dbType: PostgreSQL}
	case MySQL:
		return &SQLDatabase{dbType: MySQL}
	case SQLite:
		return &SQLDatabase{dbType: SQLite}
	default:
		return &SQLDatabase{dbType: SQLite} // Default to SQLite for development
	}
}

// Connect establishes database connection
func (s *SQLDatabase) Connect(config DatabaseConfig) error {
	s.config = config
	s.dbType = config.Type

	var dsn string
	switch s.dbType {
	case PostgreSQL:
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)
	case MySQL:
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			config.Username, config.Password, config.Host, config.Port, config.Database)
	case SQLite:
		dsn = config.Database
	}

	driverName := "postgres"
	switch s.dbType {
	case MySQL:
		driverName = "mysql"
	case SQLite:
		driverName = "sqlite"
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", s.dbType, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping %s: %w", s.dbType, err)
	}

	s.db = db
	log.Printf("Successfully connected to %s database", s.dbType)
	return nil
}

// Close closes database connection
func (s *SQLDatabase) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Ping tests database connection
func (s *SQLDatabase) Ping() error {
	if s.db != nil {
		return s.db.Ping()
	}
	return fmt.Errorf("database not connected")
}

// GetType returns database type
func (s *SQLDatabase) GetType() DatabaseType {
	return s.dbType
}

// GetConnection returns the underlying database connection
func (s *SQLDatabase) GetConnection() interface{} {
	return s.db
}

// GetDB returns SQL database connection (helper method)
func (s *SQLDatabase) GetDB() *sql.DB {
	return s.db
}
