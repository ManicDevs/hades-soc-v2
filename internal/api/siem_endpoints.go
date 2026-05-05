package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hades-v2/internal/database"
	"hades-v2/internal/siem"
)

// SIEMEndpoints provides SIEM API endpoints
type SIEMEndpoints struct {
	siemEngine *siem.SIEMEngine
	router     *http.ServeMux
}

// NewSIEMEndpoints creates new SIEM endpoints
func NewSIEMEndpoints(db interface{}) (*SIEMEndpoints, error) {
	// Create SIEM engine
	siemEngine, err := siem.NewSIEMEngine(db.(database.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to create SIEM engine: %w", err)
	}

	endpoints := &SIEMEndpoints{
		siemEngine: siemEngine,
		router:     http.NewServeMux(),
	}

	// Register SIEM routes
	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers SIEM API routes
func (se *SIEMEndpoints) registerRoutes() {
	se.router.HandleFunc("/api/v2/siem/collectors", se.handleGetCollectors)
	se.router.HandleFunc("/api/v2/siem/rules", se.handleGetRules)
	se.router.HandleFunc("/api/v2/siem/threat-feeds", se.handleGetThreatFeeds)
	se.router.HandleFunc("/api/v2/siem/alerts", se.handleGetAlerts)
	se.router.HandleFunc("/api/v2/siem/incidents", se.handleGetIncidents)
	se.router.HandleFunc("/api/v2/siem/events", se.handleProcessEvent)
	se.router.HandleFunc("/api/v2/siem/status", se.handleGetStatus)
}

// handleGetCollectors handles getting collectors
func (se *SIEMEndpoints) handleGetCollectors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	collectors := se.siemEngine.GetCollectors()

	response := map[string]interface{}{
		"collectors": collectors,
		"count":      len(collectors),
		"timestamp":  time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetRules handles getting rules
func (se *SIEMEndpoints) handleGetRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rules := se.siemEngine.GetRules()

	response := map[string]interface{}{
		"rules":     rules,
		"count":     len(rules),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetThreatFeeds handles getting threat feeds
func (se *SIEMEndpoints) handleGetThreatFeeds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	threatFeeds := se.siemEngine.GetThreatFeeds()

	response := map[string]interface{}{
		"threat_feeds": threatFeeds,
		"count":        len(threatFeeds),
		"timestamp":    time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetAlerts handles getting alerts
func (se *SIEMEndpoints) handleGetAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	alerts := se.siemEngine.GetAlerts()

	response := map[string]interface{}{
		"alerts":    alerts,
		"count":     len(alerts),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetIncidents handles getting incidents
func (se *SIEMEndpoints) handleGetIncidents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	incidents := se.siemEngine.GetIncidents()

	response := map[string]interface{}{
		"incidents": incidents,
		"count":     len(incidents),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleProcessEvent handles processing events
func (se *SIEMEndpoints) handleProcessEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event siem.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if event.EventType == "" {
		http.Error(w, "Event type is required", http.StatusBadRequest)
		return
	}
	if event.Source == "" {
		http.Error(w, "Source is required", http.StatusBadRequest)
		return
	}

	// Process event
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err := se.siemEngine.ProcessEvent(ctx, &event)
	if err != nil {
		http.Error(w, "Failed to process event", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"event_id":  event.ID,
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetStatus handles getting SIEM status
func (se *SIEMEndpoints) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := se.siemEngine.GetEngineStatus()

	WriteJSONResponse(w, status)
}

// GetRouter returns the SIEM endpoints router
func (se *SIEMEndpoints) GetRouter() *http.ServeMux {
	return se.router
}

// GetSIEMEngine returns the SIEM engine
func (se *SIEMEndpoints) GetSIEMEngine() *siem.SIEMEngine {
	return se.siemEngine
}
