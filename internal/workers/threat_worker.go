package workers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

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
			if err := json.Unmarshal([]byte(indicatorsJSON), &pattern.Indicators); err != nil {
				log.Printf("Warning: failed to unmarshal indicators JSON: %v", err)
			}
		}
		if actionsJSON != "" {
			if err := json.Unmarshal([]byte(actionsJSON), &pattern.Actions); err != nil {
				log.Printf("Warning: failed to unmarshal actions JSON: %v", err)
			}
		}

		tw.patterns = append(tw.patterns, &pattern)
		tw.learningData[pattern.Pattern] = &pattern
	}

	log.Printf("Loaded %d threat patterns for worker %s", len(tw.patterns), tw.Name)
	return nil
}

type DataSource struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Target  string `json:"target"`
}
