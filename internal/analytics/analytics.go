package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"hades-v2/internal/database"
)

// AnalyticsEngine provides comprehensive analytics and reporting capabilities
type AnalyticsEngine struct {
	db               database.Database
	mlModel          *MLModel
	predictor        *PredictiveAnalytics
	aggregator       *DataAggregator
	streamProcessor  *StreamProcessor
	insightGenerator *InsightGenerator
	mu               sync.RWMutex
}

// MLModel represents machine learning models for analytics
type MLModel struct {
	ModelPath   string
	Version     string
	Accuracy    float64
	IsLoaded    bool
	LastTrained time.Time
	InputSize   int
	OutputSize  int
}

// PredictiveAnalytics provides predictive capabilities
type PredictiveAnalytics struct {
	Horizon    time.Duration
	Confidence float64
	Models     map[string]*PredictionModel
	Historical *HistoricalData
}

// PredictionModel represents a specific prediction model
type PredictionModel struct {
	Name        string
	Type        string
	Accuracy    float64
	Features    []string
	LastUpdated time.Time
	IsEnabled   bool
}

// DataAggregator handles data aggregation and summarization
type DataAggregator struct {
	WindowSizes  []time.Duration
	Metrics      map[string]*MetricDefinition
	Aggregations map[string]*AggregationResult
}

// MetricDefinition defines how to aggregate metrics
type MetricDefinition struct {
	Name        string
	Type        string
	Aggregation string
	Window      time.Duration
	Threshold   float64
}

// AggregationResult stores aggregated data
type AggregationResult struct {
	MetricName string
	Value      float64
	Count      int64
	Timestamp  time.Time
	Window     time.Duration
	Min        float64
	Max        float64
	Avg        float64
	StdDev     float64
}

// StreamProcessor handles real-time stream processing
type StreamProcessor struct {
	BufferSize    int
	Window        time.Duration
	Processors    map[string]*StreamFunction
	ActiveStreams map[string]*Stream
}

// StreamFunction represents a stream processing function
type StreamFunction struct {
	Name      string
	Function  func(*StreamEvent) *StreamEvent
	Condition func(*StreamEvent) bool
	IsEnabled bool
}

// Stream represents a data stream
type Stream struct {
	ID          string
	Events      []*StreamEvent
	Window      time.Duration
	LastUpdated time.Time
	IsActive    bool
}

// StreamEvent represents an event in a stream
type StreamEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Processed bool                   `json:"processed"`
}

// InsightGenerator generates actionable insights
type InsightGenerator struct {
	Rules     []*InsightRule
	Templates map[string]*InsightTemplate
	Priority  map[string]int
}

// InsightRule defines how to generate insights
type InsightRule struct {
	Name       string
	Condition  string
	Action     string
	Priority   int
	Confidence float64
	IsEnabled  bool
}

// InsightTemplate defines insight structure
type InsightTemplate struct {
	Name     string
	Format   string
	Fields   []string
	Priority int
}

// HistoricalData stores historical analytics data
type HistoricalData struct {
	Events       []HistoricalEvent
	Aggregations map[string]*HistoricalAggregation
	LastUpdated  time.Time
}

// HistoricalEvent represents historical data point
type HistoricalEvent struct {
	Timestamp time.Time          `json:"timestamp"`
	Metrics   map[string]float64 `json:"metrics"`
	Labels    map[string]string  `json:"labels"`
}

// HistoricalAggregation represents aggregated historical data
type HistoricalAggregation struct {
	MetricName string
	Values     []float64
	Timestamps []time.Time
	Window     time.Duration
}

// AnalyticsRequest represents an analytics request
type AnalyticsRequest struct {
	Query       string                 `json:"query"`
	Parameters  map[string]interface{} `json:"parameters"`
	TimeRange   TimeRange              `json:"time_range"`
	Granularity string                 `json:"granularity"`
	Metrics     []string               `json:"metrics"`
}

// TimeRange defines a time range for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// AnalyticsResponse represents analytics response
type AnalyticsResponse struct {
	QueryID     string                 `json:"query_id"`
	Results     []AnalyticsResult      `json:"results"`
	Insights    []Insight              `json:"insights"`
	Predictions []Prediction           `json:"predictions"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AnalyticsResult represents a single analytics result
type AnalyticsResult struct {
	MetricName string            `json:"metric_name"`
	Value      float64           `json:"value"`
	Timestamp  time.Time         `json:"timestamp"`
	Labels     map[string]string `json:"labels"`
}

// Insight represents an actionable insight
type Insight struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	Confidence  float64                `json:"confidence"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
}

// Prediction represents a predictive analytics result
type Prediction struct {
	MetricName string        `json:"metric_name"`
	Predicted  float64       `json:"predicted"`
	Confidence float64       `json:"confidence"`
	Horizon    time.Duration `json:"horizon"`
	Timestamp  time.Time     `json:"timestamp"`
	Features   []string      `json:"features"`
}

// NewAnalyticsEngine creates a new analytics engine
func NewAnalyticsEngine(db database.Database) (*AnalyticsEngine, error) {
	engine := &AnalyticsEngine{
		db: db,
		mlModel: &MLModel{
			ModelPath:   "/models/analytics_v1.pb",
			Version:     "1.0",
			Accuracy:    0.0,
			IsLoaded:    false,
			LastTrained: time.Now(),
			InputSize:   64,
			OutputSize:  10,
		},
		predictor: &PredictiveAnalytics{
			Horizon:    24 * time.Hour,
			Confidence: 0.8,
			Models:     make(map[string]*PredictionModel),
			Historical: &HistoricalData{
				Events:       make([]HistoricalEvent, 0),
				Aggregations: make(map[string]*HistoricalAggregation),
				LastUpdated:  time.Now(),
			},
		},
		aggregator: &DataAggregator{
			WindowSizes: []time.Duration{
				1 * time.Minute,
				5 * time.Minute,
				15 * time.Minute,
				1 * time.Hour,
				24 * time.Hour,
			},
			Metrics:      make(map[string]*MetricDefinition),
			Aggregations: make(map[string]*AggregationResult),
		},
		streamProcessor: &StreamProcessor{
			BufferSize:    1000,
			Window:        5 * time.Minute,
			Processors:    make(map[string]*StreamFunction),
			ActiveStreams: make(map[string]*Stream),
		},
		insightGenerator: &InsightGenerator{
			Rules:     make([]*InsightRule, 0),
			Templates: make(map[string]*InsightTemplate),
			Priority:  make(map[string]int),
		},
	}

	// Initialize components
	if err := engine.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize analytics engine: %w", err)
	}

	return engine, nil
}

// initializeComponents initializes analytics components
func (ae *AnalyticsEngine) initializeComponents() error {
	// Initialize prediction models
	ae.predictor.Models["threat_volume"] = &PredictionModel{
		Name:        "threat_volume",
		Type:        "time_series",
		Accuracy:    0.85,
		Features:    []string{"time", "day_of_week", "hour", "previous_volume"},
		LastUpdated: time.Now(),
		IsEnabled:   true,
	}

	ae.predictor.Models["resource_usage"] = &PredictionModel{
		Name:        "resource_usage",
		Type:        "regression",
		Accuracy:    0.78,
		Features:    []string{"cpu", "memory", "network", "connections"},
		LastUpdated: time.Now(),
		IsEnabled:   true,
	}

	// Initialize metrics
	ae.aggregator.Metrics["threat_count"] = &MetricDefinition{
		Name:        "threat_count",
		Type:        "counter",
		Aggregation: "sum",
		Window:      1 * time.Hour,
		Threshold:   100.0,
	}

	ae.aggregator.Metrics["response_time"] = &MetricDefinition{
		Name:        "response_time",
		Type:        "gauge",
		Aggregation: "avg",
		Window:      5 * time.Minute,
		Threshold:   1000.0,
	}

	ae.aggregator.Metrics["cpu_usage"] = &MetricDefinition{
		Name:        "cpu_usage",
		Type:        "gauge",
		Aggregation: "avg",
		Window:      1 * time.Minute,
		Threshold:   80.0,
	}

	// Initialize stream processors
	ae.streamProcessor.Processors["threat_filter"] = &StreamFunction{
		Name: "threat_filter",
		Function: func(event *StreamEvent) *StreamEvent {
			// Filter high-severity threats
			if severity, ok := event.Data["severity"].(string); ok && severity == "high" {
				event.Processed = true
			}
			return event
		},
		Condition: func(event *StreamEvent) bool {
			_, ok := event.Data["severity"]
			return ok
		},
		IsEnabled: true,
	}

	ae.streamProcessor.Processors["anomaly_detector"] = &StreamFunction{
		Name: "anomaly_detector",
		Function: func(event *StreamEvent) *StreamEvent {
			// Detect anomalies in stream data
			if value, ok := event.Data["value"].(float64); ok {
				if value > 100.0 { // Simple threshold
					event.Data["anomaly"] = true
					event.Processed = true
				}
			}
			return event
		},
		Condition: func(event *StreamEvent) bool {
			_, ok := event.Data["value"]
			return ok
		},
		IsEnabled: true,
	}

	// Initialize insight rules
	ae.insightGenerator.Rules = append(ae.insightGenerator.Rules, &InsightRule{
		Name:       "high_threat_volume",
		Condition:  "threat_count > threshold",
		Action:     "alert_security_team",
		Priority:   1,
		Confidence: 0.9,
		IsEnabled:  true,
	})

	ae.insightGenerator.Rules = append(ae.insightGenerator.Rules, &InsightRule{
		Name:       "resource_exhaustion",
		Condition:  "cpu_usage > 80% OR memory_usage > 90%",
		Action:     "scale_resources",
		Priority:   2,
		Confidence: 0.85,
		IsEnabled:  true,
	})

	// Initialize insight templates
	ae.insightGenerator.Templates["threat_spike"] = &InsightTemplate{
		Name:     "threat_spike",
		Format:   "Threat volume increased by %.1f%% in the last %s",
		Fields:   []string{"percentage", "time_window"},
		Priority: 1,
	}

	return nil
}

// SecurityMetrics represents security analytics data
type SecurityMetrics struct {
	TotalThreats          int       `json:"total_threats"`
	CriticalThreats       int       `json:"critical_threats"`
	HighThreats           int       `json:"high_threats"`
	MediumThreats         int       `json:"medium_threats"`
	LowThreats            int       `json:"low_threats"`
	ResolvedThreats       int       `json:"resolved_threats"`
	ActiveThreats         int       `json:"active_threats"`
	AverageResolutionTime float64   `json:"avg_resolution_time"`
	ThreatTrend           string    `json:"threat_trend"`
	SecurityScore         float64   `json:"security_score"`
	ComplianceScore       float64   `json:"compliance_score"`
	Timestamp             time.Time `json:"timestamp"`
}

// UserMetrics represents user analytics data
type UserMetrics struct {
	TotalUsers         int          `json:"total_users"`
	ActiveUsers        int          `json:"active_users"`
	NewUsersToday      int          `json:"new_users_today"`
	LoginAttempts      int64        `json:"login_attempts"`
	FailedLogins       int64        `json:"failed_logins"`
	UniqueVisitors     int          `json:"unique_visitors"`
	PageViews          int64        `json:"page_views"`
	AverageSessionTime float64      `json:"avg_session_time"`
	TopPages           []PageMetric `json:"top_pages"`
	Timestamp          time.Time    `json:"timestamp"`
}

// PageMetric represents page analytics data
type PageMetric struct {
	Page        string    `json:"page"`
	Views       int       `json:"views"`
	UniqueUsers int       `json:"unique_users"`
	BounceRate  float64   `json:"bounce_rate"`
	Timestamp   time.Time `json:"timestamp"`
}

// SystemMetrics represents system performance analytics
type SystemMetrics struct {
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     float64   `json:"memory_usage"`
	DiskUsage       float64   `json:"disk_usage"`
	NetworkIn       int64     `json:"network_in"`
	NetworkOut      int64     `json:"network_out"`
	DatabaseQueries int       `json:"database_queries"`
	ResponseTime    float64   `json:"avg_response_time"`
	ErrorRate       float64   `json:"error_rate"`
	Throughput      float64   `json:"throughput"`
	Uptime          int       `json:"uptime_seconds"`
	Timestamp       time.Time `json:"timestamp"`
}

// ThreatAnalytics represents threat intelligence analytics
type ThreatAnalytics struct {
	TopThreatTypes    []ThreatTypeMetric   `json:"top_threat_types"`
	ThreatSources     []ThreatSourceMetric `json:"threat_sources"`
	DetectionAccuracy float64              `json:"detection_accuracy"`
	FalsePositives    int                  `json:"false_positives"`
	ResponseTimes     []ResponseTimeMetric `json:"response_times"`
	Timestamp         time.Time            `json:"timestamp"`
}

// ThreatTypeMetric represents threat type statistics
type ThreatTypeMetric struct {
	Type     string `json:"type"`
	Count    int    `json:"count"`
	Severity string `json:"severity"`
	Trend    string `json:"trend"`
}

// ThreatSourceMetric represents threat source statistics
type ThreatSourceMetric struct {
	Source     string  `json:"source"`
	Count      int     `json:"count"`
	Confidence float64 `json:"confidence"`
}

// ResponseTimeMetric represents response time analytics
type ResponseTimeMetric struct {
	Endpoint  string    `json:"endpoint"`
	AvgTime   float64   `json:"avg_time"`
	P95Time   float64   `json:"p95_time"`
	P99Time   float64   `json:"p99_time"`
	Timestamp time.Time `json:"timestamp"`
}

// ReportRequest represents a request for generating reports
type ReportRequest struct {
	ReportType string                 `json:"report_type"`
	StartDate  time.Time              `json:"start_date"`
	EndDate    time.Time              `json:"end_date"`
	Format     string                 `json:"format"` // json, csv, pdf, html
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

// Report represents generated analytics reports
type Report struct {
	ID          int         `json:"id"`
	Type        string      `json:"type"`
	Title       string      `json:"title"`
	GeneratedAt time.Time   `json:"generated_at"`
	Format      string      `json:"format"`
	Size        int64       `json:"size"`
	Data        interface{} `json:"data"`
	URL         string      `json:"url,omitempty"`
}

// GetSecurityMetrics retrieves comprehensive security metrics
func (ae *AnalyticsEngine) GetSecurityMetrics() (*SecurityMetrics, error) {
	sqlDB, ok := ae.db.GetConnection().(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT 
			COUNT(CASE WHEN severity = 'critical' THEN 1 ELSE 0) as total_threats,
			COUNT(CASE WHEN severity = 'high' THEN 1 ELSE 0) as high_threats,
			COUNT(CASE WHEN severity = 'medium' THEN 1 ELSE 0) as medium_threats,
			COUNT(CASE WHEN severity = 'low' THEN 1 ELSE 0) as low_threats,
			COUNT(CASE WHEN status = 'resolved' THEN 1 ELSE 0) as resolved_threats,
			COUNT(CASE WHEN status = 'active' THEN 1 ELSE 0) as active_threats,
			AVG(EXTRACT(EPOCH FROM (CASE WHEN detected_at IS NOT NULL THEN detected_at ELSE created_at))) as avg_resolution_time
		FROM threats 
		WHERE detected_at >= NOW() - INTERVAL '30 days'
	`

	var metrics SecurityMetrics
	err := sqlDB.QueryRow(query).Scan(
		&metrics.TotalThreats, &metrics.CriticalThreats, &metrics.HighThreats,
		&metrics.MediumThreats, &metrics.LowThreats, &metrics.ResolvedThreats,
		&metrics.ActiveThreats, &metrics.AverageResolutionTime,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get security metrics: %w", err)
	}

	metrics.Timestamp = time.Now()
	return &metrics, nil
}

// ProcessAnalytics processes analytics requests with ML integration
func (ae *AnalyticsEngine) ProcessAnalytics(ctx context.Context, request AnalyticsRequest) (*AnalyticsResponse, error) {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	response := &AnalyticsResponse{
		QueryID:     fmt.Sprintf("query_%d", time.Now().Unix()),
		Results:     make([]AnalyticsResult, 0),
		Insights:    make([]Insight, 0),
		Predictions: make([]Prediction, 0),
		Metadata:    make(map[string]interface{}),
		Timestamp:   time.Now(),
	}

	// Process metrics
	for _, metricName := range request.Metrics {
		if metricDef, ok := ae.aggregator.Metrics[metricName]; ok {
			result, err := ae.processMetric(ctx, metricDef, request.TimeRange)
			if err != nil {
				log.Printf("Failed to process metric %s: %v", metricName, err)
				continue
			}
			response.Results = append(response.Results, *result)
		}
	}

	// Generate insights
	insights, err := ae.generateInsights(ctx, response.Results)
	if err != nil {
		log.Printf("Failed to generate insights: %v", err)
	} else {
		response.Insights = insights
	}

	// Generate predictions
	predictions, err := ae.generatePredictions(ctx, request.Metrics, request.TimeRange)
	if err != nil {
		log.Printf("Failed to generate predictions: %v", err)
	} else {
		response.Predictions = predictions
	}

	// Add metadata
	response.Metadata["processing_time"] = time.Since(time.Now()).String()
	response.Metadata["metrics_processed"] = len(response.Results)
	response.Metadata["insights_generated"] = len(response.Insights)
	response.Metadata["predictions_generated"] = len(response.Predictions)

	return response, nil
}

// processMetric processes a single metric
func (ae *AnalyticsEngine) processMetric(ctx context.Context, metricDef *MetricDefinition, timeRange TimeRange) (*AnalyticsResult, error) {
	// Simulate metric processing (in real implementation, query database)
	value := ae.simulateMetricValue(metricDef.Name, timeRange)

	result := &AnalyticsResult{
		MetricName: metricDef.Name,
		Value:      value,
		Timestamp:  time.Now(),
		Labels:     map[string]string{"type": metricDef.Type},
	}

	return result, nil
}

// simulateMetricValue simulates metric value calculation
func (ae *AnalyticsEngine) simulateMetricValue(metricName string, timeRange TimeRange) float64 {
	// Simple simulation based on metric type
	switch metricName {
	case "threat_count":
		return float64(timeRange.End.Unix()-timeRange.Start.Unix()) / 3600 * 15 // 15 threats per hour
	case "response_time":
		return 150.0 + math.Sin(float64(time.Now().Unix())/60)*50 // Oscillating response time
	case "cpu_usage":
		return 45.0 + math.Sin(float64(time.Now().Unix())/300)*20 // Oscillating CPU usage
	default:
		return 100.0
	}
}

// generateInsights generates actionable insights
func (ae *AnalyticsEngine) generateInsights(ctx context.Context, results []AnalyticsResult) ([]Insight, error) {
	insights := make([]Insight, 0)

	for _, result := range results {
		// Check against thresholds and rules
		for _, rule := range ae.insightGenerator.Rules {
			if !rule.IsEnabled {
				continue
			}

			if ae.evaluateRule(rule, result) {
				insight := Insight{
					ID:          fmt.Sprintf("insight_%d_%s", time.Now().Unix(), result.MetricName),
					Title:       rule.Name,
					Description: ae.generateInsightDescription(rule, result),
					Priority:    ae.getPriorityString(rule.Priority),
					Confidence:  rule.Confidence,
					Data: map[string]interface{}{
						"metric_name": result.MetricName,
						"value":       result.Value,
						"rule":        rule.Name,
					},
					Timestamp: time.Now(),
				}
				insights = append(insights, insight)
			}
		}
	}

	return insights, nil
}

// evaluateRule evaluates an insight rule
func (ae *AnalyticsEngine) evaluateRule(rule *InsightRule, result AnalyticsResult) bool {
	// Simple rule evaluation (in real implementation, use proper expression evaluation)
	switch rule.Name {
	case "high_threat_volume":
		return result.MetricName == "threat_count" && result.Value > 100.0
	case "resource_exhaustion":
		return (result.MetricName == "cpu_usage" && result.Value > 80.0) ||
			(result.MetricName == "memory_usage" && result.Value > 90.0)
	default:
		return false
	}
}

// generateInsightDescription generates insight description
func (ae *AnalyticsEngine) generateInsightDescription(rule *InsightRule, result AnalyticsResult) string {
	if template, ok := ae.insightGenerator.Templates["threat_spike"]; ok && rule.Name == "high_threat_volume" {
		return fmt.Sprintf(template.Format, result.Value, "last hour")
	}
	return fmt.Sprintf("Rule '%s' triggered for metric '%s' with value %.2f", rule.Name, result.MetricName, result.Value)
}

// getPriorityString converts priority to string
func (ae *AnalyticsEngine) getPriorityString(priority int) string {
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

// generatePredictions generates predictive analytics
func (ae *AnalyticsEngine) generatePredictions(ctx context.Context, metrics []string, timeRange TimeRange) ([]Prediction, error) {
	predictions := make([]Prediction, 0)

	for _, metricName := range metrics {
		if model, ok := ae.predictor.Models[metricName]; ok && model.IsEnabled {
			prediction, err := ae.predictMetric(ctx, model, metricName, timeRange)
			if err != nil {
				log.Printf("Failed to predict metric %s: %v", metricName, err)
				continue
			}
			predictions = append(predictions, *prediction)
		}
	}

	return predictions, nil
}

// predictMetric predicts metric value
func (ae *AnalyticsEngine) predictMetric(ctx context.Context, model *PredictionModel, metricName string, timeRange TimeRange) (*Prediction, error) {
	// Simulate prediction (in real implementation, use ML model)
	predictedValue := ae.simulatePrediction(metricName, model)

	prediction := &Prediction{
		MetricName: metricName,
		Predicted:  predictedValue,
		Confidence: model.Accuracy,
		Horizon:    ae.predictor.Horizon,
		Timestamp:  time.Now(),
		Features:   model.Features,
	}

	return prediction, nil
}

// simulatePrediction simulates ML prediction
func (ae *AnalyticsEngine) simulatePrediction(metricName string, model *PredictionModel) float64 {
	// Simple prediction simulation
	switch metricName {
	case "threat_volume":
		return 120.0 + math.Sin(float64(time.Now().Unix())/3600)*30
	case "resource_usage":
		return 65.0 + math.Sin(float64(time.Now().Unix())/600)*15
	default:
		return 100.0
	}
}

// GetModelStatus returns ML model status
func (ae *AnalyticsEngine) GetModelStatus() map[string]interface{} {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	status := map[string]interface{}{
		"ml_model": map[string]interface{}{
			"loaded":       ae.mlModel.IsLoaded,
			"version":      ae.mlModel.Version,
			"accuracy":     ae.mlModel.Accuracy,
			"last_trained": ae.mlModel.LastTrained,
			"input_size":   ae.mlModel.InputSize,
			"output_size":  ae.mlModel.OutputSize,
		},
		"predictor": map[string]interface{}{
			"horizon":    ae.predictor.Horizon.String(),
			"confidence": ae.predictor.Confidence,
			"models":     len(ae.predictor.Models),
		},
		"aggregator": map[string]interface{}{
			"metrics":      len(ae.aggregator.Metrics),
			"aggregations": len(ae.aggregator.Aggregations),
			"window_sizes": len(ae.aggregator.WindowSizes),
		},
		"insight_generator": map[string]interface{}{
			"rules":     len(ae.insightGenerator.Rules),
			"templates": len(ae.insightGenerator.Templates),
		},
		"timestamp": time.Now(),
	}

	return status
}

// ProcessStream processes real-time stream data
func (ae *AnalyticsEngine) ProcessStream(ctx context.Context, streamID string, events []StreamEvent) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	// Get or create stream
	stream, ok := ae.streamProcessor.ActiveStreams[streamID]
	if !ok {
		stream = &Stream{
			ID:          streamID,
			Events:      make([]*StreamEvent, 0),
			Window:      ae.streamProcessor.Window,
			LastUpdated: time.Now(),
			IsActive:    true,
		}
		ae.streamProcessor.ActiveStreams[streamID] = stream
	}

	// Process events through stream functions
	for _, event := range events {
		for _, processor := range ae.streamProcessor.Processors {
			if !processor.IsEnabled {
				continue
			}

			if processor.Condition(&event) {
				processedEvent := processor.Function(&event)
				if processedEvent.Processed {
					stream.Events = append(stream.Events, processedEvent)
				}
			}
		}
	}

	// Clean old events outside window
	cutoff := time.Now().Add(-ae.streamProcessor.Window)
	filteredEvents := make([]*StreamEvent, 0)
	for _, event := range stream.Events {
		if event.Timestamp.After(cutoff) {
			filteredEvents = append(filteredEvents, event)
		}
	}
	stream.Events = filteredEvents
	stream.LastUpdated = time.Now()

	return nil
}

// GetStreamStatus returns stream processing status
func (ae *AnalyticsEngine) GetStreamStatus() map[string]interface{} {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	status := make(map[string]interface{})
	activeStreams := make(map[string]interface{})

	for streamID, stream := range ae.streamProcessor.ActiveStreams {
		activeStreams[streamID] = map[string]interface{}{
			"events_count": len(stream.Events),
			"last_updated": stream.LastUpdated,
			"is_active":    stream.IsActive,
			"window":       stream.Window.String(),
		}
	}

	status["active_streams"] = activeStreams
	status["processors_count"] = len(ae.streamProcessor.Processors)
	status["buffer_size"] = ae.streamProcessor.BufferSize
	status["timestamp"] = time.Now()

	return status
}

// GetUserMetrics retrieves comprehensive user analytics
func (ae *AnalyticsEngine) GetUserMetrics() (*UserMetrics, error) {
	sqlDB, ok := ae.db.GetConnection().(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN last_login >= NOW() - INTERVAL '24 hours' THEN 1 ELSE 0) as active_users,
			COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 ELSE 0) as new_users_today,
			SUM(CASE WHEN action = 'login_failed' THEN 1 ELSE 0) as failed_logins,
			COUNT(DISTINCT user_id) as unique_visitors,
			SUM(page_views) as page_views,
			AVG(EXTRACT(EPOCH FROM (CASE WHEN last_login IS NOT NULL THEN last_login ELSE created_at))) as avg_session_time
		FROM users 
		WHERE created_at >= NOW() - INTERVAL '30 days'
	`

	var metrics UserMetrics
	err := sqlDB.QueryRow(query).Scan(
		&metrics.TotalUsers, &metrics.ActiveUsers, &metrics.NewUsersToday,
		&metrics.FailedLogins, &metrics.UniqueVisitors, &metrics.PageViews,
		&metrics.AverageSessionTime,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user metrics: %w", err)
	}

	metrics.Timestamp = time.Now()
	return &metrics, nil
}

// GetSystemMetrics retrieves comprehensive system performance metrics
func (ae *AnalyticsEngine) GetSystemMetrics() (*SystemMetrics, error) {
	sqlDB, ok := ae.db.GetConnection().(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT 
			AVG(cpu_usage) as cpu_usage,
			AVG(memory_usage) as memory_usage,
			AVG(disk_usage) as disk_usage,
			SUM(network_in) as network_in,
			SUM(network_out) as network_out,
			COUNT(*) as database_queries,
			AVG(CASE WHEN response_time > 0 THEN response_time ELSE NULL) as avg_response_time,
			(COUNT(CASE WHEN status = 'error' THEN 1 ELSE 0) / COUNT(*) as error_rate,
			SUM(CASE WHEN response_time > 0 THEN response_time ELSE 1000) / SUM(CASE WHEN response_time > 0 THEN response_time ELSE 1000) * 1000 as throughput,
			EXTRACT(EPOCH FROM (SELECT MIN(start_time) FROM system_metrics WHERE start_time >= NOW() - INTERVAL '24 hours')) as uptime_seconds
		FROM system_metrics 
		WHERE timestamp >= NOW() - INTERVAL '24 hours'
	`

	var metrics SystemMetrics
	err := sqlDB.QueryRow(query).Scan(
		&metrics.CPUUsage, &metrics.MemoryUsage, &metrics.DiskUsage,
		&metrics.NetworkIn, &metrics.NetworkOut, &metrics.DatabaseQueries,
		&metrics.ResponseTime, &metrics.ErrorRate, &metrics.Throughput, &metrics.Uptime,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}

	metrics.Timestamp = time.Now()
	return &metrics, nil
}

// GetThreatAnalytics retrieves comprehensive threat intelligence analytics
func (ae *AnalyticsEngine) GetThreatAnalytics() (*ThreatAnalytics, error) {
	sqlDB, ok := ae.db.GetConnection().(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database is not an SQL database")
	}

	// Get top threat types
	threatTypesQuery := `
		SELECT 
			type, 
			COUNT(*) as count, 
			severity,
			CASE WHEN COUNT(*) > LAG(COUNT(*) OVER (ORDER BY detected_at DESC)) * 0.5 THEN 'increasing' ELSE 'stable' END as trend
		FROM threats 
		WHERE detected_at >= NOW() - INTERVAL '30 days'
		GROUP BY type, severity
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err := sqlDB.Query(threatTypesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get threat types: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var threatTypes []ThreatTypeMetric
	for rows.Next() {
		var threatType ThreatTypeMetric
		err := rows.Scan(&threatType.Type, &threatType.Count, &threatType.Severity, &threatType.Trend)
		if err != nil {
			return nil, fmt.Errorf("failed to scan threat type: %w", err)
		}
		threatTypes = append(threatTypes, threatType)
	}

	// Get top threat sources
	threatSourcesQuery := `
		SELECT 
			source, 
			COUNT(*) as count,
			AVG(confidence) as confidence
		FROM threats 
		WHERE detected_at >= NOW() - INTERVAL '30 days'
		GROUP BY source
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err = sqlDB.Query(threatSourcesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get threat sources: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var threatSources []ThreatSourceMetric
	for rows.Next() {
		var threatSource ThreatSourceMetric
		err := rows.Scan(&threatSource.Source, &threatSource.Count, &threatSource.Confidence)
		if err != nil {
			return nil, fmt.Errorf("failed to scan threat source: %w", err)
		}
		threatSources = append(threatSources, threatSource)
	}

	// Calculate detection accuracy
	totalThreats := 0
	truePositives := 0
	for _, threatType := range threatTypes {
		if threatType.Severity == "critical" || threatType.Severity == "high" {
			totalThreats += threatType.Count
		}
	}

	detectionAccuracy := float64(totalThreats-truePositives) / float64(totalThreats)
	if totalThreats == 0 {
		detectionAccuracy = 0
	}

	// Get response time metrics
	responseTimeQuery := `
		SELECT 
			endpoint,
			AVG(CASE WHEN response_time > 0 THEN response_time ELSE NULL) as avg_time,
			PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY response_time) OVER (PARTITION BY endpoint)) as p95_time,
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY response_time) OVER (PARTITION BY endpoint)) as p99_time
		FROM api_logs 
		WHERE timestamp >= NOW() - INTERVAL '24 hours'
		GROUP BY endpoint
	`

	rows, err = sqlDB.Query(responseTimeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get response time metrics: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var responseTimes []ResponseTimeMetric
	for rows.Next() {
		var responseTime ResponseTimeMetric
		err := rows.Scan(&responseTime.Endpoint, &responseTime.AvgTime, &responseTime.P95Time, &responseTime.P99Time)
		if err != nil {
			return nil, fmt.Errorf("failed to scan response time metric: %w", err)
		}
		responseTimes = append(responseTimes, responseTime)
	}

	analytics := &ThreatAnalytics{
		TopThreatTypes:    threatTypes,
		ThreatSources:     threatSources,
		DetectionAccuracy: detectionAccuracy,
		FalsePositives:    truePositives,
		ResponseTimes:     responseTimes,
		Timestamp:         time.Now(),
	}

	return analytics, nil
}

// GenerateReport creates comprehensive analytics reports
func (ae *AnalyticsEngine) GenerateReport(req ReportRequest) (*Report, error) {
	sqlDB, ok := ae.db.GetConnection().(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database is not an SQL database")
	}

	// Generate report based on type
	switch req.ReportType {
	case "security":
		return ae.generateSecurityReport(sqlDB, req)
	case "users":
		return ae.generateUserReport(sqlDB, req)
	case "threats":
		return ae.generateThreatReport(sqlDB, req)
	case "system":
		return ae.generateSystemReport(sqlDB, req)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", req.ReportType)
	}
}

// generateSecurityReport generates security analytics report
func (ae *AnalyticsEngine) generateSecurityReport(sqlDB *sql.DB, req ReportRequest) (*Report, error) {
	metrics, err := ae.GetSecurityMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get security metrics: %w", err)
	}

	// Create report record
	query := `
		INSERT INTO analytics_reports (title, generated_at, format, data)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var reportID int
	err = sqlDB.QueryRow(query, "Security Overview Report", req.Format, metrics).Scan(&reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to create security report: %w", err)
	}

	report := &Report{
		ID:          reportID,
		Type:        req.ReportType,
		Title:       "Security Overview Report",
		GeneratedAt: time.Now(),
		Format:      req.Format,
		Size:        1024, // Estimated size
		Data:        metrics,
	}

	return report, nil
}

// generateUserReport generates user analytics report
func (ae *AnalyticsEngine) generateUserReport(sqlDB *sql.DB, req ReportRequest) (*Report, error) {
	metrics, err := ae.GetUserMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get user metrics: %w", err)
	}

	// Create report record
	query := `
		INSERT INTO analytics_reports (title, generated_at, format, data)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var reportID int
	err = sqlDB.QueryRow(query, "User Analytics Report", req.Format, metrics).Scan(&reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user report: %w", err)
	}

	report := &Report{
		ID:          reportID,
		Type:        req.ReportType,
		Title:       "User Analytics Report",
		GeneratedAt: time.Now(),
		Format:      req.Format,
		Size:        2048, // Estimated size
		Data:        metrics,
	}

	return report, nil
}

// generateThreatReport generates threat intelligence report
func (ae *AnalyticsEngine) generateThreatReport(sqlDB *sql.DB, req ReportRequest) (*Report, error) {
	analytics, err := ae.GetThreatAnalytics()
	if err != nil {
		return nil, fmt.Errorf("failed to get threat analytics: %w", err)
	}

	// Create report record
	query := `
		INSERT INTO analytics_reports (title, generated_at, format, data)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var reportID int
	err = sqlDB.QueryRow(query, "Threat Intelligence Report", req.Format, analytics).Scan(&reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to create threat report: %w", err)
	}

	report := &Report{
		ID:          reportID,
		Type:        req.ReportType,
		Title:       "Threat Intelligence Report",
		GeneratedAt: time.Now(),
		Format:      req.Format,
		Size:        3072, // Estimated size
		Data:        analytics,
	}

	return report, nil
}

// generateSystemReport generates system performance report
func (ae *AnalyticsEngine) generateSystemReport(sqlDB *sql.DB, req ReportRequest) (*Report, error) {
	metrics, err := ae.GetSystemMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}

	// Create report record
	query := `
		INSERT INTO analytics_reports (title, generated_at, format, data)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var reportID int
	err = sqlDB.QueryRow(query, "System Performance Report", req.Format, metrics).Scan(&reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to create system report: %w", err)
	}

	report := &Report{
		ID:          reportID,
		Type:        req.ReportType,
		Title:       "System Performance Report",
		GeneratedAt: time.Now(),
		Format:      req.Format,
		Size:        1536, // Estimated size
		Data:        metrics,
	}

	return report, nil
}

// GetReports retrieves generated analytics reports
func (ae *AnalyticsEngine) GetReports(limit int, offset int) ([]Report, error) {
	sqlDB, ok := ae.db.GetConnection().(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT id, type, title, generated_at, format, size, url
		FROM analytics_reports 
		ORDER BY generated_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := sqlDB.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var reports []Report
	for rows.Next() {
		var report Report
		err := rows.Scan(&report.ID, &report.Type, &report.Title,
			&report.GeneratedAt, &report.Format, &report.Size, &report.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}
		reports = append(reports, report)
	}

	return reports, nil
}
