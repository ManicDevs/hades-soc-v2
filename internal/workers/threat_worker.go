package workers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"

	"hades-v2/internal/database"
)

// ThreatWorker specializes in real-time threat detection and analysis
type ThreatWorker struct {
	*Worker
	learningData map[string]*ThreatPattern
	patterns     []*ThreatPattern
	lastScan     time.Time
}

// ThreatPattern represents a learned threat pattern
type ThreatPattern struct {
	ID         int                    `json:"id"`
	Pattern    string                 `json:"pattern"`
	Severity   string                 `json:"severity"`
	Category   string                 `json:"category"`
	Confidence float64                `json:"confidence"`
	Count      int                    `json:"count"`
	LastSeen   time.Time              `json:"last_seen"`
	Indicators map[string]interface{} `json:"indicators"`
	Actions    []string               `json:"actions"`
}

// ThreatIntelligence represents external threat intelligence data
type ThreatIntelligence struct {
	Source     string            `json:"source"`
	Threats    []ThreatIndicator `json:"threats"`
	Confidence float64           `json:"confidence"`
	Timestamp  time.Time         `json:"timestamp"`
}

// ThreatIndicator represents a specific threat indicator
type ThreatIndicator struct {
	Type        string  `json:"type"` // ip, domain, hash, url, pattern
	Value       string  `json:"value"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Source      string  `json:"source"`
}

// NewThreatWorker creates a specialized threat detection worker
func NewThreatWorker(id int, name string, db database.Database, maxRetries int, retryDelay time.Duration) *ThreatWorker {
	baseWorker := NewWorker(id, name, db, maxRetries, retryDelay)

	return &ThreatWorker{
		Worker:       baseWorker,
		learningData: make(map[string]*ThreatPattern),
		patterns:     make([]*ThreatPattern, 0),
		lastScan:     time.Now().Add(-24 * time.Hour), // Start with scan 24 hours ago
	}
}

// LoadThreatPatterns loads existing threat patterns from database
func (tw *ThreatWorker) LoadThreatPatterns() error {
	sqlDB, ok := tw.Database.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT id, pattern, severity, category, confidence, count, last_seen, indicators, actions
		FROM threat_patterns
		ORDER BY confidence DESC
	`

	rows, err := sqlDB.Query(query)
	if err != nil {
		return fmt.Errorf("failed to load threat patterns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pattern ThreatPattern
		var indicatorsJSON, actionsJSON string

		err := rows.Scan(&pattern.ID, &pattern.Pattern, &pattern.Severity, &pattern.Category,
			&pattern.Confidence, &pattern.Count, &pattern.LastSeen, &indicatorsJSON, &actionsJSON)
		if err != nil {
			return fmt.Errorf("failed to scan threat pattern: %w", err)
		}

		// Parse JSON fields
		if indicatorsJSON != "" {
			json.Unmarshal([]byte(indicatorsJSON), &pattern.Indicators)
		}
		if actionsJSON != "" {
			json.Unmarshal([]byte(actionsJSON), &pattern.Actions)
		}

		tw.patterns = append(tw.patterns, &pattern)
		tw.learningData[pattern.Pattern] = &pattern
	}

	log.Printf("Loaded %d threat patterns for worker %s", len(tw.patterns), tw.Name)
	return nil
}

// executeThreatScan performs intelligent threat scanning
func (tw *ThreatWorker) executeThreatScan(task database.WorkerTask) error {
	log.Printf("ThreatWorker %s performing intelligent threat scan", tw.Name)

	// Load threat patterns if not already loaded
	if len(tw.learningData) == 0 {
		if err := tw.LoadThreatPatterns(); err != nil {
			log.Printf("Warning: Failed to load threat patterns: %v", err)
		}
	}

	// Perform multi-layered threat analysis
	results := tw.performMultiLayeredAnalysis(task)

	// Update patterns based on findings
	tw.updateThreatPatterns(results)

	// Store detected threats
	if err := tw.storeDetectedThreats(results); err != nil {
		return fmt.Errorf("failed to store detected threats: %w", err)
	}

	// Trigger automated responses
	tw.triggerAutomatedResponses(results)

	tw.lastScan = time.Now()
	log.Printf("ThreatWorker %s completed intelligent scan, found %d potential threats", tw.Name, len(results))

	return nil
}

// performMultiLayeredAnalysis conducts comprehensive threat analysis
func (tw *ThreatWorker) performMultiLayeredAnalysis(task database.WorkerTask) []database.Threat {
	var threats []database.Threat

	// Layer 1: Pattern-based detection
	patternThreats := tw.detectPatternBasedThreats()
	threats = append(threats, patternThreats...)

	// Layer 2: Anomaly detection
	anomalyThreats := tw.detectAnomalies()
	threats = append(threats, anomalyThreats...)

	// Layer 3: Behavioral analysis
	behaviorThreats := tw.detectBehavioralAnomalies()
	threats = append(threats, behaviorThreats...)

	// Layer 4: External intelligence integration
	intelThreats := tw.integrateThreatIntelligence()
	threats = append(threats, intelThreats...)

	return threats
}

// detectPatternBasedThreats uses learned patterns to detect threats
func (tw *ThreatWorker) detectPatternBasedThreats() []database.Threat {
	var threats []database.Threat

	// Simulate scanning various data sources
	dataSources := tw.getDataSourceData()

	for _, data := range dataSources {
		for _, pattern := range tw.patterns {
			if tw.matchesPattern(data, pattern) {
				threat := database.Threat{
					Title:       fmt.Sprintf("Pattern Match: %s", pattern.Category),
					Description: fmt.Sprintf("Detected %s pattern: %s", pattern.Category, pattern.Pattern),
					Severity:    pattern.Severity,
					Status:      "active",
					Source:      "pattern_detection",
					Target:      data.Target,
					DetectedAt:  time.Now(),
					CreatedAt:   time.Now(),
				}

				threats = append(threats, threat)

				// Update pattern statistics
				pattern.Count++
				pattern.LastSeen = time.Now()
			}
		}
	}

	return threats
}

// detectAnomalies identifies unusual patterns and behaviors
func (tw *ThreatWorker) detectAnomalies() []database.Threat {
	var threats []database.Threat

	// Get recent system metrics
	metrics := tw.getSystemMetrics()

	// Detect anomalies in metrics
	if metrics.CPUUsage > 90 {
		threats = append(threats, database.Threat{
			Title:       "High CPU Usage Anomaly",
			Description: fmt.Sprintf("CPU usage spike detected at %.1f%%", metrics.CPUUsage),
			Severity:    "medium",
			Status:      "investigating",
			Source:      "anomaly_detection",
			DetectedAt:  time.Now(),
			CreatedAt:   time.Now(),
		})
	}

	if metrics.MemoryUsage > 95 {
		threats = append(threats, database.Threat{
			Title:       "High Memory Usage Anomaly",
			Description: fmt.Sprintf("Memory usage spike detected at %.1f%%", metrics.MemoryUsage),
			Severity:    "high",
			Status:      "investigating",
			Source:      "anomaly_detection",
			DetectedAt:  time.Now(),
			CreatedAt:   time.Now(),
		})
	}

	// Detect unusual network patterns
	if metrics.ErrorRate > 5.0 {
		threats = append(threats, database.Threat{
			Title:       "High Error Rate Anomaly",
			Description: fmt.Sprintf("Error rate spike detected at %.1f%%", metrics.ErrorRate),
			Severity:    "high",
			Status:      "investigating",
			Source:      "anomaly_detection",
			DetectedAt:  time.Now(),
			CreatedAt:   time.Now(),
		})
	}

	return threats
}

// detectBehavioralAnomalies analyzes user and system behaviors
func (tw *ThreatWorker) detectBehavioralAnomalies() []database.Threat {
	var threats []database.Threat

	// Get recent audit logs
	auditLogs := tw.getRecentAuditLogs()

	// Analyze for suspicious patterns
	userActivity := make(map[int]int)
	failedLogins := make(map[int]int)

	for _, log := range auditLogs {
		userActivity[log.UserID]++
		if log.Action == "login_failed" {
			failedLogins[log.UserID]++
		}
	}

	// Detect brute force attempts
	for userID, failures := range failedLogins {
		if failures > 5 {
			threats = append(threats, database.Threat{
				Title:       "Potential Brute Force Attack",
				Description: fmt.Sprintf("Multiple failed login attempts detected for user %d: %d failures", userID, failures),
				Severity:    "high",
				Status:      "active",
				Source:      "behavioral_analysis",
				Target:      fmt.Sprintf("user_%d", userID),
				DetectedAt:  time.Now(),
				CreatedAt:   time.Now(),
			})
		}
	}

	// Detect unusual activity patterns
	for userID, activity := range userActivity {
		if activity > 1000 { // Unusually high activity
			threats = append(threats, database.Threat{
				Title:       "Unusual User Activity",
				Description: fmt.Sprintf("User %d showing unusual activity pattern: %d actions", userID, activity),
				Severity:    "medium",
				Status:      "investigating",
				Source:      "behavioral_analysis",
				Target:      fmt.Sprintf("user_%d", userID),
				DetectedAt:  time.Now(),
				CreatedAt:   time.Now(),
			})
		}
	}

	return threats
}

// integrateThreatIntelligence incorporates external threat feeds
func (tw *ThreatWorker) integrateThreatIntelligence() []database.Threat {
	var threats []database.Threat

	// Simulate threat intelligence feeds
	feeds := []string{
		"malware_domains",
		"malicious_ips",
		"vulnerability_database",
		"security_advisories",
	}

	for _, feed := range feeds {
		intel := tw.fetchThreatIntelligence(feed)
		for _, indicator := range intel.Threats {
			if tw.isThreatRelevant(indicator) {
				threat := database.Threat{
					Title:       fmt.Sprintf("Threat Intelligence: %s", feed),
					Description: fmt.Sprintf("External threat detected: %s (%s)", indicator.Value, indicator.Description),
					Severity:    indicator.Severity,
					Status:      "active",
					Source:      fmt.Sprintf("threat_intelligence_%s", intel.Source),
					Target:      indicator.Value,
					DetectedAt:  time.Now(),
					CreatedAt:   time.Now(),
				}

				threats = append(threats, threat)
			}
		}
	}

	return threats
}

// Helper methods
func (tw *ThreatWorker) matchesPattern(data interface{}, pattern *ThreatPattern) bool {
	// Implement pattern matching logic
	dataStr := fmt.Sprintf("%v", data)

	// Use regex for pattern matching
	matched, _ := regexp.MatchString(pattern.Pattern, dataStr)
	return matched && pattern.Confidence > 0.7 // Only high confidence matches
}

type DataSource struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Target  string `json:"target"`
}

// getDataSourceData returns simulated data from various sources
func (tw *ThreatWorker) getDataSourceData() []DataSource {
	// Simulate getting data from various sources
	return []DataSource{
		{Type: "log", Content: "Failed login attempt from 192.168.1.100", Target: "auth_system"},
		{Type: "network", Content: "Unusual traffic to port 4444", Target: "firewall"},
		{Type: "file", Content: "Suspicious file detected: malware.exe", Target: "file_system"},
	}
}

func (tw *ThreatWorker) getSystemMetrics() database.SystemMetrics {
	// Simulate getting system metrics
	return database.SystemMetrics{
		CPUUsage:      75.5,
		MemoryUsage:   82.3,
		DiskUsage:     45.2,
		NetworkIn:     1024000,
		NetworkOut:    512000,
		ActiveUsers:   15,
		TotalRequests: 15420,
		ErrorRate:     2.1,
		Timestamp:     time.Now(),
	}
}

func (tw *ThreatWorker) getRecentAuditLogs() []database.AuditLog {
	// Simulate getting recent audit logs
	return []database.AuditLog{
		{UserID: 1, Action: "login_failed", Resource: "/auth/login", Timestamp: time.Now().Add(-1 * time.Hour)},
		{UserID: 2, Action: "login_failed", Resource: "/auth/login", Timestamp: time.Now().Add(-30 * time.Minute)},
		{UserID: 1, Action: "login_failed", Resource: "/auth/login", Timestamp: time.Now().Add(-15 * time.Minute)},
		{UserID: 2, Action: "login_failed", Resource: "/auth/login", Timestamp: time.Now().Add(-10 * time.Minute)},
		{UserID: 1, Action: "login_failed", Resource: "/auth/login", Timestamp: time.Now().Add(-5 * time.Minute)},
		{UserID: 2, Action: "login_failed", Resource: "/auth/login", Timestamp: time.Now().Add(-2 * time.Minute)},
	}
}

func (tw *ThreatWorker) fetchThreatIntelligence(feed string) ThreatIntelligence {
	// Simulate fetching threat intelligence
	return ThreatIntelligence{
		Source: feed,
		Threats: []ThreatIndicator{
			{Type: "ip", Value: "192.168.1.100", Severity: "high", Description: "Known malicious IP", Confidence: 0.9},
			{Type: "domain", Value: "malicious-site.com", Severity: "critical", Description: "C&C domain", Confidence: 0.95},
			{Type: "hash", Value: "a1b2c3d4e5f6", Severity: "medium", Description: "Known malware hash", Confidence: 0.8},
		},
		Confidence: 0.85,
		Timestamp:  time.Now(),
	}
}

func (tw *ThreatWorker) isThreatRelevant(indicator ThreatIndicator) bool {
	// Check if threat is relevant to current environment
	// This would include IP whitelisting, domain filtering, etc.
	return indicator.Confidence > 0.8
}

func (tw *ThreatWorker) updateThreatPatterns(threats []database.Threat) {
	// Update patterns based on detected threats
	for _, threat := range threats {
		if threat.Source == "pattern_detection" {
			// Extract new patterns from detected threats
			newPattern := tw.extractPatternFromThreat(threat)
			if newPattern != nil {
				tw.learningData[newPattern.Pattern] = newPattern
				tw.patterns = append(tw.patterns, newPattern)
			}
		}
	}
}

func (tw *ThreatWorker) extractPatternFromThreat(threat database.Threat) *ThreatPattern {
	// Extract new patterns from detected threats
	return &ThreatPattern{
		ID:         len(tw.patterns) + 1,
		Pattern:    threat.Description,
		Severity:   threat.Severity,
		Category:   "auto_generated",
		Confidence: 0.8,
		Count:      1,
		LastSeen:   time.Now(),
		Indicators: map[string]interface{}{"source": threat.Source},
		Actions:    []string{"investigate", "block"},
	}
}

func (tw *ThreatWorker) storeDetectedThreats(threats []database.Threat) error {
	sqlDB, ok := tw.Database.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	for _, threat := range threats {
		query := `
			INSERT INTO threats (title, description, severity, status, source, target, detected_at, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err := sqlDB.Exec(query, threat.Title, threat.Description, threat.Severity,
			threat.Status, threat.Source, threat.Target, threat.DetectedAt, threat.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to store threat: %w", err)
		}
	}

	return nil
}

func (tw *ThreatWorker) triggerAutomatedResponses(threats []database.Threat) {
	for _, threat := range threats {
		switch threat.Severity {
		case "critical":
			// Trigger immediate automated response
			tw.executeAutomatedResponse(threat, "immediate")
		case "high":
			// Trigger delayed automated response
			go func() {
				time.Sleep(5 * time.Minute)
				tw.executeAutomatedResponse(threat, "delayed")
			}()
		}
	}
}

func (tw *ThreatWorker) executeAutomatedResponse(threat database.Threat, responseType string) {
	log.Printf("Executing %s automated response for threat: %s", responseType, threat.Title)

	// Implement automated response logic
	switch threat.Source {
	case "malicious_ip":
		tw.blockIP(threat.Target)
	case "malware_detected":
		tw.quarantineFile(threat.Target)
	case "brute_force":
		tw.lockAccount(threat.Target)
	default:
		tw.createAlert(threat)
	}
}

func (tw *ThreatWorker) blockIP(ip string) {
	// Implement IP blocking logic
	log.Printf("Blocking malicious IP: %s", ip)
}

func (tw *ThreatWorker) quarantineFile(path string) {
	// Implement file quarantine logic
	log.Printf("Quarantining malicious file: %s", path)
}

func (tw *ThreatWorker) lockAccount(account string) {
	// Implement account locking logic
	log.Printf("Locking account due to suspicious activity: %s", account)
}

func (tw *ThreatWorker) createAlert(threat database.Threat) {
	// Implement alert creation logic
	log.Printf("Creating security alert for threat: %s", threat.Title)
}
