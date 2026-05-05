package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"hades-v2/internal/analytics" // Import for response helper
	"hades-v2/internal/database"
)

// AnalyticsEndpoints provides advanced analytics API endpoints
type AnalyticsEndpoints struct {
	analyticsEngine *analytics.AnalyticsEngine
	router          *http.ServeMux
}

// NewAnalyticsEndpoints creates new analytics endpoints
func NewAnalyticsEndpoints(db interface{}) (*AnalyticsEndpoints, error) {
	// Create analytics engine
	analyticsEngine, err := analytics.NewAnalyticsEngine(db.(database.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to create analytics engine: %w", err)
	}

	endpoints := &AnalyticsEndpoints{
		analyticsEngine: analyticsEngine,
		router:          http.NewServeMux(),
	}

	// Register analytics routes
	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers analytics API routes
func (ae *AnalyticsEndpoints) registerRoutes() {
	ae.router.HandleFunc("/api/v2/analytics/query", ae.handleAnalyticsQuery)
	ae.router.HandleFunc("/api/v2/analytics/metrics", ae.handleGetMetrics)
	ae.router.HandleFunc("/api/v2/analytics/insights", ae.handleGetInsights)
	ae.router.HandleFunc("/api/v2/analytics/predictions", ae.handleGetPredictions)
	ae.router.HandleFunc("/api/v2/analytics/model/status", ae.handleModelStatus)
	ae.router.HandleFunc("/api/v2/analytics/model/train", ae.handleModelTraining)
	ae.router.HandleFunc("/api/v2/analytics/stream", ae.handleStreamProcessing)
	ae.router.HandleFunc("/api/v2/analytics/security", ae.handleSecurityMetrics)
	ae.router.HandleFunc("/api/v2/analytics/users", ae.handleUserMetrics)
}

// handleAnalyticsQuery handles analytics queries
func (ae *AnalyticsEndpoints) handleAnalyticsQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request analytics.AnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(request.Metrics) == 0 {
		http.Error(w, "At least one metric is required", http.StatusBadRequest)
		return
	}

	// Set default time range if not provided
	if request.TimeRange.Start.IsZero() {
		request.TimeRange.Start = time.Now().Add(-24 * time.Hour)
	}
	if request.TimeRange.End.IsZero() {
		request.TimeRange.End = time.Now()
	}

	// Process analytics query
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	response, err := ae.analyticsEngine.ProcessAnalytics(ctx, request)
	if err != nil {
		log.Printf("Analytics query failed: %v", err)
		http.Error(w, "Analytics query failed", http.StatusInternalServerError)
		return
	}

	// Return response
	WriteJSONResponse(w, response)
}

// handleGetMetrics handles metrics retrieval
func (ae *AnalyticsEndpoints) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	metrics := r.URL.Query()["metric"]
	if len(metrics) == 0 {
		metrics = []string{"threat_count", "response_time", "cpu_usage"}
	}

	// Create analytics request
	request := analytics.AnalyticsRequest{
		Metrics: metrics,
		TimeRange: analytics.TimeRange{
			Start: time.Now().Add(-1 * time.Hour),
			End:   time.Now(),
		},
		Granularity: "1m",
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	response, err := ae.analyticsEngine.ProcessAnalytics(ctx, request)
	if err != nil {
		http.Error(w, "Failed to get metrics", http.StatusInternalServerError)
		return
	}

	// Return only results
	WriteJSONResponse(w, map[string]interface{}{
		"metrics":   response.Results,
		"timestamp": response.Timestamp,
	})
}

// handleGetInsights handles insights retrieval
func (ae *AnalyticsEndpoints) handleGetInsights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create analytics request for insights
	request := analytics.AnalyticsRequest{
		Metrics: []string{"threat_count", "cpu_usage", "response_time"},
		TimeRange: analytics.TimeRange{
			Start: time.Now().Add(-1 * time.Hour),
			End:   time.Now(),
		},
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	response, err := ae.analyticsEngine.ProcessAnalytics(ctx, request)
	if err != nil {
		http.Error(w, "Failed to get insights", http.StatusInternalServerError)
		return
	}

	// Return only insights
	WriteJSONResponse(w, map[string]interface{}{
		"insights":  response.Insights,
		"timestamp": response.Timestamp,
	})
}

// handleGetPredictions handles predictions retrieval
func (ae *AnalyticsEndpoints) handleGetPredictions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create analytics request for predictions
	request := analytics.AnalyticsRequest{
		Metrics: []string{"threat_volume", "resource_usage"},
		TimeRange: analytics.TimeRange{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
		},
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	response, err := ae.analyticsEngine.ProcessAnalytics(ctx, request)
	if err != nil {
		http.Error(w, "Failed to get predictions", http.StatusInternalServerError)
		return
	}

	// Return only predictions
	WriteJSONResponse(w, map[string]interface{}{
		"predictions": response.Predictions,
		"timestamp":   response.Timestamp,
	})
}

// handleModelStatus handles model status requests
func (ae *AnalyticsEndpoints) handleModelStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := ae.analyticsEngine.GetModelStatus()
	WriteJSONResponse(w, status)
}

// handleModelTraining handles model training requests
func (ae *AnalyticsEndpoints) handleModelTraining(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ModelName    string                   `json:"model_name"`
		TrainingData []map[string]interface{} `json:"training_data"`
		Epochs       int                      `json:"epochs"`
		LearningRate float64                  `json:"learning_rate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate training parameters
	if request.ModelName == "" {
		http.Error(w, "Model name is required", http.StatusBadRequest)
		return
	}

	if len(request.TrainingData) == 0 {
		http.Error(w, "Training data is required", http.StatusBadRequest)
		return
	}

	// In real implementation, trigger async training
	response := map[string]interface{}{
		"training_id":    fmt.Sprintf("train_%d", time.Now().Unix()),
		"status":         "started",
		"model_name":     request.ModelName,
		"data_samples":   len(request.TrainingData),
		"estimated_time": "10m",
		"timestamp":      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStreamProcessing handles stream processing requests
func (ae *AnalyticsEndpoints) handleStreamProcessing(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ae.getStreamStatus(w, r)
	case http.MethodPost:
		ae.processStream(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getStreamStatus returns stream processing status
func (ae *AnalyticsEndpoints) getStreamStatus(w http.ResponseWriter, r *http.Request) {
	status := ae.analyticsEngine.GetStreamStatus()
	WriteJSONResponse(w, status)
}

// processStream processes stream data
func (ae *AnalyticsEndpoints) processStream(w http.ResponseWriter, r *http.Request) {
	var request struct {
		StreamID string                  `json:"stream_id"`
		Events   []analytics.StreamEvent `json:"events"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.StreamID == "" {
		http.Error(w, "Stream ID is required", http.StatusBadRequest)
		return
	}

	if len(request.Events) == 0 {
		http.Error(w, "At least one event is required", http.StatusBadRequest)
		return
	}

	// Process stream
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err := ae.analyticsEngine.ProcessStream(ctx, request.StreamID, request.Events)
	if err != nil {
		log.Printf("Stream processing failed: %v", err)
		http.Error(w, "Stream processing failed", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":           "success",
		"stream_id":        request.StreamID,
		"events_processed": len(request.Events),
		"timestamp":        time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleSecurityMetrics handles security metrics requests
func (ae *AnalyticsEndpoints) handleSecurityMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get security metrics from analytics engine
	metrics, err := ae.analyticsEngine.GetSecurityMetrics()
	if err != nil {
		log.Printf("Failed to get security metrics: %v", err)
		http.Error(w, "Failed to get security metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleUserMetrics handles user metrics requests
func (ae *AnalyticsEndpoints) handleUserMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user metrics from analytics engine
	metrics, err := ae.analyticsEngine.GetUserMetrics()
	if err != nil {
		log.Printf("Failed to get user metrics: %v", err)
		http.Error(w, "Failed to get user metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetRouter returns the analytics endpoints router
func (ae *AnalyticsEndpoints) GetRouter() *http.ServeMux {
	return ae.router
}

// GetAnalyticsEngine returns the analytics engine
func (ae *AnalyticsEndpoints) GetAnalyticsEngine() *analytics.AnalyticsEngine {
	return ae.analyticsEngine
}
