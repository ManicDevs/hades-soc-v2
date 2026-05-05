package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hades-v2/internal/database"
	"hades-v2/internal/threat"
)

// ThreatEndpoints provides threat modeling API endpoints
type ThreatEndpoints struct {
	modelingEngine *threat.ModelingEngine
	router         *http.ServeMux
}

// NewThreatEndpoints creates new threat endpoints
func NewThreatEndpoints(db interface{}) (*ThreatEndpoints, error) {
	// Create modeling engine
	modelingEngine, err := threat.NewModelingEngine(db.(database.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to create modeling engine: %w", err)
	}

	endpoints := &ThreatEndpoints{
		modelingEngine: modelingEngine,
		router:         http.NewServeMux(),
	}

	// Register threat routes
	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers threat API routes
func (te *ThreatEndpoints) registerRoutes() {
	te.router.HandleFunc("/api/v2/threat/models", te.handleGetThreatModels)
	te.router.HandleFunc("/api/v2/threat/scenarios", te.handleGetAttackScenarios)
	te.router.HandleFunc("/api/v2/threat/vulnerabilities", te.handleGetVulnerabilities)
	te.router.HandleFunc("/api/v2/threat/mitigations", te.handleGetMitigations)
	te.router.HandleFunc("/api/v2/threat/simulations", te.handleSimulations)
	te.router.HandleFunc("/api/v2/threat/status", te.handleGetStatus)
}

// handleGetThreatModels handles getting threat models
func (te *ThreatEndpoints) handleGetThreatModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	models := te.modelingEngine.GetThreatModels()

	response := map[string]interface{}{
		"models":    models,
		"count":     len(models),
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetAttackScenarios handles getting attack scenarios
func (te *ThreatEndpoints) handleGetAttackScenarios(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scenarios := te.modelingEngine.GetAttackScenarios()

	response := map[string]interface{}{
		"scenarios": scenarios,
		"count":     len(scenarios),
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetVulnerabilities handles getting vulnerabilities
func (te *ThreatEndpoints) handleGetVulnerabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vulnerabilities := te.modelingEngine.GetVulnerabilities()

	response := map[string]interface{}{
		"vulnerabilities": vulnerabilities,
		"count":           len(vulnerabilities),
		"timestamp":       time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetMitigations handles getting mitigations
func (te *ThreatEndpoints) handleGetMitigations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mitigations := te.modelingEngine.GetMitigations()

	response := map[string]interface{}{
		"mitigations": mitigations,
		"count":       len(mitigations),
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSimulations handles simulation requests (GET and POST)
func (te *ThreatEndpoints) handleSimulations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		te.handleGetSimulations(w, r)
	case http.MethodPost:
		te.handleRunSimulation(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetSimulations handles getting simulations
func (te *ThreatEndpoints) handleGetSimulations(w http.ResponseWriter, r *http.Request) {
	simulations := te.modelingEngine.GetSimulations()

	response := map[string]interface{}{
		"simulations": simulations,
		"count":       len(simulations),
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRunSimulation handles running simulations
func (te *ThreatEndpoints) handleRunSimulation(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ScenarioID string `json:"scenario_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.ScenarioID == "" {
		http.Error(w, "Scenario ID is required", http.StatusBadRequest)
		return
	}

	// Run simulation
	simulation, err := te.modelingEngine.RunSimulation(request.ScenarioID)
	if err != nil {
		http.Error(w, "Failed to run simulation", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":       true,
		"simulation_id": simulation.ID,
		"simulation":    simulation,
		"timestamp":     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetStatus handles getting threat modeling status
func (te *ThreatEndpoints) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := te.modelingEngine.GetEngineStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetRouter returns the threat endpoints router
func (te *ThreatEndpoints) GetRouter() *http.ServeMux {
	return te.router
}

// GetModelingEngine returns the modeling engine
func (te *ThreatEndpoints) GetModelingEngine() *threat.ModelingEngine {
	return te.modelingEngine
}
