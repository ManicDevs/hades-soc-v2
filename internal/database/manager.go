package database

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/scrypt"
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
	encryption  *DBEncryptionService
}

type ManagerConfig struct {
	PrimaryDSN   string
	Database     string
	MaxOpenConns int
	MaxIdleConns int
	ConnLifetime time.Duration
	UseSQLite    bool
	SQLitePath   string
	DBType       DatabaseType
}

func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		Database:     "hades",
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		ConnLifetime: 5 * time.Minute,
		UseSQLite:    false,
		SQLitePath:   "hades.db",
		DBType:       PostgreSQL,
	}
}

func GetManager() *DatabaseManager {
	managerOnce.Do(func() {
		defaultManager = NewDatabaseManager(DefaultManagerConfig())
	})
	return defaultManager
}

// DBEncryptionService provides encryption for database fields
type DBEncryptionService struct {
	masterKey []byte
}

// NewDBEncryptionService creates a new database encryption service
func NewDBEncryptionService(masterPassword string) (*DBEncryptionService, error) {
	// Derive master key from password using scrypt
	salt := []byte("hades-db-encryption-salt")
	masterKey, err := scrypt.Key([]byte(masterPassword), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("hades.database.encryption: failed to derive master key: %w", err)
	}

	return &DBEncryptionService{
		masterKey: masterKey,
	}, nil
}

// Encrypt encrypts data using AES-256-GCM
func (des *DBEncryptionService) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(des.masterKey)
	if err != nil {
		return nil, fmt.Errorf("hades.database.encryption: failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("hades.database.encryption: failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("hades.database.encryption: failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-256-GCM
func (des *DBEncryptionService) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(des.masterKey)
	if err != nil {
		return nil, fmt.Errorf("hades.database.encryption: failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("hades.database.encryption: failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("hades.database.encryption: ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("hades.database.encryption: failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns base64 encoded result
func (des *DBEncryptionService) EncryptString(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	ciphertext, err := des.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts a base64 encoded string
func (des *DBEncryptionService) DecryptString(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	plaintext, err := des.Decrypt(data)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func NewDatabaseManager(config *ManagerConfig) *DatabaseManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	// Initialize encryption service
	encryptionKey := os.Getenv("HADES_DB_ENCRYPTION_KEY")
	if encryptionKey == "" {
		if os.Getenv("HADES_ALLOW_INSECURE_DEV_DB_KEY") == "true" {
			log.Printf("WARNING: using insecure development database encryption key")
			encryptionKey = "hades-insecure-dev-key"
		} else {
			log.Printf("CRITICAL: HADES_DB_ENCRYPTION_KEY not set")
			return nil
		}
	}

	encryption, err := NewDBEncryptionService(encryptionKey)
	if err != nil {
		log.Printf("CRITICAL: Failed to initialize encryption service: %v", err)
		return nil
	}

	return &DatabaseManager{
		config:      config,
		connections: make(map[string]*sql.DB),
		encryption:  encryption,
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

	switch dm.config.DBType {
	case SQLite:
		db, err = dm.connectSQLite()
	case MySQL:
		db, err = dm.connectMySQL()
	case PostgreSQL:
		db, err = dm.connectPostgreSQL()
	default:
		// Fallback to SQLite for unknown types
		log.Printf("Warning: Unknown DB type %v, falling back to SQLite", dm.config.DBType)
		dm.config.DBType = SQLite
		db, err = dm.connectSQLite()
	}

	if err != nil {
		log.Printf("Warning: Primary connection failed, falling back to SQLite: %v", err)
		dm.config.DBType = SQLite
		db, err = dm.connectSQLite()
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
	}

	dm.primary = db
	dm.connections["primary"] = db
	dm.initialized = true

	log.Printf("DatabaseManager: Initialized with primary connection to %s (%s)",
		dm.config.Database, dm.config.DBType)

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

func (dm *DatabaseManager) connectMySQL() (*sql.DB, error) {
	dsn := dm.config.PrimaryDSN
	if dsn == "" {
		dsn = "hades:hades@tcp(localhost:3306)/hades?parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	dm.configurePool(db)
	log.Println("DatabaseManager: Connected to MySQL")

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
	// Get optimal pool settings for the database type
	maxOpenConns, maxIdleConns, connMaxLifetime := dm.GetOptimalPoolConfig()

	// Special case for in-memory SQLite
	if dm.config.DBType == SQLite && dm.config.SQLitePath == ":memory:" {
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(0) // No idle connections for in-memory DB
	} else {
		db.SetMaxOpenConns(maxOpenConns)
		db.SetMaxIdleConns(maxIdleConns)
	}

	db.SetConnMaxLifetime(connMaxLifetime)

	log.Printf("DatabaseManager: Connection pool configured - MaxOpen: %d, MaxIdle: %d, Lifetime: %v",
		maxOpenConns, maxIdleConns, connMaxLifetime)
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

// GetPlaceholder returns the appropriate parameter placeholder for the database type
func (dm *DatabaseManager) GetPlaceholder() string {
	switch dm.config.DBType {
	case PostgreSQL:
		return "$"
	case MySQL, SQLite:
		return "?"
	default:
		return "?" // Default to ? for unknown types
	}
}

// GetPlaceholderForIndex returns the placeholder for a specific parameter index (1-based)
func (dm *DatabaseManager) GetPlaceholderForIndex(index int) string {
	switch dm.config.DBType {
	case PostgreSQL:
		return fmt.Sprintf("$%d", index)
	case MySQL, SQLite:
		return "?"
	default:
		return "?" // Default to ? for unknown types
	}
}

// BuildQuery builds a driver-agnostic query with proper placeholders
// Takes a query template with ? placeholders and returns the appropriate query for the database type
func (dm *DatabaseManager) BuildQuery(queryTemplate string, args ...interface{}) (string, []interface{}) {
	if dm.IsPostgreSQL() {
		// For PostgreSQL, convert ? placeholders to $1, $2, etc.
		return dm.convertToPostgreSQLPlaceholders(queryTemplate, args...)
	}
	// For MySQL and SQLite, keep ? placeholders
	return queryTemplate, args
}

// convertToPostgreSQLPlaceholders converts ? placeholders to $1, $2, etc.
func (dm *DatabaseManager) convertToPostgreSQLPlaceholders(queryTemplate string, args ...interface{}) (string, []interface{}) {
	result := []rune(queryTemplate)
	placeholderCount := 0

	// Count the number of ? placeholders
	for _, char := range result {
		if char == '?' {
			placeholderCount++
		}
	}

	// Replace ? with $1, $2, etc. in order
	currentIndex := 1
	for i := 0; i < len(result) && currentIndex <= placeholderCount; i++ {
		if result[i] == '?' {
			// Replace this ? with $N
			replacement := []rune(fmt.Sprintf("$%d", currentIndex))
			newResult := make([]rune, 0, len(result)+len(replacement)-1)
			newResult = append(newResult, result[:i]...)
			newResult = append(newResult, replacement...)
			newResult = append(newResult, result[i+1:]...)
			result = newResult
			// Adjust index since we changed the string length
			i += len(replacement) - 1
			currentIndex++
		}
	}

	return string(result), args
}

// GetOptimalPoolConfig returns optimal connection pool settings for the database type
func (dm *DatabaseManager) GetOptimalPoolConfig() (maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) {
	switch dm.config.DBType {
	case PostgreSQL:
		// PostgreSQL optimized for high concurrency
		// Allow more connections for parallel processing
		maxOpenConns = 25
		maxIdleConns = 5
		connMaxLifetime = 30 * time.Minute
	case MySQL:
		// MySQL moderate concurrency
		maxOpenConns = 20
		maxIdleConns = 10
		connMaxLifetime = 1 * time.Hour
	case SQLite:
		// SQLite limited concurrency (file-based)
		maxOpenConns = 1
		maxIdleConns = 0
		connMaxLifetime = 1 * time.Hour
	default:
		// Sensible defaults
		maxOpenConns = 10
		maxIdleConns = 5
		connMaxLifetime = 30 * time.Minute
	}

	return maxOpenConns, maxIdleConns, connMaxLifetime
}

// IsPostgreSQL returns true if the database type is PostgreSQL
func (dm *DatabaseManager) IsPostgreSQL() bool {
	return dm.config.DBType == PostgreSQL
}

// IsMySQL returns true if the database type is MySQL
func (dm *DatabaseManager) IsMySQL() bool {
	return dm.config.DBType == MySQL
}

// IsSQLite returns true if the database type is SQLite
func (dm *DatabaseManager) IsSQLite() bool {
	return dm.config.DBType == SQLite
}

func (dm *DatabaseManager) GetConfig() *ManagerConfig {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.config
}

// encryptField encrypts a field using the database encryption service
// Returns the original string if encryption is not available or fails
func (dm *DatabaseManager) encryptField(plaintext string) string {
	if plaintext == "" {
		return plaintext
	}

	if dm.encryption == nil {
		log.Printf("WARNING: Encryption service not available, storing plain-text data")
		return plaintext
	}

	encrypted, err := dm.encryption.EncryptString(plaintext)
	if err != nil {
		log.Printf("WARNING: Failed to encrypt field, storing plain-text: %v", err)
		return plaintext
	}

	return encrypted
}

// decryptField decrypts a field using the database encryption service
// Returns the original string if decryption is not available or fails (graceful fallback for legacy data)
func (dm *DatabaseManager) decryptField(ciphertext string) string {
	if ciphertext == "" {
		return ciphertext
	}

	if dm.encryption == nil {
		log.Printf("WARNING: Encryption service not available, returning potentially encrypted data")
		return ciphertext
	}

	decrypted, err := dm.encryption.DecryptString(ciphertext)
	if err != nil {
		// Graceful fallback - assume it's legacy plaintext data
		log.Printf("INFO: Field appears to be legacy plaintext (decryption failed): %v", err)
		return ciphertext
	}

	return decrypted
}
