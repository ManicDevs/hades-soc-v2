package threat

import (
	"context"
	"sync"
	"time"
)

// ThreatIntelligence manages threat intelligence data
type ThreatIntelligence struct {
	indicators map[string]ThreatIndicator
	feeds      map[string]ThreatFeed
	mu         sync.RWMutex
	lastUpdate time.Time
}

// ThreatFeed represents a threat intelligence feed
type ThreatFeed struct {
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	LastUpdate time.Time `json:"last_update"`
	Indicators int       `json:"indicators"`
	Active     bool      `json:"active"`
	Priority   int       `json:"priority"`
}

// NewThreatIntelligence creates a new threat intelligence manager
func NewThreatIntelligence() *ThreatIntelligence {
	return &ThreatIntelligence{
		indicators: make(map[string]ThreatIndicator),
		feeds:      make(map[string]ThreatFeed),
		lastUpdate: time.Now(),
	}
}

// UpdateIndicators updates threat indicators
func (ti *ThreatIntelligence) UpdateIndicators(ctx context.Context, indicators []ThreatIndicator) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	for _, indicator := range indicators {
		key := indicator.Type + "_" + indicator.Value
		ti.indicators[key] = indicator
	}

	ti.lastUpdate = time.Now()
	return nil
}

// GetIndicator gets a specific threat indicator
func (ti *ThreatIntelligence) GetIndicator(indicatorType, value string) (ThreatIndicator, bool) {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	key := indicatorType + "_" + value
	indicator, exists := ti.indicators[key]
	return indicator, exists
}

// GetAllIndicators returns all threat indicators
func (ti *ThreatIntelligence) GetAllIndicators() []ThreatIndicator {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	indicators := make([]ThreatIndicator, 0, len(ti.indicators))
	for _, indicator := range ti.indicators {
		indicators = append(indicators, indicator)
	}

	return indicators
}

// AddFeed adds a threat intelligence feed
func (ti *ThreatIntelligence) AddFeed(feed ThreatFeed) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	ti.feeds[feed.Name] = feed
}

// GetFeeds returns all threat feeds
func (ti *ThreatIntelligence) GetFeeds() map[string]ThreatFeed {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	feeds := make(map[string]ThreatFeed)
	for k, v := range ti.feeds {
		feeds[k] = v
	}

	return feeds
}
