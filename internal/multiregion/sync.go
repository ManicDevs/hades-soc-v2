package multiregion

import (
	"fmt"
	"sync"
	"time"
)

// SyncManager handles cross-region data synchronization
type SyncManager struct {
	regionManager *RegionManager
	config        *SyncConfig
	activeSyncs   map[string]*SyncSession
	syncMutex     sync.RWMutex
}

// SyncConfig holds synchronization configuration
type SyncConfig struct {
	Enabled            bool          `yaml:"enabled"`
	SyncInterval       time.Duration `yaml:"sync_interval"`
	MaxSyncWorkers     int           `yaml:"max_sync_workers"`
	SyncTimeout        time.Duration `yaml:"sync_timeout"`
	ConflictResolution string        `yaml:"conflict_resolution"` // "latest_wins", "manual", "merge"
	CompressionEnabled bool          `yaml:"compression_enabled"`
}

// SyncSession represents an active synchronization session
type SyncSession struct {
	ID            string    `json:"id"`
	SourceRegion  string    `json:"source_region"`
	TargetRegions []string  `json:"target_regions"`
	StartTime     time.Time `json:"start_time"`
	Status        string    `json:"status"` // "pending", "running", "completed", "failed"
	Progress      float64   `json:"progress"`
	ItemsSynced   int       `json:"items_synced"`
	TotalItems    int       `json:"total_items"`
	LastError     string    `json:"last_error"`
}

// SyncItem represents an item to be synchronized
type SyncItem struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"` // "user", "config", "threat_data", "analytics"
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Checksum  string      `json:"checksum"`
	Region    string      `json:"source_region"`
}

// NewSyncManager creates a new synchronization manager
func NewSyncManager(regionManager *RegionManager, config *SyncConfig) *SyncManager {
	sm := &SyncManager{
		regionManager: regionManager,
		config:        config,
		activeSyncs:   make(map[string]*SyncSession),
	}

	if config.Enabled {
		go sm.startSyncMonitoring()
	}

	return sm
}

// startSyncMonitoring begins continuous synchronization monitoring
func (sm *SyncManager) startSyncMonitoring() {
	ticker := time.NewTicker(sm.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.performSyncCycle()
		}
	}
}

// performSyncCycle executes a full synchronization cycle
func (sm *SyncManager) performSyncCycle() {
	regions := sm.regionManager.ListRegions()

	// Find the primary region (highest priority, active)
	var primaryRegion *Region
	for _, region := range regions {
		if region.Status == "active" && region.Priority == 1 {
			primaryRegion = region
			break
		}
	}

	if primaryRegion == nil {
		return // No primary region found
	}

	// Get items to sync from primary region
	items := sm.getSyncItems(primaryRegion.ID)
	if len(items) == 0 {
		return // No items to sync
	}

	// Sync to other active regions
	for _, region := range regions {
		if region.ID != primaryRegion.ID && region.Status == "active" {
			sm.syncToRegion(primaryRegion.ID, region.ID, items)
		}
	}
}

// getSyncItems retrieves items that need synchronization
func (sm *SyncManager) getSyncItems(regionID string) []*SyncItem {
	// Simulate getting sync items from database or cache
	items := make([]*SyncItem, 0)

	// Generate sample sync items
	for i := 0; i < 10; i++ {
		item := &SyncItem{
			ID:        fmt.Sprintf("item_%s_%d", regionID, i),
			Type:      "user",
			Data:      fmt.Sprintf("Sample data from %s - item %d", regionID, i),
			Timestamp: time.Now().Add(-time.Duration(i) * time.Minute),
			Checksum:  fmt.Sprintf("checksum_%d", i),
			Region:    regionID,
		}
		items = append(items, item)
	}

	return items
}

// syncToRegion synchronizes items to a target region
func (sm *SyncManager) syncToRegion(sourceRegion, targetRegion string, items []*SyncItem) {
	sessionID := fmt.Sprintf("sync_%s_%s_%d", sourceRegion, targetRegion, time.Now().Unix())

	session := &SyncSession{
		ID:            sessionID,
		SourceRegion:  sourceRegion,
		TargetRegions: []string{targetRegion},
		StartTime:     time.Now(),
		Status:        "running",
		TotalItems:    len(items),
		ItemsSynced:   0,
		Progress:      0.0,
	}

	sm.syncMutex.Lock()
	sm.activeSyncs[sessionID] = session
	sm.syncMutex.Unlock()

	// Simulate sync process
	go sm.performSync(session, items)
}

// performSync executes the actual synchronization
func (sm *SyncManager) performSync(session *SyncSession, items []*SyncItem) {
	for i, item := range items {
		// Simulate sync delay
		time.Sleep(100 * time.Millisecond)

		// Update progress
		sm.syncMutex.Lock()
		session.ItemsSynced = i + 1
		session.Progress = float64(i+1) / float64(len(items)) * 100
		sm.syncMutex.Unlock()

		// Simulate occasional sync failure
		if i%7 == 0 && time.Now().Unix()%10 == 0 {
			sm.syncMutex.Lock()
			session.Status = "failed"
			session.LastError = fmt.Sprintf("Failed to sync item %s", item.ID)
			sm.syncMutex.Unlock()
			return
		}
	}

	// Mark as completed
	sm.syncMutex.Lock()
	session.Status = "completed"
	session.Progress = 100.0
	sm.syncMutex.Unlock()
}

// StartSync initiates a manual synchronization
func (sm *SyncManager) StartSync(sourceRegion string, targetRegions []string) (*SyncSession, error) {
	if !sm.config.Enabled {
		return nil, fmt.Errorf("synchronization is disabled")
	}

	// Validate source region
	if _, err := sm.regionManager.GetRegion(sourceRegion); err != nil {
		return nil, fmt.Errorf("source region not found: %v", err)
	}

	// Validate target regions
	for _, targetRegion := range targetRegions {
		if _, err := sm.regionManager.GetRegion(targetRegion); err != nil {
			return nil, fmt.Errorf("target region %s not found: %v", targetRegion, err)
		}
	}

	// Get items to sync
	items := sm.getSyncItems(sourceRegion)
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to sync")
	}

	// Create sync session
	sessionID := fmt.Sprintf("manual_sync_%s_%d", sourceRegion, time.Now().Unix())
	session := &SyncSession{
		ID:            sessionID,
		SourceRegion:  sourceRegion,
		TargetRegions: targetRegions,
		StartTime:     time.Now(),
		Status:        "pending",
		TotalItems:    len(items),
		ItemsSynced:   0,
		Progress:      0.0,
	}

	sm.syncMutex.Lock()
	sm.activeSyncs[sessionID] = session
	sm.syncMutex.Unlock()

	// Start sync in background
	go func() {
		session.Status = "running"
		for _, targetRegion := range targetRegions {
			sm.syncToRegion(sourceRegion, targetRegion, items)
		}
		session.Status = "completed"
	}()

	return session, nil
}

// GetSyncStatus returns the status of all active sync sessions
func (sm *SyncManager) GetSyncStatus() map[string]interface{} {
	sm.syncMutex.RLock()
	defer sm.syncMutex.RUnlock()

	status := make(map[string]interface{})
	status["config_enabled"] = sm.config.Enabled
	status["active_syncs"] = len(sm.activeSyncs)
	status["max_sync_workers"] = sm.config.MaxSyncWorkers
	status["sync_interval"] = sm.config.SyncInterval.String()

	sessions := make([]map[string]interface{}, 0)
	for _, session := range sm.activeSyncs {
		sessionInfo := map[string]interface{}{
			"id":             session.ID,
			"source_region":  session.SourceRegion,
			"target_regions": session.TargetRegions,
			"start_time":     session.StartTime,
			"status":         session.Status,
			"progress":       session.Progress,
			"items_synced":   session.ItemsSynced,
			"total_items":    session.TotalItems,
			"last_error":     session.LastError,
		}
		sessions = append(sessions, sessionInfo)
	}
	status["sync_sessions"] = sessions

	return status
}

// GetSyncSession returns a specific sync session
func (sm *SyncManager) GetSyncSession(sessionID string) (*SyncSession, error) {
	sm.syncMutex.RLock()
	defer sm.syncMutex.RUnlock()

	session, exists := sm.activeSyncs[sessionID]
	if !exists {
		return nil, fmt.Errorf("sync session %s not found", sessionID)
	}

	return session, nil
}

// CancelSync cancels an active sync session
func (sm *SyncManager) CancelSync(sessionID string) error {
	sm.syncMutex.Lock()
	defer sm.syncMutex.Unlock()

	session, exists := sm.activeSyncs[sessionID]
	if !exists {
		return fmt.Errorf("sync session %s not found", sessionID)
	}

	if session.Status == "completed" {
		return fmt.Errorf("cannot cancel completed sync session")
	}

	session.Status = "cancelled"
	return nil
}

// CleanupCompletedSyncs removes completed sync sessions older than the specified duration
func (sm *SyncManager) CleanupCompletedSyncs(maxAge time.Duration) {
	sm.syncMutex.Lock()
	defer sm.syncMutex.Unlock()

	for sessionID, session := range sm.activeSyncs {
		if (session.Status == "completed" || session.Status == "failed" || session.Status == "cancelled") &&
			time.Since(session.StartTime) > maxAge {
			delete(sm.activeSyncs, sessionID)
		}
	}
}

// ResolveConflict handles synchronization conflicts
func (sm *SyncManager) ResolveConflict(sourceItem, targetItem *SyncItem) (*SyncItem, error) {
	switch sm.config.ConflictResolution {
	case "latest_wins":
		if sourceItem.Timestamp.After(targetItem.Timestamp) {
			return sourceItem, nil
		}
		return targetItem, nil
	case "manual":
		return nil, fmt.Errorf("manual conflict resolution required")
	case "merge":
		// Simple merge strategy - combine data
		mergedItem := *sourceItem
		mergedItem.Data = fmt.Sprintf("MERGED: %s + %s", sourceItem.Data, targetItem.Data)
		mergedItem.Timestamp = time.Now()
		return &mergedItem, nil
	default:
		return sourceItem, nil
	}
}

// GetSyncMetrics returns synchronization metrics
func (sm *SyncManager) GetSyncMetrics() map[string]interface{} {
	sm.syncMutex.RLock()
	defer sm.syncMutex.RUnlock()

	metrics := make(map[string]interface{})

	totalSessions := len(sm.activeSyncs)
	completedSessions := 0
	failedSessions := 0
	runningSessions := 0
	totalItemsSynced := 0
	totalItems := 0

	for _, session := range sm.activeSyncs {
		switch session.Status {
		case "completed":
			completedSessions++
		case "failed":
			failedSessions++
		case "running":
			runningSessions++
		}

		totalItemsSynced += session.ItemsSynced
		totalItems += session.TotalItems
	}

	metrics["total_sessions"] = totalSessions
	metrics["completed_sessions"] = completedSessions
	metrics["failed_sessions"] = failedSessions
	metrics["running_sessions"] = runningSessions
	metrics["total_items_synced"] = totalItemsSynced
	metrics["total_items"] = totalItems
	metrics["success_rate"] = float64(completedSessions) / float64(totalSessions) * 100
	metrics["average_progress"] = float64(totalItemsSynced) / float64(totalItems) * 100

	return metrics
}
