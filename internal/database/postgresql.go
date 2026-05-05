package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQL specific implementation
type PostgreSQLDatabase struct {
	SQLDatabase
}

// NewPostgreSQLDatabase creates a new PostgreSQL database instance
func NewPostgreSQLDatabase() *PostgreSQLDatabase {
	return &PostgreSQLDatabase{
		SQLDatabase: SQLDatabase{dbType: PostgreSQL},
	}
}

// Connect establishes connection to PostgreSQL
func (p *PostgreSQLDatabase) Connect(config DatabaseConfig) error {
	p.config = config

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	p.db = db
	log.Printf("Successfully connected to PostgreSQL database")
	return nil
}

// GetTables returns list of tables in the database
func (p *PostgreSQLDatabase) GetTables() ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// ExecuteMigration runs a migration file
func (p *PostgreSQLDatabase) ExecuteMigration(migration string) error {
	_, err := p.db.Exec(migration)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}
	return nil
}

// CreateMigrationsTable creates the migrations tracking table
func (p *PostgreSQLDatabase) CreateMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`

	_, err := p.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// IsMigrationApplied checks if a migration has been applied
func (p *PostgreSQLDatabase) IsMigrationApplied(version string) (bool, error) {
	query := `SELECT COUNT(*) FROM schema_migrations WHERE version = $1`

	var count int
	err := p.db.QueryRow(query, version).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check migration: %w", err)
	}

	return count > 0, nil
}

// RecordMigration records that a migration has been applied
func (p *PostgreSQLDatabase) RecordMigration(version string) error {
	query := `INSERT INTO schema_migrations (version) VALUES ($1)`

	_, err := p.db.Exec(query, version)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}
