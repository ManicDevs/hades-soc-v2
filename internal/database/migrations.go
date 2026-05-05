package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// MigrationFile represents a database migration file
type MigrationFile struct {
	Version string
	Name    string
	SQL     string
}

// Migrator handles database migrations
type Migrator struct {
	db Database
}

// NewMigrator creates a new migrator instance
func NewMigrator(db Database) *Migrator {
	return &Migrator{db: db}
}

// RunMigrations executes all pending migrations
func (m *Migrator) RunMigrations(migrationsPath string) error {
	log.Printf("Starting database migrations from: %s", migrationsPath)

	// Create migrations table
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get migration files
	migrationFiles, err := m.getMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Run each migration
	for _, migration := range migrationFiles {
		if err := m.runMigration(migration); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Version, err)
		}
	}

	log.Printf("All migrations completed successfully")
	return nil
}

// createMigrationsTable creates the schema_migrations table
func (m *Migrator) createMigrationsTable() error {
	sqlDB, ok := m.db.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`

	_, err := sqlDB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// getMigrationFiles reads migration files from directory
func (m *Migrator) getMigrationFiles(migrationsPath string) ([]MigrationFile, error) {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []MigrationFile
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Extract version from filename (e.g., 001_create_users.sql)
		parts := strings.Split(file.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		version := parts[0]
		name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")

		// Read SQL content
		sqlPath := filepath.Join(migrationsPath, file.Name())
		sqlContent, err := os.ReadFile(sqlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		migrations = append(migrations, MigrationFile{
			Version: version,
			Name:    name,
			SQL:     string(sqlContent),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		vi, _ := strconv.Atoi(migrations[i].Version)
		vj, _ := strconv.Atoi(migrations[j].Version)
		return vi < vj
	})

	return migrations, nil
}

// runMigration executes a single migration
func (m *Migrator) runMigration(migration MigrationFile) error {
	// Check if migration already applied
	applied, err := m.isMigrationApplied(migration.Version)
	if err != nil {
		return fmt.Errorf("failed to check if migration %s is applied: %w", migration.Version, err)
	}

	if applied {
		log.Printf("Migration %s already applied, skipping", migration.Version)
		return nil
	}

	log.Printf("Running migration %s: %s", migration.Version, migration.Name)

	sqlDB, ok := m.db.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	// Begin transaction
	tx, err := sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	if err := m.recordMigration(tx, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	log.Printf("Migration %s completed successfully", migration.Version)
	return nil
}

// isMigrationApplied checks if a migration has been applied
func (m *Migrator) isMigrationApplied(version string) (bool, error) {
	sqlDB, ok := m.db.GetConnection().(*sql.DB)
	if !ok {
		return false, fmt.Errorf("database is not an SQL database")
	}

	query := `SELECT COUNT(*) FROM schema_migrations WHERE version = $1`
	var count int
	err := sqlDB.QueryRow(query, version).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}

	return count > 0, nil
}

// recordMigration records that a migration has been applied
func (m *Migrator) recordMigration(tx *sql.Tx, version, name string) error {
	query := `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`
	_, err := tx.Exec(query, version, name)
	if err != nil {
		return fmt.Errorf("failed to insert migration record: %w", err)
	}
	return nil
}
