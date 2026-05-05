package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestDatabaseManagerDriverSupport(t *testing.T) {
	tests := []struct {
		name     string
		dbType   DatabaseType
		expected string
	}{
		{"PostgreSQL placeholders", PostgreSQL, "$"},
		{"MySQL placeholders", MySQL, "?"},
		{"SQLite placeholders", SQLite, "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ManagerConfig{
				DBType:     tt.dbType,
				UseSQLite:  tt.dbType == SQLite,
				SQLitePath: ":memory:",
			}

			dm := NewDatabaseManager(config)

			// Test placeholder methods
			if dm.GetPlaceholder() != tt.expected {
				t.Errorf("GetPlaceholder() = %v, want %v", dm.GetPlaceholder(), tt.expected)
			}
			// For PostgreSQL, expect numbered placeholders; for MySQL/SQLite, expect just "?"
			expectedIndex1 := tt.expected
			expectedIndex2 := tt.expected
			if tt.dbType == PostgreSQL {
				expectedIndex1 = tt.expected + "1"
				expectedIndex2 = tt.expected + "2"
			}

			if dm.GetPlaceholderForIndex(1) != expectedIndex1 {
				t.Errorf("GetPlaceholderForIndex(1) = %v, want %v", dm.GetPlaceholderForIndex(1), expectedIndex1)
			}
			if dm.GetPlaceholderForIndex(2) != expectedIndex2 {
				t.Errorf("GetPlaceholderForIndex(2) = %v, want %v", dm.GetPlaceholderForIndex(2), expectedIndex2)
			}

			// Test type detection methods
			if dm.IsPostgreSQL() != (tt.dbType == PostgreSQL) {
				t.Errorf("IsPostgreSQL() = %v, want %v", dm.IsPostgreSQL(), tt.dbType == PostgreSQL)
			}
			if dm.IsMySQL() != (tt.dbType == MySQL) {
				t.Errorf("IsMySQL() = %v, want %v", dm.IsMySQL(), tt.dbType == MySQL)
			}
			if dm.IsSQLite() != (tt.dbType == SQLite) {
				t.Errorf("IsSQLite() = %v, want %v", dm.IsSQLite(), tt.dbType == SQLite)
			}
		})
	}
}

func TestGovernorActionCrossDB(t *testing.T) {
	// Test with SQLite (in-memory)
	t.Run("SQLite", func(t *testing.T) {
		config := &ManagerConfig{
			DBType:     SQLite,
			UseSQLite:  true,
			SQLitePath: fmt.Sprintf("/tmp/test_hades_%d.db", time.Now().UnixNano()),
		}

		dm := NewDatabaseManager(config)
		ctx := context.Background()

		err := dm.Initialize(ctx)
		if err != nil {
			t.Fatalf("Failed to initialize database: %v", err)
		}
		defer dm.Close()

		// Test table creation
		err = dm.CreateGovernorActionTable(ctx)
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		// Test recording an action
		action := &GovernorAction{
			ActionID:          "test-action-1",
			ActionName:        "Test Block",
			Target:            "192.168.1.100",
			Reasoning:         "Malicious activity detected",
			Requester:         "threat-engine",
			Status:            string(GovernorActionStatusApproved),
			RequiresApproval:  false,
			Approved:          true,
			RequiresManualAck: false,
			ExecutionTime:     100,
			Metadata:          `{"test": true}`,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err = dm.RecordGovernorAction(ctx, action)
		if err != nil {
			t.Fatalf("Failed to record action: %v", err)
		}

		// Test getting block count (uses Go-calculated timestamps)
		count, err := dm.GetBlockCountInLastHour(ctx)
		if err != nil {
			t.Fatalf("Failed to get block count: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected block count 1, got %d", count)
		}

		// Test getting recent actions
		actions, err := dm.GetRecentActions(ctx, time.Now().Add(-1*time.Hour), 10)
		if err != nil {
			t.Fatalf("Failed to get recent actions: %v", err)
		}
		if len(actions) != 1 {
			t.Errorf("Expected 1 action, got %d", len(actions))
		}
		if actions[0].ActionID != "test-action-1" {
			t.Errorf("Expected action ID test-action-1, got %s", actions[0].ActionID)
		}

		// Test getting stats (uses Go-calculated timestamps)
		stats, err := dm.GetGovernorStats(ctx)
		if err != nil {
			t.Fatalf("Failed to get stats: %v", err)
		}
		if stats["approved_last_hour"] != 1 {
			t.Errorf("Expected approved_last_hour 1, got %v", stats["approved_last_hour"])
		}
		if stats["total_last_24h"] != 1 {
			t.Errorf("Expected total_last_24h 1, got %v", stats["total_last_24h"])
		}
		if stats["approved_last_24h"] != 1 {
			t.Errorf("Expected approved_last_24h 1, got %v", stats["approved_last_24h"])
		}

		// Test getting action by ID
		retrieved, err := dm.GetGovernorActionByID(ctx, "test-action-1")
		if err != nil {
			t.Fatalf("Failed to get action by ID: %v", err)
		}
		if retrieved == nil {
			t.Fatal("Expected action, got nil")
		}
		if retrieved.ActionName != "Test Block" {
			t.Errorf("Expected action name 'Test Block', got %s", retrieved.ActionName)
		}

		// Test cleanup (uses Go-calculated timestamps)
		err = dm.CleanupOldGovernorActions(ctx, 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to cleanup: %v", err)
		}
	})
}

func TestGovernorActionWithEnvironmentDB(t *testing.T) {
	// Only run if database environment variables are set
	dbType := os.Getenv("TEST_DB_TYPE")
	if dbType == "" {
		t.Skip("TEST_DB_TYPE not set, skipping integration test")
	}

	var config *ManagerConfig

	switch dbType {
	case "postgres":
		config = &ManagerConfig{
			DBType:     PostgreSQL,
			PrimaryDSN: os.Getenv("TEST_POSTGRES_DSN"),
		}
	case "mysql":
		config = &ManagerConfig{
			DBType:     MySQL,
			PrimaryDSN: os.Getenv("TEST_MYSQL_DSN"),
		}
	default:
		t.Skipf("Unsupported TEST_DB_TYPE: %s", dbType)
	}

	if config.PrimaryDSN == "" {
		t.Skip("TEST_*_DSN not set, skipping integration test")
	}

	dm := NewDatabaseManager(config)
	ctx := context.Background()

	err := dm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer dm.Close()

	// Test basic functionality works with the real database
	err = dm.CreateGovernorActionTable(ctx)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Verify table exists using driver-aware method
	exists, err := dm.EnsureTableExists(ctx, "governor_actions")
	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}
	if !exists {
		t.Error("Expected table to exist")
	}

	t.Logf("Successfully tested with %s database", dbType)
}

func TestTimeLogicConsistency(t *testing.T) {
	// Test that time calculations are consistent across database types
	config := &ManagerConfig{
		DBType:     SQLite,
		UseSQLite:  true,
		SQLitePath: fmt.Sprintf("/tmp/test_hades_%d.db", time.Now().UnixNano()),
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

	// Create actions at different times
	now := time.Now()
	actions := []*GovernorAction{
		{
			ActionID:   "action-59m-ago",
			ActionName: "Recent Action",
			Status:     string(GovernorActionStatusApproved),
			Approved:   true,
			CreatedAt:  now.Add(-59 * time.Minute),
			UpdatedAt:  now.Add(-59 * time.Minute),
		},
		{
			ActionID:   "action-30m-ago",
			ActionName: "Very Recent Action",
			Status:     string(GovernorActionStatusApproved),
			Approved:   true,
			CreatedAt:  now.Add(-30 * time.Minute),
			UpdatedAt:  now.Add(-30 * time.Minute),
		},
		{
			ActionID:   "action-2h-ago",
			ActionName: "Very Old Action",
			Status:     string(GovernorActionStatusApproved),
			Approved:   true,
			CreatedAt:  now.Add(-2 * time.Hour),
			UpdatedAt:  now.Add(-2 * time.Hour),
		},
	}

	for _, action := range actions {
		err = dm.RecordGovernorAction(ctx, action)
		if err != nil {
			t.Fatalf("Failed to record action %s: %v", action.ActionID, err)
		}
	}

	// Test that GetBlockCountInLastHour only counts actions within the last hour
	count, err := dm.GetBlockCountInLastHour(ctx)
	if err != nil {
		t.Fatalf("Failed to get block count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected block count 2, got %d", count)
	}

	// Test that GetRecentActions respects the time window
	recent, err := dm.GetRecentActions(ctx, now.Add(-45*time.Minute), 10)
	if err != nil {
		t.Fatalf("Failed to get recent actions: %v", err)
	}
	if len(recent) != 1 {
		t.Errorf("Expected 1 recent action, got %d", len(recent))
	}

	// Test cleanup with time-based logic
	err = dm.CleanupOldGovernorActions(ctx, 90*time.Minute)
	if err != nil {
		t.Fatalf("Failed to cleanup: %v", err)
	}

	// Verify the 2h old action was cleaned up
	remaining, err := dm.GetRecentActions(ctx, now.Add(-3*time.Hour), 10)
	if err != nil {
		t.Fatalf("Failed to get remaining actions: %v", err)
	}
	if len(remaining) != 2 {
		t.Errorf("Expected 2 remaining actions, got %d", len(remaining))
	}
}
