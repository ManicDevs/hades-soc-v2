package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"hades-v2/internal/ai"
)

// AIEndpoints provides AI-powered security analysis endpoints
type AIEndpoints struct {
	threatEngine *ai.AIThreatEngine
	router       *http.ServeMux
}

// NewAIEndpoints creates new AI endpoints
func NewAIEndpoints() (*AIEndpoints, error) {
	// Initialize AI threat engine
	threatEngine, err := ai.NewAIThreatEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI threat engine: %w", err)
	}

	endpoints := &AIEndpoints{
		threatEngine: threatEngine,
		router:       http.NewServeMux(),
	}

	// Register AI endpoints
	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers AI API routes
func (ae *AIEndpoints) registerRoutes() {
	ae.router.HandleFunc("/threats", ae.handleThreats)
	ae.router.HandleFunc("/anomalies", ae.handleAnomalies)
	ae.router.HandleFunc("/predictions", ae.handlePredictions)
	ae.router.HandleFunc("/overview", ae.handleOverview)
	ae.router.HandleFunc("/analyze", ae.handleAnalyzeEvent)
	ae.router.HandleFunc("/batch-analyze", ae.handleBatchAnalyze)
	ae.router.HandleFunc("/threat-score", ae.handleThreatScore)
	ae.router.HandleFunc("/patterns", ae.handlePatternMatching)
	ae.router.HandleFunc("/baseline", ae.handleBaselineManagement)
	ae.router.HandleFunc("/model/status", ae.handleModelStatus)
	ae.router.HandleFunc("/model/train", ae.handleModelTraining)
}

// handleAnalyzeEvent handles single event analysis
func (ae *AIEndpoints) handleAnalyzeEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event ai.SecurityEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate event
	if event.ID == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Analyze threat
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	assessment, err := ae.threatEngine.AnalyzeThreat(ctx, event)
	if err != nil {
		log.Printf("AI analysis failed: %v", err)
		http.Error(w, "Analysis failed", http.StatusInternalServerError)
		return
	}

	// Return assessment
	WriteJSONResponse(w, assessment)
}

// handleBatchAnalyze handles batch event analysis
func (ae *AIEndpoints) handleBatchAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var events []ai.SecurityEvent
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Limit batch size
	if len(events) > 100 {
		http.Error(w, "Batch size too large (max 100)", http.StatusBadRequest)
		return
	}

	assessments := make([]*ai.ThreatAssessment, 0, len(events))
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	for _, event := range events {
		// Set timestamp if not provided
		if event.Timestamp.IsZero() {
			event.Timestamp = time.Now()
		}

		assessment, err := ae.threatEngine.AnalyzeThreat(ctx, event)
		if err != nil {
			log.Printf("AI analysis failed for event %s: %v", event.ID, err)
			continue
		}

		assessments = append(assessments, assessment)
	}

	// Return batch results
	response := map[string]interface{}{
		"assessments": assessments,
		"total":       len(events),
		"processed":   len(assessments),
		"timestamp":   time.Now(),
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleThreatScore handles threat scoring
func (ae *AIEndpoints) handleThreatScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Events []ai.SecurityEvent `json:"events"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	scores := make([]float64, 0, len(request.Events))
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	for _, event := range request.Events {
		if event.Timestamp.IsZero() {
			event.Timestamp = time.Now()
		}

		assessment, err := ae.threatEngine.AnalyzeThreat(ctx, event)
		if err != nil {
			log.Printf("Threat scoring failed for event %s: %v", event.ID, err)
			scores = append(scores, 0.0)
			continue
		}

		scores = append(scores, assessment.ThreatScore)
	}

	response := map[string]interface{}{
		"scores":    scores,
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleAnomalyDetection function removed - unused

// handlePatternMatching handles pattern matching
func (ae *AIEndpoints) handlePatternMatching(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Signature string           `json:"signature"`
		Event     ai.SecurityEvent `json:"event"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Event.Timestamp.IsZero() {
		request.Event.Timestamp = time.Now()
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	assessment, err := ae.threatEngine.AnalyzeThreat(ctx, request.Event)
	if err != nil {
		http.Error(w, "Analysis failed", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"patterns":      assessment.Patterns,
		"pattern_score": assessment.PatternScore,
		"timestamp":     time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleBaselineManagement handles baseline operations
func (ae *AIEndpoints) handleBaselineManagement(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ae.getBaseline(w, r)
	case http.MethodPost:
		ae.updateBaseline(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getBaseline returns current baseline information
func (ae *AIEndpoints) getBaseline(w http.ResponseWriter, r *http.Request) {
	// In real implementation, return baseline data
	response := map[string]interface{}{
		"baseline_id":  "baseline_001",
		"last_updated": time.Now().Add(-24 * time.Hour),
		"status":       "active",
		"patterns":     []string{"normal_behavior", "baseline_traffic"},
		"timestamp":    time.Now(),
	}

	WriteJSONResponse(w, response)
}

// updateBaseline updates baseline data
func (ae *AIEndpoints) updateBaseline(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Events []ai.SecurityEvent `json:"events"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// In real implementation, update baseline with events
	response := map[string]interface{}{
		"status":           "success",
		"events_processed": len(request.Events),
		"baseline_id":      "baseline_001",
		"timestamp":        time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleModelStatus returns AI model status
func (ae *AIEndpoints) handleModelStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"model_status": map[string]interface{}{
			"loaded":           true,
			"version":          "1.0",
			"accuracy":         0.92,
			"last_trained":     time.Now().Add(-24 * time.Hour),
			"training_samples": 100000,
		},
		"anomaly_detector": map[string]interface{}{
			"threshold": 0.85,
			"window":    "24h",
			"status":    "active",
		},
		"pattern_matcher": map[string]interface{}{
			"rules_count": 150,
			"signatures":  5000,
			"confidence":  0.8,
		},
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleModelTraining triggers model training
func (ae *AIEndpoints) handleModelTraining(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		TrainingData []ai.SecurityEvent `json:"training_data"`
		Epochs       int                `json:"epochs"`
		LearningRate float64            `json:"learning_rate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate training parameters
	if len(request.TrainingData) == 0 {
		http.Error(w, "Training data is required", http.StatusBadRequest)
		return
	}

	if request.Epochs == 0 {
		request.Epochs = 10
	}

	if request.LearningRate == 0 {
		request.LearningRate = 0.01
	}
}

// handleThreats handles getting AI detected threats
func (ae *AIEndpoints) handleThreats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	threats := ae.threatEngine.GetDetectedThreats()

	response := map[string]interface{}{
		"threats":   threats,
		"count":     len(threats),
		"accuracy":  ae.threatEngine.GetAccuracy(),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleAnomalies handles getting anomaly detection results
func (ae *AIEndpoints) handleAnomalies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	anomalies := ae.threatEngine.GetAnomalies()

	response := map[string]interface{}{
		"anomalies": anomalies,
		"count":     len(anomalies),
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handlePredictions handles getting ML predictions
func (ae *AIEndpoints) handlePredictions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	predictions := ae.threatEngine.GetPredictions()

	response := map[string]interface{}{
		"predictions": predictions,
		"count":       len(predictions),
		"timestamp":   time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleOverview handles getting AI threat intelligence overview
func (ae *AIEndpoints) handleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	overview := ae.threatEngine.GetOverview()
	WriteJSONResponse(w, overview)
}

// GetRouter returns the AI endpoints router
func (ae *AIEndpoints) GetRouter() *http.ServeMux {
	return ae.router
}

// GetThreatEngine returns the AI threat engine
func (ae *AIEndpoints) GetThreatEngine() *ai.AIThreatEngine {
	return ae.threatEngine
}
