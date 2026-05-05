package threat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// ThreatFeedManager manages multiple threat intelligence feeds
type ThreatFeedManager struct {
	feeds     map[string]*ThreatFeedClient
	processor *ThreatProcessor
	mu        sync.RWMutex
	active    bool
}

// ThreatFeedClient represents a client for a threat intelligence feed
type ThreatFeedClient struct {
	Name           string
	URL            string
	APIKey         string
	Format         string // "json", "xml", "stix", "taxii"
	UpdateInterval time.Duration
	LastUpdate     time.Time
	Status         string
	Client         *http.Client
	Parser         FeedParser
}

// FeedParser interface for different feed formats
type FeedParser interface {
	Parse(data []byte) ([]ThreatIndicator, error)
	Validate(indicator ThreatIndicator) bool
}

// ThreatProcessor processes and enriches threat indicators
type ThreatProcessor struct {
	enrichers []IndicatorEnricher
	cache     map[string]ThreatIndicator
	mu        sync.RWMutex
}

// IndicatorEnricher interface for enriching threat indicators
type IndicatorEnricher interface {
	Enrich(ctx context.Context, indicator ThreatIndicator) (ThreatIndicator, error)
	Name() string
}

// NewThreatFeedManager creates a new threat feed manager
func NewThreatFeedManager() *ThreatFeedManager {
	tfm := &ThreatFeedManager{
		feeds:     make(map[string]*ThreatFeedClient),
		processor: NewThreatProcessor(),
		active:    false,
	}

	// Initialize default feeds
	tfm.initializeDefaultFeeds()
	return tfm
}

// initializeDefaultFeeds initializes common threat intelligence feeds
func (tfm *ThreatFeedManager) initializeDefaultFeeds() {
	defaultFeeds := []struct {
		name   string
		url    string
		format string
	}{
		{
			name:   "VirusTotal",
			url:    "https://www.virustotal.com/vtapi/v2/",
			format: "json",
		},
		{
			name:   "AbuseIPDB",
			url:    "https://api.abuseipdb.com/api/v2/",
			format: "json",
		},
		{
			name:   "OTX AlienVault",
			url:    "https://otx.alienvault.com/api/v1/",
			format: "json",
		},
		{
			name:   "MISP",
			url:    "https://misp.example.com/",
			format: "json",
		},
		{
			name:   "ThreatFox",
			url:    "https://threatfox.abuse.ch/api/v1/",
			format: "json",
		},
	}

	for _, feed := range defaultFeeds {
		client := &ThreatFeedClient{
			Name:           feed.name,
			URL:            feed.url,
			Format:         feed.format,
			UpdateInterval: 1 * time.Hour,
			Client:         &http.Client{Timeout: 30 * time.Second},
			Status:         "inactive",
		}

		// Set appropriate parser based on format
		switch feed.format {
		case "json":
			client.Parser = &JSONParser{}
		case "xml":
			client.Parser = &XMLParser{}
		case "stix":
			client.Parser = &STIXParser{}
		case "taxii":
			client.Parser = &TAXIIParser{}
		default:
			client.Parser = &JSONParser{}
		}

		tfm.feeds[feed.name] = client
	}
}

// Start starts the threat feed manager
func (tfm *ThreatFeedManager) Start(ctx context.Context) error {
	tfm.mu.Lock()
	defer tfm.mu.Unlock()

	if tfm.active {
		return fmt.Errorf("threat feed manager already active")
	}

	tfm.active = true

	// Start feed collectors
	for name, feed := range tfm.feeds {
		go tfm.runFeedCollector(ctx, name, feed)
	}

	log.Println("Threat feed manager started")
	return nil
}

// Stop stops the threat feed manager
func (tfm *ThreatFeedManager) Stop() error {
	tfm.mu.Lock()
	defer tfm.mu.Unlock()

	tfm.active = false
	log.Println("Threat feed manager stopped")
	return nil
}

// runFeedCollector runs a feed collector for a specific feed
func (tfm *ThreatFeedManager) runFeedCollector(ctx context.Context, name string, feed *ThreatFeedClient) {
	ticker := time.NewTicker(feed.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := tfm.collectFromFeed(ctx, feed); err != nil {
				log.Printf("Failed to collect from feed %s: %v", name, err)
				feed.Status = "error"
			} else {
				feed.Status = "active"
				feed.LastUpdate = time.Now()
			}
		}
	}
}

// collectFromFeed collects threat indicators from a feed
func (tfm *ThreatFeedManager) collectFromFeed(ctx context.Context, feed *ThreatFeedClient) error {
	req, err := http.NewRequestWithContext(ctx, "GET", feed.URL, nil)
	if err != nil {
		return err
	}

	if feed.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+feed.APIKey)
	}
	req.Header.Set("User-Agent", "Hades-Threat-Intelligence/1.0")

	resp, err := feed.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	indicators, err := feed.Parser.Parse(data)
	if err != nil {
		return err
	}

	// Process and enrich indicators
	processedIndicators := make([]ThreatIndicator, 0, len(indicators))
	for _, indicator := range indicators {
		if feed.Parser.Validate(indicator) {
			enriched, err := tfm.processor.EnrichIndicator(ctx, indicator)
			if err != nil {
				log.Printf("Failed to enrich indicator: %v", err)
				processedIndicators = append(processedIndicators, indicator)
			} else {
				processedIndicators = append(processedIndicators, enriched)
			}
		}
	}

	// Update cache
	tfm.processor.UpdateCache(processedIndicators)

	log.Printf("Collected %d indicators from %s", len(processedIndicators), feed.Name)
	return nil
}

// GetFeedStatus returns the status of all feeds
func (tfm *ThreatFeedManager) GetFeedStatus() map[string]FeedStatus {
	tfm.mu.RLock()
	defer tfm.mu.RUnlock()

	status := make(map[string]FeedStatus)
	for name, feed := range tfm.feeds {
		status[name] = FeedStatus{
			Name:           feed.Name,
			URL:            feed.URL,
			Format:         feed.Format,
			Status:         feed.Status,
			LastUpdate:     feed.LastUpdate,
			UpdateInterval: feed.UpdateInterval,
		}
	}

	return status
}

// FeedStatus represents the status of a threat feed
type FeedStatus struct {
	Name           string        `json:"name"`
	URL            string        `json:"url"`
	Format         string        `json:"format"`
	Status         string        `json:"status"`
	LastUpdate     time.Time     `json:"last_update"`
	UpdateInterval time.Duration `json:"update_interval"`
}

// NewThreatProcessor creates a new threat processor
func NewThreatProcessor() *ThreatProcessor {
	tp := &ThreatProcessor{
		enrichers: make([]IndicatorEnricher, 0),
		cache:     make(map[string]ThreatIndicator),
	}

	// Initialize enrichers
	tp.enrichers = append(tp.enrichers,
		&GeoIPEnricher{},
		&DomainEnricher{},
		&ReputationEnricher{},
		&MalwareEnricher{},
	)

	return tp
}

// EnrichIndicator enriches a threat indicator
func (tp *ThreatProcessor) EnrichIndicator(ctx context.Context, indicator ThreatIndicator) (ThreatIndicator, error) {
	enriched := indicator

	for _, enricher := range tp.enrichers {
		var err error
		enriched, err = enricher.Enrich(ctx, enriched)
		if err != nil {
			log.Printf("Enricher %s failed: %v", enricher.Name(), err)
		}
	}

	return enriched, nil
}

// UpdateCache updates the indicator cache
func (tp *ThreatProcessor) UpdateCache(indicators []ThreatIndicator) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	for _, indicator := range indicators {
		key := indicator.Type + "_" + indicator.Value
		tp.cache[key] = indicator
	}
}

// GetFromCache gets an indicator from cache
func (tp *ThreatProcessor) GetFromCache(indicatorType, value string) (ThreatIndicator, bool) {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	key := indicatorType + "_" + value
	indicator, exists := tp.cache[key]
	return indicator, exists
}

// JSONParser implements JSON feed parsing
type JSONParser struct{}

func (jp *JSONParser) Parse(data []byte) ([]ThreatIndicator, error) {
	var feedData struct {
		Indicators []struct {
			Type        string    `json:"type"`
			Value       string    `json:"value"`
			Confidence  float64   `json:"confidence"`
			Source      string    `json:"source"`
			Description string    `json:"description"`
			FirstSeen   time.Time `json:"first_seen"`
			LastSeen    time.Time `json:"last_seen"`
		} `json:"indicators"`
	}

	if err := json.Unmarshal(data, &feedData); err != nil {
		return nil, err
	}

	indicators := make([]ThreatIndicator, 0, len(feedData.Indicators))
	for _, item := range feedData.Indicators {
		indicators = append(indicators, ThreatIndicator{
			Type:        item.Type,
			Value:       item.Value,
			Confidence:  item.Confidence,
			Source:      item.Source,
			FirstSeen:   item.FirstSeen,
			LastSeen:    item.LastSeen,
			Description: item.Description,
		})
	}

	return indicators, nil
}

func (jp *JSONParser) Validate(indicator ThreatIndicator) bool {
	return indicator.Type != "" && indicator.Value != ""
}

// XMLParser implements XML feed parsing
type XMLParser struct{}

func (xp *XMLParser) Parse(data []byte) ([]ThreatIndicator, error) {
	// Simplified XML parsing - in production, use proper XML parser
	return []ThreatIndicator{}, nil
}

func (xp *XMLParser) Validate(indicator ThreatIndicator) bool {
	return indicator.Type != "" && indicator.Value != ""
}

// STIXParser implements STIX format parsing
type STIXParser struct{}

func (sp *STIXParser) Parse(data []byte) ([]ThreatIndicator, error) {
	// Simplified STIX parsing - in production, use STIX library
	return []ThreatIndicator{}, nil
}

func (sp *STIXParser) Validate(indicator ThreatIndicator) bool {
	return indicator.Type != "" && indicator.Value != ""
}

// TAXIIParser implements TAXII format parsing
type TAXIIParser struct{}

func (tp *TAXIIParser) Parse(data []byte) ([]ThreatIndicator, error) {
	// Simplified TAXII parsing - in production, use TAXII library
	return []ThreatIndicator{}, nil
}

func (tp *TAXIIParser) Validate(indicator ThreatIndicator) bool {
	return indicator.Type != "" && indicator.Value != ""
}

// GeoIPEnricher enriches indicators with geo-location data
type GeoIPEnricher struct{}

func (ge *GeoIPEnricher) Name() string {
	return "GeoIP Enricher"
}

func (ge *GeoIPEnricher) Enrich(ctx context.Context, indicator ThreatIndicator) (ThreatIndicator, error) {
	// Simplified geo-location enrichment
	// In production, use actual GeoIP service
	if indicator.Type == "ip" {
		// Add mock geo-location data
		indicator.Description += " [GeoIP: US, California]"
	}
	return indicator, nil
}

// DomainEnricher enriches domain indicators
type DomainEnricher struct{}

func (de *DomainEnricher) Name() string {
	return "Domain Enricher"
}

func (de *DomainEnricher) Enrich(ctx context.Context, indicator ThreatIndicator) (ThreatIndicator, error) {
	if indicator.Type == "domain" {
		// Add mock domain analysis
		indicator.Description += " [Domain: Suspicious TLD]"
	}
	return indicator, nil
}

// ReputationEnricher enriches indicators with reputation data
type ReputationEnricher struct{}

func (re *ReputationEnricher) Name() string {
	return "Reputation Enricher"
}

func (re *ReputationEnricher) Enrich(ctx context.Context, indicator ThreatIndicator) (ThreatIndicator, error) {
	// Add mock reputation data
	indicator.Description += " [Reputation: Malicious]"
	return indicator, nil
}

// MalwareEnricher enriches malware indicators
type MalwareEnricher struct{}

func (me *MalwareEnricher) Name() string {
	return "Malware Enricher"
}

func (me *MalwareEnricher) Enrich(ctx context.Context, indicator ThreatIndicator) (ThreatIndicator, error) {
	if indicator.Type == "hash" {
		// Add mock malware analysis
		indicator.Description += " [Malware: Trojan]"
	}
	return indicator, nil
}
