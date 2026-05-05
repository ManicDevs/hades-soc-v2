package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"hades-v2/internal/incident"
)

// IncidentEndpoints provides incident response API endpoints
type IncidentEndpoints struct {
	incidentManager *incident.IncidentResponseManager
	responseEngine  *incident.ResponseEngine
	router          *http.ServeMux
}

// NewIncidentEndpoints creates new incident endpoints
func NewIncidentEndpoints(db interface{}) (*IncidentEndpoints, error) {
	// Create incident response manager
	incidentManager := incident.NewIncidentResponseManager(nil) // Pass nil for threat detector for now

	endpoints := &IncidentEndpoints{
		incidentManager: incidentManager,
		router:          http.NewServeMux(),
	}

	// Register incident routes
	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers incident API routes
func (ie *IncidentEndpoints) registerRoutes() {
	ie.router.HandleFunc("/api/v2/incident/playbooks", ie.handleGetPlaybooks)
	ie.router.HandleFunc("/api/v2/incident/incidents", ie.handleIncidents)
	ie.router.HandleFunc("/api/v2/incident/actions", ie.handleGetResponseActions)
	ie.router.HandleFunc("/api/v2/incident/status", ie.handleGetStatus)
}

// handleGetPlaybooks handles getting playbooks
func (ie *IncidentEndpoints) handleGetPlaybooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	playbooks := []incident.Workflow{} // Return empty workflows for now

	response := map[string]interface{}{
		"playbooks": playbooks,
		"count":     len(playbooks),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleIncidents handles incident requests (GET and POST)
func (ie *IncidentEndpoints) handleIncidents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ie.handleGetIncidents(w, r)
	case http.MethodPost:
		ie.handleCreateIncident(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetIncidents handles getting incidents
func (ie *IncidentEndpoints) handleGetIncidents(w http.ResponseWriter, _ *http.Request) {
	incidents := ie.incidentManager.GetActiveIncidents()

	response := map[string]interface{}{
		"incidents": incidents,
		"count":     len(incidents),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleCreateIncident handles creating incidents
func (ie *IncidentEndpoints) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	var incident incident.Incident
	if err := json.NewDecoder(r.Body).Decode(&incident); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if incident.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if incident.Description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}

	// Create incident (simplified for now)
	// TODO: Implement proper incident creation
	// For now, just return success without actual processing
	_ = context.Background()

	response := map[string]interface{}{
		"success":     true,
		"incident_id": incident.ID,
		"timestamp":   time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetResponseActions handles getting response actions
func (ie *IncidentEndpoints) handleGetResponseActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actions := []string{} // Return empty actions for now

	response := map[string]interface{}{
		"actions":   actions,
		"count":     len(actions),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetStatus handles getting engine status
func (ie *IncidentEndpoints) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]interface{}{
		"active":    true,
		"incidents": 0,
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// GetRouter returns the incident endpoints router
func (ie *IncidentEndpoints) GetRouter() *http.ServeMux {
	return ie.router
}

// GetResponseEngine returns the response engine
func (ie *IncidentEndpoints) GetResponseEngine() *incident.ResponseEngine {
	return ie.responseEngine
}
