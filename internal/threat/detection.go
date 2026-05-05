package threat

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"hades-v2/internal/database"
)

// ThreatDetector represents the main threat detection engine
type ThreatDetector struct {
	db                database.Database
	algorithms        []DetectionAlgorithm
	mlModel           *MLModel
	threatIntel       *ThreatIntelligence
	patternMatcher    *PatternMatcher
	anomalyDetector   *AnomalyDetector
	mu                sync.RWMutex
	knownSignatures   map[string]bool
	whitelistIPs      map[string]bool
	blacklistIPs      map[string]bool
	suspiciousDomains map[string]float64
}

// DetectionAlgorithm interface for different detection methods
type DetectionAlgorithm interface {
	Name() string
	Detect(ctx context.Context, event SecurityEvent) (*ThreatAlert, error)
	Confidence() float64
	Train(data []SecurityEvent) error
}

// SecurityEvent represents a security event to analyze
type SecurityEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	EventType string                 `json:"event_type"`
	SourceIP  string                 `json:"source_ip"`
	DestIP    string                 `json:"dest_ip"`
	Port      int                    `json:"port"`
	Protocol  string                 `json:"protocol"`
	Payload   string                 `json:"payload"`
	UserAgent string                 `json:"user_agent"`
	Method    string                 `json:"method"`
	Path      string                 `json:"path"`
	Query     string                 `json:"query"`
	Headers   map[string]string      `json:"headers"`
	Metadata  map[string]interface{} `json:"metadata"`
	Severity  string                 `json:"severity"`
	RiskScore float64                `json:"risk_score"`
}

// ThreatAlert represents a detected threat
type ThreatAlert struct {
	ID              string                 `json:"id"`
	Timestamp       time.Time              `json:"timestamp"`
	ThreatType      string                 `json:"threat_type"`
	Severity        string                 `json:"severity"`
	Confidence      float64                `json:"confidence"`
	SourceIP        string                 `json:"source_ip"`
	DestIP          string                 `json:"dest_ip"`
	Description     string                 `json:"description"`
	Indicators      []ThreatIndicator      `json:"indicators"`
	Mitigation      MitigationAction       `json:"mitigation"`
	RelatedEvents   []string               `json:"related_events"`
	Metadata        map[string]interface{} `json:"metadata"`
	Status          string                 `json:"status"`
	AssignedTo      string                 `json:"assigned_to"`
	EscalationLevel int                    `json:"escalation_level"`
}

// ThreatIndicator represents indicators of compromise
type ThreatIndicator struct {
	Type        string    `json:"type"`
	Value       string    `json:"value"`
	Confidence  float64   `json:"confidence"`
	Source      string    `json:"source"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Description string    `json:"description"`
}

// MitigationAction represents automated mitigation actions
type MitigationAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Automated   bool                   `json:"automated"`
	Executed    bool                   `json:"executed"`
	Parameters  map[string]interface{} `json:"parameters"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewThreatDetector creates a new threat detection engine
func NewThreatDetector(db database.Database) *ThreatDetector {
	td := &ThreatDetector{
		db:                db,
		algorithms:        make([]DetectionAlgorithm, 0),
		mlModel:           NewMLModel(),
		threatIntel:       NewThreatIntelligence(),
		patternMatcher:    NewPatternMatcher(),
		anomalyDetector:   NewAnomalyDetector(),
		knownSignatures:   make(map[string]bool),
		whitelistIPs:      make(map[string]bool),
		blacklistIPs:      make(map[string]bool),
		suspiciousDomains: make(map[string]float64),
	}

	// Initialize detection algorithms
	td.algorithms = append(td.algorithms,
		&SignatureBasedDetection{td: td},
		&AnomalyBasedDetection{td: td, detector: td.anomalyDetector},
		&BehavioralAnalysis{td: td, ml: td.mlModel},
		&NetworkTrafficAnalysis{td: td},
		&MalwareDetection{td: td},
		&PhishingDetection{td: td},
		&DDoSDetection{td: td},
		&SQLInjectionDetection{td: td},
		&XSSDetection{td: td},
	)

	return td
}

// AnalyzeEvent analyzes a security event for threats
func (td *ThreatDetector) AnalyzeEvent(ctx context.Context, event SecurityEvent) ([]*ThreatAlert, error) {
	td.mu.RLock()
	defer td.mu.RUnlock()

	var alerts []*ThreatAlert
	var wg sync.WaitGroup
	alertsChan := make(chan *ThreatAlert, len(td.algorithms))

	// Run detection algorithms in parallel
	for _, algorithm := range td.algorithms {
		wg.Add(1)
		go func(alg DetectionAlgorithm) {
			defer wg.Done()
			alert, err := alg.Detect(ctx, event)
			if err != nil {
				log.Printf("Detection algorithm %s failed: %v", alg.Name(), err)
				return
			}
			if alert != nil {
				alertsChan <- alert
			}
		}(algorithm)
	}

	// Wait for all algorithms to complete
	go func() {
		wg.Wait()
		close(alertsChan)
	}()

	// Collect alerts
	for alert := range alertsChan {
		alerts = append(alerts, alert)
	}

	// Correlate and prioritize alerts
	correlatedAlerts := td.correlateAlerts(alerts)
	prioritizedAlerts := td.prioritizeAlerts(correlatedAlerts)

	return prioritizedAlerts, nil
}

// correlateAlerts correlates related alerts
func (td *ThreatDetector) correlateAlerts(alerts []*ThreatAlert) []*ThreatAlert {
	if len(alerts) <= 1 {
		return alerts
	}

	// Group alerts by source IP and time window
	alertGroups := make(map[string][]*ThreatAlert)
	timeWindow := 5 * time.Minute

	for _, alert := range alerts {
		key := fmt.Sprintf("%s_%d", alert.SourceIP, alert.Timestamp.Truncate(timeWindow).Unix())
		alertGroups[key] = append(alertGroups[key], alert)
	}

	// Create correlated alerts
	var correlated []*ThreatAlert
	for _, group := range alertGroups {
		if len(group) == 1 {
			correlated = append(correlated, group[0])
			continue
		}

		// Merge alerts into a single correlated alert
		mergedAlert := td.mergeAlerts(group)
		correlated = append(correlated, mergedAlert)
	}

	return correlated
}

// mergeAlerts merges multiple alerts into a single correlated alert
func (td *ThreatDetector) mergeAlerts(alerts []*ThreatAlert) *ThreatAlert {
	if len(alerts) == 0 {
		return nil
	}

	if len(alerts) == 1 {
		return alerts[0]
	}

	// Create merged alert
	merged := &ThreatAlert{
		ID:              fmt.Sprintf("correlated_%d", time.Now().UnixNano()),
		Timestamp:       alerts[0].Timestamp,
		ThreatType:      "correlated_attack",
		Severity:        td.calculateMergedSeverity(alerts),
		Confidence:      td.calculateMergedConfidence(alerts),
		SourceIP:        alerts[0].SourceIP,
		DestIP:          alerts[0].DestIP,
		Description:     fmt.Sprintf("Correlated attack with %d indicators", len(alerts)),
		Indicators:      td.mergeIndicators(alerts),
		RelatedEvents:   td.mergeRelatedEvents(alerts),
		Metadata:        td.mergeMetadata(alerts),
		Status:          "new",
		EscalationLevel: td.calculateEscalationLevel(alerts),
	}

	return merged
}

// prioritizeAlerts prioritizes alerts based on risk
func (td *ThreatDetector) prioritizeAlerts(alerts []*ThreatAlert) []*ThreatAlert {
	sort.Slice(alerts, func(i, j int) bool {
		scoreI := td.calculateRiskScore(alerts[i])
		scoreJ := td.calculateRiskScore(alerts[j])
		return scoreI > scoreJ
	})

	return alerts
}

// calculateRiskScore calculates overall risk score for an alert
func (td *ThreatDetector) calculateRiskScore(alert *ThreatAlert) float64 {
	severityWeight := map[string]float64{
		"critical": 10.0,
		"high":     7.0,
		"medium":   4.0,
		"low":      2.0,
	}

	baseScore := severityWeight[alert.Severity]
	confidenceBonus := alert.Confidence * 3.0
	escalationBonus := float64(alert.EscalationLevel) * 2.0

	return baseScore + confidenceBonus + escalationBonus
}

// calculateMergedSeverity calculates merged severity from multiple alerts
func (td *ThreatDetector) calculateMergedSeverity(alerts []*ThreatAlert) string {
	severityLevels := map[string]int{
		"critical": 4,
		"high":     3,
		"medium":   2,
		"low":      1,
	}

	maxLevel := 0
	for _, alert := range alerts {
		if level, exists := severityLevels[alert.Severity]; exists && level > maxLevel {
			maxLevel = level
		}
	}

	severityNames := []string{"low", "medium", "high", "critical"}
	if maxLevel > 0 && maxLevel <= len(severityNames) {
		return severityNames[maxLevel-1]
	}
	return "medium"
}

// calculateMergedConfidence calculates merged confidence from multiple alerts
func (td *ThreatDetector) calculateMergedConfidence(alerts []*ThreatAlert) float64 {
	if len(alerts) == 0 {
		return 0
	}

	var totalConfidence float64
	for _, alert := range alerts {
		totalConfidence += alert.Confidence
	}

	// Use weighted average with more weight for higher confidence
	avgConfidence := totalConfidence / float64(len(alerts))
	return math.Min(avgConfidence*1.2, 1.0) // Boost confidence for correlated alerts
}

// mergeIndicators merges indicators from multiple alerts
func (td *ThreatDetector) mergeIndicators(alerts []*ThreatAlert) []ThreatIndicator {
	indicatorMap := make(map[string]ThreatIndicator)

	for _, alert := range alerts {
		for _, indicator := range alert.Indicators {
			key := fmt.Sprintf("%s_%s", indicator.Type, indicator.Value)
			if existing, exists := indicatorMap[key]; exists {
				// Merge with higher confidence
				if indicator.Confidence > existing.Confidence {
					indicatorMap[key] = indicator
				}
			} else {
				indicatorMap[key] = indicator
			}
		}
	}

	var merged []ThreatIndicator
	for _, indicator := range indicatorMap {
		merged = append(merged, indicator)
	}

	return merged
}

// mergeRelatedEvents merges related events from multiple alerts
func (td *ThreatDetector) mergeRelatedEvents(alerts []*ThreatAlert) []string {
	eventMap := make(map[string]bool)

	for _, alert := range alerts {
		for _, event := range alert.RelatedEvents {
			eventMap[event] = true
		}
	}

	var merged []string
	for event := range eventMap {
		merged = append(merged, event)
	}

	return merged
}

// mergeMetadata merges metadata from multiple alerts
func (td *ThreatDetector) mergeMetadata(alerts []*ThreatAlert) map[string]interface{} {
	merged := make(map[string]interface{})

	for _, alert := range alerts {
		for key, value := range alert.Metadata {
			merged[key] = value
		}
	}

	// Add correlation metadata
	merged["correlated_alerts_count"] = len(alerts)
	merged["correlation_timestamp"] = time.Now()

	return merged
}

// calculateEscalationLevel calculates escalation level for correlated alerts
func (td *ThreatDetector) calculateEscalationLevel(alerts []*ThreatAlert) int {
	level := 0
	for _, alert := range alerts {
		if alert.EscalationLevel > level {
			level = alert.EscalationLevel
		}
	}

	// Escalate if multiple high-severity alerts
	highSeverityCount := 0
	for _, alert := range alerts {
		if alert.Severity == "high" || alert.Severity == "critical" {
			highSeverityCount++
		}
	}

	if highSeverityCount >= 2 {
		level++
	}

	return level
}

// UpdateThreatIntelligence updates threat intelligence data
func (td *ThreatDetector) UpdateThreatIntelligence(ctx context.Context, indicators []ThreatIndicator) error {
	return td.threatIntel.UpdateIndicators(ctx, indicators)
}

// TrainModels trains the machine learning models
func (td *ThreatDetector) TrainModels(ctx context.Context, events []SecurityEvent) error {
	for _, algorithm := range td.algorithms {
		if err := algorithm.Train(events); err != nil {
			log.Printf("Failed to train algorithm %s: %v", algorithm.Name(), err)
		}
	}
	return nil
}

// GetDetectionStats returns detection statistics
func (td *ThreatDetector) GetDetectionStats(ctx context.Context) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"algorithms_count":   len(td.algorithms),
		"known_signatures":   len(td.knownSignatures),
		"whitelist_ips":      len(td.whitelistIPs),
		"blacklist_ips":      len(td.blacklistIPs),
		"suspicious_domains": len(td.suspiciousDomains),
		"ml_model_trained":   td.mlModel.IsTrained(),
		"last_update":        time.Now(),
	}

	return stats, nil
}

// AddSignature adds a known threat signature
func (td *ThreatDetector) AddSignature(signature string) {
	td.mu.Lock()
	defer td.mu.Unlock()

	hash := sha256.Sum256([]byte(signature))
	td.knownSignatures[hex.EncodeToString(hash[:])] = true
}

// AddWhitelistIP adds an IP to the whitelist
func (td *ThreatDetector) AddWhitelistIP(ip string) {
	td.mu.Lock()
	defer td.mu.Unlock()
	td.whitelistIPs[ip] = true
}

// AddBlacklistIP adds an IP to the blacklist
func (td *ThreatDetector) AddBlacklistIP(ip string) {
	td.mu.Lock()
	defer td.mu.Unlock()
	td.blacklistIPs[ip] = true
}

// IsWhitelisted checks if an IP is whitelisted
func (td *ThreatDetector) IsWhitelisted(ip string) bool {
	td.mu.RLock()
	defer td.mu.RUnlock()
	return td.whitelistIPs[ip]
}

// IsBlacklisted checks if an IP is blacklisted
func (td *ThreatDetector) IsBlacklisted(ip string) bool {
	td.mu.RLock()
	defer td.mu.RUnlock()
	return td.blacklistIPs[ip]
}
