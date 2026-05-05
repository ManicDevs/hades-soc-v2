package multiregion

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// FailoverManager handles automatic failover between regions
type FailoverManager struct {
	regionManager   *RegionManager
	config          *FailoverConfig
	activeFailovers map[string]*FailoverSession
	failoverMutex   sync.RWMutex
}

// FailoverConfig holds failover configuration
type FailoverConfig struct {
	Enabled               bool          `yaml:"enabled"`
	HealthCheckInterval   time.Duration `yaml:"health_check_interval"`
	FailoverThreshold     int           `yaml:"failover_threshold"` // consecutive failures
	RecoveryCheckInterval time.Duration `yaml:"recovery_check_interval"`
	MaxFailoverAttempts   int           `yaml:"max_failover_attempts"`
	FailoverTimeout       time.Duration `yaml:"failover_timeout"`
}

// FailoverSession represents an active failover session
type FailoverSession struct {
	ID              string    `json:"id"`
	PrimaryRegion   string    `json:"primary_region"`
	BackupRegion    string    `json:"backup_region"`
	StartTime       time.Time `json:"start_time"`
	FailureCount    int       `json:"failure_count"`
	Status          string    `json:"status"` // "failing_over", "active", "recovering"
	LastHealthCheck time.Time `json:"last_health_check"`
}

// FailoverEvent represents a failover event
type FailoverEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"` // "failure_detected", "failover_initiated", "failover_completed", "recovery_detected"
	RegionID    string    `json:"region_id"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"` // "low", "medium", "high", "critical"
}

// NewFailoverManager creates a new failover manager
func NewFailoverManager(regionManager *RegionManager, config *FailoverConfig) *FailoverManager {
	fm := &FailoverManager{
		regionManager:   regionManager,
		config:          config,
		activeFailovers: make(map[string]*FailoverSession),
	}

	if config.Enabled {
		go fm.startFailoverMonitoring()
	}

	return fm
}

// startFailoverMonitoring begins continuous failover monitoring
func (fm *FailoverManager) startFailoverMonitoring() {
	healthTicker := time.NewTicker(fm.config.HealthCheckInterval)
	recoveryTicker := time.NewTicker(fm.config.RecoveryCheckInterval)
	defer healthTicker.Stop()
	defer recoveryTicker.Stop()

	for {
		select {
		case <-healthTicker.C:
			fm.checkRegionHealth()
		case <-recoveryTicker.C:
			fm.checkRecovery()
		}
	}
}

// checkRegionHealth monitors all regions for health issues
func (fm *FailoverManager) checkRegionHealth() {
	regions := fm.regionManager.ListRegions()

	for _, region := range regions {
		if region.Status == "active" {
			if time.Since(region.LastHealth) > fm.config.HealthCheckInterval*2 {
				fm.handleRegionFailure(region.ID)
			}
		}
	}
}

// handleRegionFailure processes a region failure
func (fm *FailoverManager) handleRegionFailure(regionID string) {
	fm.failoverMutex.Lock()
	defer fm.failoverMutex.Unlock()

	// Check if failover session already exists
	if session, exists := fm.activeFailovers[regionID]; exists {
		session.FailureCount++
		session.LastHealthCheck = time.Now()

		// Log failure event
		fm.logFailoverEvent(&FailoverEvent{
			Timestamp:   time.Now(),
			EventType:   "failure_detected",
			RegionID:    regionID,
			Description: fmt.Sprintf("Region %s has failed %d times", regionID, session.FailureCount),
			Severity:    determineSeverity(session.FailureCount),
		})

		// Initiate failover if threshold reached
		if session.FailureCount >= fm.config.FailoverThreshold {
			fm.initiateFailover(session)
		}
	} else {
		// Create new failover session
		session := &FailoverSession{
			ID:              fmt.Sprintf("failover_%s_%d", regionID, time.Now().Unix()),
			PrimaryRegion:   regionID,
			StartTime:       time.Now(),
			FailureCount:    1,
			Status:          "failing_over",
			LastHealthCheck: time.Now(),
		}

		fm.activeFailovers[regionID] = session

		fm.logFailoverEvent(&FailoverEvent{
			Timestamp:   time.Now(),
			EventType:   "failure_detected",
			RegionID:    regionID,
			Description: fmt.Sprintf("First failure detected for region %s", regionID),
			Severity:    "medium",
		})
	}
}

// initiateFailover starts the failover process
func (fm *FailoverManager) initiateFailover(session *FailoverSession) {
	if session.Status == "active" {
		return // Already failed over
	}

	// Find best backup region
	backupRegion, err := fm.findBackupRegion(session.PrimaryRegion)
	if err != nil {
		log.Printf("Failed to find backup region for %s: %v", session.PrimaryRegion, err)
		return
	}

	session.BackupRegion = backupRegion.ID
	session.Status = "active"

	// Update region status
	if primaryRegion, err := fm.regionManager.GetRegion(session.PrimaryRegion); err == nil {
		primaryRegion.Status = "failed"
	}

	if backupRegion.Status == "standby" {
		backupRegion.Status = "active"
	}

	fm.logFailoverEvent(&FailoverEvent{
		Timestamp:   time.Now(),
		EventType:   "failover_completed",
		RegionID:    session.PrimaryRegion,
		Description: fmt.Sprintf("Failover completed: %s -> %s", session.PrimaryRegion, backupRegion.ID),
		Severity:    "high",
	})

	log.Printf("Failover initiated: %s -> %s", session.PrimaryRegion, backupRegion.ID)
}

// findBackupRegion finds the best backup region for failover
func (fm *FailoverManager) findBackupRegion(failedRegionID string) (*Region, error) {
	regions := fm.regionManager.ListRegions()

	var candidates []*Region
	for _, region := range regions {
		if region.ID != failedRegionID && (region.Status == "standby" || region.Status == "active") {
			candidates = append(candidates, region)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no backup regions available")
	}

	// Select region with lowest load and highest priority
	bestRegion := candidates[0]
	bestScore := float64(bestRegion.Capacity-bestRegion.Load) / float64(bestRegion.Priority)

	for _, region := range candidates[1:] {
		score := float64(region.Capacity-region.Load) / float64(region.Priority)
		if score > bestScore {
			bestScore = score
			bestRegion = region
		}
	}

	return bestRegion, nil
}

// checkRecovery monitors failed regions for recovery
func (fm *FailoverManager) checkRecovery() {
	fm.failoverMutex.RLock()
	defer fm.failoverMutex.RUnlock()

	for regionID, session := range fm.activeFailovers {
		if session.Status == "active" {
			// Check if primary region has recovered
			if region, err := fm.regionManager.GetRegion(regionID); err == nil {
				if time.Since(region.LastHealth) < fm.config.RecoveryCheckInterval {
					fm.handleRegionRecovery(session, region)
				}
			}
		}
	}
}

// handleRegionRecovery processes region recovery
func (fm *FailoverManager) handleRegionRecovery(session *FailoverSession, recoveredRegion *Region) {
	fm.failoverMutex.Lock()
	defer fm.failoverMutex.Unlock()

	session.Status = "recovering"

	// Switch back to primary region if it's stable
	if backupRegion, err := fm.regionManager.GetRegion(session.BackupRegion); err == nil {
		backupRegion.Status = "standby"
	}

	recoveredRegion.Status = "active"

	// Remove failover session after successful recovery
	delete(fm.activeFailovers, session.PrimaryRegion)

	fm.logFailoverEvent(&FailoverEvent{
		Timestamp:   time.Now(),
		EventType:   "recovery_detected",
		RegionID:    session.PrimaryRegion,
		Description: fmt.Sprintf("Region %s has recovered", session.PrimaryRegion),
		Severity:    "low",
	})

	log.Printf("Region %s has recovered, switching back", session.PrimaryRegion)
}

// logFailoverEvent logs a failover event
func (fm *FailoverManager) logFailoverEvent(event *FailoverEvent) {
	// In a real implementation, this would log to a file, database, or monitoring system
	log.Printf("FAILOVER EVENT: %s - %s (%s): %s",
		event.Timestamp.Format("2006-01-02 15:04:05"),
		event.EventType,
		event.RegionID,
		event.Description)
}

// determineSeverity determines the severity level based on failure count
func determineSeverity(failureCount int) string {
	switch {
	case failureCount >= 5:
		return "critical"
	case failureCount >= 3:
		return "high"
	case failureCount >= 2:
		return "medium"
	default:
		return "low"
	}
}

// GetFailoverStatus returns current failover status
func (fm *FailoverManager) GetFailoverStatus() map[string]interface{} {
	fm.failoverMutex.RLock()
	defer fm.failoverMutex.RUnlock()

	status := make(map[string]interface{})
	status["active_failovers"] = len(fm.activeFailovers)
	status["config_enabled"] = fm.config.Enabled
	status["failover_threshold"] = fm.config.FailoverThreshold
	status["max_failover_attempts"] = fm.config.MaxFailoverAttempts

	failovers := make([]map[string]interface{}, 0)
	for _, session := range fm.activeFailovers {
		failoverInfo := map[string]interface{}{
			"id":                session.ID,
			"primary_region":    session.PrimaryRegion,
			"backup_region":     session.BackupRegion,
			"start_time":        session.StartTime,
			"failure_count":     session.FailureCount,
			"status":            session.Status,
			"last_health_check": session.LastHealthCheck,
		}
		failovers = append(failovers, failoverInfo)
	}
	status["failover_sessions"] = failovers

	return status
}

// GetFailoverEvents returns recent failover events
func (fm *FailoverManager) GetFailoverEvents(limit int) []*FailoverEvent {
	// In a real implementation, this would query from a database or log file
	// For now, return empty slice
	return []*FailoverEvent{}
}

// ManualFailover allows manual failover initiation
func (fm *FailoverManager) ManualFailover(primaryRegionID, backupRegionID string) error {
	if !fm.config.Enabled {
		return fmt.Errorf("failover is disabled")
	}

	primaryRegion, err := fm.regionManager.GetRegion(primaryRegionID)
	if err != nil {
		return fmt.Errorf("primary region not found: %v", err)
	}

	backupRegion, err := fm.regionManager.GetRegion(backupRegionID)
	if err != nil {
		return fmt.Errorf("backup region not found: %v", err)
	}

	if backupRegion.Status != "standby" {
		return fmt.Errorf("backup region %s is not in standby mode", backupRegionID)
	}

	// Create manual failover session
	session := &FailoverSession{
		ID:              fmt.Sprintf("manual_failover_%s_%d", primaryRegionID, time.Now().Unix()),
		PrimaryRegion:   primaryRegionID,
		BackupRegion:    backupRegionID,
		StartTime:       time.Now(),
		FailureCount:    0,
		Status:          "active",
		LastHealthCheck: time.Now(),
	}

	fm.failoverMutex.Lock()
	fm.activeFailovers[primaryRegionID] = session
	fm.failoverMutex.Unlock()

	// Perform the failover
	primaryRegion.Status = "maintenance"
	backupRegion.Status = "active"

	fm.logFailoverEvent(&FailoverEvent{
		Timestamp:   time.Now(),
		EventType:   "failover_completed",
		RegionID:    primaryRegionID,
		Description: fmt.Sprintf("Manual failover completed: %s -> %s", primaryRegionID, backupRegionID),
		Severity:    "high",
	})

	return nil
}

// CancelFailover cancels an active failover session
func (fm *FailoverManager) CancelFailover(regionID string) error {
	fm.failoverMutex.Lock()
	defer fm.failoverMutex.Unlock()

	session, exists := fm.activeFailovers[regionID]
	if !exists {
		return fmt.Errorf("no active failover for region %s", regionID)
	}

	// Restore primary region
	if primaryRegion, err := fm.regionManager.GetRegion(regionID); err == nil {
		primaryRegion.Status = "active"
	}

	// Set backup region back to standby
	if session.BackupRegion != "" {
		if backupRegion, err := fm.regionManager.GetRegion(session.BackupRegion); err == nil {
			backupRegion.Status = "standby"
		}
	}

	// Remove failover session
	delete(fm.activeFailovers, regionID)

	fm.logFailoverEvent(&FailoverEvent{
		Timestamp:   time.Now(),
		EventType:   "recovery_detected",
		RegionID:    regionID,
		Description: fmt.Sprintf("Failover cancelled for region %s", regionID),
		Severity:    "medium",
	})

	return nil
}
