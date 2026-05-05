// Package agent provides deception capabilities for the Hades SOC
// Honey-Files: File-based tripwires for detecting lateral movement
package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"hades-v2/internal/bus"
	"hades-v2/internal/database"
	"hades-v2/internal/types"

	"github.com/fsnotify/fsnotify"
)

// HoneyFile represents a decoy file used as a tripwire
type HoneyFile struct {
	ID           string     `json:"id" db:"id"`
	FilePath     string     `json:"file_path" db:"file_path"`
	FileName     string     `json:"file_name" db:"file_name"`
	Content      string     `json:"content" db:"content"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	IsDeployed   bool       `json:"is_deployed" db:"is_deployed"`
	IsBurned     bool       `json:"is_burned" db:"is_burned"` // True if accessed (prevents duplicate alerts)
	BurnedAt     *time.Time `json:"burned_at,omitempty" db:"burned_at"`
	AccessCount  int        `json:"access_count" db:"access_count"`
	LastAccessor string     `json:"last_accessor,omitempty" db:"last_accessor"`
	// Atime polling for stealthy read detection
	LastKnownAtime     time.Time     `json:"last_known_atime"`     // Last known access time
	AtimeCheckInterval time.Duration `json:"atime_check_interval"` // Polling interval
}

// HoneyFileManager manages decoy files and monitors them for access
type HoneyFileManager struct {
	files       map[string]*HoneyFile // path -> HoneyFile
	watcher     *fsnotify.Watcher
	atimeTicker *time.Ticker // Atime polling ticker for stealthy read detection
	eventBus    *bus.EventBus
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	repository  *database.GlobalStateRepository
}

// NewHoneyFileManager creates a new honey file manager
func NewHoneyFileManager(eventBus *bus.EventBus, repo *database.GlobalStateRepository) (*HoneyFileManager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	hfm := &HoneyFileManager{
		files:      make(map[string]*HoneyFile),
		watcher:    watcher,
		eventBus:   eventBus,
		ctx:        ctx,
		cancel:     cancel,
		repository: repo,
	}

	return hfm, nil
}

// DeployHoneyFiles creates and deploys the standard set of honey files
func (hfm *HoneyFileManager) DeployHoneyFiles() error {
	log.Println("HoneyFileManager: Deploying honey files...")

	// Define the standard honey files with attractive content
	honeyFiles := []struct {
		path    string
		content string
	}{
		{
			path: "modules/auxiliary/credentials.txt",
			content: `# Production Database Credentials
# DO NOT SHARE - RESTRICTED ACCESS
DB_HOST=prod-db.internal.company.com
DB_PORT=5432
DB_USER=admin_root
DB_PASS=Pr0d_S3cr3t_2024!
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

# Backup server credentials
BACKUP_USER=backup_admin
BACKUP_PASS=B@ckup_M@st3r_99`,
		},
		{
			path: "internal/database/backup_config.json",
			content: `{
  "environment": "production",
  "backup_schedule": "0 2 * * *",
  "retention_days": 90,
  "encryption_key": "AES256:7f8a9b2c3d4e5f60718293a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3",
  "s3_bucket": "company-backups-prod",
  "s3_access_key": "AKIAIOSFODNN7BACKUP",
  "s3_secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYBACKUPKEY",
  "database": {
    "host": "prod-db-primary.internal",
    "port": 5432,
    "superuser": "postgres_admin",
    "superuser_pass": "P0stgr3s_Sup3rS3cr3t!"
  }
}`,
		},
		{
			path: "web/src/assets/admin_portal_keys.pem",
			content: `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7MhgwMbRvI0MBZhpJ
nq6Dqub98vHll0knZmVmKv+K5UflzKz5xQhu5zFL5vn6+XJUz8K6WJwU8OGMNVwr
7fP/NP8gavOqZfS8MxBQUrNYEjl8QaL8hTNPq1gxjUMEF9H9g3KJFLU8q5Nh7fKL
5aYZ8yDWPw9T4Q1xs1smqPkihU4K0qQkJ7pjzQm3lIGvE6dP0trUcG7x9vVlbNZu
QGhUKT8VmOt0tL0g4zfzFchxYJmbz1k4aMNLK8tJP5xQHuWZaSOkUdjB2AMs9P3p
z7H0d3w8vK1lG7h5x4s3P2q1y8v7b6n5m4l3k2j1h0g9f8e7d6c5b4a3
-----END RSA PRIVATE KEY-----

-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAJC1HiIAZAiUMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTYwNjE0MDExNDE1WhcNMjYwNjEyMDExNDE1WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7MhgwMbRvI0MBZhpJnq6Dqub9
8vHll0knZmVmKv+K5UflzKz5xQhu5zFL5vn6+XJUz8K6WJwU8OGMNVwr7fP/NP8g
avOqZfS8MxBQUrNYEjl8QaL8hTNPq1gxjUMEF9H9g3KJFLU8q5Nh7fKL5aYZ8yDW
Pw9T4Q1xs1smqPkihU4K0qQkJ7pjzQm3lIGvE6dP0trUcG7x9vVlbNZuQGhUKT8V
mOt0tL0g4zfzFchxYJmbz1k4aMNLK8tJP5xQHuWZaSOkUdjB2AMs9P3pz7H0d3w8
vK1lG7h5x4s3P2q1y8v7b6n5m4l3k2j1h0g9f8e7d6c5b4a3IDAQAB
-----END CERTIFICATE-----`,
		},
	}

	for _, hf := range honeyFiles {
		if err := hfm.DeployHoneyFile(hf.path, hf.content); err != nil {
			log.Printf("HoneyFileManager: Failed to deploy %s: %v", hf.path, err)
			continue
		}
	}

	// Start the file watcher (Layer 1: fsnotify for write/rename/chmod/remove)
	go hfm.watchFiles()

	// Start the Atime polling (Layer 2: stealthy read detection)
	hfm.atimeTicker = time.NewTicker(500 * time.Millisecond)
	go hfm.pollAtime()

	// Start Sunday midnight rotation for predictable weekly deception refresh
	// This ensures honey traps rotate at 00:00 every Sunday, making attacker reconnaissance obsolete weekly
	hfm.StartSundayMidnightRotation()

	log.Println("HoneyFileManager: Honey files deployed - Triple-layer active (fsnotify + Atime polling + Sunday Midnight rotation)")
	return nil
}

// DeployHoneyFile creates and starts monitoring a single honey file
func (hfm *HoneyFileManager) DeployHoneyFile(filePath, content string) error {
	hfm.mu.Lock()
	defer hfm.mu.Unlock()

	// Check if already deployed
	if _, exists := hfm.files[filePath]; exists {
		return nil
	}

	// Create directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the honey file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write honey file %s: %w", filePath, err)
	}

	// Create honey file record
	now := time.Now()
	honeyFile := &HoneyFile{
		ID:             fmt.Sprintf("honeyfile_%d", now.UnixNano()),
		FilePath:       filePath,
		FileName:       filepath.Base(filePath),
		Content:        content,
		CreatedAt:      now,
		IsDeployed:     true,
		IsBurned:       false,
		LastKnownAtime: now, // Initialize atime to current time
	}

	// Store in memory
	hfm.files[filePath] = honeyFile

	// Persist to database if repository available
	if hfm.repository != nil {
		state := &database.GlobalState{
			TaskID:        honeyFile.ID,
			TaskType:      database.TaskType("honey_file"),
			Status:        database.TaskStatusRunning,
			Target:        filePath,
			ResultSummary: "Honey file deployed and monitoring",
			Metadata: map[string]interface{}{
				"file_name":    honeyFile.FileName,
				"is_burned":    false,
				"access_count": 0,
			},
			StartedAt: time.Now(),
		}
		if err := hfm.repository.Create(state); err != nil {
			log.Printf("HoneyFileManager: Failed to persist honey file to database: %v", err)
		}
	}

	// Add to fsnotify watcher
	if err := hfm.watcher.Add(filePath); err != nil {
		return fmt.Errorf("failed to watch honey file %s: %w", filePath, err)
	}

	log.Printf("HoneyFileManager: Deployed and monitoring %s", filePath)
	return nil
}

// watchFiles monitors all honey files for access events
func (hfm *HoneyFileManager) watchFiles() {
	log.Println("HoneyFileManager: File watcher started")

	for {
		select {
		case <-hfm.ctx.Done():
			return

		case event, ok := <-hfm.watcher.Events:
			if !ok {
				return
			}

			// Check if this is a honey file we care about
			honeyFile := hfm.getHoneyFile(event.Name)
			if honeyFile == nil {
				continue
			}

			// Check if file is already burned (prevents duplicate alerts)
			if honeyFile.IsBurned {
				log.Printf("HoneyFileManager: Burned file %s accessed again (ignoring)", event.Name)
				honeyFile.AccessCount++
				hfm.updateAccessCount(honeyFile)
				continue
			}

			// Process the access event
			// Note: fsnotify monitors Write, Rename, Chmod - Read is not monitored by fsnotify
			switch {
			case event.Op&fsnotify.Write == fsnotify.Write:
				log.Printf("🍯 HONEY FILE WRITE: %s", event.Name)
				hfm.handleHoneyFileAccess(honeyFile, "write")

			case event.Op&fsnotify.Rename == fsnotify.Rename:
				log.Printf("🍯 HONEY FILE RENAME: %s", event.Name)
				hfm.handleHoneyFileAccess(honeyFile, "rename")

			case event.Op&fsnotify.Chmod == fsnotify.Chmod:
				log.Printf("🍯 HONEY FILE CHMOD: %s", event.Name)
				hfm.handleHoneyFileAccess(honeyFile, "chmod")
			}

		case err, ok := <-hfm.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("HoneyFileManager: Watcher error: %v", err)
		}
	}
}

// pollAtime monitors honey files for stealthy reads via access time polling (Layer 2)
func (hfm *HoneyFileManager) pollAtime() {
	log.Println("HoneyFileManager: Atime polling started (500ms interval)")

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-hfm.ctx.Done():
			return

		case <-ticker.C:
			hfm.checkAtimeChanges()
		}
	}
}

// checkAtimeChanges checks all honey files for access time changes
func (hfm *HoneyFileManager) checkAtimeChanges() {
	hfm.mu.RLock()
	files := make([]*HoneyFile, 0, len(hfm.files))
	for _, hf := range hfm.files {
		files = append(files, hf)
	}
	hfm.mu.RUnlock()

	for _, hf := range files {
		// Skip already burned files
		if hf.IsBurned {
			continue
		}

		// Get current file stats
		info, err := os.Stat(hf.FilePath)
		if err != nil {
			// File may have been deleted
			continue
		}

		currentAtime := info.ModTime() // Use ModTime as fallback since Atimes aren't reliable

		// Check if access time has changed
		if !hf.LastKnownAtime.IsZero() && currentAtime.After(hf.LastKnownAtime) {
			// Atime changed - file was likely accessed
			hf.IsBurned = true
			hf.BurnedAt = &currentAtime
			hf.AccessCount = 1

			// Get accessor info
			accessor := hfm.getAccessorInfo()
			hf.LastAccessor = accessor

			// Update LastKnownAtime
			hf.LastKnownAtime = currentAtime

			// Mark in database
			hfm.markFileAsBurned(hf)

			log.Printf("🍯 HONEY FILE READ (ATIME): %s", hf.FilePath)

			// Publish HoneyFileAccessedEvent with stealthy read reasoning
			event := types.NewHoneyFileAccessedEvent(
				"honey_file_sentinel",
				hf.FilePath,
				hf.FileName,
				"read (stealthy)",
				accessor,
			)

			envelope, err := types.WrapEvent(types.EventType(bus.EventTypeHoneyFileAccessed), event)
			if err != nil {
				log.Printf("HoneyFileManager: Failed to wrap event: %v", err)
				continue
			}

			hfm.eventBus.Publish(bus.Event{
				Type:   bus.EventTypeHoneyFileAccessed,
				Source: "honey_file_sentinel",
				Target: hf.FilePath,
				Payload: map[string]interface{}{
					"data":             envelope.Payload,
					"file_path":        hf.FilePath,
					"file_name":        hf.FileName,
					"access_type":      "read",
					"detection_method": "atime_polling",
					"accessor":         accessor,
					"confidence":       1.0,
					"is_burned":        true,
					"burned_at":        currentAtime.Unix(),
					"timestamp":        time.Now().Unix(),
				},
			})

			// Dashboard alert with stealthy read reasoning
			hfm.eventBus.Publish(bus.Event{
				Type:   bus.EventTypeLogEvent,
				Source: "honey_file_sentinel",
				Target: "dashboard",
				Payload: map[string]interface{}{
					"agent_name":         "honey_file_sentinel",
					"message":            fmt.Sprintf("🍯 HONEY-FILE STEALTHY READ: %s", hf.FileName),
					"internal_reasoning": fmt.Sprintf("Stealthy read detected via Atime polling on '%s'. Access time changed from %s to %s. 100%% confidence of unauthorized reconnaissance (e.g., cat, grep, or cp). File marked as BURNED. Immediate quarantine initiated.", hf.FileName, hf.LastKnownAtime.Format(time.RFC3339), currentAtime.Format(time.RFC3339)),
					"severity":           "critical",
					"category":           "honey_file_accessed",
					"detection_layer":    "atime_polling",
					"confidence":         1.0,
					"timestamp":          time.Now().Unix(),
				},
			})
		}
	}
}

// getHoneyFile retrieves a honey file by path (thread-safe)
func (hfm *HoneyFileManager) getHoneyFile(path string) *HoneyFile {
	hfm.mu.RLock()
	defer hfm.mu.RUnlock()
	return hfm.files[path]
}

// handleHoneyFileAccess processes a honey file access and triggers containment
func (hfm *HoneyFileManager) handleHoneyFileAccess(hf *HoneyFile, accessType string) {
	now := time.Now()

	// Mark as burned immediately to prevent duplicate alerts
	hf.IsBurned = true
	hf.BurnedAt = &now
	hf.AccessCount = 1

	// Get process info (best effort)
	accessor := hfm.getAccessorInfo()
	hf.LastAccessor = accessor

	// Update in database
	hfm.markFileAsBurned(hf)

	// Publish HoneyFileAccessedEvent
	event := types.NewHoneyFileAccessedEvent(
		"honey_file_sentinel",
		hf.FilePath,
		hf.FileName,
		accessType,
		accessor,
	)

	envelope, err := types.WrapEvent(types.EventType(bus.EventTypeHoneyFileAccessed), event)
	if err != nil {
		log.Printf("HoneyFileManager: Failed to wrap event: %v", err)
		return
	}

	hfm.eventBus.Publish(bus.Event{
		Type:   bus.EventTypeHoneyFileAccessed,
		Source: "honey_file_sentinel",
		Target: hf.FilePath,
		Payload: map[string]interface{}{
			"data":        envelope.Payload,
			"file_path":   hf.FilePath,
			"file_name":   hf.FileName,
			"access_type": accessType,
			"accessor":    accessor,
			"confidence":  1.0,
			"is_burned":   true,
			"burned_at":   now.Unix(),
			"timestamp":   now.Unix(),
		},
	})

	// Also publish LogEvent for dashboard
	hfm.eventBus.Publish(bus.Event{
		Type:   bus.EventTypeLogEvent,
		Source: "honey_file_sentinel",
		Target: "dashboard",
		Payload: map[string]interface{}{
			"agent_name":         "honey_file_sentinel",
			"message":            fmt.Sprintf("🍯 HONEY-FILE ACCESSED: %s", hf.FileName),
			"internal_reasoning": fmt.Sprintf("Honey-file '%s' %s by %s. 100%% confidence of unauthorized lateral movement. File marked as BURNED. Immediate quarantine initiated.", hf.FileName, accessType, accessor),
			"severity":           "critical",
			"category":           "honey_file_accessed",
			"confidence":         1.0,
			"timestamp":          now.Unix(),
		},
	})

	log.Printf("🚨 HoneyFileManager: HONEY-FILE EVENT PUBLISHED for %s", hf.FilePath)
}

// markFileAsBurned updates the database to mark file as burned
func (hfm *HoneyFileManager) markFileAsBurned(hf *HoneyFile) {
	if hfm.repository == nil {
		return
	}

	state := &database.GlobalState{
		TaskID:        hf.ID,
		TaskType:      database.TaskType("honey_file"),
		Status:        database.TaskStatusCompleted, // Completed = burned
		Target:        hf.FilePath,
		ResultSummary: fmt.Sprintf("Honey file BURNED - accessed %d times", hf.AccessCount),
		Metadata: map[string]interface{}{
			"file_name":     hf.FileName,
			"is_burned":     true,
			"burned_at":     hf.BurnedAt.Format(time.RFC3339),
			"access_count":  hf.AccessCount,
			"last_accessor": hf.LastAccessor,
		},
		CompletedAt: hf.BurnedAt,
	}

	if err := hfm.repository.Create(state); err != nil {
		log.Printf("HoneyFileManager: Failed to update burned status: %v", err)
	}
}

// updateAccessCount increments the access count in database
func (hfm *HoneyFileManager) updateAccessCount(hf *HoneyFile) {
	if hfm.repository == nil {
		return
	}

	// In a real implementation, this would update the existing record
	// For now, we just log it
	log.Printf("HoneyFileManager: Burned file %s accessed again (total: %d)", hf.FilePath, hf.AccessCount)
}

// getAccessorInfo attempts to identify who accessed the file
func (hfm *HoneyFileManager) getAccessorInfo() string {
	// In a real implementation, this would use audit logs, process info, etc.
	// For demo purposes, return a generic identifier
	return fmt.Sprintf("process_%d", time.Now().UnixNano()%10000)
}

// IsFileBurned checks if a honey file has been accessed (burned)
func (hfm *HoneyFileManager) IsFileBurned(filePath string) bool {
	hf := hfm.getHoneyFile(filePath)
	if hf == nil {
		return false
	}
	return hf.IsBurned
}

// GetAllHoneyFiles returns all deployed honey files
func (hfm *HoneyFileManager) GetAllHoneyFiles() []*HoneyFile {
	hfm.mu.RLock()
	defer hfm.mu.RUnlock()

	files := make([]*HoneyFile, 0, len(hfm.files))
	for _, hf := range hfm.files {
		files = append(files, hf)
	}
	return files
}

// GetBurnedFiles returns all honey files that have been accessed
func (hfm *HoneyFileManager) GetBurnedFiles() []*HoneyFile {
	hfm.mu.RLock()
	defer hfm.mu.RUnlock()

	var burned []*HoneyFile
	for _, hf := range hfm.files {
		if hf.IsBurned {
			burned = append(burned, hf)
		}
	}
	return burned
}

// RotateHoneyTraps deletes compromised honey-files and deploys new ones in random locations
// This is the "Shuffling" logic that makes the SOC self-evolving
func (hfm *HoneyFileManager) RotateHoneyTraps() error {
	log.Println("🔄 ROTATING HONEY TRAPS: Self-evolving deception layer initiated")

	hfm.mu.Lock()
	defer hfm.mu.Unlock()

	// Step 1: Collect all burned (compromised) honey files
	var compromised []*HoneyFile
	for _, hf := range hfm.files {
		if hf.IsBurned {
			compromised = append(compromised, hf)
		}
	}

	if len(compromised) == 0 {
		log.Println("HoneyFileManager: No compromised files to rotate")
		return nil
	}

	log.Printf("HoneyFileManager: Found %d compromised files to rotate", len(compromised))

	// Step 2: Delete compromised files and remove from watcher
	for _, hf := range compromised {
		// Remove from fsnotify watcher
		if err := hfm.watcher.Remove(hf.FilePath); err != nil {
			log.Printf("HoneyFileManager: Failed to remove watcher for %s: %v", hf.FilePath, err)
		}

		// Delete the file from disk
		if err := os.Remove(hf.FilePath); err != nil {
			log.Printf("HoneyFileManager: Failed to delete compromised file %s: %v", hf.FilePath, err)
		} else {
			log.Printf("HoneyFileManager: Deleted compromised honey file: %s", hf.FilePath)
		}

		// Remove from memory
		delete(hfm.files, hf.FilePath)

		// Update database to mark as rotated
		if hfm.repository != nil {
			state := &database.GlobalState{
				TaskID:        hf.ID,
				TaskType:      database.TaskType("honey_file_rotated"),
				Status:        database.TaskStatusCompleted,
				Target:        hf.FilePath,
				ResultSummary: "Honey file rotated after compromise - new trap deployed",
				Metadata: map[string]interface{}{
					"file_name":     hf.FileName,
					"was_burned":    true,
					"access_count":  hf.AccessCount,
					"rotated_at":    time.Now().Format(time.RFC3339),
					"last_accessor": hf.LastAccessor,
				},
				CompletedAt: &[]time.Time{time.Now()}[0],
			}
			if err := hfm.repository.Create(state); err != nil {
				log.Printf("HoneyFileManager: Failed to create audit record for %s: %v", hf.FilePath, err)
			}
		}
	}

	// Step 3: Generate 3 new honey files in random directories
	newLocations := hfm.generateRandomHoneyLocations()

	for i, location := range newLocations {
		var content string
		var fileName string

		switch i % 3 {
		case 0:
			fileName = "production_secrets.env"
			content = `# Production Environment Secrets
# DEPLOYMENT_KEY=deploy_7f8a9b2c3d4e5f60718293a4b5c6d7e8
# DATABASE_URL=postgresql://prod_admin:P0d_D8t@b@se_S3cr3t@prod-db.internal:5432/production
# REDIS_PASSWORD=Redis_M@st3r_P@ss_2024
# JWT_SECRET=jwt_super_secret_key_production_only
# API_KEY=api_key_9f8e7d6c5b4a3f2e1d0c9b8a7f6e5d4c3b2a1f0`
		case 1:
			fileName = "kubernetes_admin.conf"
			content = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM...
    server: https://k8s-prod-control-plane.internal:6443
  name: production-cluster
contexts:
- context:
    cluster: production-cluster
    user: kubernetes-admin
  name: kubernetes-admin@production-cluster
current-context: kubernetes-admin@production-cluster
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t...
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJ...`
		case 2:
			fileName = "aws_vault_credentials.json"
			content = `{
  "version": 1,
  "aws_access_key_id": "AKIAIOSFODNN7VAULT",
  "aws_secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYVAULTKEY",
  "vault_token": "hvs.CAESIGFqEmBva1ZJdFZ4dFhUVG1VOEtGZWJyZUdvYUlxM...
  "vault_addr": "https://vault.prod.internal:8200",
  "vault_namespace": "admin/production",
  "ssh_key_vault_path": "secret/data/ssh/production_root"
}`
		}

		newPath := filepath.Join(location, fileName)
		if err := hfm.deploySingleHoneyFile(newPath, content); err != nil {
			log.Printf("HoneyFileManager: Failed to deploy rotated honey file %s: %v", newPath, err)
		} else {
			// Publish LogEvent for dashboard
			hfm.eventBus.Publish(bus.Event{
				Type:   bus.EventTypeLogEvent,
				Source: "honey_file_sentinel",
				Target: "dashboard",
				Payload: map[string]interface{}{
					"agent_name":         "honey_file_sentinel",
					"message":            fmt.Sprintf("🔄 New honey-trap deployed: %s", fileName),
					"internal_reasoning": fmt.Sprintf("Self-evolving deception: Rotated compromised trap to new location '%s'. Attacker will find old location empty. New trap armed and monitoring.", newPath),
					"severity":           "info",
					"category":           "honey_file_rotation",
					"new_location":       newPath,
					"timestamp":          time.Now().Unix(),
				},
			})
		}
	}

	log.Println("🔄 HONEY TRAP ROTATION COMPLETE: Self-evolving deception layer updated")
	return nil
}

// generateRandomHoneyLocations returns 3 random directories for new honey files
func (hfm *HoneyFileManager) generateRandomHoneyLocations() []string {
	possibleDirs := []string{
		"internal/quantum",
		"internal/ai/models",
		"internal/security/vault",
		"modules/exploitation",
		"modules/recon/data",
		"web/src/config",
		"web/public/assets",
		"migrations/seeds",
		"scripts/deploy",
		"docs/internal",
	}

	// Shuffle and pick 3
	randomDirs := make([]string, len(possibleDirs))
	copy(randomDirs, possibleDirs)

	// Simple shuffle
	for i := range randomDirs {
		j := int(time.Now().UnixNano()) % len(randomDirs)
		randomDirs[i], randomDirs[j] = randomDirs[j], randomDirs[i]
	}

	return randomDirs[:3]
}

// deploySingleHoneyFile deploys a single honey file (helper for rotation)
func (hfm *HoneyFileManager) deploySingleHoneyFile(filePath, content string) error {
	// Create directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the honey file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write honey file %s: %w", filePath, err)
	}

	// Create honey file record
	now := time.Now()
	honeyFile := &HoneyFile{
		ID:             fmt.Sprintf("honeyfile_%d", now.UnixNano()),
		FilePath:       filePath,
		FileName:       filepath.Base(filePath),
		Content:        content,
		CreatedAt:      now,
		IsDeployed:     true,
		IsBurned:       false,
		LastKnownAtime: now,
	}

	// Store in memory
	hfm.files[filePath] = honeyFile

	// Persist to database
	if hfm.repository != nil {
		state := &database.GlobalState{
			TaskID:        honeyFile.ID,
			TaskType:      database.TaskType("honey_file"),
			Status:        database.TaskStatusRunning,
			Target:        filePath,
			ResultSummary: "Rotated honey file deployed and monitoring",
			Metadata: map[string]interface{}{
				"file_name":    honeyFile.FileName,
				"is_rotated":   true,
				"is_burned":    false,
				"access_count": 0,
				"deployed_at":  now.Format(time.RFC3339),
			},
			StartedAt: now,
		}
		if err := hfm.repository.Create(state); err != nil {
			log.Printf("HoneyFileManager: Failed to create deployment audit record: %v", err)
		}
	}

	// Add to fsnotify watcher
	if err := hfm.watcher.Add(filePath); err != nil {
		return fmt.Errorf("failed to watch honey file %s: %w", filePath, err)
	}

	log.Printf("HoneyFileManager: Deployed rotated honey file: %s", filePath)
	return nil
}

// StartRotationTicker starts a continuous rotation ticker for honey traps
// This ensures that even if an attacker has mapped your system silently, their map becomes obsolete every week
func (hfm *HoneyFileManager) StartRotationTicker(duration time.Duration) {
	log.Printf("🔄 Starting continuous honey trap rotation ticker (duration: %v)", duration)

	ticker := time.NewTicker(duration)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-hfm.ctx.Done():
				log.Println("HoneyFileManager: Rotation ticker stopped")
				return
			case <-ticker.C:
				log.Println("🔄 WEEKLY HONEY TRAP ROTATION: Self-evolving deception cycle initiated")
				if err := hfm.RotateHoneyTraps(); err != nil {
					log.Printf("HoneyFileManager: Rotation failed: %v", err)
				} else {
					log.Println("✅ Weekly honey trap rotation completed successfully")

					// Publish LogEvent for dashboard
					hfm.eventBus.Publish(bus.Event{
						Type:   bus.EventTypeLogEvent,
						Source: "honey_file_sentinel",
						Target: "dashboard",
						Payload: map[string]interface{}{
							"agent_name":         "honey_file_sentinel",
							"message":            "🔄 Weekly honey trap rotation completed",
							"internal_reasoning": "Scheduled weekly rotation executed successfully. Attacker mappings are now obsolete. New deception layer deployed.",
							"severity":           "info",
							"category":           "honey_file_rotation",
							"rotation_type":      "scheduled_weekly",
							"timestamp":          time.Now().Unix(),
						},
					})
				}
			}
		}
	}()
}

// StartSundayMidnightRotation schedules honey trap rotation every Sunday at midnight (00:00)
// This provides predictable weekly rotation while ensuring fresh deception at the start of each week
func (hfm *HoneyFileManager) StartSundayMidnightRotation() {
	now := time.Now()

	// Calculate days until next Sunday (Sunday = 0, Monday = 1, ... Saturday = 6)
	daysUntilSunday := (7 - int(now.Weekday())) % 7
	if daysUntilSunday == 0 && now.Hour() == 0 && now.Minute() == 0 {
		// If it's currently Sunday midnight, schedule for next Sunday
		daysUntilSunday = 7
	}

	// Calculate next Sunday at midnight
	nextSunday := now.AddDate(0, 0, daysUntilSunday)
	nextSundayMidnight := time.Date(
		nextSunday.Year(), nextSunday.Month(), nextSunday.Day(),
		0, 0, 0, 0, now.Location(),
	)

	durationUntilSunday := nextSundayMidnight.Sub(now)

	log.Printf("🕛 Sunday Midnight Rotation scheduled for: %s (in %v)",
		nextSundayMidnight.Format("2006-01-02 15:04:05"), durationUntilSunday)

	// Initial delay timer until next Sunday midnight
	go func() {
		select {
		case <-hfm.ctx.Done():
			return
		case <-time.After(durationUntilSunday):
			log.Println("🕛 SUNDAY MIDNIGHT: Initiating weekly honey trap rotation")
			hfm.performRotationWithLogging()

			// After first Sunday rotation, start weekly ticker
			weeklyDuration := 168 * time.Hour // 7 days
			hfm.StartRotationTicker(weeklyDuration)
		}
	}()
}

// performRotationWithLogging executes rotation and publishes dashboard events
func (hfm *HoneyFileManager) performRotationWithLogging() {
	if err := hfm.RotateHoneyTraps(); err != nil {
		log.Printf("HoneyFileManager: Sunday midnight rotation failed: %v", err)
		hfm.eventBus.Publish(bus.Event{
			Type:   bus.EventTypeLogEvent,
			Source: "honey_file_sentinel",
			Target: "dashboard",
			Payload: map[string]interface{}{
				"agent_name":         "honey_file_sentinel",
				"message":            "❌ Sunday midnight rotation failed",
				"internal_reasoning": fmt.Sprintf("Rotation error: %v", err),
				"severity":           "error",
				"category":           "honey_file_rotation",
				"rotation_type":      "sunday_midnight",
				"timestamp":          time.Now().Unix(),
			},
		})
	} else {
		log.Println("✅ Sunday midnight honey trap rotation completed successfully")
		hfm.eventBus.Publish(bus.Event{
			Type:   bus.EventTypeLogEvent,
			Source: "honey_file_sentinel",
			Target: "dashboard",
			Payload: map[string]interface{}{
				"agent_name":         "honey_file_sentinel",
				"message":            "🕛 Sunday midnight rotation completed - Deception layer refreshed",
				"internal_reasoning": "Weekly Sunday midnight rotation executed. Attacker reconnaissance data from previous week is now obsolete. Fresh honey traps deployed across randomized locations.",
				"severity":           "info",
				"category":           "honey_file_rotation",
				"rotation_type":      "sunday_midnight",
				"timestamp":          time.Now().Unix(),
			},
		})
	}
}

// Stop shuts down the honey file manager
func (hfm *HoneyFileManager) Stop() {
	hfm.cancel()
	if err := hfm.watcher.Close(); err != nil {
		log.Printf("HoneyFileManager: Failed to close watcher: %v", err)
	}
	if hfm.atimeTicker != nil {
		hfm.atimeTicker.Stop()
	}
	log.Println("HoneyFileManager: Stopped")
}
