package testing

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"hades-v2/internal/database"
)

// TestDatabaseConnections tests all database connections
func TestDatabaseConnections(t *testing.T) {
	// Test SQLite connection
	t.Run("SQLite", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open SQLite database: %v", err)
		}
		defer func() {
			if err := db.Close(); err != nil {
				t.Logf("Warning: failed to close database: %v", err)
			}
		}()

		err = db.Ping()
		if err != nil {
			t.Fatalf("Failed to ping SQLite database: %v", err)
		}

		// Test basic operations
		_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		_, err = db.Exec("INSERT INTO test (name) VALUES (?)", "test_value")
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}

		var name string
		err = db.QueryRow("SELECT name FROM test WHERE id = ?", 1).Scan(&name)
		if err != nil {
			t.Fatalf("Failed to query data: %v", err)
		}

		if name != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", name)
		}
	})

	// Test PostgreSQL connection (if available)
	t.Run("PostgreSQL", func(t *testing.T) {
		connStr := "host=localhost port=5432 user=postgres dbname=test sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			t.Skipf("PostgreSQL not available: %v", err)
		}
		defer func() {
			if err := db.Close(); err != nil {
				t.Logf("Warning: failed to close database: %v", err)
			}
		}()

		err = db.Ping()
		if err != nil {
			t.Skipf("PostgreSQL connection failed: %v", err)
		}

		// Test basic operations
		_, err = db.Exec("CREATE TEMP TABLE test (id SERIAL PRIMARY KEY, name TEXT)")
		if err != nil {
			t.Fatalf("Failed to create temp table: %v", err)
		}

		_, err = db.Exec("INSERT INTO test (name) VALUES ($1)", "test_value")
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}

		var name string
		err = db.QueryRow("SELECT name FROM test WHERE id = $1", 1).Scan(&name)
		if err != nil {
			t.Fatalf("Failed to query data: %v", err)
		}

		if name != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", name)
		}
	})

	// Test MySQL connection (if available)
	t.Run("MySQL", func(t *testing.T) {
		connStr := "root:@tcp(localhost:3306)/test"
		db, err := sql.Open("mysql", connStr)
		if err != nil {
			t.Skipf("MySQL not available: %v", err)
		}
		defer func() {
			if err := db.Close(); err != nil {
				t.Logf("Warning: failed to close database: %v", err)
			}
		}()

		err = db.Ping()
		if err != nil {
			t.Skipf("MySQL connection failed: %v", err)
		}

		// Test basic operations
		_, err = db.Exec("CREATE TEMPORARY TABLE test (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255))")
		if err != nil {
			t.Fatalf("Failed to create temp table: %v", err)
		}

		_, err = db.Exec("INSERT INTO test (name) VALUES (?)", "test_value")
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}

		var name string
		err = db.QueryRow("SELECT name FROM test WHERE id = ?", 1).Scan(&name)
		if err != nil {
			t.Fatalf("Failed to query data: %v", err)
		}

		if name != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", name)
		}
	})
}

// TestSQLDatabase tests the SQL database implementation
func TestSQLDatabase(t *testing.T) {
	// Test SQLite database
	t.Run("SQLite", func(t *testing.T) {
		config := database.DatabaseConfig{
			Type:     database.SQLite,
			Database: ":memory:",
		}

		sqlDB := &database.SQLDatabase{}
		err := sqlDB.Connect(config)
		if err != nil {
			t.Fatalf("Failed to connect to SQLite: %v", err)
		}
		defer func() {
			if err := sqlDB.Close(); err != nil {
				t.Logf("Warning: failed to close database: %v", err)
			}
		}()

		// Test ping
		err = sqlDB.Ping()
		if err != nil {
			t.Fatalf("Failed to ping SQLite: %v", err)
		}

		// Test type
		if sqlDB.GetType() != database.SQLite {
			t.Errorf("Expected SQLite type, got %s", sqlDB.GetType())
		}

		// Test get connection
		conn := sqlDB.GetConnection()
		if conn == nil {
			t.Fatal("Expected database connection, got nil")
		}

		// Test basic operations
		db := sqlDB.GetDB()
		_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		_, err = db.Exec("INSERT INTO test (name) VALUES (?)", "test_value")
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}

		var name string
		err = db.QueryRow("SELECT name FROM test WHERE id = ?", 1).Scan(&name)
		if err != nil {
			t.Fatalf("Failed to query data: %v", err)
		}

		if name != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", name)
		}
	})
}

// TestDatabasePerformance tests database performance
func TestDatabasePerformance(t *testing.T) {
	// Create SQLite database
	config := database.DatabaseConfig{
		Type:     database.SQLite,
		Database: ":memory:",
	}

	sqlDB := &database.SQLDatabase{}
	err := sqlDB.Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	}()

	db := sqlDB.GetDB()

	// Create test table
	_, err = db.Exec("CREATE TABLE perf_test (id INTEGER PRIMARY KEY, data TEXT, timestamp DATETIME)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Benchmark inserts
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := db.Exec("INSERT INTO perf_test (data, timestamp) VALUES (?, ?)",
			fmt.Sprintf("data_%d", i), time.Now())
		if err != nil {
			t.Fatalf("Failed to insert record %d: %v", i, err)
		}
	}
	insertDuration := time.Since(start)

	// Benchmark selects
	start = time.Now()
	for i := 0; i < 1000; i++ {
		var data string
		err := db.QueryRow("SELECT data FROM perf_test WHERE id = ?", i+1).Scan(&data)
		if err != nil {
			t.Fatalf("Failed to select record %d: %v", i, err)
		}
	}
	selectDuration := time.Since(start)

	// Performance assertions
	if insertDuration > 5*time.Second {
		t.Errorf("Insert performance too slow: %v for 1000 records", insertDuration)
	}

	if selectDuration > 1*time.Second {
		t.Errorf("Select performance too slow: %v for 1000 records", selectDuration)
	}

	t.Logf("Insert performance: %v for 1000 records", insertDuration)
	t.Logf("Select performance: %v for 1000 records", selectDuration)
}

// TestDatabaseTransactions tests database transactions
func TestDatabaseTransactions(t *testing.T) {
	// Create SQLite database
	config := database.DatabaseConfig{
		Type:     database.SQLite,
		Database: ":memory:",
	}

	sqlDB := &database.SQLDatabase{}
	err := sqlDB.Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	}()

	db := sqlDB.GetDB()

	// Create test table
	_, err = db.Exec("CREATE TABLE tx_test (id INTEGER PRIMARY KEY, value INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test successful transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	_, err = tx.Exec("INSERT INTO tx_test (value) VALUES (?)", 100)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Logf("Warning: failed to rollback transaction: %v", err)
		}
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tx_test").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}

	// Test failed transaction
	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	_, err = tx.Exec("INSERT INTO tx_test (value) VALUES (?)", 200)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Logf("Warning: failed to rollback transaction: %v", err)
		}
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	// Intentionally cause an error
	_, err = tx.Exec("INSERT INTO tx_test (nonexistent_column) VALUES (?)", 100)
	if err == nil {
		if err := tx.Rollback(); err != nil {
			t.Logf("Warning: failed to rollback transaction: %v", err)
		}
		t.Fatal("Expected error for invalid insert")
	}

	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// Verify no new data was added
	err = db.QueryRow("SELECT COUNT(*) FROM tx_test").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 record after rollback, got %d", count)
	}
}

// TestDatabaseConcurrency tests concurrent database access
func TestDatabaseConcurrency(t *testing.T) {
	// Clean up any existing database file
	if err := os.Remove("/tmp/test_concurrency.db"); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: failed to remove test database file: %v", err)
	}

	// Create SQLite database
	config := database.DatabaseConfig{
		Type:     database.SQLite,
		Database: "/tmp/test_concurrency.db",
	}

	sqlDB := &database.SQLDatabase{}
	err := sqlDB.Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	}()

	db := sqlDB.GetDB()

	// Create test table
	_, err = db.Exec("CREATE TABLE concurrent_test (id INTEGER PRIMARY KEY, worker_id INTEGER, value INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test concurrent inserts
	const numWorkers = 10
	const insertsPerWorker = 100

	done := make(chan bool, numWorkers)
	errors := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			for j := 0; j < insertsPerWorker; j++ {
				_, err := db.Exec("INSERT INTO concurrent_test (worker_id, value) VALUES (?, ?)",
					workerID, j)
				if err != nil {
					errors <- fmt.Errorf("Worker %d insert %d failed: %v", workerID, j, err)
					return
				}
			}
		}(i)
	}

	// Wait for all workers
	for i := 0; i < numWorkers; i++ {
		select {
		case <-done:
		case err := <-errors:
			t.Errorf("Concurrent operation failed: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	// Verify all records were inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM concurrent_test").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	expectedCount := numWorkers * insertsPerWorker
	if count != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, count)
	}
}

// RunDatabaseTests runs all database tests
func RunDatabaseTests() {
	log.Println("Running comprehensive database integration tests...")

	// Run tests
	tests := []testing.InternalTest{
		{Name: "TestDatabaseConnections", F: TestDatabaseConnections},
		{Name: "TestSQLDatabase", F: TestSQLDatabase},
		{Name: "TestDatabasePerformance", F: TestDatabasePerformance},
		{Name: "TestDatabaseTransactions", F: TestDatabaseTransactions},
		{Name: "TestDatabaseConcurrency", F: TestDatabaseConcurrency},
	}

	for _, test := range tests {
		log.Printf("Running %s...", test.Name)
		t := &testing.T{}
		test.F(t)
		if t.Failed() {
			log.Printf("❌ %s FAILED", test.Name)
		} else {
			log.Printf("✅ %s PASSED", test.Name)
		}
	}

	log.Println("Database integration tests completed!")
}
