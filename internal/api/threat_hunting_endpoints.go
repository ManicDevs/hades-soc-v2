package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"hades-v2/internal/ai"
	"hades-v2/internal/analytics"
	"hades-v2/internal/threathunting"
)

// ThreatHuntingEndpoints provides threat hunting API endpoints
type ThreatHuntingEndpoints struct {
	threatHuntingEngine *threathunting.ThreatHuntingEngine
	router              *http.ServeMux
}

// NewThreatHuntingEndpoints creates new threat hunting endpoints
func NewThreatHuntingEndpoints(aiEngine interface{}, analyticsEngine interface{}) (*ThreatHuntingEndpoints, error) {
	// Create threat hunting engine
	threatHuntingEngine, err := threathunting.NewThreatHuntingEngine(
		aiEngine.(*ai.AIThreatEngine),
		analyticsEngine.(*analytics.AnalyticsEngine),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create threat hunting engine: %w", err)
	}

	endpoints := &ThreatHuntingEndpoints{
		threatHuntingEngine: threatHuntingEngine,
		router:              http.NewServeMux(),
	}

	// Register threat hunting routes
	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers threat hunting API routes
func (the *ThreatHuntingEndpoints) registerRoutes() {
	the.router.HandleFunc("/api/v2/threat-hunting/hunts", the.handleHunts)
	the.router.HandleFunc("/api/v2/threat-hunting/hunts/start", the.handleStartHunt)
	the.router.HandleFunc("/api/v2/threat-hunting/hunts/{id}", the.handleGetHunt)
	the.router.HandleFunc("/api/v2/threat-hunting/strategies", the.handleGetStrategies)
	the.router.HandleFunc("/api/v2/threat-hunting/intelligence", the.handleGetThreatIntel)
	the.router.HandleFunc("/api/v2/threat-hunting/automation/status", the.handleGetAutomationStatus)
	the.router.HandleFunc("/api/v2/threat-hunting/findings", the.handleGetFindings)
	the.router.HandleFunc("/api/v2/threat-hunting/artifacts", the.handleGetArtifacts)
}

// handleHunts handles hunts management
func (the *ThreatHuntingEndpoints) handleHunts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		the.getActiveHunts(w, r)
	case http.MethodPost:
		the.createHunt(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getActiveHunts returns all active hunts
func (the *ThreatHuntingEndpoints) getActiveHunts(w http.ResponseWriter, r *http.Request) {
	activeHunts := the.threatHuntingEngine.GetActiveHunts()

	response := map[string]interface{}{
		"active_hunts": activeHunts,
		"count":        len(activeHunts),
		"timestamp":    time.Now(),
	}

	WriteJSONResponse(w, response)
}

// createHunt creates a new hunt
func (the *ThreatHuntingEndpoints) createHunt(w http.ResponseWriter, r *http.Request) {
	var request struct {
		StrategyID string                 `json:"strategy_id"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.StrategyID == "" {
		http.Error(w, "Strategy ID is required", http.StatusBadRequest)
		return
	}

	// Start hunt
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	hunt, err := the.threatHuntingEngine.StartHunt(ctx, request.StrategyID)
	if err != nil {
		log.Printf("Failed to start hunt: %v", err)
		http.Error(w, "Failed to start hunt", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"hunt":      hunt,
		"status":    "started",
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleStartHunt handles starting a new hunt
func (the *ThreatHuntingEndpoints) handleStartHunt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		StrategyID string `json:"strategy_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.StrategyID == "" {
		http.Error(w, "Strategy ID is required", http.StatusBadRequest)
		return
	}

	// Start hunt
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	hunt, err := the.threatHuntingEngine.StartHunt(ctx, request.StrategyID)
	if err != nil {
		log.Printf("Failed to start hunt: %v", err)
		http.Error(w, "Failed to start hunt", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"hunt":      hunt,
		"status":    "started",
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetHunt handles getting a specific hunt
func (the *ThreatHuntingEndpoints) handleGetHunt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get hunt ID from URL path
	huntID := r.URL.Path[len("/api/v2/threat-hunting/hunts/"):]
	if huntID == "" {
		http.Error(w, "Hunt ID is required", http.StatusBadRequest)
		return
	}

	// Get active hunts
	activeHunts := the.threatHuntingEngine.GetActiveHunts()
	hunt, ok := activeHunts[huntID]
	if !ok {
		http.Error(w, "Hunt not found", http.StatusNotFound)
		return
	}

	WriteJSONResponse(w, hunt)
}

// handleGetStrategies handles getting hunting strategies
func (the *ThreatHuntingEndpoints) handleGetStrategies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	strategies := the.threatHuntingEngine.GetHuntStrategies()

	response := map[string]interface{}{
		"strategies": strategies,
		"count":      len(strategies),
		"timestamp":  time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetThreatIntel handles getting threat intelligence
func (the *ThreatHuntingEndpoints) handleGetThreatIntel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	threatIntel := the.threatHuntingEngine.GetThreatIntelligence()

	WriteJSONResponse(w, threatIntel)
}

// handleGetAutomationStatus handles getting automation status
func (the *ThreatHuntingEndpoints) handleGetAutomationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := the.threatHuntingEngine.GetAutomationStatus()

	WriteJSONResponse(w, status)
}

// handleGetFindings handles getting hunt findings
func (the *ThreatHuntingEndpoints) handleGetFindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all findings from active hunts
	activeHunts := the.threatHuntingEngine.GetActiveHunts()
	allFindings := make([]interface{}, 0)

	for _, hunt := range activeHunts {
		for _, finding := range hunt.Findings {
			allFindings = append(allFindings, finding)
		}
	}

	response := map[string]interface{}{
		"findings":  allFindings,
		"count":     len(allFindings),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetArtifacts handles getting hunt artifacts
func (the *ThreatHuntingEndpoints) handleGetArtifacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all artifacts from active hunts
	activeHunts := the.threatHuntingEngine.GetActiveHunts()
	allArtifacts := make([]interface{}, 0)

	for _, hunt := range activeHunts {
		for _, artifact := range hunt.Artifacts {
			allArtifacts = append(allArtifacts, artifact)
		}
	}

	response := map[string]interface{}{
		"artifacts": allArtifacts,
		"count":     len(allArtifacts),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// GetRouter returns the threat hunting endpoints router
func (the *ThreatHuntingEndpoints) GetRouter() *http.ServeMux {
	return the.router
}

// GetThreatHuntingEngine returns the threat hunting engine
func (the *ThreatHuntingEndpoints) GetThreatHuntingEngine() *threathunting.ThreatHuntingEngine {
	return the.threatHuntingEngine
}
