package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"hades-v2/internal/database"
)

// AuditLogger manages comprehensive audit logging
type AuditLogger struct {
	db            database.Database
	loggers       map[string]LogWriter
	buffer        chan AuditEvent
	filters       map[string]AuditFilter
	retention     RetentionPolicy
	compression   CompressionEngine
	encryption    EncryptionEngine
	mu            sync.RWMutex
	enabled       bool
	bufferSize    int
	flushInterval time.Duration
}

// AuditEvent represents an audit event
type AuditEvent struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	EventType    string                 `json:"event_type"`
	Category     string                 `json:"category"`
	Severity     string                 `json:"severity"` // "info", "warning", "error", "critical"
	UserID       string                 `json:"user_id"`
	SessionID    string                 `json:"session_id"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Resource     string                 `json:"resource"`
	Action       string                 `json:"action"`
	Description  string                 `json:"description"`
	Outcome      string                 `json:"outcome"` // "success", "failure", "partial"
	Details      map[string]interface{} `json:"details"`
	Metadata     map[string]interface{} `json:"metadata"`
	Hash         string                 `json:"hash"`
	Signature    string                 `json:"signature"`
	ChainID      string                 `json:"chain_id"` // For audit trail chaining
	PreviousHash string                 `json:"previous_hash"`
}

// LogWriter interface for different log destinations
type LogWriter interface {
	Name() string
	Write(ctx context.Context, event AuditEvent) error
	Flush(ctx context.Context) error
	Close() error
	HealthCheck() error
}

// AuditFilter represents an audit log filter
type AuditFilter struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Enabled     bool           `json:"enabled"`
	Rules       []FilterRule   `json:"rules"`
	Actions     []FilterAction `json:"actions"`
	Created     time.Time      `json:"created"`
	Updated     time.Time      `json:"updated"`
}

// FilterRule represents a filtering rule
type FilterRule struct {
	Field         string      `json:"field"`
	Operator      string      `json:"operator"` // "equals", "contains", "regex", "greater_than", "less_than"
	Value         interface{} `json:"value"`
	Negate        bool        `json:"negate"`
	CaseSensitive bool        `json:"case_sensitive"`
}

// FilterAction represents an action to take when filter matches
type FilterAction struct {
	Type       string                 `json:"type"` // "alert", "block", "log", "forward", "transform"
	Parameters map[string]interface{} `json:"parameters"`
	Enabled    bool                   `json:"enabled"`
}

// RetentionPolicy manages audit log retention
type RetentionPolicy struct {
	DefaultRetention   time.Duration            `json:"default_retention"`
	CategoryRetention  map[string]time.Duration `json:"category_retention"`
	MaxStorageSize     int64                    `json:"max_storage_size"`
	CompressionEnabled bool                     `json:"compression_enabled"`
	ArchivalEnabled    bool                     `json:"archival_enabled"`
	ArchivalLocation   string                   `json:"archival_location"`
}

// CompressionEngine handles log compression
type CompressionEngine struct {
	Algorithm string `json:"algorithm"`
	Level     int    `json:"level"`
	Enabled   bool   `json:"enabled"`
}

// EncryptionEngine handles log encryption
type EncryptionEngine struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id"`
	Enabled   bool   `json:"enabled"`
}

// AuditQuery represents a query for audit logs
type AuditQuery struct {
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	EventTypes []string               `json:"event_types"`
	Categories []string               `json:"categories"`
	Users      []string               `json:"users"`
	Resources  []string               `json:"resources"`
	Severity   []string               `json:"severity"`
	Outcomes   []string               `json:"outcomes"`
	Keywords   string                 `json:"keywords"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
	SortBy     string                 `json:"sort_by"`
	SortOrder  string                 `json:"sort_order"`
	Filters    map[string]interface{} `json:"filters"`
}

// AuditResult represents the result of an audit query
type AuditResult struct {
	Events      []AuditEvent           `json:"events"`
	TotalCount  int                    `json:"total_count"`
	PageCount   int                    `json:"page_count"`
	CurrentPage int                    `json:"current_page"`
	HasNext     bool                   `json:"has_next"`
	HasPrev     bool                   `json:"has_prev"`
	QueryTime   time.Duration          `json:"query_time"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db database.Database) *AuditLogger {
	al := &AuditLogger{
		db:            db,
		loggers:       make(map[string]LogWriter),
		buffer:        make(chan AuditEvent, 10000),
		filters:       make(map[string]AuditFilter),
		enabled:       true,
		bufferSize:    10000,
		flushInterval: 5 * time.Second,
		retention: RetentionPolicy{
			DefaultRetention:   365 * 24 * time.Hour, // 1 year
			CategoryRetention:  make(map[string]time.Duration),
			MaxStorageSize:     10 * 1024 * 1024 * 1024, // 10GB
			CompressionEnabled: true,
			ArchivalEnabled:    true,
			ArchivalLocation:   "/var/log/hades/archive/",
		},
		compression: CompressionEngine{
			Algorithm: "gzip",
			Level:     6,
			Enabled:   true,
		},
		encryption: EncryptionEngine{
			Algorithm: "AES-256-GCM",
			KeyID:     "audit-key-1",
			Enabled:   true,
		},
	}

	// Initialize default loggers
	al.initializeDefaultLoggers()

	// Initialize default filters
	al.initializeDefaultFilters()

	return al
}

// initializeDefaultLogners initializes default log writers
func (al *AuditLogger) initializeDefaultLoggers() {
	// Database logger
	al.loggers["database"] = &DatabaseLogger{
		db:     al.db,
		table:  "audit_logs",
		buffer: make([]AuditEvent, 0, 1000),
	}

	// File logger
	al.loggers["file"] = &FileLogger{
		filePath: "/var/log/hades/audit.log",
		rotation: true,
		maxSize:  100 * 1024 * 1024, // 100MB
	}

	// Syslog logger
	al.loggers["syslog"] = &SyslogLogger{
		facility: "local0",
		tag:      "hades-audit",
	}

	// Remote logger (for SIEM integration)
	al.loggers["remote"] = &RemoteLogger{
		endpoint: "https://siem.company.com/api/logs",
		apiKey:   "audit-api-key",
	}
}

// initializeDefaultFilters initializes default audit filters
func (al *AuditLogger) initializeDefaultFilters() {
	defaultFilters := []*AuditFilter{
		{
			ID:          "security_events",
			Name:        "Security Events Filter",
			Description: "Filters and alerts on security-related events",
			Enabled:     true,
			Rules: []FilterRule{
				{
					Field:         "category",
					Operator:      "equals",
					Value:         "security",
					Negate:        false,
					CaseSensitive: false,
				},
			},
			Actions: []FilterAction{
				{
					Type: "alert",
					Parameters: map[string]interface{}{
						"channel":  "security-team",
						"priority": "high",
					},
					Enabled: true,
				},
			},
			Created: time.Now(),
			Updated: time.Now(),
		},
		{
			ID:          "failed_logins",
			Name:        "Failed Login Filter",
			Description: "Detects and alerts on failed login attempts",
			Enabled:     true,
			Rules: []FilterRule{
				{
					Field:         "event_type",
					Operator:      "equals",
					Value:         "login_attempt",
					Negate:        false,
					CaseSensitive: false,
				},
				{
					Field:         "outcome",
					Operator:      "equals",
					Value:         "failure",
					Negate:        false,
					CaseSensitive: false,
				},
			},
			Actions: []FilterAction{
				{
					Type: "alert",
					Parameters: map[string]interface{}{
						"channel":  "security-team",
						"priority": "medium",
					},
					Enabled: true,
				},
			},
			Created: time.Now(),
			Updated: time.Now(),
		},
		{
			ID:          "privilege_escalation",
			Name:        "Privilege Escalation Filter",
			Description: "Detects privilege escalation attempts",
			Enabled:     true,
			Rules: []FilterRule{
				{
					Field:         "event_type",
					Operator:      "equals",
					Value:         "privilege_change",
					Negate:        false,
					CaseSensitive: false,
				},
			},
			Actions: []FilterAction{
				{
					Type: "alert",
					Parameters: map[string]interface{}{
						"channel":  "security-team",
						"priority": "critical",
					},
					Enabled: true,
				},
				{
					Type: "block",
					Parameters: map[string]interface{}{
						"duration": "1h",
					},
					Enabled: true,
				},
			},
			Created: time.Now(),
			Updated: time.Now(),
		},
	}

	for _, filter := range defaultFilters {
		al.filters[filter.ID] = *filter
	}
}

// Start starts the audit logger
func (al *AuditLogger) Start(ctx context.Context) error {
	if !al.enabled {
		return fmt.Errorf("audit logger is disabled")
	}

	// Start log writers
	for name, logger := range al.loggers {
		if err := logger.HealthCheck(); err != nil {
			log.Printf("Logger %s health check failed: %v", name, err)
		}
	}

	// Start buffer processor
	go al.processBuffer(ctx)

	// Start periodic flush
	go al.periodicFlush(ctx)

	// Start retention manager
	go al.retentionManager(ctx)

	log.Println("Audit logger started")
	return nil
}

// Stop stops the audit logger
func (al *AuditLogger) Stop() error {
	al.enabled = false
	close(al.buffer)

	// Close all loggers
	for name, logger := range al.loggers {
		if err := logger.Close(); err != nil {
			log.Printf("Error closing logger %s: %v", name, err)
		}
	}

	log.Println("Audit logger stopped")
	return nil
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(ctx context.Context, event AuditEvent) error {
	if !al.enabled {
		return fmt.Errorf("audit logger is disabled")
	}

	// Generate event ID and hash
	if event.ID == "" {
		event.ID = fmt.Sprintf("audit_%d", time.Now().UnixNano())
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Generate hash for integrity
	event.Hash = al.generateEventHash(event)

	// Apply filters
	filtered, err := al.applyFilters(event)
	if err != nil {
		log.Printf("Error applying filters: %v", err)
	}

	if !filtered {
		return nil // Event was filtered out
	}

	// Add to buffer
	select {
	case al.buffer <- event:
	default:
		log.Printf("Audit buffer full, dropping event %s", event.ID)
		return fmt.Errorf("audit buffer full")
	}

	return nil
}

// generateEventHash generates a hash for the event
func (al *AuditLogger) generateEventHash(event AuditEvent) string {
	data := fmt.Sprintf("%d|%s|%s|%s|%s|%s|%s",
		event.Timestamp.UnixNano(),
		event.EventType,
		event.Category,
		event.UserID,
		event.Resource,
		event.Action,
		event.Outcome,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// applyFilters applies audit filters to an event
func (al *AuditLogger) applyFilters(event AuditEvent) (bool, error) {
	al.mu.RLock()
	defer al.mu.RUnlock()

	for _, filter := range al.filters {
		if !filter.Enabled {
			continue
		}

		matches, err := al.evaluateFilter(filter, event)
		if err != nil {
			return false, err
		}

		if matches {
			// Execute filter actions
			for _, action := range filter.Actions {
				if !action.Enabled {
					continue
				}

				err := al.executeFilterAction(action, event)
				if err != nil {
					log.Printf("Error executing filter action: %v", err)
				}
			}
		}
	}

	return true, nil
}

// evaluateFilter evaluates a filter against an event
func (al *AuditLogger) evaluateFilter(filter AuditFilter, event AuditEvent) (bool, error) {
	for _, rule := range filter.Rules {
		matches, err := al.evaluateRule(rule, event)
		if err != nil {
			return false, err
		}

		if rule.Negate {
			matches = !matches
		}

		if !matches {
			return false, nil // All rules must match (AND logic)
		}
	}

	return true, nil
}

// evaluateRule evaluates a single filter rule
func (al *AuditLogger) evaluateRule(rule FilterRule, event AuditEvent) (bool, error) {
	var fieldValue interface{}

	switch rule.Field {
	case "event_type":
		fieldValue = event.EventType
	case "category":
		fieldValue = event.Category
	case "severity":
		fieldValue = event.Severity
	case "user_id":
		fieldValue = event.UserID
	case "resource":
		fieldValue = event.Resource
	case "action":
		fieldValue = event.Action
	case "outcome":
		fieldValue = event.Outcome
	case "ip_address":
		fieldValue = event.IPAddress
	default:
		return false, fmt.Errorf("unknown field: %s", rule.Field)
	}

	return al.evaluateCondition(rule.Operator, fieldValue, rule.Value, rule.CaseSensitive)
}

// evaluateCondition evaluates a condition
func (al *AuditLogger) evaluateCondition(operator string, fieldValue, conditionValue interface{}, caseSensitive bool) (bool, error) {
	switch operator {
	case "equals":
		return al.compareEquals(fieldValue, conditionValue, caseSensitive)
	case "contains":
		return al.compareContains(fieldValue, conditionValue, caseSensitive)
	case "greater_than":
		return al.compareGreaterThan(fieldValue, conditionValue)
	case "less_than":
		return al.compareLessThan(fieldValue, conditionValue)
	case "regex":
		return al.compareRegex(fieldValue, conditionValue, caseSensitive)
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

// compareEquals compares values for equality
func (al *AuditLogger) compareEquals(fieldValue, conditionValue interface{}, caseSensitive bool) (bool, error) {
	fieldStr, fieldOk := fieldValue.(string)
	conditionStr, conditionOk := conditionValue.(string)

	// Guard clause: if not both strings, use direct comparison
	if !fieldOk || !conditionOk {
		return fieldValue == conditionValue, nil
	}

	// Guard clause: case-insensitive comparison
	if !caseSensitive {
		return strings.EqualFold(fieldStr, conditionStr), nil
	}

	// Default: case-sensitive string comparison
	return fieldStr == conditionStr, nil
}

// compareContains checks if field contains condition
func (al *AuditLogger) compareContains(fieldValue, conditionValue interface{}, caseSensitive bool) (bool, error) {
	fieldStr, fieldOk := fieldValue.(string)
	conditionStr, conditionOk := conditionValue.(string)

	if !fieldOk || !conditionOk {
		return false, fmt.Errorf("both values must be strings for contains comparison")
	}

	if !caseSensitive {
		return strings.Contains(strings.ToLower(fieldStr), strings.ToLower(conditionStr)), nil
	}

	return strings.Contains(fieldStr, conditionStr), nil
}

// compareGreaterThan compares numeric values
func (al *AuditLogger) compareGreaterThan(fieldValue, conditionValue interface{}) (bool, error) {
	fieldFloat, fieldOk := fieldValue.(float64)
	conditionFloat, conditionOk := conditionValue.(float64)

	if fieldOk && conditionOk {
		return fieldFloat > conditionFloat, nil
	}

	fieldInt, fieldOk := fieldValue.(int)
	conditionInt, conditionOk := conditionValue.(int)

	if fieldOk && conditionOk {
		return fieldInt > conditionInt, nil
	}

	return false, fmt.Errorf("both values must be numeric for greater_than comparison")
}

// compareLessThan compares numeric values
func (al *AuditLogger) compareLessThan(fieldValue, conditionValue interface{}) (bool, error) {
	fieldFloat, fieldOk := fieldValue.(float64)
	conditionFloat, conditionOk := conditionValue.(float64)

	if fieldOk && conditionOk {
		return fieldFloat < conditionFloat, nil
	}

	fieldInt, fieldOk := fieldValue.(int)
	conditionInt, conditionOk := conditionValue.(int)

	if fieldOk && conditionOk {
		return fieldInt < conditionInt, nil
	}

	return false, fmt.Errorf("both values must be numeric for less_than comparison")
}

// compareRegex compares using regex
func (al *AuditLogger) compareRegex(fieldValue, conditionValue interface{}, caseSensitive bool) (bool, error) {
	fieldStr, fieldOk := fieldValue.(string)
	conditionStr, conditionOk := conditionValue.(string)

	if !fieldOk || !conditionOk {
		return false, fmt.Errorf("both values must be strings for regex comparison")
	}

	// In production, use actual regex matching
	if !caseSensitive {
		return strings.Contains(strings.ToLower(fieldStr), strings.ToLower(conditionStr)), nil
	}

	return strings.Contains(fieldStr, conditionStr), nil
}

// executeFilterAction executes a filter action
func (al *AuditLogger) executeFilterAction(action FilterAction, event AuditEvent) error {
	switch action.Type {
	case "alert":
		return al.executeAlertAction(action, event)
	case "block":
		return al.executeBlockAction(action, event)
	case "log":
		return al.executeLogAction(action, event)
	case "forward":
		return al.executeForwardAction(action, event)
	case "transform":
		return al.executeTransformAction(action, event)
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// executeAlertAction executes an alert action
func (al *AuditLogger) executeAlertAction(action FilterAction, event AuditEvent) error {
	// Send alert to security team using action parameters
	if priority, ok := action.Parameters["priority"].(string); ok {
		log.Printf("ALERT [%s]: %s - %s", priority, event.EventType, event.Description)
	} else {
		log.Printf("ALERT: %s - %s", event.EventType, event.Description)
	}
	return nil
}

// executeBlockAction executes a block action
func (al *AuditLogger) executeBlockAction(action FilterAction, event AuditEvent) error {
	// Block IP or user using action parameters
	if duration, ok := action.Parameters["duration"].(string); ok {
		log.Printf("BLOCK [%s]: %s - %s", duration, event.IPAddress, event.UserID)
	} else {
		log.Printf("BLOCK: %s - %s", event.IPAddress, event.UserID)
	}
	return nil
}

// executeLogAction executes a log action
func (al *AuditLogger) executeLogAction(action FilterAction, event AuditEvent) error {
	// Additional logging using action parameters
	if channel, ok := action.Parameters["channel"].(string); ok {
		log.Printf("FILTERED LOG [%s]: %s - %s", channel, event.EventType, event.Description)
	} else {
		log.Printf("FILTERED LOG: %s - %s", event.EventType, event.Description)
	}
	return nil
}

// executeForwardAction executes a forward action
func (al *AuditLogger) executeForwardAction(action FilterAction, event AuditEvent) error {
	// Forward to external system using action parameters
	if endpoint, ok := action.Parameters["endpoint"].(string); ok {
		log.Printf("FORWARD: %s to %s", event.ID, endpoint)
	} else {
		log.Printf("FORWARD: %s to external system", event.ID)
	}
	return nil
}

// executeTransformAction executes a transform action
func (al *AuditLogger) executeTransformAction(action FilterAction, event AuditEvent) error {
	// Transform event data using action parameters
	if transformType, ok := action.Parameters["type"].(string); ok {
		log.Printf("TRANSFORM [%s]: %s", transformType, event.ID)
	} else {
		log.Printf("TRANSFORM: %s", event.ID)
	}
	return nil
}

// processBuffer processes the audit event buffer
func (al *AuditLogger) processBuffer(ctx context.Context) {
	batch := make([]AuditEvent, 0, 100)

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-al.buffer:
			if !ok {
				// Channel closed, flush remaining events
				al.flushBatch(batch)
				return
			}

			batch = append(batch, event)

			// Flush batch when it reaches a certain size
			if len(batch) >= 100 {
				al.flushBatch(batch)
				batch = batch[:0] // Reset slice
			}
		}
	}
}

// flushBatch flushes a batch of events to all loggers
func (al *AuditLogger) flushBatch(batch []AuditEvent) {
	if len(batch) == 0 {
		return
	}

	for name, logger := range al.loggers {
		for _, event := range batch {
			if err := logger.Write(context.Background(), event); err != nil {
				log.Printf("Error writing to logger %s: %v", name, err)
			}
		}

		// Flush the logger
		if err := logger.Flush(context.Background()); err != nil {
			log.Printf("Error flushing logger %s: %v", name, err)
		}
	}
}

// periodicFlush periodically flushes the buffer
func (al *AuditLogger) periodicFlush(ctx context.Context) {
	ticker := time.NewTicker(al.flushInterval)
	defer ticker.Stop()

	batch := make([]AuditEvent, 0, 100)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Collect events from buffer
		InnerLoop:
			for len(batch) < 100 && len(al.buffer) > 0 {
				select {
				case event := <-al.buffer:
					batch = append(batch, event)
				default:
					break InnerLoop
				}
			}

			if len(batch) > 0 {
				al.flushBatch(batch)
				batch = batch[:0] // Reset slice
			}
		}
	}
}

// retentionManager manages log retention
func (al *AuditLogger) retentionManager(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			al.cleanupOldLogs()
		}
	}
}

// cleanupOldLogs cleans up old audit logs based on retention policy
func (al *AuditLogger) cleanupOldLogs() {
	// In production, implement actual cleanup logic
	log.Println("Running audit log cleanup")
}

// Query queries audit logs
func (al *AuditLogger) Query(ctx context.Context, query AuditQuery) (*AuditResult, error) {
	start := time.Now()

	// In production, implement actual database query
	events := make([]AuditEvent, 0)

	result := &AuditResult{
		Events:      events,
		TotalCount:  0,
		PageCount:   0,
		CurrentPage: query.Offset / query.Limit,
		HasNext:     false,
		HasPrev:     query.Offset > 0,
		QueryTime:   time.Since(start),
		Metadata:    make(map[string]interface{}),
	}

	return result, nil
}

// GetStats returns audit logging statistics
func (al *AuditLogger) GetStats() map[string]interface{} {
	al.mu.RLock()
	defer al.mu.RUnlock()

	stats := map[string]interface{}{
		"enabled":             al.enabled,
		"buffer_size":         len(al.buffer),
		"buffer_capacity":     al.bufferSize,
		"active_loggers":      len(al.loggers),
		"active_filters":      len(al.filters),
		"compression_enabled": al.compression.Enabled,
		"encryption_enabled":  al.encryption.Enabled,
		"flush_interval":      al.flushInterval.String(),
	}

	// Add logger-specific stats
	loggerStats := make(map[string]interface{})
	for name, logger := range al.loggers {
		loggerStats[name] = map[string]interface{}{
			"healthy": logger.HealthCheck() == nil,
		}
	}
	stats["loggers"] = loggerStats

	// Add filter stats
	filterStats := make(map[string]interface{})
	for id, filter := range al.filters {
		filterStats[id] = map[string]interface{}{
			"name":    filter.Name,
			"enabled": filter.Enabled,
		}
	}
	stats["filters"] = filterStats

	return stats
}

// LogWriter implementations (simplified for demonstration)

type DatabaseLogger struct {
	db     database.Database
	table  string
	buffer []AuditEvent
	mu     sync.Mutex
}

func (dl *DatabaseLogger) Name() string {
	return "Database Logger"
}

func (dl *DatabaseLogger) Write(ctx context.Context, event AuditEvent) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.buffer = append(dl.buffer, event)

	// Flush buffer if it gets too large
	if len(dl.buffer) >= 1000 {
		return dl.flush(ctx)
	}

	return nil
}

func (dl *DatabaseLogger) Flush(ctx context.Context) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	return dl.flush(ctx)
}

func (dl *DatabaseLogger) flush(ctx context.Context) error {
	if len(dl.buffer) == 0 {
		return nil
	}

	// Check if context is cancelled before flushing
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue with flush
	}

	// In production, insert events into database
	log.Printf("Flushing %d audit events to database", len(dl.buffer))
	dl.buffer = dl.buffer[:0] // Reset buffer

	return nil
}

func (dl *DatabaseLogger) Close() error {
	return dl.flush(context.Background())
}

func (dl *DatabaseLogger) HealthCheck() error {
	// In production, check database connection
	return nil
}

type FileLogger struct {
	filePath string
	rotation bool
	maxSize  int64
	mu       sync.Mutex
}

func (fl *FileLogger) Name() string {
	return "File Logger"
}

func (fl *FileLogger) Write(ctx context.Context, event AuditEvent) error {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	// In production, write to file with rotation
	log.Printf("Writing audit event to file: %s", event.ID)
	return nil
}

func (fl *FileLogger) Flush(ctx context.Context) error {
	return nil
}

func (fl *FileLogger) Close() error {
	return nil
}

func (fl *FileLogger) HealthCheck() error {
	// In production, check file permissions and disk space
	return nil
}

type SyslogLogger struct {
	facility string
	tag      string
}

func (sl *SyslogLogger) Name() string {
	return "Syslog Logger"
}

func (sl *SyslogLogger) Write(ctx context.Context, event AuditEvent) error {
	// In production, write to syslog
	log.Printf("Writing audit event to syslog: %s", event.ID)
	return nil
}

func (sl *SyslogLogger) Flush(ctx context.Context) error {
	return nil
}

func (sl *SyslogLogger) Close() error {
	return nil
}

func (sl *SyslogLogger) HealthCheck() error {
	return nil
}

type RemoteLogger struct {
	endpoint string
	apiKey   string
}

func (rl *RemoteLogger) Name() string {
	return "Remote Logger"
}

func (rl *RemoteLogger) Write(ctx context.Context, event AuditEvent) error {
	// In production, send to remote endpoint
	log.Printf("Sending audit event to remote endpoint: %s", event.ID)
	return nil
}

func (rl *RemoteLogger) Flush(ctx context.Context) error {
	return nil
}

func (rl *RemoteLogger) Close() error {
	return nil
}

func (rl *RemoteLogger) HealthCheck() error {
	// In production, check endpoint connectivity
	return nil
}
