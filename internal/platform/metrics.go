package platform

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// SOC Metrics for Hades Security Operations Center
var (
	// Global risk level gauge (0-100 scale)
	hadesGlobalRiskLevel = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hades_global_risk_level",
		Help: "Current global risk level across all SOC operations (0-100 scale)",
	})

	// Autonomous actions counter
	hadesAutonomousActionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "hades_autonomous_actions_total",
		Help: "Total number of autonomous security actions taken by the SOC",
	})

	// Threat detection counter by severity
	hadesThreatsDetectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hades_threats_detected_total",
		Help: "Total number of threats detected by severity level",
	}, []string{"severity"})

	// Orchestrator decisions counter
	hadesOrchestratorDecisionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hades_orchestrator_decisions_total",
		Help: "Total number of decisions made by the orchestrator by action type",
	}, []string{"action_type", "status"})

	// Active sessions gauge
	hadesActiveSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hades_active_sessions",
		Help: "Number of active security sessions",
	})

	// Worker pool status gauge
	hadesWorkerPoolActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hades_worker_pool_active",
		Help: "Number of active worker pool processes",
	})

	// Event processing duration histogram
	hadesEventProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "hades_event_processing_duration_seconds",
		Help:    "Time taken to process security events",
		Buckets: prometheus.DefBuckets,
	})

	// Database operations counter
	hadesDatabaseOperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hades_database_operations_total",
		Help: "Total number of database operations by type",
	}, []string{"operation", "status"})
)

// MetricsCollector manages SOC metrics collection and reporting
type MetricsCollector struct {
	mu                sync.RWMutex
	currentRiskLevel  float64
	lastRiskUpdate    time.Time
	actionCounts      map[string]int64
	threatCounts      map[string]int64
	sessionCount      int
	workerCount       int
	startTime         time.Time
}

// NewMetricsCollector creates a new metrics collector instance
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		currentRiskLevel: 0.0,
		lastRiskUpdate:   time.Now(),
		actionCounts:     make(map[string]int64),
		threatCounts:     make(map[string]int64),
		sessionCount:     0,
		workerCount:      0,
		startTime:        time.Now(),
	}
}

// UpdateGlobalRiskLevel updates the global risk level metric
func (mc *MetricsCollector) UpdateGlobalRiskLevel(level float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.currentRiskLevel = level
	mc.lastRiskUpdate = time.Now()
	hadesGlobalRiskLevel.Set(level)
}

// IncrementAutonomousActions increments the autonomous actions counter
func (mc *MetricsCollector) IncrementAutonomousActions() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.actionCounts["total"]++
	hadesAutonomousActionsTotal.Inc()
}

// IncrementThreatDetected increments threat detection counter by severity
func (mc *MetricsCollector) IncrementThreatDetected(severity string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.threatCounts[severity]++
	hadesThreatsDetectedTotal.WithLabelValues(severity).Inc()
}

// RecordOrchestratorDecision records an orchestrator decision
func (mc *MetricsCollector) RecordOrchestratorDecision(actionType, status string) {
	hadesOrchestratorDecisionsTotal.WithLabelValues(actionType, status).Inc()
	mc.IncrementAutonomousActions()
}

// UpdateActiveSessions updates the active sessions gauge
func (mc *MetricsCollector) UpdateActiveSessions(count int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.sessionCount = count
	hadesActiveSessions.Set(float64(count))
}

// UpdateWorkerPoolStatus updates the worker pool status gauge
func (mc *MetricsCollector) UpdateWorkerPoolStatus(count int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.workerCount = count
	hadesWorkerPoolActive.Set(float64(count))
}

// RecordEventProcessingDuration records event processing duration
func (mc *MetricsCollector) RecordEventProcessingDuration(duration time.Duration) {
	hadesEventProcessingDuration.Observe(duration.Seconds())
}

// RecordDatabaseOperation records a database operation
func (mc *MetricsCollector) RecordDatabaseOperation(operation, status string) {
	hadesDatabaseOperationsTotal.WithLabelValues(operation, status).Inc()
}

// GetCurrentRiskLevel returns the current risk level
func (mc *MetricsCollector) GetCurrentRiskLevel() float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.currentRiskLevel
}

// GetMetricsSummary returns a summary of current metrics
func (mc *MetricsCollector) GetMetricsSummary() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	uptime := time.Since(mc.startTime)

	return map[string]interface{}{
		"global_risk_level":     mc.currentRiskLevel,
		"last_risk_update":      mc.lastRiskUpdate.Format(time.RFC3339),
		"autonomous_actions":    mc.actionCounts["total"],
		"threats_by_severity":   mc.threatCounts,
		"active_sessions":       mc.sessionCount,
		"active_workers":        mc.workerCount,
		"uptime_seconds":        uptime.Seconds(),
		"uptime_human":          uptime.String(),
	}
}

// CalculateRiskLevel calculates risk level based on threat activity and system state
func (mc *MetricsCollector) CalculateRiskLevel(threatCount int, criticalThreats int, systemLoad float64) float64 {
	// Base risk calculation: (threats * severity_weight) + system_load_factor
	baseRisk := float64(threatCount) * 2.0
	criticalRisk := float64(criticalThreats) * 10.0
	systemRisk := systemLoad * 5.0

	totalRisk := baseRisk + criticalRisk + systemRisk

	// Cap at 100
	if totalRisk > 100 {
		totalRisk = 100
	}

	// Update the metric
	mc.UpdateGlobalRiskLevel(totalRisk)

	return totalRisk
}

// Global metrics collector instance
var globalMetricsCollector *MetricsCollector
var metricsOnce sync.Once

// GetGlobalMetrics returns the singleton metrics collector
func GetGlobalMetrics() *MetricsCollector {
	metricsOnce.Do(func() {
		globalMetricsCollector = NewMetricsCollector()
	})
	return globalMetricsCollector
}
