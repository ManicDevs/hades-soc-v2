package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hades-v2/internal/database"
	"hades-v2/internal/zerotrust"
)

// ZeroTrustEndpoints provides zero-trust network API endpoints
type ZeroTrustEndpoints struct {
	zeroTrustEngine *zerotrust.ZeroTrustEngine
	router          *http.ServeMux
}

// NewZeroTrustEndpoints creates new zero-trust endpoints
func NewZeroTrustEndpoints(db interface{}) (*ZeroTrustEndpoints, error) {
	// Create zero-trust engine
	zeroTrustEngine, err := zerotrust.NewZeroTrustEngine(db.(database.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to create zero-trust engine: %w", err)
	}

	endpoints := &ZeroTrustEndpoints{
		zeroTrustEngine: zeroTrustEngine,
		router:          http.NewServeMux(),
	}

	// Register zero-trust routes
	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers zero-trust API routes
func (zte *ZeroTrustEndpoints) registerRoutes() {
	zte.router.HandleFunc("/api/v2/zerotrust/access/evaluate", zte.handleEvaluateAccess)
	zte.router.HandleFunc("/api/v2/zerotrust/devices/register", zte.handleRegisterDevice)
	zte.router.HandleFunc("/api/v2/zerotrust/sessions/create", zte.handleCreateSession)
	zte.router.HandleFunc("/api/v2/zerotrust/sessions/validate", zte.handleValidateSession)
	zte.router.HandleFunc("/api/v2/zerotrust/segments", zte.handleGetSegments)
	zte.router.HandleFunc("/api/v2/zerotrust/devices", zte.handleGetDevices)
	zte.router.HandleFunc("/api/v2/zerotrust/policies", zte.handleGetPolicies)
	zte.router.HandleFunc("/api/v2/zerotrust/trust/status", zte.handleGetTrustStatus)
}

// handleEvaluateAccess handles access evaluation requests
func (zte *ZeroTrustEndpoints) handleEvaluateAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request zerotrust.AccessRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.DeviceID == "" {
		http.Error(w, "Device ID is required", http.StatusBadRequest)
		return
	}
	if request.Resource == "" {
		http.Error(w, "Resource is required", http.StatusBadRequest)
		return
	}

	// Generate request ID if not provided
	if request.ID == "" {
		request.ID = fmt.Sprintf("access_req_%d", time.Now().UnixNano())
	}
	request.Timestamp = time.Now()

	// Evaluate access
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	decision, err := zte.zeroTrustEngine.EvaluateAccess(ctx, request)
	if err != nil {
		http.Error(w, "Failed to evaluate access", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, decision)
}

// handleRegisterDevice handles device registration
func (zte *ZeroTrustEndpoints) handleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var device zerotrust.Device
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if device.Name == "" {
		http.Error(w, "Device name is required", http.StatusBadRequest)
		return
	}
	if device.IPAddress == "" {
		http.Error(w, "IP address is required", http.StatusBadRequest)
		return
	}

	// Register device
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err := zte.zeroTrustEngine.RegisterDevice(ctx, &device)
	if err != nil {
		http.Error(w, "Failed to register device", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"device_id":   device.ID,
		"trust_score": device.TrustScore,
		"risk_level":  device.RiskLevel,
		"timestamp":   time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleCreateSession handles session creation
func (zte *ZeroTrustEndpoints) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		DeviceID string        `json:"device_id"`
		UserID   string        `json:"user_id"`
		Duration time.Duration `json:"duration"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.DeviceID == "" {
		http.Error(w, "Device ID is required", http.StatusBadRequest)
		return
	}
	if request.UserID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Set default duration if not provided
	if request.Duration == 0 {
		request.Duration = 8 * time.Hour // 8 hours default
	}

	// Create session
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	session, err := zte.zeroTrustEngine.CreateSession(ctx, request.DeviceID, request.UserID, request.Duration)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"session_id":  session.ID,
		"expires_at":  session.ExpiresAt,
		"trust_level": session.TrustLevel,
		"timestamp":   time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleValidateSession handles session validation
func (zte *ZeroTrustEndpoints) handleValidateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		SessionID string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.SessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// Validate session
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	session, err := zte.zeroTrustEngine.ValidateSession(ctx, request.SessionID)
	if err != nil {
		response := map[string]interface{}{
			"valid":     false,
			"error":     err.Error(),
			"timestamp": time.Now(),
		}
		WriteJSONResponse(w, response)
		return
	}

	response := map[string]interface{}{
		"valid":     true,
		"session":   session,
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetSegments handles getting network segments
func (zte *ZeroTrustEndpoints) handleGetSegments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	segments := zte.zeroTrustEngine.GetNetworkSegments()

	response := map[string]interface{}{
		"segments":  segments,
		"count":     len(segments),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetDevices handles getting devices
func (zte *ZeroTrustEndpoints) handleGetDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devices := zte.zeroTrustEngine.GetDevices()

	response := map[string]interface{}{
		"devices":   devices,
		"count":     len(devices),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetPolicies handles getting policies
func (zte *ZeroTrustEndpoints) handleGetPolicies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	policies := zte.zeroTrustEngine.GetPolicies()

	response := map[string]interface{}{
		"policies":  policies,
		"count":     len(policies),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetTrustStatus handles getting trust engine status
func (zte *ZeroTrustEndpoints) handleGetTrustStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := zte.zeroTrustEngine.GetTrustEngineStatus()

	WriteJSONResponse(w, status)
}

// GetRouter returns the zero-trust endpoints router
func (zte *ZeroTrustEndpoints) GetRouter() *http.ServeMux {
	return zte.router
}

// GetZeroTrustEngine returns the zero-trust engine
func (zte *ZeroTrustEndpoints) GetZeroTrustEngine() *zerotrust.ZeroTrustEngine {
	return zte.zeroTrustEngine
}
