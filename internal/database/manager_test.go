package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hades-v2/internal/database"
)

func TestDatabaseManager_NewManager(t *testing.T) {
	config := database.ManagerConfig{
		PrimaryDSN: "postgres://test_user:test_password@localhost:5432/test_db?sslmode=disable",
		DBType:     database.PostgreSQL,
	}
	manager := database.NewDatabaseManager(&config)
	assert.NotNil(t, manager)
}

func TestDatabaseManager_NewManager_SQLite(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	config := database.ManagerConfig{
		DBType:     database.SQLite,
		UseSQLite:  true,
		SQLitePath: tmpFile,
	}
	manager := database.NewDatabaseManager(&config)
	assert.NotNil(t, manager)
}

func TestDatabaseManager_Initialize(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	config := database.ManagerConfig{
		DBType:     database.SQLite,
		UseSQLite:  true,
		SQLitePath: tmpFile,
	}

	manager := database.NewDatabaseManager(&config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, manager.IsInitialized())

	err = manager.Close()
	assert.NoError(t, err)
}

func TestDatabaseManager_ConnectionPooling(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	config := database.ManagerConfig{
		DBType:       database.SQLite,
		UseSQLite:    true,
		SQLitePath:   tmpFile,
		MaxOpenConns: 25,
		MaxIdleConns: 10,
		ConnLifetime: 5 * time.Minute,
	}

	manager := database.NewDatabaseManager(&config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)

	conn := manager.GetPrimary()
	assert.NotNil(t, conn)

	err = manager.Close()
	assert.NoError(t, err)
}

func TestDatabaseManager_QueryAndExec(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	config := database.ManagerConfig{
		DBType:     database.SQLite,
		UseSQLite:  true,
		SQLitePath: tmpFile,
	}

	manager := database.NewDatabaseManager(&config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)
	defer manager.Close()

	_, err = manager.Exec(ctx, `
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	result, err := manager.Exec(ctx, `
		INSERT INTO test_table (name) VALUES (?)
	`, "test_name")
	assert.NoError(t, err)
	rowsAff, _ := result.RowsAffected()
	assert.Equal(t, int64(1), rowsAff)

	rows, err := manager.Query(ctx, `
		SELECT id, name FROM test_table WHERE name = ?
	`, "test_name")
	assert.NoError(t, err)
	defer rows.Close()

	var id int
	var name string
	assert.True(t, rows.Next())
	err = rows.Scan(&id, &name)
	assert.NoError(t, err)
	assert.Equal(t, "test_name", name)
}

func TestDatabaseManager_Transaction(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	config := database.ManagerConfig{
		DBType:     database.SQLite,
		UseSQLite:  true,
		SQLitePath: tmpFile,
	}

	manager := database.NewDatabaseManager(&config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)
	defer manager.Close()

	_, err = manager.Exec(ctx, `
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	tx, err := manager.Begin(ctx)
	require.NoError(t, err)

	_, err = tx.Exec(`INSERT INTO test_table (name) VALUES (?)`, "test1")
	require.NoError(t, err)

	_, err = tx.Exec(`INSERT INTO test_table (name) VALUES (?)`, "test2")
	require.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	var count int
	err = manager.QueryRow(ctx, `SELECT COUNT(*) FROM test_table`).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestDatabaseManager_Encryption(t *testing.T) {
	os.Setenv("HADES_ALLOW_INSECURE_DEV_DB_KEY", "true")
	defer os.Unsetenv("HADES_ALLOW_INSECURE_DEV_DB_KEY")

	tmpFile := t.TempDir() + "/test.db"
	config := database.ManagerConfig{
		DBType:     database.SQLite,
		UseSQLite:  true,
		SQLitePath: tmpFile,
	}

	manager := database.NewDatabaseManager(&config)
	assert.NotNil(t, manager)
}

func TestDatabaseManager_Ping(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	config := database.ManagerConfig{
		DBType:     database.SQLite,
		UseSQLite:  true,
		SQLitePath: tmpFile,
	}

	manager := database.NewDatabaseManager(&config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)
	defer manager.Close()

	err = manager.Ping(ctx)
	assert.NoError(t, err)
}

func TestDatabaseManager_Stats(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	config := database.ManagerConfig{
		DBType:     database.SQLite,
		UseSQLite:  true,
		SQLitePath: tmpFile,
	}

	manager := database.NewDatabaseManager(&config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)
	defer manager.Close()

	stats := manager.GetStats()
	assert.NotNil(t, stats)
}

func TestDatabaseManager_QueryRow(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	config := database.ManagerConfig{
		DBType:     database.SQLite,
		UseSQLite:  true,
		SQLitePath: tmpFile,
	}

	manager := database.NewDatabaseManager(&config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)
	defer manager.Close()

	_, err = manager.Exec(ctx, `CREATE TABLE test_query_row (value TEXT)`)
	require.NoError(t, err)

	_, err = manager.Exec(ctx, `INSERT INTO test_query_row VALUES (?)`, "test_value")
	require.NoError(t, err)

	var value string
	err = manager.QueryRow(ctx, `SELECT value FROM test_query_row`).Scan(&value)
	assert.NoError(t, err)
	assert.Equal(t, "test_value", value)
}
