package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ThreatLevel represents threat intelligence levels
type ThreatLevel string

const (
	ThreatLevelLow      ThreatLevel = "low"
	ThreatLevelMedium   ThreatLevel = "medium"
	ThreatLevelHigh     ThreatLevel = "high"
	ThreatLevelCritical ThreatLevel = "critical"
)

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID               string                 `json:"id"`
	CVEID            string                 `json:"cve_id"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Severity         string                 `json:"severity"`
	CVSSScore        float64                `json:"cvss_score"`
	PublishedDate    time.Time              `json:"published_date"`
	ModifiedDate     time.Time              `json:"modified_date"`
	AffectedProducts []string               `json:"affected_products"`
	References       []string               `json:"references"`
	Tags             []string               `json:"tags"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ThreatIntel represents threat intelligence data
type ThreatIntel struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Level       ThreatLevel            `json:"level"`
	Confidence  int                    `json:"confidence"`
	Source      string                 `json:"source"`
	Indicators  map[string]interface{} `json:"indicators"`
	Timestamp   time.Time              `json:"timestamp"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ThreatIntelConfig holds threat intelligence configuration
type ThreatIntelConfig struct {
	EnableCVEFeeds    bool          `json:"enable_cve_feeds"`
	EnableThreatFeeds bool          `json:"enable_threat_feeds"`
	UpdateInterval    time.Duration `json:"update_interval"`
	CacheExpiry       time.Duration `json:"cache_expiry"`
	MaxCacheSize      int           `json:"max_cache_size"`
	APITimeout        time.Duration `json:"api_timeout"`
	RetryCount        int           `json:"retry_count"`
}

// DefaultThreatIntelConfig returns sensible threat intelligence defaults
func DefaultThreatIntelConfig() *ThreatIntelConfig {
	return &ThreatIntelConfig{
		EnableCVEFeeds:    true,
		EnableThreatFeeds: true,
		UpdateInterval:    time.Hour,
		CacheExpiry:       24 * time.Hour,
		MaxCacheSize:      10000,
		APITimeout:        30 * time.Second,
		RetryCount:        3,
	}
}

// ThreatIntelManager provides vulnerability database and threat intelligence
type ThreatIntelManager struct {
	config          *ThreatIntelConfig
	vulnerabilities map[string]*Vulnerability
	threatIntel     map[string]*ThreatIntel
	mu              sync.RWMutex
	httpClient      *http.Client
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewThreatIntelManager creates a new threat intelligence manager
func NewThreatIntelManager(config *ThreatIntelConfig) *ThreatIntelManager {
	if config == nil {
		config = DefaultThreatIntelConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	tim := &ThreatIntelManager{
		config:          config,
		vulnerabilities: make(map[string]*Vulnerability),
		threatIntel:     make(map[string]*ThreatIntel),
		httpClient: &http.Client{
			Timeout: config.APITimeout,
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Start background updates
	go tim.backgroundUpdater()

	return tim
}

// SearchVulnerabilities searches for vulnerabilities by various criteria
func (tim *ThreatIntelManager) SearchVulnerabilities(ctx context.Context, query string, severity string, limit int) ([]*Vulnerability, error) {
	tim.mu.RLock()
	defer tim.mu.RUnlock()

	var results []*Vulnerability
	count := 0

	for _, vuln := range tim.vulnerabilities {
		if limit > 0 && count >= limit {
			break
		}

		// Apply filters
		if query != "" && !tim.matchesQuery(vuln, query) {
			continue
		}

		if severity != "" && vuln.Severity != severity {
			continue
		}

		results = append(results, vuln)
		count++
	}

	return results, nil
}

// GetVulnerability retrieves a specific vulnerability by CVE ID
func (tim *ThreatIntelManager) GetVulnerability(ctx context.Context, cveID string) (*Vulnerability, error) {
	tim.mu.RLock()
	defer tim.mu.RUnlock()

	vuln, exists := tim.vulnerabilities[cveID]
	if !exists {
		return nil, fmt.Errorf("hades.platform.threat_intel: vulnerability not found: %s", cveID)
	}

	return vuln, nil
}

// SearchThreatIntel searches threat intelligence data
func (tim *ThreatIntelManager) SearchThreatIntel(ctx context.Context, threatType string, level ThreatLevel, limit int) ([]*ThreatIntel, error) {
	tim.mu.RLock()
	defer tim.mu.RUnlock()

	var results []*ThreatIntel
	count := 0

	for _, intel := range tim.threatIntel {
		if limit > 0 && count >= limit {
			break
		}

		// Apply filters
		if threatType != "" && intel.Type != threatType {
			continue
		}

		if level != "" && intel.Level != level {
			continue
		}

		results = append(results, intel)
		count++
	}

	return results, nil
}

// GetThreatIntel retrieves specific threat intelligence
func (tim *ThreatIntelManager) GetThreatIntel(ctx context.Context, id string) (*ThreatIntel, error) {
	tim.mu.RLock()
	defer tim.mu.RUnlock()

	intel, exists := tim.threatIntel[id]
	if !exists {
		return nil, fmt.Errorf("hades.platform.threat_intel: threat intel not found: %s", id)
	}

	return intel, nil
}

// AddVulnerability adds a new vulnerability to the database
func (tim *ThreatIntelManager) AddVulnerability(vuln *Vulnerability) error {
	tim.mu.Lock()
	defer tim.mu.Unlock()

	if vuln.CVEID == "" {
		return fmt.Errorf("hades.platform.threat_intel: CVE ID cannot be empty")
	}

	tim.vulnerabilities[vuln.CVEID] = vuln
	return nil
}

// AddThreatIntel adds new threat intelligence
func (tim *ThreatIntelManager) AddThreatIntel(intel *ThreatIntel) error {
	tim.mu.Lock()
	defer tim.mu.Unlock()

	if intel.ID == "" {
		return fmt.Errorf("hades.platform.threat_intel: threat intel ID cannot be empty")
	}

	tim.threatIntel[intel.ID] = intel
	return nil
}

// GetStatistics returns threat intelligence statistics
func (tim *ThreatIntelManager) GetStatistics() map[string]interface{} {
	tim.mu.RLock()
	defer tim.mu.RUnlock()

	stats := map[string]interface{}{
		"total_vulnerabilities":  len(tim.vulnerabilities),
		"total_threat_intel":     len(tim.threatIntel),
		"severity_breakdown":     make(map[string]int),
		"threat_level_breakdown": make(map[string]int),
		"last_update":            time.Now(),
	}

	// Severity breakdown
	for _, vuln := range tim.vulnerabilities {
		stats["severity_breakdown"].(map[string]int)[vuln.Severity]++
	}

	// Threat level breakdown
	for _, intel := range tim.threatIntel {
		stats["threat_level_breakdown"].(map[string]int)[string(intel.Level)]++
	}

	return stats
}

// backgroundUpdater periodically updates threat intelligence data
func (tim *ThreatIntelManager) backgroundUpdater() {
	ticker := time.NewTicker(tim.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tim.ctx.Done():
			return
		case <-ticker.C:
			tim.updateData()
		}
	}
}

// updateData fetches fresh threat intelligence data
func (tim *ThreatIntelManager) updateData() {
	if tim.config.EnableCVEFeeds {
		tim.updateCVEData()
	}

	if tim.config.EnableThreatFeeds {
		tim.updateThreatData()
	}
}

// updateCVEData fetches CVE data (simulated)
func (tim *ThreatIntelManager) updateCVEData() {
	// Simulated CVE data fetch
	sampleCVEs := []*Vulnerability{
		{
			ID:               "vuln_1",
			CVEID:            "CVE-2024-0001",
			Title:            "Critical Remote Code Execution Vulnerability",
			Description:      "A critical vulnerability allows remote code execution",
			Severity:         "critical",
			CVSSScore:        9.8,
			PublishedDate:    time.Now().AddDate(0, 0, -7),
			ModifiedDate:     time.Now(),
			AffectedProducts: []string{"Apache", "nginx", "OpenSSL"},
			References:       []string{"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2024-0001"},
			Tags:             []string{"rce", "critical", "web"},
		},
		{
			ID:               "vuln_2",
			CVEID:            "CVE-2024-0002",
			Title:            "SQL Injection Vulnerability",
			Description:      "SQL injection vulnerability in database layer",
			Severity:         "high",
			CVSSScore:        8.1,
			PublishedDate:    time.Now().AddDate(0, 0, -14),
			ModifiedDate:     time.Now(),
			AffectedProducts: []string{"MySQL", "PostgreSQL", "MongoDB"},
			References:       []string{"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2024-0002"},
			Tags:             []string{"sql-injection", "database", "high"},
		},
		{
			ID:               "vuln_3",
			CVEID:            "CVE-2024-0003",
			Title:            "Cross-Site Scripting (XSS) Vulnerability",
			Description:      "Reflected XSS vulnerability in web application",
			Severity:         "medium",
			CVSSScore:        6.1,
			PublishedDate:    time.Now().AddDate(0, 0, -21),
			ModifiedDate:     time.Now(),
			AffectedProducts: []string{"Chrome", "Firefox", "Safari"},
			References:       []string{"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2024-0003"},
			Tags:             []string{"xss", "web", "browser"},
		},
	}

	tim.mu.Lock()
	defer tim.mu.Unlock()

	for _, vuln := range sampleCVEs {
		tim.vulnerabilities[vuln.CVEID] = vuln
	}
}

// updateThreatData fetches threat intelligence data (simulated)
func (tim *ThreatIntelManager) updateThreatData() {
	// Simulated threat intelligence data
	sampleThreats := []*ThreatIntel{
		{
			ID:          "threat_1",
			Type:        "malware",
			Title:       "New Ransomware Variant Detected",
			Description: "A new ransomware variant targeting enterprise networks",
			Level:       ThreatLevelCritical,
			Confidence:  95,
			Source:      "internal_analysis",
			Indicators: map[string]interface{}{
				"file_hashes": []string{"a1b2c3d4e5f6", "f6e5d4c3b2a1"},
				"domains":     []string{"malicious.example.com", "badsite.example.org"},
				"ips":         []string{"192.168.1.100", "10.0.0.50"},
			},
			Timestamp: time.Now(),
			Tags:      []string{"ransomware", "enterprise", "critical"},
		},
		{
			ID:          "threat_2",
			Type:        "phishing",
			Title:       "Sophisticated Phishing Campaign",
			Description: "Phishing campaign targeting financial institutions",
			Level:       ThreatLevelHigh,
			Confidence:  85,
			Source:      "threat_feeds",
			Indicators: map[string]interface{}{
				"domains":  []string{"secure-bank.example.com", "login-financial.example.org"},
				"emails":   []string{"noreply@secure-bank.example.com"},
				"subjects": []string{"Urgent: Account Verification Required"},
			},
			Timestamp: time.Now().Add(-2 * time.Hour),
			Tags:      []string{"phishing", "financial", "high"},
		},
		{
			ID:          "threat_3",
			Type:        "apt",
			Title:       "APT Group Activity Detected",
			Description: "Advanced persistent threat group targeting government agencies",
			Level:       ThreatLevelHigh,
			Confidence:  90,
			Source:      "intelligence_partners",
			Indicators: map[string]interface{}{
				"tools":        []string{"mimikatz", "cobalt_strike"},
				"techniques":   []string{"T1059.001", "T1078.002"},
				"attributions": []string{"APT-28", "Fancy Bear"},
			},
			Timestamp: time.Now().Add(-6 * time.Hour),
			Tags:      []string{"apt", "government", "espionage"},
		},
	}

	tim.mu.Lock()
	defer tim.mu.Unlock()

	for _, threat := range sampleThreats {
		tim.threatIntel[threat.ID] = threat
	}
}

// matchesQuery checks if vulnerability matches search query
func (tim *ThreatIntelManager) matchesQuery(vuln *Vulnerability, query string) bool {
	query = strings.ToLower(query)

	return strings.Contains(strings.ToLower(vuln.Title), query) ||
		strings.Contains(strings.ToLower(vuln.Description), query) ||
		strings.Contains(strings.ToLower(vuln.CVEID), query)
}

// Close shuts down the threat intelligence manager
func (tim *ThreatIntelManager) Close() error {
	tim.cancel()
	return nil
}

// ImportFromJSON imports threat intelligence from JSON data
func (tim *ThreatIntelManager) ImportFromJSON(ctx context.Context, jsonData []byte) error {
	var data struct {
		Vulnerabilities []*Vulnerability `json:"vulnerabilities"`
		ThreatIntel     []*ThreatIntel   `json:"threat_intel"`
	}

	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("hades.platform.threat_intel: failed to parse JSON: %w", err)
	}

	tim.mu.Lock()
	defer tim.mu.Unlock()

	// Import vulnerabilities
	for _, vuln := range data.Vulnerabilities {
		if vuln.CVEID != "" {
			tim.vulnerabilities[vuln.CVEID] = vuln
		}
	}

	// Import threat intelligence
	for _, intel := range data.ThreatIntel {
		if intel.ID != "" {
			tim.threatIntel[intel.ID] = intel
		}
	}

	return nil
}

// ExportToJSON exports threat intelligence to JSON format
func (tim *ThreatIntelManager) ExportToJSON(ctx context.Context) ([]byte, error) {
	tim.mu.RLock()
	defer tim.mu.RUnlock()

	data := struct {
		Vulnerabilities []*Vulnerability `json:"vulnerabilities"`
		ThreatIntel     []*ThreatIntel   `json:"threat_intel"`
		ExportedAt      time.Time        `json:"exported_at"`
	}{
		ExportedAt: time.Now(),
	}

	// Convert maps to slices
	for _, vuln := range tim.vulnerabilities {
		data.Vulnerabilities = append(data.Vulnerabilities, vuln)
	}

	for _, intel := range tim.threatIntel {
		data.ThreatIntel = append(data.ThreatIntel, intel)
	}

	return json.MarshalIndent(data, "", "  ")
}
