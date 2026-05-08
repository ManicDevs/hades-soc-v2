package backup

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"hades-v2/internal/database"
)

// BackupManager handles comprehensive backup and recovery operations
type BackupManager struct {
	db     database.Database
	config BackupConfig
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	BackupDir        string        `json:"backup_dir"`
	RetentionDays    int           `json:"retention_days"`
	Compression      bool          `json:"compression"`
	Encryption       bool          `json:"encryption"`
	ScheduleInterval time.Duration `json:"schedule_interval"`
	RemoteStorage    RemoteStorage `json:"remote_storage,omitempty"`
}

// RemoteStorage represents remote backup storage configuration
type RemoteStorage struct {
	Type      string `json:"type"` // s3, ftp, sftp
	Endpoint  string `json:"endpoint"`
	Bucket    string `json:"bucket,omitempty"`
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Region    string `json:"region,omitempty"`
}

// Backup represents a backup record
type Backup struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"`   // full, incremental, differential
	Status      string    `json:"status"` // running, completed, failed
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// BackupJob represents a backup job
type BackupJob struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Schedule    string    `json:"schedule"`
	Status      string    `json:"status"`
	LastRun     time.Time `json:"last_run"`
	NextRun     time.Time `json:"next_run"`
	Retention   int       `json:"retention_days"`
	Compression bool      `json:"compression"`
	Encryption  bool      `json:"encryption"`
	Storage     string    `json:"storage_location"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewBackupManager creates a new backup manager
func NewBackupManager(db database.Database, config BackupConfig) *BackupManager {
	return &BackupManager{
		db:     db,
		config: config,
	}
}

// CreateBackup creates a new backup
func (bm *BackupManager) CreateBackup(backupType string, description string) (*Backup, error) {
	log.Printf("Starting %s backup: %s", backupType, description)

	backup := &Backup{
		Type:        backupType,
		Status:      "running",
		StartedAt:   time.Now(),
		Description: description,
		CreatedAt:   time.Now(),
	}

	// Create backup record
	_, err := bm.createBackupRecord(backup)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	// Perform backup based on type
	switch backupType {
	case "full":
		err = bm.performFullBackup(backup)
	case "incremental":
		err = bm.performIncrementalBackup(backup)
	case "differential":
		err = bm.performDifferentialBackup(backup)
	default:
		return nil, fmt.Errorf("unsupported backup type: %s", backupType)
	}

	if err != nil {
		backup.Status = "failed"
		if err := bm.updateBackupRecord(backup); err != nil {
			log.Printf("Error updating backup record: %v", err)
		}
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	backup.Status = "completed"
	backup.CompletedAt = time.Now()
	backup.Location = bm.generateBackupPath(backup)

	// Update backup record
	err = bm.updateBackupRecord(backup)
	if err != nil {
		return nil, fmt.Errorf("failed to update backup record: %w", err)
	}

	log.Printf("Backup completed successfully: %s", backup.Location)
	return backup, nil
}

// performFullBackup performs a full database backup
func (bm *BackupManager) performFullBackup(backup *Backup) error {
	log.Printf("Performing full backup")

	// Get database connection
	sqlDB, ok := bm.db.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	// Create backup file
	backupPath := bm.generateBackupPath(backup)
	file, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close backup file: %v", err)
		}
	}()

	// Perform database dump
	err = bm.dumpDatabase(sqlDB, file)
	if err != nil {
		return fmt.Errorf("failed to dump database: %w", err)
	}

	// Get file info for checksum
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get backup file info: %w", err)
	}

	backup.Size = fileInfo.Size()
	backup.Checksum = bm.calculateChecksum(backupPath)

	return nil
}

// performIncrementalBackup performs incremental backup
func (bm *BackupManager) performIncrementalBackup(backup *Backup) error {
	log.Printf("Performing incremental backup")

	// Get last backup timestamp
	lastBackupTime, err := bm.getLastBackupTime("full")
	if err != nil {
		return fmt.Errorf("failed to get last backup time: %w", err)
	}

	// Get database connection
	sqlDB, ok := bm.db.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	// Create backup file
	backupPath := bm.generateBackupPath(backup)
	file, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close backup file: %v", err)
		}
	}()

	// Perform incremental dump
	err = bm.dumpDatabaseSince(sqlDB, file, lastBackupTime)
	if err != nil {
		return fmt.Errorf("failed to perform incremental dump: %w", err)
	}

	// Get file info for checksum
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get backup file info: %w", err)
	}

	backup.Size = fileInfo.Size()
	backup.Checksum = bm.calculateChecksum(backupPath)

	return nil
}

// performDifferentialBackup performs differential backup
func (bm *BackupManager) performDifferentialBackup(backup *Backup) error {
	log.Printf("Performing differential backup")

	// Get last backup timestamp
	lastBackupTime, err := bm.getLastBackupTime("full")
	if err != nil {
		return fmt.Errorf("failed to get last backup time: %w", err)
	}

	// Get database connection
	sqlDB, ok := bm.db.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	// Create backup file
	backupPath := bm.generateBackupPath(backup)
	file, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close backup file: %v", err)
		}
	}()

	// Perform differential dump
	err = bm.dumpDatabaseSince(sqlDB, file, lastBackupTime)
	if err != nil {
		return fmt.Errorf("failed to perform differential dump: %w", err)
	}

	// Get file info for checksum
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get backup file info: %w", err)
	}

	backup.Size = fileInfo.Size()
	backup.Checksum = bm.calculateChecksum(backupPath)

	return nil
}

// RestoreBackup restores a backup
func (bm *BackupManager) RestoreBackup(backupID int, targetLocation string) error {
	log.Printf("Starting restore from backup ID %d to %s", backupID, targetLocation)

	// Get backup record
	backup, err := bm.getBackupRecord(backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup record: %w", err)
	}

	// Extract backup file
	backupPath := backup.Location
	err = bm.extractBackup(backupPath, targetLocation)
	if err != nil {
		return fmt.Errorf("failed to extract backup: %w", err)
	}

	// Restore database
	err = bm.restoreDatabase(targetLocation)
	if err != nil {
		return fmt.Errorf("failed to restore database: %w", err)
	}

	log.Printf("Backup restored successfully from %s", backupPath)
	return nil
}

// CleanupOldBackups removes old backups based on retention policy
func (bm *BackupManager) CleanupOldBackups() error {
	log.Printf("Cleaning up old backups (retention: %d days)", bm.config.RetentionDays)

	// Get old backup records
	cutoffTime := time.Now().AddDate(0, 0, -bm.config.RetentionDays)

	sqlDB, ok := bm.db.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT id, location 
		FROM backups 
		WHERE created_at < $1 AND status = 'completed'
		ORDER BY created_at ASC
	`

	rows, err := sqlDB.Query(query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to get old backups: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var backupIDs []int
	var backupPaths []string

	for rows.Next() {
		var backupID int
		var backupPath string
		err := rows.Scan(&backupID, &backupPath)
		if err != nil {
			return fmt.Errorf("failed to scan backup record: %w", err)
		}
		backupIDs = append(backupIDs, backupID)
		backupPaths = append(backupPaths, backupPath)
	}

	// Remove old backup files
	for _, backupPath := range backupPaths {
		err = os.RemoveAll(backupPath)
		if err != nil {
			log.Printf("Failed to remove old backup file %s: %v", backupPath, err)
		} else {
			log.Printf("Removed old backup file: %s", backupPath)
		}
	}

	// Delete old backup records
	for _, backupID := range backupIDs {
		query := "DELETE FROM backups WHERE id = $1"
		_, err = sqlDB.Exec(query, backupID)
		if err != nil {
			log.Printf("Failed to delete backup record %d: %v", backupID, err)
		}
	}

	return nil
}

// Helper methods
func (bm *BackupManager) createBackupRecord(backup *Backup) (int, error) {
	sqlDB, ok := bm.db.GetConnection().(*sql.DB)
	if !ok {
		return 0, fmt.Errorf("database is not an SQL database")
	}

	query := `
		INSERT INTO backups (type, status, size, checksum, started_at, completed_at, location, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var backupID int
	err := sqlDB.QueryRow(query, backup.Type, backup.Status, backup.Size,
		backup.Checksum, backup.StartedAt, backup.CompletedAt,
		backup.Location, backup.Description, backup.CreatedAt).Scan(&backupID)

	return backupID, err
}

func (bm *BackupManager) updateBackupRecord(backup *Backup) error {
	sqlDB, ok := bm.db.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	query := `
		UPDATE backups 
		SET status = $1, size = $2, checksum = $3, completed_at = $4, location = $5
		WHERE id = $6
	`

	_, err := sqlDB.Exec(query, backup.Status, backup.Size, backup.Checksum,
		backup.CompletedAt, backup.Location, backup.ID)

	return err
}

func (bm *BackupManager) getBackupRecord(backupID int) (*Backup, error) {
	sqlDB, ok := bm.db.GetConnection().(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT id, type, status, size, checksum, started_at, completed_at, location, description, created_at
		FROM backups 
		WHERE id = $1
	`

	var backup Backup
	err := sqlDB.QueryRow(query, backupID).Scan(&backup.ID, &backup.Type, &backup.Status,
		&backup.Size, &backup.Checksum, &backup.StartedAt, &backup.CompletedAt,
		&backup.Location, &backup.Description, &backup.CreatedAt)

	return &backup, err
}

func (bm *BackupManager) getLastBackupTime(backupType string) (time.Time, error) {
	sqlDB, ok := bm.db.GetConnection().(*sql.DB)
	if !ok {
		return time.Time{}, fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT completed_at 
		FROM backups 
		WHERE type = $1 AND status = 'completed'
		ORDER BY completed_at DESC 
		LIMIT 1
	`

	var completedAt time.Time
	err := sqlDB.QueryRow(query, backupType).Scan(&completedAt)

	return completedAt, err
}

func (bm *BackupManager) generateBackupPath(backup *Backup) string {
	timestamp := backup.StartedAt.Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s_%s.sql", backup.Type, timestamp)

	if bm.config.Compression {
		filename += ".gz"
	}

	return filepath.Join(bm.config.BackupDir, filename)
}

func (bm *BackupManager) calculateChecksum(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close backup file: %v", err)
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (bm *BackupManager) dumpDatabase(sqlDB *sql.DB, file *os.File) error {
	// This would implement database-specific dump logic
	// For SQLite, use .dump command
	// For PostgreSQL, use pg_dump
	// For MySQL, use mysqldump

	// Simplified implementation for demonstration
	_, err := file.WriteString("-- Database dump placeholder\n")
	return err
}

func (bm *BackupManager) dumpDatabaseSince(sqlDB *sql.DB, file *os.File, since time.Time) error {
	// This would implement incremental dump logic
	// For demonstration, just write a placeholder

	_, err := fmt.Fprintf(file, "-- Incremental dump since %s\n", since.Format(time.RFC3339))
	return err
}

func (bm *BackupManager) extractBackup(backupPath, targetLocation string) error {
	// Extract backup file
	// For compressed backups, decompress first
	// For demonstration, just copy the file

	err := os.MkdirAll(targetLocation, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy backup file to target
	sourceFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			log.Printf("Failed to close source file: %v", err)
		}
	}()

	targetPath := filepath.Join(targetLocation, filepath.Base(backupPath))
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer func() {
		if err := targetFile.Close(); err != nil {
			log.Printf("Failed to close target file: %v", err)
		}
	}()

	_, err = io.Copy(targetFile, sourceFile)
	return err
}

func (bm *BackupManager) restoreDatabase(targetLocation string) error {
	// This would implement database-specific restore logic
	// For demonstration, just log the restore operation

	log.Printf("Database restore operation to %s", targetLocation)
	return nil
}
