package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type QuotaStatus struct {
	Model       string    `json:"model"`
	Provider    string    `json:"provider"`
	Used        int       `json:"used"`
	Limit       int       `json:"limit"`
	Remaining   int       `json:"remaining"`
	LastReset   time.Time `json:"last_reset"`
	IsExhausted bool      `json:"is_exhausted"`
}

type QuotaMonitor struct {
	mu           sync.RWMutex
	statusFile   string
	monitoring   bool
	stopChan     chan struct{}
	quotaManager *QuotaManager
}

func NewQuotaMonitor(quotaManager *QuotaManager) *QuotaMonitor {
	return &QuotaMonitor{
		statusFile:   "/tmp/hades_quota_status.json",
		stopChan:     make(chan struct{}),
		quotaManager: quotaManager,
	}
}

func (qm *QuotaMonitor) StartMonitoring(interval time.Duration) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if qm.monitoring {
		return
	}

	qm.monitoring = true
	go qm.monitorLoop(interval)

	log.Printf("Started quota monitoring with %v interval", interval)
}

func (qm *QuotaMonitor) StopMonitoring() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if !qm.monitoring {
		return
	}

	close(qm.stopChan)
	qm.monitoring = false

	log.Printf("Stopped quota monitoring")
}

func (qm *QuotaMonitor) monitorLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			qm.UpdateStatus()
		case <-qm.stopChan:
			return
		}
	}
}

func (qm *QuotaMonitor) UpdateStatus() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	statuses := qm.getCurrentStatuses()

	// Write to status file
	if data, err := json.MarshalIndent(statuses, "", "  "); err == nil {
		if err := os.WriteFile(qm.statusFile, data, 0644); err != nil {
			log.Printf("Failed to write quota status file: %v", err)
		}
	} else {
		log.Printf("Failed to marshal quota status: %v", err)
	}

	// Log warnings for exhausted quotas
	for _, status := range statuses {
		if status.IsExhausted {
			log.Printf("WARNING: Quota exhausted for %s (%s)", status.Model, status.Provider)
		} else if status.Remaining < 5 {
			log.Printf("WARNING: Low quota for %s (%s): %d remaining", status.Model, status.Provider, status.Remaining)
		}
	}
}

func (qm *QuotaMonitor) getCurrentStatuses() []QuotaStatus {
	qm.quotaManager.mu.RLock()
	defer qm.quotaManager.mu.RUnlock()

	var statuses []QuotaStatus

	for model, limit := range qm.quotaManager.dailyLimits {
		used := qm.quotaManager.requestCounts[model]
		remaining := limit - used
		isExhausted := remaining <= 0

		// Find provider for this model
		provider := "unknown"
		for _, modelConfig := range qm.quotaManager.availableModels {
			if modelConfig.Model == model {
				provider = string(modelConfig.Provider)
				break
			}
		}

		statuses = append(statuses, QuotaStatus{
			Model:       model,
			Provider:    provider,
			Used:        used,
			Limit:       limit,
			Remaining:   remaining,
			LastReset:   qm.quotaManager.lastResetTime,
			IsExhausted: isExhausted,
		})
	}

	return statuses
}

func (qm *QuotaMonitor) GetStatus() []QuotaStatus {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	return qm.getCurrentStatuses()
}

func (qm *QuotaMonitor) GetStatusString() string {
	statuses := qm.GetStatus()

	var result string
	result += "╔════════════════════════════════════════════════════════════╗\n"
	result += "║                  API QUOTA STATUS                        ║\n"
	result += "╚════════════════════════════════════════════════════════════╝\n\n"

	for _, status := range statuses {
		statusStr := "✅ OK"
		if status.IsExhausted {
			statusStr = "❌ EXHAUSTED"
		} else if status.Remaining < 5 {
			statusStr = "⚠️  LOW"
		}

		result += fmt.Sprintf("Model: %s (%s)\n", status.Model, status.Provider)
		result += fmt.Sprintf("  Status: %s\n", statusStr)
		result += fmt.Sprintf("  Usage: %d/%d (%d remaining)\n", status.Used, status.Limit, status.Remaining)
		result += fmt.Sprintf("  Last Reset: %s\n\n", status.LastReset.Format("2006-01-02 15:04:05"))
	}

	return result
}
