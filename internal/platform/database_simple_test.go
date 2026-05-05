package platform

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Simple database implementation for testing without DatabaseManager
type SimpleDatabase struct {
	db *sql.DB
}

func NewSimpleDatabase() (*SimpleDatabase, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &SimpleDatabase{db: db}, nil
}

func (sd *SimpleDatabase) Close() error {
	return sd.db.Close()
}

func createTables(db *sql.DB) error {
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
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

// Test simple database operations
func TestSimpleDatabase(t *testing.T) {
	ctx := context.Background()

	t.Run("ConnectAndBasicOperations", func(t *testing.T) {
		db, err := NewSimpleDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		// Test user CRUD
		user := &User{
			ID:           "test-user-1",
			Username:     "testuser",
			Email:        "test@example.com",
			Role:         RoleAdmin,
			PasswordHash: "hashed-password",
			IsActive:     true,
			CreatedAt:    time.Now(),
		}

		// Insert user
		query := `INSERT INTO users (id, username, email, role, password_hash, is_active, created_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = db.db.ExecContext(ctx, query,
			user.ID, user.Username, user.Email, user.Role,
			user.PasswordHash, user.IsActive, user.CreatedAt)
		if err != nil {
			t.Fatalf("Failed to insert user: %v", err)
		}

		// Retrieve user
		query = `SELECT id, username, email, role, password_hash, is_active, last_login, created_at 
			FROM users WHERE username = ?`
		row := db.db.QueryRowContext(ctx, query, "testuser")

		var retrievedUser User
		var lastLogin sql.NullTime
		err = row.Scan(&retrievedUser.ID, &retrievedUser.Username, &retrievedUser.Email, &retrievedUser.Role,
			&retrievedUser.PasswordHash, &retrievedUser.IsActive, &lastLogin, &retrievedUser.CreatedAt)
		// Note: LastLogin is not tested in this basic CRUD test
		// if lastLogin.Valid {
		//     retrievedUser.LastLogin = lastLogin.Time
		// }
		if err != nil {
			t.Fatalf("Failed to retrieve user: %v", err)
		}

		if retrievedUser.Username != user.Username {
			t.Errorf("Expected username %s, got %s", user.Username, retrievedUser.Username)
		}
		if retrievedUser.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, retrievedUser.Email)
		}

		// Test session CRUD
		session := &Session{
			ID:        "test-session-1",
			UserID:    "test-user-1",
			Token:     "test-token-123",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			IPAddress: "127.0.0.1",
			UserAgent: "test-agent",
			CreatedAt: time.Now(),
		}

		// Insert session
		query = `INSERT INTO sessions (id, user_id, token, expires_at, ip_address, user_agent, created_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = db.db.ExecContext(ctx, query,
			session.ID, session.UserID, session.Token, session.ExpiresAt,
			session.IPAddress, session.UserAgent, session.CreatedAt)
		if err != nil {
			t.Fatalf("Failed to insert session: %v", err)
		}

		// Retrieve session
		query = `SELECT id, user_id, token, expires_at, ip_address, user_agent, created_at 
			FROM sessions WHERE token = ?`
		row = db.db.QueryRowContext(ctx, query, "test-token-123")

		var retrievedSession Session
		err = row.Scan(&retrievedSession.ID, &retrievedSession.UserID, &retrievedSession.Token, &retrievedSession.ExpiresAt,
			&retrievedSession.IPAddress, &retrievedSession.UserAgent, &retrievedSession.CreatedAt)
		if err != nil {
			t.Fatalf("Failed to retrieve session: %v", err)
		}

		if retrievedSession.Token != session.Token {
			t.Errorf("Expected token %s, got %s", session.Token, retrievedSession.Token)
		}
		if retrievedSession.UserID != session.UserID {
			t.Errorf("Expected user ID %s, got %s", session.UserID, retrievedSession.UserID)
		}
	})

	t.Run("ConnectionPooling", func(t *testing.T) {
		db, err := NewSimpleDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		t.Cleanup(func() {
			if err := db.Close(); err != nil {
				t.Logf("Warning: failed to close simple database: %v", err)
			}
		})

		// Verify users table exists before concurrent operations
		var result int
		err = db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&result)
		if err != nil {
			t.Fatalf("Failed to query users table: %v", err)
		}

		// Test concurrent operations
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer func() { done <- true }()

				// Insert and retrieve a user
				user := &User{
					ID:           fmt.Sprintf("user-%d", goroutineID),
					Username:     fmt.Sprintf("user%d", goroutineID),
					Email:        fmt.Sprintf("user%d@example.com", goroutineID),
					Role:         RoleViewer,
					PasswordHash: "hash",
					IsActive:     true,
					CreatedAt:    time.Now(),
				}

				query := `INSERT INTO users (id, username, email, role, password_hash, is_active, created_at) 
					VALUES (?, ?, ?, ?, ?, ?, ?)`
				_, err := db.db.ExecContext(ctx, query,
					user.ID, user.Username, user.Email, user.Role,
					user.PasswordHash, user.IsActive, user.CreatedAt)
				if err != nil {
					t.Errorf("Failed to insert user %d: %v", goroutineID, err)
					return
				}

				// Retrieve user
				query = `SELECT id, username, email, role, password_hash, is_active, last_login, created_at 
					FROM users WHERE username = ?`
				row := db.db.QueryRowContext(ctx, query, user.Username)

				var retrievedUser User
				var lastLogin sql.NullTime
				err = row.Scan(&retrievedUser.ID, &retrievedUser.Username, &retrievedUser.Email, &retrievedUser.Role,
					&retrievedUser.PasswordHash, &retrievedUser.IsActive, &lastLogin, &retrievedUser.CreatedAt)
				// Note: LastLogin is not tested in this connection pooling test
				// if lastLogin.Valid {
				//     retrievedUser.LastLogin = lastLogin.Time
				// }
				if err != nil {
					t.Errorf("Failed to retrieve user %d: %v", goroutineID, err)
					return
				}

				if retrievedUser.Username != user.Username {
					t.Errorf("User %d: expected username %s, got %s", goroutineID, user.Username, retrievedUser.Username)
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	t.Run("ClosedConnectionHandling", func(t *testing.T) {
		db, err := NewSimpleDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}

		// Close the database
		err = db.Close()
		if err != nil {
			t.Fatalf("Failed to close database: %v", err)
		}

		// Try to use closed database
		_, err = db.db.ExecContext(ctx, "SELECT 1")
		if err == nil {
			t.Error("Expected error when using closed database")
		}
	})

	t.Run("TimeoutHandling", func(t *testing.T) {
		db, err := NewSimpleDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		// Test with timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		// This should work within timeout
		_, err = db.db.ExecContext(timeoutCtx, "SELECT 1")
		if err != nil {
			if timeoutCtx.Err() == context.DeadlineExceeded {
				t.Error("Simple query should not timeout")
			} else {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	})
}
