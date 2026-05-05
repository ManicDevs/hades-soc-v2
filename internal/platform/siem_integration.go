package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// SIEMProvider represents supported SIEM/EDR providers
type SIEMProvider string

const (
	ProviderSplunk      SIEMProvider = "splunk"
	ProviderElastic     SIEMProvider = "elastic"
	ProviderSentinelOne SIEMProvider = "sentinelone"
	ProviderCrowdStrike SIEMProvider = "crowdstrike"
	ProviderQRadar      SIEMProvider = "qradar"
)

// SIEMEvent represents a security event for SIEM integration
type SIEMEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Severity    string                 `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target"`
	User        string                 `json:"user,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// SIEMConfig holds SIEM integration configuration
type SIEMConfig struct {
	Provider     SIEMProvider  `json:"provider"`
	Endpoint     string        `json:"endpoint"`
	APIKey       string        `json:"api_key"`
	SecretKey    string        `json:"secret_key,omitempty"`
	Index        string        `json:"index,omitempty"`
	Timeout      time.Duration `json:"timeout"`
	RetryCount   int           `json:"retry_count"`
	BatchSize    int           `json:"batch_size"`
	BatchTimeout time.Duration `json:"batch_timeout"`
}

// DefaultSIEMConfig returns sensible SIEM defaults
func DefaultSIEMConfig(provider SIEMProvider) *SIEMConfig {
	return &SIEMConfig{
		Provider:     provider,
		Timeout:      30 * time.Second,
		RetryCount:   3,
		BatchSize:    100,
		BatchTimeout: 5 * time.Second,
		Index:        "hades_events",
	}
}

// SIEMIntegration provides SIEM/EDR integration functionality
type SIEMIntegration struct {
	config     *SIEMConfig
	httpClient *http.Client
	eventQueue chan *SIEMEvent
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewSIEMIntegration creates a new SIEM integration instance
func NewSIEMIntegration(config *SIEMConfig) (*SIEMIntegration, error) {
	if config == nil {
		return nil, fmt.Errorf("hades.platform.siem: config cannot be nil")
	}

	ctx, cancel := context.WithCancel(context.Background())

	siem := &SIEMIntegration{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		eventQueue: make(chan *SIEMEvent, config.BatchSize),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start batch processor
	go siem.batchProcessor()

	return siem, nil
}

// SendEvent sends a single event to the SIEM
func (si *SIEMIntegration) SendEvent(ctx context.Context, event *SIEMEvent) error {
	select {
	case si.eventQueue <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("hades.platform.siem: event queue is full")
	}
}

// SendEventBatch sends multiple events to the SIEM
func (si *SIEMIntegration) SendEventBatch(ctx context.Context, events []*SIEMEvent) error {
	switch si.config.Provider {
	case ProviderSplunk:
		return si.sendToSplunk(ctx, events)
	case ProviderElastic:
		return si.sendToElastic(ctx, events)
	case ProviderSentinelOne:
		return si.sendToSentinelOne(ctx, events)
	case ProviderCrowdStrike:
		return si.sendToCrowdStrike(ctx, events)
	case ProviderQRadar:
		return si.sendToQRadar(ctx, events)
	default:
		return fmt.Errorf("hades.platform.siem: unsupported provider: %s", si.config.Provider)
	}
}

// batchProcessor processes events in batches
func (si *SIEMIntegration) batchProcessor() {
	ticker := time.NewTicker(si.config.BatchTimeout)
	defer ticker.Stop()

	var batch []*SIEMEvent

	for {
		select {
		case event := <-si.eventQueue:
			batch = append(batch, event)
			if len(batch) >= si.config.BatchSize {
				si.processBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				si.processBatch(batch)
				batch = batch[:0]
			}

		case <-si.ctx.Done():
			// Process remaining events before exit
			if len(batch) > 0 {
				si.processBatch(batch)
			}
			return
		}
	}
}

// processBatch sends a batch of events
func (si *SIEMIntegration) processBatch(events []*SIEMEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), si.config.Timeout)
	defer cancel()

	if err := si.SendEventBatch(ctx, events); err != nil {
		// Log error but continue processing
		fmt.Printf("hades.platform.siem: failed to send batch: %v\n", err)
	}
}

// sendToSplunk sends events to Splunk
func (si *SIEMIntegration) sendToSplunk(ctx context.Context, events []*SIEMEvent) error {
	url := fmt.Sprintf("%s/services/collector/event", si.config.Endpoint)

	for _, event := range events {
		payload := map[string]interface{}{
			"time":       event.Timestamp.Unix(),
			"index":      si.config.Index,
			"source":     "hades",
			"sourcetype": "hades:security",
			"event":      event,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("hades.platform.siem: failed to marshal Splunk event: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("hades.platform.siem: failed to create Splunk request: %w", err)
		}

		req.Header.Set("Authorization", "Splunk "+si.config.APIKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := si.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("hades.platform.siem: failed to send to Splunk: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("Warning: failed to close response body: %v", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("hades.platform.siem: Splunk returned status %d", resp.StatusCode)
		}
	}

	return nil
}

// sendToElastic sends events to Elasticsearch
func (si *SIEMIntegration) sendToElastic(ctx context.Context, events []*SIEMEvent) error {
	url := fmt.Sprintf("%s/%s/_bulk", si.config.Endpoint, si.config.Index)

	var bulkBuffer bytes.Buffer
	for _, event := range events {
		indexAction := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": si.config.Index,
			},
		}

		actionJSON, err := json.Marshal(indexAction)
		if err != nil {
			fmt.Printf("Warning: failed to marshal index action: %v\n", err)
			actionJSON = []byte("{}")
		}
		eventJSON, err := json.Marshal(event)
		if err != nil {
			fmt.Printf("Warning: failed to marshal event: %v\n", err)
			eventJSON = []byte("{}")
		}

		bulkBuffer.Write(actionJSON)
		bulkBuffer.WriteByte('\n')
		bulkBuffer.Write(eventJSON)
		bulkBuffer.WriteByte('\n')
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &bulkBuffer)
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to create Elastic request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+si.config.APIKey)
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := si.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to send to Elastic: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hades.platform.siem: Elastic returned status %d", resp.StatusCode)
	}

	return nil
}

// sendToSentinelOne sends events to SentinelOne
func (si *SIEMIntegration) sendToSentinelOne(ctx context.Context, events []*SIEMEvent) error {
	url := fmt.Sprintf("%s/web/api/v2.1/custom-events", si.config.Endpoint)

	payload := map[string]interface{}{
		"data": events,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to marshal SentinelOne events: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to create SentinelOne request: %w", err)
	}

	req.Header.Set("Authorization", "ApiToken "+si.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := si.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to send to SentinelOne: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("hades.platform.siem: SentinelOne returned status %d", resp.StatusCode)
	}

	return nil
}

// sendToCrowdStrike sends events to CrowdStrike
func (si *SIEMIntegration) sendToCrowdStrike(ctx context.Context, events []*SIEMEvent) error {
	url := fmt.Sprintf("%s/sensors/entities/datafeed/v1", si.config.Endpoint)

	payload := map[string]interface{}{
		"events": events,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to marshal CrowdStrike events: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to create CrowdStrike request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+si.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := si.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to send to CrowdStrike: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("hades.platform.siem: CrowdStrike returned status %d", resp.StatusCode)
	}

	return nil
}

// sendToQRadar sends events to IBM QRadar
func (si *SIEMIntegration) sendToQRadar(ctx context.Context, events []*SIEMEvent) error {
	url := fmt.Sprintf("%s/api/ariel/storedsearches", si.config.Endpoint)

	// QRadar uses different event format
	qrEvents := make([]map[string]interface{}, len(events))
	for i, event := range events {
		qrEvents[i] = map[string]interface{}{
			"devicetime":    event.Timestamp.Format("2006-01-02T15:04:05.000Z"),
			"eventname":     event.EventType,
			"severity":      si.mapSeverityToQRadar(event.Severity),
			"sourceaddress": event.IPAddress,
			"username":      event.User,
			"description":   event.Description,
			"customfields": map[string]string{
				"hades_source": event.Source,
				"hades_target": event.Target,
			},
		}
	}

	payload := map[string]interface{}{
		"events": qrEvents,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to marshal QRadar events: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to create QRadar request: %w", err)
	}

	req.Header.Set("Authorization", "QRadar "+si.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := si.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to send to QRadar: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("hades.platform.siem: QRadar returned status %d", resp.StatusCode)
	}

	return nil
}

// mapSeverityToQRadar maps severity levels to QRadar format
func (si *SIEMIntegration) mapSeverityToQRadar(severity string) int {
	switch severity {
	case "low":
		return 2
	case "medium":
		return 5
	case "high":
		return 8
	case "critical":
		return 10
	default:
		return 3
	}
}

// CreateSecurityEvent creates a standardized security event
func (si *SIEMIntegration) CreateSecurityEvent(eventType, severity, title, description, source, target, user, ipAddress, userAgent string, details map[string]interface{}) *SIEMEvent {
	tags := []string{"hades", eventType}
	if severity == "critical" || severity == "high" {
		tags = append(tags, "urgent")
	}

	return &SIEMEvent{
		Timestamp:   time.Now().UTC(),
		EventType:   eventType,
		Severity:    severity,
		Title:       title,
		Description: description,
		Source:      source,
		Target:      target,
		User:        user,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Details:     details,
		Tags:        tags,
	}
}

// Close shuts down the SIEM integration
func (si *SIEMIntegration) Close() error {
	si.cancel()
	return nil
}

// Health checks the SIEM connection
func (si *SIEMIntegration) Health(ctx context.Context) error {
	var url string
	switch si.config.Provider {
	case ProviderSplunk:
		url = fmt.Sprintf("%s/services/server/info", si.config.Endpoint)
	case ProviderElastic:
		url = si.config.Endpoint + "/"
	case ProviderSentinelOne:
		url = fmt.Sprintf("%s/web/api/v1.3/users", si.config.Endpoint)
	default:
		return fmt.Errorf("hades.platform.siem: health check not implemented for provider: %s", si.config.Provider)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("hades.platform.siem: failed to create health check request: %w", err)
	}

	switch si.config.Provider {
	case ProviderSplunk:
		req.Header.Set("Authorization", "Splunk "+si.config.APIKey)
	case ProviderElastic:
		req.Header.Set("Authorization", "Bearer "+si.config.APIKey)
	case ProviderSentinelOne:
		req.Header.Set("Authorization", "ApiToken "+si.config.APIKey)
	}

	resp, err := si.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("hades.platform.siem: health check failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hades.platform.siem: health check returned status %d", resp.StatusCode)
	}

	return nil
}
