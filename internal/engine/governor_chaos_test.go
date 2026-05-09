package engine

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"hades-v2/internal/database"
)

// cleanupTestDatabase removes the test database file to ensure clean state
func cleanupTestDatabase() {
	dbFiles := []string{
		"hades_test.db",
		"hades_test.db-shm",
		"hades_test.db-wal",
	}
	for _, file := range dbFiles {
		os.Remove(file)
	}
}

// TestDatabaseManager creates a test database manager with SQLite
func NewTestDatabaseManager(t *testing.T) *database.DatabaseManager {
	// Clean up any existing test database
	cleanupTestDatabase()

if os.Getenv("HADES_DB_ENCRYPTION_KEY") == "" && os.Getenv("HADES_ALLOW_INSECURE_DEV_DB_KEY") != "true" {
		t.Skip("Requires HADES_DB_ENCRYPTION_KEY or HADES_ALLOW_INSECURE_DEV_DB_KEY=true")
	}

	config := &database.ManagerConfig{
		UseSQLite:  true,
		SQLitePath: "file:hades_test.db?cache=shared&_journal_mode=WAL",
	}

	dbManager := database.NewDatabaseManager(config)
	if dbManager == nil {
		t.Skip("Database manager initialization failed (encryption key required)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbManager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Create governor actions table for testing
	if err := dbManager.CreateGovernorActionTable(ctx); err != nil {
		t.Fatalf("Failed to create governor actions table: %v", err)
	}

	// Wait for table creation to settle
	time.Sleep(100 * time.Millisecond)

	return dbManager
}

// TestAutomatedStorm simulates 10 block actions within 10 minutes
func TestAutomatedStorm(t *testing.T) {
	// Setup
	testDB := NewTestDatabaseManager(t)
	defer testDB.Close()

	governor := NewSafetyGovernor(testDB)
	governor.Start()
	defer governor.Stop()

	// Test parameters
	totalActions := 10
	stormDuration := 10 * time.Minute
	intervalBetweenActions := stormDuration / time.Duration(totalActions)

	var wg sync.WaitGroup
	results := make(chan *ActionResponse, totalActions)
	actionRequests := make([]ActionRequest, totalActions)

	// Create action requests
	for i := 0; i < totalActions; i++ {
		actionRequests[i] = ActionRequest{
			ActionName:       fmt.Sprintf("block_action_%d", i+1),
			Target:           fmt.Sprintf("target_%d", i+1),
			Reasoning:        fmt.Sprintf("Automated storm test action %d", i+1),
			RequiresApproval: i >= 5, // Actions 6-10 require manual approval
			Requester:        "chaos_test",
			Timestamp:        time.Now(),
			Metadata:         map[string]interface{}{"test_id": i + 1},
		}
	}

	// Start the automated storm
	t.Logf("Starting automated storm: %d actions over %v", totalActions, stormDuration)
	startTime := time.Now()

	for i := 0; i < totalActions; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Simulate realistic timing between actions
			time.Sleep(time.Duration(index) * intervalBetweenActions)

			response, err := governor.RequestAction(actionRequests[index])
			if err != nil {
				t.Errorf("Action %d returned error: %v", index+1, err)
				return
			}

			results <- response
		}(i)
	}

	// Wait for all actions to complete
	wg.Wait()
	close(results)

	// Collect and analyze results
	var approvedActions, blockedActions, manualAckActions int
	var responses []*ActionResponse

	for response := range results {
		responses = append(responses, response)
		if response.Approved {
			approvedActions++
		} else if response.RequiresManualAck {
			manualAckActions++
		} else {
			blockedActions++
		}
	}

	// Verify the Safety Governor behavior
	t.Logf("Storm completed in %v", time.Since(startTime))
	t.Logf("Results: %d approved, %d blocked, %d require manual ACK",
		approvedActions, blockedActions, manualAckActions)

	// Test assertions
	if approvedActions != 5 {
		t.Errorf("Expected exactly 5 approved actions, got %d", approvedActions)
	}

	if manualAckActions != 5 {
		t.Errorf("Expected exactly 5 manual ACK actions (6-10), got %d", manualAckActions)
	}

	if blockedActions != 0 {
		t.Errorf("Expected 0 blocked actions (all should be approved or require manual ACK), got %d", blockedActions)
	}

	// Verify that actions 1-5 are approved
	for i, response := range responses[:5] {
		if !response.Approved {
			t.Errorf("Action %d should be approved but was blocked", i+1)
		}
		if response.RequiresManualAck {
			t.Errorf("Action %d should not require manual ACK but does", i+1)
		}
	}

	// Verify that actions 6-10 require manual ACK
	for i, response := range responses[5:] {
		if response.Approved {
			t.Errorf("Action %d should require manual ACK but was approved", i+6)
		}
		if !response.RequiresManualAck {
			t.Errorf("Action %d should require manual ACK but does not", i+6)
		}
	}

	// Test governor status
	status := governor.GetStatus()
	if currentCount, ok := status["current_block_count"].(int); !ok || currentCount != 5 {
		t.Errorf("Expected current_block_count=5, got %v", status["current_block_count"])
	}
	if remaining, ok := status["remaining_blocks"].(int); !ok || remaining != 0 {
		t.Errorf("Expected remaining_blocks=0, got %v", status["remaining_blocks"])
	}
}

// TestGovernorHourlyReset tests that the governor properly resets after an hour
func TestGovernorHourlyReset(t *testing.T) {
	testDB := NewTestDatabaseManager(t)
	defer testDB.Close()

	// Explicitly create table before inserting data
	if err := testDB.CreateGovernorActionTable(context.Background()); err != nil {
		t.Fatalf("Failed to create governor actions table: %v", err)
	}

	governor := NewSafetyGovernor(testDB)
	governor.Start()
	defer governor.Stop()

	// Simulate an action from more than 1 hour ago by directly inserting into database
	oldAction := &database.GovernorAction{
		ActionID:          "old_action_test",
		ActionName:        "old_action",
		Target:            "old_target",
		Reasoning:         "Old action for reset test",
		Requester:         "test",
		Status:            string(database.GovernorActionStatusApproved),
		RequiresApproval:  false,
		Approved:          true,
		RequiresManualAck: false,
		CreatedAt:         time.Now().Add(-61 * time.Minute), // More than 1 hour ago
		UpdatedAt:         time.Now().Add(-61 * time.Minute),
	}

	// Record the old action
	if err := testDB.RecordGovernorAction(context.Background(), oldAction); err != nil {
		t.Fatalf("Failed to record old action: %v", err)
	}

	// Test action that should be approved since old action is outside 1-hour window
	action := ActionRequest{
		ActionName:       "test_action",
		Target:           "test_target",
		Reasoning:        "Testing hourly reset",
		RequiresApproval: false,
		Requester:        "test",
		Timestamp:        time.Now(),
	}

	response, err := governor.RequestAction(action)
	if err != nil {
		t.Fatalf("RequestAction failed: %v", err)
	}

	if !response.Approved {
		t.Error("Action should be approved after hourly reset")
	}
}

// TestConcurrentRequests tests concurrent action requests
func TestConcurrentRequests(t *testing.T) {
	testDB := NewTestDatabaseManager(t)
	defer testDB.Close()

	governor := NewSafetyGovernor(testDB)
	governor.Start()
	defer governor.Stop()

	// Send 5 concurrent requests that should all be approved
	var wg sync.WaitGroup
	results := make(chan *ActionResponse, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			action := ActionRequest{
				ActionName:       fmt.Sprintf("concurrent_action_%d", id),
				Target:           fmt.Sprintf("target_%d", id),
				Reasoning:        "Concurrent test",
				RequiresApproval: false,
				Requester:        "concurrent_test",
				Timestamp:        time.Now(),
			}

			response, err := governor.RequestAction(action)
			if err != nil {
				t.Errorf("Concurrent action %d failed: %v", id, err)
				return
			}

			results <- response
		}(i)
	}

	wg.Wait()
	close(results)

	// Count approved actions
	approvedCount := 0
	for response := range results {
		if response.Approved {
			approvedCount++
		}
	}

	if approvedCount != 5 {
		t.Errorf("Expected all 5 concurrent actions to be approved, got %d", approvedCount)
	}

	// Verify governor status
	status := governor.GetStatus()
	if currentCount, ok := status["current_block_count"].(int); !ok || currentCount != 5 {
		t.Errorf("Expected current_block_count=5, got %v", status["current_block_count"])
	}
}

// BenchmarkGovernorRequestAction benchmarks the RequestAction performance
func BenchmarkGovernorRequestAction(b *testing.B) {
	// Create a test database manager for benchmarking
	config := &database.ManagerConfig{
		UseSQLite:  true,
		SQLitePath: ":memory:",
	}

	testDB := database.NewDatabaseManager(config)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := testDB.Initialize(ctx); err != nil {
		b.Fatalf("Failed to initialize benchmark database: %v", err)
	}
	defer testDB.Close()

	governor := NewSafetyGovernor(testDB)
	governor.Start()
	defer governor.Stop()

	action := ActionRequest{
		ActionName:       "benchmark_action",
		Target:           "benchmark_target",
		Reasoning:        "Benchmark test",
		RequiresApproval: false,
		Requester:        "benchmark",
		Timestamp:        time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := governor.RequestAction(action)
		if err != nil {
			b.Fatalf("RequestAction failed: %v", err)
		}
	}
}

// TestServiceRestartPersistence validates that block count persists across service restarts
func TestServiceRestartPersistence(t *testing.T) {
	if os.Getenv("HADES_DB_ENCRYPTION_KEY") == "" && os.Getenv("HADES_ALLOW_INSECURE_DEV_DB_KEY") != "true" {
		t.Skip("Requires HADES_DB_ENCRYPTION_KEY or HADES_ALLOW_INSECURE_DEV_DB_KEY=true")
	}

	// Create a persistent database (using file instead of memory)
	dbFile := "/tmp/test_governor_restart.db"

	// Clean up any existing database file
	os.Remove(dbFile)

	config := &database.ManagerConfig{
		UseSQLite:  true,
		SQLitePath: dbFile,
	}

	// Cleanup function
	defer func() {
		if testDB := database.NewDatabaseManager(config); testDB != nil {
			testDB.Close()
		}
		os.Remove(dbFile) // Clean up test database file
	}()

	// Phase 1: Create initial governor and perform 3 actions
	t.Log("Phase 1: Initial governor instance - performing 3 actions")

	testDB1 := database.NewDatabaseManager(config)
	ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel1()

	if err := testDB1.Initialize(ctx1); err != nil {
		t.Fatalf("Failed to initialize first database: %v", err)
	}
	defer testDB1.Close()

	governor1 := NewSafetyGovernor(testDB1)
	governor1.Start()
	defer governor1.Stop()

	// Perform 3 approved actions
	for i := 0; i < 3; i++ {
		action := ActionRequest{
			ActionName:       fmt.Sprintf("restart_action_%d", i+1),
			Target:           fmt.Sprintf("target_%d", i+1),
			Reasoning:        "Service restart test action",
			RequiresApproval: false,
			Requester:        "restart_test",
			Timestamp:        time.Now(),
		}

		response, err := governor1.RequestAction(action)
		if err != nil {
			t.Fatalf("Action %d failed in phase 1: %v", i+1, err)
		}

		if !response.Approved {
			t.Errorf("Action %d should be approved in phase 1", i+1)
		}
	}

	// Check status after first phase
	status1 := governor1.GetStatus()
	if approvedCount, ok := status1["approved_last_hour"].(int); !ok || approvedCount != 3 {
		t.Errorf("Expected 3 approved actions in phase 1, got %v", status1["approved_last_hour"])
	}

	t.Logf("Phase 1 complete: %v", status1)

	// Phase 2: Simulate service restart - create new governor instance with same database
	t.Log("Phase 2: Service restart - creating new governor instance")

	// Wait a moment to ensure timestamps are different
	time.Sleep(100 * time.Millisecond)

	testDB2 := database.NewDatabaseManager(config)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	if err := testDB2.Initialize(ctx2); err != nil {
		t.Fatalf("Failed to initialize second database: %v", err)
	}
	defer testDB2.Close()

	governor2 := NewSafetyGovernor(testDB2)
	governor2.Start()
	defer governor2.Stop()

	// Check that the new governor sees the previous actions
	status2 := governor2.GetStatus()
	if approvedCount, ok := status2["approved_last_hour"].(int); !ok || approvedCount != 3 {
		t.Errorf("Expected 3 approved actions after restart, got %v", status2["approved_last_hour"])
	}

	if remaining, ok := status2["remaining_blocks"].(int); !ok || remaining != 2 {
		t.Errorf("Expected 2 remaining blocks after restart, got %v", status2["remaining_blocks"])
	}

	t.Logf("Phase 2 complete (after restart): %v", status2)

	// Phase 3: Perform 2 more actions to reach the limit
	t.Log("Phase 3: Performing 2 more actions to reach limit")

	for i := 3; i < 5; i++ {
		action := ActionRequest{
			ActionName:       fmt.Sprintf("restart_action_%d", i+1),
			Target:           fmt.Sprintf("target_%d", i+1),
			Reasoning:        "Service restart test action",
			RequiresApproval: false,
			Requester:        "restart_test",
			Timestamp:        time.Now(),
		}

		response, err := governor2.RequestAction(action)
		if err != nil {
			t.Fatalf("Action %d failed in phase 3: %v", i+1, err)
		}

		if !response.Approved {
			t.Errorf("Action %d should be approved in phase 3", i+1)
		}
	}

	// Check status after reaching limit
	status3 := governor2.GetStatus()
	if approvedCount, ok := status3["approved_last_hour"].(int); !ok || approvedCount != 5 {
		t.Errorf("Expected 5 approved actions after phase 3, got %v", status3["approved_last_hour"])
	}

	if remaining, ok := status3["remaining_blocks"].(int); !ok || remaining != 0 {
		t.Errorf("Expected 0 remaining blocks after phase 3, got %v", status3["remaining_blocks"])
	}

	t.Logf("Phase 3 complete (limit reached): %v", status3)

	// Phase 4: Try one more action - should be blocked
	t.Log("Phase 4: Attempting action beyond limit - should be blocked")

	action := ActionRequest{
		ActionName:       "restart_action_6",
		Target:           "target_6",
		Reasoning:        "Service restart test action - should be blocked",
		RequiresApproval: false,
		Requester:        "restart_test",
		Timestamp:        time.Now(),
	}

	response, err := governor2.RequestAction(action)
	if err == nil {
		t.Error("Expected error for action beyond limit, but got none")
	}

	if response != nil && response.Approved {
		t.Error("Action beyond limit should not be approved")
	}

	t.Log("Phase 4 complete: Action correctly blocked")

	// Phase 5: Third restart - should still see the same state
	t.Log("Phase 5: Third restart - validating persistent state")

	testDB3 := database.NewDatabaseManager(config)
	ctx3, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel3()

	if err := testDB3.Initialize(ctx3); err != nil {
		t.Fatalf("Failed to initialize third database: %v", err)
	}
	defer testDB3.Close()

	governor3 := NewSafetyGovernor(testDB3)
	governor3.Start()
	defer governor3.Stop()

	// Check that the third governor sees the same state
	status4 := governor3.GetStatus()
	if approvedCount, ok := status4["approved_last_hour"].(int); !ok || approvedCount != 5 {
		t.Errorf("Expected 5 approved actions after third restart, got %v", status4["approved_last_hour"])
	}

	if remaining, ok := status4["remaining_blocks"].(int); !ok || remaining != 0 {
		t.Errorf("Expected 0 remaining blocks after third restart, got %v", status4["remaining_blocks"])
	}

	t.Logf("Phase 5 complete (third restart): %v", status4)

	// Test database statistics
	stats, err := testDB3.GetGovernorStats(context.Background())
	if err != nil {
		t.Fatalf("Failed to get governor stats: %v", err)
	}

	t.Logf("Final database stats: %+v", stats)

	// Verify total actions recorded (includes blocked actions)
	if total, ok := stats["total_last_24h"].(int); !ok || total != 6 {
		t.Errorf("Expected 6 total actions in database (5 approved + 1 blocked), got %v", total)
	}

	// Test persistence validation complete
	t.Log("✅ Service restart persistence test passed - state correctly preserved across restarts")
}
