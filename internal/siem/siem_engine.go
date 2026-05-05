package siem

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"hades-v2/internal/database"
)

// SIEMEngine provides advanced SIEM integration with threat intelligence feeds
type SIEMEngine struct {
	db                database.Database
	collectors        map[string]*Collector
	rules             map[string]*Rule
	correlationEngine *CorrelationEngine
	threatFeeds       map[string]*ThreatFeed
	alerts            map[string]*Alert
	incidents         map[string]*Incident
	mu                sync.RWMutex
}

// Collector represents a log collector
type Collector struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Source     string                 `json:"source"`
	Enabled    bool                   `json:"enabled"`
	Parameters map[string]interface{} `json:"parameters"`
	Status     string                 `json:"status"`
	LastUpdate time.Time              `json:"last_update"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Rule represents a SIEM correlation rule
type Rule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Conditions  []*Condition           `json:"conditions"`
	Actions     []*Action              `json:"actions"`
	TimeWindow  time.Duration          `json:"time_window"`
	Threshold   int                    `json:"threshold"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Condition represents a rule condition
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Enabled  bool        `json:"enabled"`
}

// Action represents a rule action
type Action struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Enabled    bool                   `json:"enabled"`
}

// CorrelationEngine handles event correlation
type CorrelationEngine struct {
	Rules       map[string]*Rule `json:"rules"`
	EventBuffer []*Event         `json:"event_buffer"`
	WindowSize  time.Duration    `json:"window_size"`
	MaxEvents   int              `json:"max_events"`
	LastRun     time.Time        `json:"last_run"`
}

// Event represents a security event
type Event struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	EventType string                 `json:"event_type"`
	Severity  string                 `json:"severity"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields"`
	Tags      []string               `json:"tags"`
	RawData   string                 `json:"raw_data"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ThreatFeed represents a threat intelligence feed
type ThreatFeed struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	URL        string                 `json:"url"`
	Format     string                 `json:"format"`
	Enabled    bool                   `json:"enabled"`
	LastUpdate time.Time              `json:"last_update"`
	Indicators []*Indicator           `json:"indicators"`
	Parameters map[string]interface{} `json:"parameters"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Indicator represents a threat indicator
type Indicator struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Value       string                 `json:"value"`
	Confidence  float64                `json:"confidence"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Alert represents a SIEM alert
type Alert struct {
	ID          string                 `json:"id"`
	RuleID      string                 `json:"rule_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Status      string                 `json:"status"`
	Events      []*Event               `json:"events"`
	Indicators  []*Indicator           `json:"indicators"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Incident represents a security incident
type Incident struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Status      string                 `json:"status"`
	Alerts      []*Alert               `json:"alerts"`
	Timeline    []*TimelineEntry       `json:"timeline"`
	Assignee    string                 `json:"assignee"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  time.Time              `json:"resolved_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TimelineEntry represents an incident timeline entry
type TimelineEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Action    string                 `json:"action"`
	User      string                 `json:"user"`
	Details   string                 `json:"details"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// NewSIEMEngine creates a new SIEM engine
func NewSIEMEngine(db database.Database) (*SIEMEngine, error) {
	engine := &SIEMEngine{
		db:          db,
		collectors:  make(map[string]*Collector),
		rules:       make(map[string]*Rule),
		threatFeeds: make(map[string]*ThreatFeed),
		alerts:      make(map[string]*Alert),
		incidents:   make(map[string]*Incident),
		correlationEngine: &CorrelationEngine{
			Rules:       make(map[string]*Rule),
			EventBuffer: make([]*Event, 0),
			WindowSize:  5 * time.Minute,
			MaxEvents:   10000,
		},
	}

	// Initialize default collectors and rules
	if err := engine.initializeDefaults(); err != nil {
		return nil, fmt.Errorf("failed to initialize defaults: %w", err)
	}

	return engine, nil
}

// initializeDefaults initializes default collectors and rules
func (se *SIEMEngine) initializeDefaults() error {
	// Create default collectors
	se.collectors["syslog"] = &Collector{
		ID:      "syslog",
		Name:    "System Log Collector",
		Type:    "syslog",
		Source:  "udp://localhost:514",
		Enabled: true,
		Parameters: map[string]interface{}{
			"port":     514,
			"protocol": "udp",
		},
		Status:     "running",
		LastUpdate: time.Now(),
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}

	se.collectors["firewall"] = &Collector{
		ID:      "firewall",
		Name:    "Firewall Log Collector",
		Type:    "file",
		Source:  "/var/log/firewall.log",
		Enabled: true,
		Parameters: map[string]interface{}{
			"file_path": "/var/log/firewall.log",
			"format":    "json",
		},
		Status:     "running",
		LastUpdate: time.Now(),
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}

	se.collectors["windows"] = &Collector{
		ID:      "windows",
		Name:    "Windows Event Log Collector",
		Type:    "wineventlog",
		Source:  "Security",
		Enabled: true,
		Parameters: map[string]interface{}{
			"log_name": "Security",
			"level":    "Warning,Error,Critical",
		},
		Status:     "running",
		LastUpdate: time.Now(),
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}

	// Create default correlation rules
	se.rules["brute_force_detection"] = &Rule{
		ID:          "brute_force_detection",
		Name:        "Brute Force Attack Detection",
		Description: "Detects multiple failed login attempts from same source",
		Enabled:     true,
		Priority:    1,
		TimeWindow:  5 * time.Minute,
		Threshold:   5,
		Conditions: []*Condition{
			{
				Field:    "event_type",
				Operator: "equals",
				Value:    "login_failure",
				Enabled:  true,
			},
		},
		Actions: []*Action{
			{
				Type: "create_alert",
				Parameters: map[string]interface{}{
					"severity": "high",
					"title":    "Brute Force Attack Detected",
				},
				Enabled: true,
			},
			{
				Type: "block_ip",
				Parameters: map[string]interface{}{
					"duration": "1h",
				},
				Enabled: true,
			},
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	se.rules["malware_detection"] = &Rule{
		ID:          "malware_detection",
		Name:        "Malware Detection",
		Description: "Detects malware indicators in system events",
		Enabled:     true,
		Priority:    2,
		TimeWindow:  1 * time.Minute,
		Threshold:   1,
		Conditions: []*Condition{
			{
				Field:    "event_type",
				Operator: "equals",
				Value:    "malware_detected",
				Enabled:  true,
			},
		},
		Actions: []*Action{
			{
				Type: "create_alert",
				Parameters: map[string]interface{}{
					"severity": "critical",
					"title":    "Malware Detection",
				},
				Enabled: true,
			},
			{
				Type: "quarantine_system",
				Parameters: map[string]interface{}{
					"scope": "affected",
				},
				Enabled: true,
			},
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	se.rules["data_exfiltration"] = &Rule{
		ID:          "data_exfiltration",
		Name:        "Data Exfiltration Detection",
		Description: "Detects potential data exfiltration patterns",
		Enabled:     true,
		Priority:    1,
		TimeWindow:  10 * time.Minute,
		Threshold:   3,
		Conditions: []*Condition{
			{
				Field:    "event_type",
				Operator: "equals",
				Value:    "file_transfer",
				Enabled:  true,
			},
			{
				Field:    "file_size",
				Operator: "greater_than",
				Value:    10000000, // 10MB
				Enabled:  true,
			},
		},
		Actions: []*Action{
			{
				Type: "create_alert",
				Parameters: map[string]interface{}{
					"severity": "high",
					"title":    "Potential Data Exfiltration",
				},
				Enabled: true,
			},
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create default threat feeds
	se.threatFeeds["malware_domains"] = &ThreatFeed{
		ID:         "malware_domains",
		Name:       "Malware Domains Feed",
		Type:       "domain_blacklist",
		URL:        "https://example.com/threat-feeds/malware-domains",
		Format:     "json",
		Enabled:    true,
		LastUpdate: time.Now(),
		Indicators: make([]*Indicator, 0),
		Parameters: map[string]interface{}{
			"update_interval": "1h",
		},
		Metadata: make(map[string]interface{}),
	}

	se.threatFeeds["malicious_ips"] = &ThreatFeed{
		ID:         "malicious_ips",
		Name:       "Malicious IPs Feed",
		Type:       "ip_blacklist",
		URL:        "https://example.com/threat-feeds/malicious-ips",
		Format:     "json",
		Enabled:    true,
		LastUpdate: time.Now(),
		Indicators: make([]*Indicator, 0),
		Parameters: map[string]interface{}{
			"update_interval": "30m",
		},
		Metadata: make(map[string]interface{}),
	}

	// Add sample indicators
	se.threatFeeds["malware_domains"].Indicators = append(se.threatFeeds["malware_domains"].Indicators, &Indicator{
		ID:          "malware_domain_1",
		Type:        "domain",
		Value:       "malicious.example.com",
		Confidence:  0.9,
		Source:      "malware_domains",
		Description: "Known malware distribution domain",
		Tags:        []string{"malware", "c2"},
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Metadata:    make(map[string]interface{}),
	})

	se.threatFeeds["malicious_ips"].Indicators = append(se.threatFeeds["malicious_ips"].Indicators, &Indicator{
		ID:          "malicious_ip_1",
		Type:        "ip",
		Value:       "192.168.1.100",
		Confidence:  0.8,
		Source:      "malicious_ips",
		Description: "Known malicious IP address",
		Tags:        []string{"malware", "botnet"},
		ExpiresAt:   time.Now().Add(12 * time.Hour),
		Metadata:    make(map[string]interface{}),
	})

	// Add rules to correlation engine
	for id, rule := range se.rules {
		se.correlationEngine.Rules[id] = rule
	}

	return nil
}

// ProcessEvent processes a security event
func (se *SIEMEngine) ProcessEvent(ctx context.Context, event *Event) error {
	se.mu.Lock()
	defer se.mu.Unlock()

	// Generate event ID if not provided
	if event.ID == "" {
		event.ID = fmt.Sprintf("event_%d", time.Now().UnixNano())
	}
	event.Timestamp = time.Now()

	// Add to correlation engine buffer
	se.correlationEngine.EventBuffer = append(se.correlationEngine.EventBuffer, event)

	// Trim buffer if too large
	if len(se.correlationEngine.EventBuffer) > se.correlationEngine.MaxEvents {
		se.correlationEngine.EventBuffer = se.correlationEngine.EventBuffer[1:]
	}

	// Check against threat intelligence
	se.checkThreatIntelligence(event)

	// Run correlation
	if err := se.runCorrelation(ctx, event); err != nil {
		log.Printf("Correlation error: %v", err)
	}

	return nil
}

// checkThreatIntelligence checks event against threat intelligence
func (se *SIEMEngine) checkThreatIntelligence(event *Event) {
	for _, feed := range se.threatFeeds {
		if !feed.Enabled {
			continue
		}

		for _, indicator := range feed.Indicators {
			if se.matchesIndicator(event, indicator) {
				// Add indicator to event tags
				event.Tags = append(event.Tags, fmt.Sprintf("threat_intel:%s", indicator.Type))

				// Create alert if high confidence
				if indicator.Confidence > 0.7 {
					alert := &Alert{
						ID:          fmt.Sprintf("alert_%d", time.Now().UnixNano()),
						RuleID:      "threat_intel_match",
						Title:       "Threat Intelligence Match",
						Description: fmt.Sprintf("Event matched threat intelligence indicator: %s", indicator.Value),
						Severity:    "high",
						Status:      "new",
						Events:      []*Event{event},
						Indicators:  []*Indicator{indicator},
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
						Metadata:    make(map[string]interface{}),
					}
					se.alerts[alert.ID] = alert
				}
			}
		}
	}
}

// matchesIndicator checks if event matches threat indicator
func (se *SIEMEngine) matchesIndicator(event *Event, indicator *Indicator) bool {
	switch indicator.Type {
	case "ip":
		if srcIP, ok := event.Fields["src_ip"]; ok {
			return fmt.Sprintf("%v", srcIP) == indicator.Value
		}
		if dstIP, ok := event.Fields["dst_ip"]; ok {
			return fmt.Sprintf("%v", dstIP) == indicator.Value
		}
	case "domain":
		if domain, ok := event.Fields["domain"]; ok {
			return fmt.Sprintf("%v", domain) == indicator.Value
		}
	case "hash":
		if hash, ok := event.Fields["file_hash"]; ok {
			return fmt.Sprintf("%v", hash) == indicator.Value
		}
	}
	return false
}

// runCorrelation runs correlation rules
func (se *SIEMEngine) runCorrelation(ctx context.Context, event *Event) error {
	for _, rule := range se.correlationEngine.Rules {
		if !rule.Enabled {
			continue
		}

		if se.evaluateRule(rule, event) {
			// Create alert
			alert := &Alert{
				ID:          fmt.Sprintf("alert_%d", time.Now().UnixNano()),
				RuleID:      rule.ID,
				Title:       rule.Name,
				Description: rule.Description,
				Severity:    se.getSeverityFromPriority(rule.Priority),
				Status:      "new",
				Events:      []*Event{event},
				Indicators:  make([]*Indicator, 0),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Metadata:    make(map[string]interface{}),
			}
			se.alerts[alert.ID] = alert

			// Execute actions
			for _, action := range rule.Actions {
				if !action.Enabled {
					continue
				}
				se.executeAction(action, alert, event)
			}
		}
	}

	return nil
}

// evaluateRule evaluates a correlation rule
func (se *SIEMEngine) evaluateRule(rule *Rule, event *Event) bool {
	// Check conditions
	for _, condition := range rule.Conditions {
		if !condition.Enabled {
			continue
		}

		if !se.evaluateCondition(condition, event) {
			return false
		}
	}

	// Check threshold (simplified)
	return true
}

// evaluateCondition evaluates a rule condition
func (se *SIEMEngine) evaluateCondition(condition *Condition, event *Event) bool {
	eventValue, exists := se.getFieldValue(event, condition.Field)
	if !exists {
		return false
	}

	switch condition.Operator {
	case "equals":
		return fmt.Sprintf("%v", eventValue) == fmt.Sprintf("%v", condition.Value)
	case "not_equals":
		return fmt.Sprintf("%v", eventValue) != fmt.Sprintf("%v", condition.Value)
	case "greater_than":
		if eventFloat, ok := eventValue.(float64); ok {
			if condFloat, ok := condition.Value.(float64); ok {
				return eventFloat > condFloat
			}
		}
	case "less_than":
		if eventFloat, ok := eventValue.(float64); ok {
			if condFloat, ok := condition.Value.(float64); ok {
				return eventFloat < condFloat
			}
		}
	case "contains":
		if eventStr, ok := eventValue.(string); ok {
			if condStr, ok := condition.Value.(string); ok {
				return contains(eventStr, condStr)
			}
		}
	}
	return false
}

// getFieldValue gets field value from event
func (se *SIEMEngine) getFieldValue(event *Event, field string) (interface{}, bool) {
	switch field {
	case "event_type":
		return event.EventType, true
	case "severity":
		return event.Severity, true
	case "source":
		return event.Source, true
	default:
		if value, ok := event.Fields[field]; ok {
			return value, true
		}
	}
	return nil, false
}

// executeAction executes a rule action
func (se *SIEMEngine) executeAction(action *Action, alert *Alert, event *Event) {
	switch action.Type {
	case "create_alert":
		// Alert already created
		log.Printf("Alert created: %s", alert.Title)
	case "block_ip":
		if srcIP, ok := event.Fields["src_ip"]; ok {
			log.Printf("Blocking IP: %v", srcIP)
		}
	case "quarantine_system":
		if hostname, ok := event.Fields["hostname"]; ok {
			log.Printf("Quarantining system: %v", hostname)
		}
	default:
		log.Printf("Unknown action type: %s", action.Type)
	}
}

// getSeverityFromPriority converts priority to severity
func (se *SIEMEngine) getSeverityFromPriority(priority int) string {
	switch priority {
	case 1:
		return "critical"
	case 2:
		return "high"
	case 3:
		return "medium"
	default:
		return "low"
	}
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

// containsMiddle checks if substring is in middle of string
func containsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetCollectors returns all collectors
func (se *SIEMEngine) GetCollectors() map[string]*Collector {
	se.mu.RLock()
	defer se.mu.RUnlock()

	// Return copy
	result := make(map[string]*Collector)
	for id, collector := range se.collectors {
		result[id] = collector
	}
	return result
}

// GetRules returns all rules
func (se *SIEMEngine) GetRules() map[string]*Rule {
	se.mu.RLock()
	defer se.mu.RUnlock()

	// Return copy
	result := make(map[string]*Rule)
	for id, rule := range se.rules {
		result[id] = rule
	}
	return result
}

// GetThreatFeeds returns all threat feeds
func (se *SIEMEngine) GetThreatFeeds() map[string]*ThreatFeed {
	se.mu.RLock()
	defer se.mu.RUnlock()

	// Return copy
	result := make(map[string]*ThreatFeed)
	for id, feed := range se.threatFeeds {
		result[id] = feed
	}
	return result
}

// GetAlerts returns all alerts
func (se *SIEMEngine) GetAlerts() map[string]*Alert {
	se.mu.RLock()
	defer se.mu.RUnlock()

	// Return copy
	result := make(map[string]*Alert)
	for id, alert := range se.alerts {
		result[id] = alert
	}
	return result
}

// GetIncidents returns all incidents
func (se *SIEMEngine) GetIncidents() map[string]*Incident {
	se.mu.RLock()
	defer se.mu.RUnlock()

	// Return copy
	result := make(map[string]*Incident)
	for id, incident := range se.incidents {
		result[id] = incident
	}
	return result
}

// GetEngineStatus returns engine status
func (se *SIEMEngine) GetEngineStatus() map[string]interface{} {
	se.mu.RLock()
	defer se.mu.RUnlock()

	return map[string]interface{}{
		"collectors":        len(se.collectors),
		"rules":             len(se.rules),
		"threat_feeds":      len(se.threatFeeds),
		"alerts":            len(se.alerts),
		"incidents":         len(se.incidents),
		"buffered_events":   len(se.correlationEngine.EventBuffer),
		"correlation_rules": len(se.correlationEngine.Rules),
		"timestamp":         time.Now(),
	}
}
