package database

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestPostgreSQLOptimizations validates PostgreSQL-specific optimizations
func TestPostgreSQLOptimizations(t *testing.T) {
	config := &ManagerConfig{
		DBType:     PostgreSQL,
		PrimaryDSN: "postgres://user:pass@localhost/hades_test?sslmode=disable", // Test DSN
	}

	dm := NewDatabaseManager(config)

	// Test connection pool optimization
	maxOpen, maxIdle, lifetime := dm.GetOptimalPoolConfig()

	if maxOpen != 25 {
		t.Errorf("Expected MaxOpenConns 25 for PostgreSQL, got %d", maxOpen)
	}
	if maxIdle != 5 {
		t.Errorf("Expected MaxIdleConns 5 for PostgreSQL, got %d", maxIdle)
	}
	if lifetime != 30*time.Minute {
		t.Errorf("Expected ConnLifetime 30m for PostgreSQL, got %v", lifetime)
	}

	// Test query building for PostgreSQL
	queryTemplate := "SELECT * FROM governor_actions WHERE status = ? AND created_at >= ?"
	args := []interface{}{"approved", time.Now().UTC()}

	query, finalArgs := dm.BuildQuery(queryTemplate, args...)

	// Should convert ? to $1, $2 for PostgreSQL
	expectedQuery := "SELECT * FROM governor_actions WHERE status = $1 AND created_at >= $2"
	if query != expectedQuery {
		t.Errorf("Expected PostgreSQL query '%s', got '%s'", expectedQuery, query)
	}

	if len(finalArgs) != 2 {
		t.Errorf("Expected 2 final args, got %d", len(finalArgs))
	}
}

// TestUniversalPlaceholderConversion validates placeholder conversion logic
func TestUniversalPlaceholderConversion(t *testing.T) {
	tests := []struct {
		name          string
		dbType        DatabaseType
		queryTemplate string
		expectedQuery string
		args          []interface{}
	}{
		{
			name:          "PostgreSQL single placeholder",
			dbType:        PostgreSQL,
			queryTemplate: "SELECT * FROM table WHERE id = ?",
			expectedQuery: "SELECT * FROM table WHERE id = $1",
			args:          []interface{}{1},
		},
		{
			name:          "PostgreSQL multiple placeholders",
			dbType:        PostgreSQL,
			queryTemplate: "INSERT INTO table (a, b, c) VALUES (?, ?, ?)",
			expectedQuery: "INSERT INTO table (a, b, c) VALUES ($1, $2, $3)",
			args:          []interface{}{"val1", "val2", "val3"},
		},
		{
			name:          "MySQL placeholders unchanged",
			dbType:        MySQL,
			queryTemplate: "SELECT * FROM table WHERE id = ?",
			expectedQuery: "SELECT * FROM table WHERE id = ?",
			args:          []interface{}{1},
		},
		{
			name:          "SQLite placeholders unchanged",
			dbType:        SQLite,
			queryTemplate: "SELECT * FROM table WHERE id = ?",
			expectedQuery: "SELECT * FROM table WHERE id = ?",
			args:          []interface{}{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ManagerConfig{DBType: tt.dbType}
			dm := NewDatabaseManager(config)

			query, finalArgs := dm.BuildQuery(tt.queryTemplate, tt.args...)

			if query != tt.expectedQuery {
				t.Errorf("Expected query '%s', got '%s'", tt.expectedQuery, query)
			}

			if len(finalArgs) != len(tt.args) {
				t.Errorf("Expected %d args, got %d", len(tt.args), len(finalArgs))
			}
		})
	}
}

// TestUTC timestamp handling validates timezone consistency
func TestUTCTimestampHandling(t *testing.T) {
	config := &ManagerConfig{
		DBType:     SQLite,
		UseSQLite:  true,
		SQLitePath: fmt.Sprintf("/tmp/test_hades_utc_%d.db", time.Now().UnixNano()),
	}

	dm := NewDatabaseManager(config)
	ctx := context.Background()

	err := dm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer dm.Close()

	err = dm.CreateGovernorActionTable(ctx)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create action with specific timestamp
	testTime := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	action := &GovernorAction{
		ActionID:   "utc-test-action",
		ActionName: "UTC Test Action",
		Status:     string(GovernorActionStatusApproved),
		Approved:   true,
		CreatedAt:  testTime,
		UpdatedAt:  testTime,
	}

	err = dm.RecordGovernorAction(ctx, action)
	if err != nil {
		t.Fatalf("Failed to record action: %v", err)
	}

	// Retrieve action and verify timestamp consistency
	retrieved, err := dm.GetGovernorActionByID(ctx, "utc-test-action")
	if err != nil {
		t.Fatalf("Failed to retrieve action: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected action, got nil")
	}

	// Verify CreatedAt is preserved in UTC
	if !retrieved.CreatedAt.Equal(testTime) {
		t.Errorf("Expected CreatedAt %v, got %v", testTime, retrieved.CreatedAt)
	}

	// UpdatedAt should be set to the current time when recording
	// Allow for small time differences (within 1 second)
	timeDiff := retrieved.UpdatedAt.Sub(time.Now().UTC())
	if timeDiff < -time.Second || timeDiff > time.Second {
		t.Errorf("Expected UpdatedAt to be close to current time, got %v (diff: %v)", retrieved.UpdatedAt, timeDiff)
	}

	// Test time-based queries with UTC
	since := testTime.Add(-1 * time.Minute)
	actions, err := dm.GetRecentActions(ctx, since, 10)
	if err != nil {
		t.Fatalf("Failed to get recent actions: %v", err)
	}

	if len(actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(actions))
	}
}

// TestConnectionPoolOptimizations validates pool settings for different database types
func TestConnectionPoolOptimizations(t *testing.T) {
	tests := []struct {
		name             DatabaseType
		expectedMaxOpen  int
		expectedMaxIdle  int
		expectedLifetime time.Duration
	}{
		{PostgreSQL, 25, 5, 30 * time.Minute},
		{MySQL, 20, 10, 1 * time.Hour},
		{SQLite, 1, 0, 1 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.name), func(t *testing.T) {
			config := &ManagerConfig{DBType: tt.name}
			dm := NewDatabaseManager(config)

			maxOpen, maxIdle, lifetime := dm.GetOptimalPoolConfig()

			if maxOpen != tt.expectedMaxOpen {
				t.Errorf("Expected MaxOpenConns %d, got %d", tt.expectedMaxOpen, maxOpen)
			}
			if maxIdle != tt.expectedMaxIdle {
				t.Errorf("Expected MaxIdleConns %d, got %d", tt.expectedMaxIdle, maxIdle)
			}
			if lifetime != tt.expectedLifetime {
				t.Errorf("Expected ConnLifetime %v, got %v", tt.expectedLifetime, lifetime)
			}
		})
	}
}
