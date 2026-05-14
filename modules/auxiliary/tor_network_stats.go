package auxiliary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"hades-v2/pkg/sdk"

	"gitlab.torproject.org/cerberus-droid/torgo/src/core"
)

// TorNetworkStats provides comprehensive Tor network statistics and monitoring
type TorNetworkStats struct {
	*sdk.BaseModule

	mu          sync.RWMutex
	torInstance *core.Tor
	httpClient  *http.Client
	lastUpdate  time.Time
	stats       *TorStats
}

// TorStats contains comprehensive Tor network statistics
type TorStats struct {
	// Circuit Statistics
	ActiveCircuits     int     `json:"active_circuits"`
	TotalCircuits      int     `json:"total_circuits"`
	CircuitBuildRate   float64 `json:"circuit_build_rate"`
	CircuitSuccessRate float64 `json:"circuit_success_rate"`
	CircuitFailureRate float64 `json:"circuit_failure_rate"`

	// Bandwidth Statistics
	BandwidthRead    int64 `json:"bandwidth_read_bytes"`
	BandwidthWritten int64 `json:"bandwidth_written_bytes"`
	BandwidthRate    int64 `json:"bandwidth_rate_bytes_per_sec"`

	// Hidden Service Statistics
	HiddenServices    int `json:"hidden_services_count"`
	ActiveServices    int `json:"active_hidden_services"`
	PublishedServices int `json:"published_hidden_services"`

	// Network Statistics
	ConnectedRelays  int  `json:"connected_relays"`
	GuardNodes       int  `json:"guard_nodes"`
	ExitNodes        int  `json:"exit_nodes"`
	DirectoryFetched bool `json:"directory_fetched"`

	// Performance Metrics
	AverageLatency      time.Duration `json:"average_latency_ms"`
	ConnectionStability float64       `json:"connection_stability"`
	Throughput          float64       `json:"throughput_kbps"`

	// Security Metrics
	ThreatLevel    string  `json:"threat_level"`
	SecurityEvents int     `json:"security_events_count"`
	AnomalyScore   float64 `json:"anomaly_score"`

	// System Health
	TorVersion  string        `json:"tor_version"`
	Uptime      time.Duration `json:"uptime_seconds"`
	MemoryUsage int64         `json:"memory_usage_bytes"`
	CPUUsage    float64       `json:"cpu_usage_percent"`
}

// NewTorNetworkStats creates a new Tor network statistics module
func NewTorNetworkStats() *TorNetworkStats {
	return &TorNetworkStats{
		BaseModule: sdk.NewBaseModule(
			"tor_network_stats",
			"Comprehensive Tor network statistics and monitoring using torgo",
			sdk.CategoryAuxiliary,
		),
		stats: &TorStats{
			ThreatLevel: "low",
		},
	}
}

// Execute starts the Tor network statistics collection
func (tns *TorNetworkStats) Execute(ctx context.Context) error {
	tns.SetStatus(sdk.StatusRunning)
	defer tns.SetStatus(sdk.StatusIdle)

	if err := tns.startStatsCollection(ctx); err != nil {
		return fmt.Errorf("hades.tor_network_stats: failed to start stats collection: %w", err)
	}

	<-ctx.Done()
	tns.stopStatsCollection()
	return nil
}

// startStatsCollection begins collecting statistics from Tor instance
func (tns *TorNetworkStats) startStatsCollection(ctx context.Context) error {
	tns.mu.Lock()
	defer tns.mu.Unlock()

	// Initialize HTTP client for Tor control interface
	tns.httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Start periodic stats collection
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := tns.collectStats(); err != nil {
				log.Printf("hades.tor_network_stats: failed to collect stats: %v", err)
			}
			tns.lastUpdate = time.Now()
		}
	}
}

// collectStats gathers comprehensive statistics from Tor instance
func (tns *TorNetworkStats) collectStats() error {
	tns.mu.Lock()
	defer tns.mu.Unlock()

	if tns.torInstance == nil {
		return fmt.Errorf("tor instance not available")
	}

	// Get Tor snapshot for comprehensive stats
	snapshot := tns.torInstance.StatusSnapshot(nil)
	if snapshot == nil {
		return fmt.Errorf("failed to get tor snapshot")
	}

	// Update circuit statistics
	tns.updateCircuitStats(snapshot)

	// Update bandwidth statistics
	tns.updateBandwidthStats(snapshot)

	// Update hidden service statistics
	tns.updateHiddenServiceStats(snapshot)

	// Update network statistics
	tns.updateNetworkStats(snapshot)

	// Update performance metrics
	tns.updatePerformanceMetrics(snapshot)

	// Update security metrics
	tns.updateSecurityMetrics(snapshot)

	// Update system health
	tns.updateSystemHealth(snapshot)

	return nil
}

// updateCircuitStats updates circuit-related statistics
func (tns *TorNetworkStats) updateCircuitStats(snapshot interface{}) {
	// Extract circuit stats from snapshot
	if snapshotMap, ok := snapshot.(map[string]interface{}); ok {
		if circuitStats, exists := snapshotMap["ClientCircuitStats"]; exists {
			if circuitStatsMap, ok := circuitStats.(map[string]interface{}); ok {
				tns.stats.ActiveCircuits = tns.getIntFromInterface(circuitStatsMap["ActiveCircuits"])
				tns.stats.TotalCircuits = tns.getIntFromInterface(circuitStatsMap["TotalCircuits"])

				// Calculate rates
				if tns.stats.TotalCircuits > 0 {
					successRate := float64(tns.getIntFromInterface(circuitStatsMap["SuccessfulCircuits"])) / float64(tns.stats.TotalCircuits) * 100
					tns.stats.CircuitSuccessRate = successRate
					tns.stats.CircuitFailureRate = 100 - successRate
				}
			}
		}
	}
}

// updateBandwidthStats updates bandwidth-related statistics
func (tns *TorNetworkStats) updateBandwidthStats(snapshot interface{}) {
	// Extract bandwidth stats from snapshot
	if snapshotMap, ok := snapshot.(map[string]interface{}); ok {
		if bandwidthStats, exists := snapshotMap["BandwidthStats"]; exists {
			if bandwidthStatsMap, ok := bandwidthStats.(map[string]interface{}); ok {
				tns.stats.BandwidthRead = tns.getInt64FromInterface(bandwidthStatsMap["BytesRead"])
				tns.stats.BandwidthWritten = tns.getInt64FromInterface(bandwidthStatsMap["BytesWritten"])
				tns.stats.BandwidthRate = tns.getInt64FromInterface(bandwidthStatsMap["BandwidthRate"])
			}
		}
	}
}

// updateHiddenServiceStats updates hidden service statistics
func (tns *TorNetworkStats) updateHiddenServiceStats(snapshot interface{}) {
	// Extract hidden service stats from snapshot
	if snapshotMap, ok := snapshot.(map[string]interface{}); ok {
		if hiddenServices, exists := snapshotMap["HiddenServices"]; exists {
			if servicesList, ok := hiddenServices.([]interface{}); ok {
				tns.stats.HiddenServices = len(servicesList)

				var activeCount, publishedCount int
				var isRunning, isPublished bool
				var serviceMap map[string]interface{}

				for _, service := range servicesList {
					if serviceMap, ok = service.(map[string]interface{}); ok {
						if isRunning, exists = serviceMap["IsRunning"].(bool); exists {
							if isRunning {
								activeCount++
							}
						}
						if isPublished, exists = serviceMap["IsPublished"].(bool); exists {
							if isPublished {
								publishedCount++
							}
						}
					}
				}

				tns.stats.ActiveServices = activeCount
				tns.stats.PublishedServices = publishedCount
			}
		}
	}
}

// updateNetworkStats updates network-related statistics
func (tns *TorNetworkStats) updateNetworkStats(snapshot interface{}) {
	// Extract network stats from snapshot
	if snapshotMap, ok := snapshot.(map[string]interface{}); ok {
		if descriptorCount, exists := snapshotMap["ConsensusDescriptorCount"]; exists {
			tns.stats.ConnectedRelays = tns.getIntFromInterface(descriptorCount)
			tns.stats.DirectoryFetched = true
		}

		// Update guard/exit node counts from consensus
		if knownRouters, exists := snapshotMap["LocalDirAuthorityKnownRouters"]; exists {
			totalRouters := tns.getIntFromInterface(knownRouters)
			tns.stats.GuardNodes = int(float64(totalRouters) * 0.1) // ~10% are guards
			tns.stats.ExitNodes = int(float64(totalRouters) * 0.05) // ~5% are exits
		}
	}
}

// updatePerformanceMetrics calculates performance metrics
func (tns *TorNetworkStats) updatePerformanceMetrics(snapshot interface{}) {
	// Extract performance metrics from snapshot
	if snapshotMap, ok := snapshot.(map[string]interface{}); ok {
		if circuitStats, exists := snapshotMap["ClientCircuitStats"]; exists {
			if circuitStatsMap, ok := circuitStats.(map[string]interface{}); ok {
				// Calculate average latency based on circuit build times
				if avgBuildTime, exists := circuitStatsMap["AvgCircuitBuildTime"]; exists {
					if duration, ok := avgBuildTime.(time.Duration); ok {
						tns.stats.AverageLatency = duration
					}
				}

				// Calculate connection stability
				if tns.stats.CircuitSuccessRate > 0 {
					tns.stats.ConnectionStability = tns.stats.CircuitSuccessRate / 100.0
				}

				// Calculate throughput based on bandwidth
				if tns.stats.BandwidthRate > 0 {
					tns.stats.Throughput = float64(tns.stats.BandwidthRate) / 1024.0 // Convert to Kbps
				}
			}
		}
	}
}

// updateSecurityMetrics updates security-related metrics
func (tns *TorNetworkStats) updateSecurityMetrics(snapshot interface{}) {
	// Analyze threat level based on circuit failures and network conditions
	if tns.stats.CircuitFailureRate > 20 {
		tns.stats.ThreatLevel = "high"
	} else if tns.stats.CircuitFailureRate > 10 {
		tns.stats.ThreatLevel = "medium"
	} else if tns.stats.CircuitFailureRate > 5 {
		tns.stats.ThreatLevel = "low"
	} else {
		tns.stats.ThreatLevel = "minimal"
	}

	// Count security events from various Tor components
	tns.stats.SecurityEvents = tns.countSecurityEvents(snapshot)

	// Calculate anomaly score based on various factors
	tns.stats.AnomalyScore = tns.calculateAnomalyScore()
}

// updateSystemHealth updates system health metrics
func (tns *TorNetworkStats) updateSystemHealth(snapshot interface{}) {
	// Extract system health from snapshot
	if snapshotMap, ok := snapshot.(map[string]interface{}); ok {
		if torVersion, exists := snapshotMap["TorVersion"]; exists {
			if version, ok := torVersion.(string); ok {
				tns.stats.TorVersion = version
			}
		}
	}

	tns.stats.Uptime = time.Since(time.Now()) // This would be tracked separately

	// Memory and CPU usage would need to be collected from system metrics
	// For now, use placeholder values
	tns.stats.MemoryUsage = 0
	tns.stats.CPUUsage = 0.0
}

// countSecurityEvents counts various security-related events
func (tns *TorNetworkStats) countSecurityEvents(snapshot interface{}) int {
	events := 0

	// Extract security events from snapshot
	if snapshotMap, ok := snapshot.(map[string]interface{}); ok {
		if circuitStats, exists := snapshotMap["ClientCircuitStats"]; exists {
			if circuitStatsMap, ok := circuitStats.(map[string]interface{}); ok {
				if failedCircuits, exists := circuitStatsMap["FailedCircuits"]; exists {
					events += tns.getIntFromInterface(failedCircuits)
				}
			}
		}
	}

	return events
}

// calculateAnomalyScore calculates an anomaly detection score
func (tns *TorNetworkStats) calculateAnomalyScore() float64 {
	score := 0.0

	// Factor in circuit failure rate
	if tns.stats.CircuitFailureRate > 0 {
		score += tns.stats.CircuitFailureRate * 0.5
	}

	// Factor in connection stability
	if tns.stats.ConnectionStability < 0.8 {
		score += (0.8 - tns.stats.ConnectionStability) * 20
	}

	// Factor in bandwidth anomalies
	if tns.stats.BandwidthRate > 0 {
		// Check for unusual bandwidth patterns
		if tns.stats.BandwidthRate > 10485760 { // > 10 MB/s
			score += 10
		}
	}

	return score
}

// stopStatsCollection stops the statistics collection
func (tns *TorNetworkStats) stopStatsCollection() {
	tns.mu.Lock()
	defer tns.mu.Unlock()

	log.Printf("Tor network statistics collection stopped")
}

// GetStats returns current Tor network statistics
func (tns *TorNetworkStats) GetStats() *TorStats {
	tns.mu.RLock()
	defer tns.mu.RUnlock()

	return tns.stats
}

// GetStatsJSON returns statistics in JSON format
func (tns *TorNetworkStats) GetStatsJSON() ([]byte, error) {
	stats := tns.GetStats()
	return json.MarshalIndent(stats, "", "  ")
}

// ExportStats exports statistics to a file
func (tns *TorNetworkStats) ExportStats(filename string) error {
	statsJSON, err := tns.GetStatsJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(filename, statsJSON, 0644)
}

// SetTorInstance sets the Tor instance for monitoring
func (tns *TorNetworkStats) SetTorInstance(torInstance *core.Tor) {
	tns.mu.Lock()
	defer tns.mu.Unlock()

	tns.torInstance = torInstance
}

// Helper methods for type conversion from interface{}
func (tns *TorNetworkStats) getIntFromInterface(value interface{}) int {
	if v, ok := value.(int); ok {
		return v
	}
	if v, ok := value.(float64); ok {
		return int(v)
	}
	return 0
}

func (tns *TorNetworkStats) getInt64FromInterface(value interface{}) int64 {
	if v, ok := value.(int64); ok {
		return v
	}
	if v, ok := value.(int); ok {
		return int64(v)
	}
	if v, ok := value.(float64); ok {
		return int64(v)
	}
	return 0
}
