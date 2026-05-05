package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hades-v2/internal/database"
	"hades-v2/internal/engine"
	"hades-v2/internal/websocket"
)

// GovernorEndpoints provides Safety Governor API endpoints
type GovernorEndpoints struct {
	dbManager      *database.DatabaseManager
	safetyGovernor *engine.SafetyGovernor
	router         *http.ServeMux
	wsManager      *websocket.WebSocketManager
}

// NewGovernorEndpoints creates new governor endpoints
func NewGovernorEndpoints(db interface{}, wsManager *websocket.WebSocketManager) (*GovernorEndpoints, error) {
	// Type assert to DatabaseManager - we need the specific implementation for governor functionality
	dbManager, ok := db.(*database.DatabaseManager)
	if !ok {
		return nil, fmt.Errorf("database must be of type *database.DatabaseManager")
	}

	// Create safety governor instance
	safetyGovernor := engine.NewSafetyGovernor(dbManager)

	endpoints := &GovernorEndpoints{
		dbManager:      dbManager,
		safetyGovernor: safetyGovernor,
		router:         http.NewServeMux(),
		wsManager:      wsManager,
	}

	// Register governor routes
	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers governor API routes
func (ge *GovernorEndpoints) registerRoutes() {
	ge.router.HandleFunc("/api/v2/governor/stats", ge.handleGetGovernorStats)
	ge.router.HandleFunc("/api/v2/governor/status", ge.handleGetGovernorStatus)
	ge.router.HandleFunc("/api/v2/governor/actions", ge.handleGetRecentActions)
	ge.router.HandleFunc("/api/v2/governor/pending", ge.handleGetPendingActions)
	ge.router.HandleFunc("/api/v2/governor/approve/", ge.handleApproveAction)
}

// handleGetGovernorStats handles getting governor statistics
func (ge *GovernorEndpoints) handleGetGovernorStats(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get governor stats from database
	stats, err := ge.dbManager.GetGovernorStats(ctx)
	if err != nil {
		log.Printf("GovernorEndpoints: Failed to get governor stats: %v", err)
		http.Error(w, "Failed to get governor statistics", http.StatusInternalServerError)
		return
	}

	// Convert to JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("GovernorEndpoints: Failed to encode stats: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleGetGovernorStatus handles getting current governor status
func (ge *GovernorEndpoints) handleGetGovernorStatus(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get governor status
	status := ge.safetyGovernor.GetStatus()

	// Convert to JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("GovernorEndpoints: Failed to encode status: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleGetRecentActions handles getting recent governor actions
func (ge *GovernorEndpoints) handleGetRecentActions(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	sinceStr := r.URL.Query().Get("since")
	limitStr := r.URL.Query().Get("limit")

	var since time.Time
	var limit int = 50 // default limit

	// Parse since parameter
	if sinceStr != "" {
		parsedSince, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			http.Error(w, "Invalid since parameter format, use RFC3339", http.StatusBadRequest)
			return
		}
		since = parsedSince
	} else {
		// Default to last 24 hours
		since = time.Now().Add(-24 * time.Hour)
	}

	// Parse limit parameter
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 || parsedLimit > 1000 {
			http.Error(w, "Invalid limit parameter, must be between 1 and 1000", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	// Get context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get recent actions from database
	actions, err := ge.dbManager.GetRecentActions(ctx, since, limit)
	if err != nil {
		log.Printf("GovernorEndpoints: Failed to get recent actions: %v", err)
		http.Error(w, "Failed to get recent actions", http.StatusInternalServerError)
		return
	}

	// Convert to JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(actions); err != nil {
		log.Printf("GovernorEndpoints: Failed to encode actions: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleGetPendingActions handles getting actions that require manual approval
func (ge *GovernorEndpoints) handleGetPendingActions(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get pending actions from database
	actions, err := ge.dbManager.GetPendingActions(ctx)
	if err != nil {
		log.Printf("GovernorEndpoints: Failed to get pending actions: %v", err)
		http.Error(w, "Failed to get pending actions", http.StatusInternalServerError)
		return
	}

	// Convert to JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(actions); err != nil {
		log.Printf("GovernorEndpoints: Failed to encode pending actions: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleApproveAction handles approving or denying a governor action
func (ge *GovernorEndpoints) handleApproveAction(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract action_id from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v2/governor/approve/")
	if path == "" {
		http.Error(w, "Action ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body to get approval decision
	type ApprovalRequest struct {
		Status string `json:"status"`  // "approved" or "denied"
		UserID string `json:"user_id"` // ID of the analyst making the decision
		Reason string `json:"reason"`  // Optional reason for the decision
	}

	var req ApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate status
	if req.Status != "approved" && req.Status != "denied" {
		http.Error(w, "Status must be 'approved' or 'denied'", http.StatusBadRequest)
		return
	}

	// Get context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Process the approval/denial
	result, err := ge.processActionApproval(ctx, path, req.Status, req.UserID, req.Reason)
	if err != nil {
		log.Printf("GovernorEndpoints: Failed to process action approval: %v", err)
		http.Error(w, "Failed to process approval", http.StatusInternalServerError)
		return
	}

	// Return the result
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("GovernorEndpoints: Failed to encode approval result: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// processActionApproval handles the business logic for approving/denying actions
func (ge *GovernorEndpoints) processActionApproval(ctx context.Context, actionID, status, userID, reason string) (map[string]interface{}, error) {
	// Get the action from database
	action, err := ge.dbManager.GetGovernorActionByID(ctx, actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get action: %w", err)
	}

	if action == nil {
		return nil, fmt.Errorf("action not found")
	}

	// Check if action is in a state that can be approved/denied
	if action.Status != string(database.GovernorActionStatusManualAckRequired) &&
		action.Status != string(database.GovernorActionStatusPending) {
		return nil, fmt.Errorf("action is not in a pending state")
	}

	// Update the action status
	var newStatus database.GovernorActionStatus
	var blockReason string

	if status == "approved" {
		newStatus = database.GovernorActionStatusApproved
		blockReason = ""

		// Execute the original action if approved
		if err := ge.executeApprovedAction(ctx, action); err != nil {
			log.Printf("GovernorEndpoints: Failed to execute approved action: %v", err)
			// Continue with approval even if execution fails
		}
	} else {
		newStatus = database.GovernorActionStatusBlocked
		blockReason = fmt.Sprintf("Manually denied by user %s: %s", userID, reason)
	}

	// Update the action in database
	updatedAction := &database.GovernorAction{
		ID:                action.ID,
		ActionID:          action.ActionID,
		ActionName:        action.ActionName,
		Target:            action.Target,
		Reasoning:         action.Reasoning,
		Requester:         action.Requester,
		Status:            string(newStatus),
		RequiresApproval:  action.RequiresApproval,
		Approved:          status == "approved",
		RequiresManualAck: false, // Clear manual ack requirement
		BlockReason:       blockReason,
		ExecutionTime:     action.ExecutionTime,
		Metadata:          action.Metadata,
		CreatedAt:         action.CreatedAt,
		UpdatedAt:         time.Now(),
	}

	if err := ge.dbManager.UpdateGovernorAction(ctx, updatedAction); err != nil {
		return nil, fmt.Errorf("failed to update action: %w", err)
	}

	// TODO: Implement audit logging
	// Log the audit trail for compliance
	log.Printf("GovernorEndpoints: Action %s %s by user %s: %s", actionID, status, userID, reason)
	// TODO: Get actual user ID from auth context and IP address from request context

	// Publish WebSocket event for dashboard
	ge.publishApprovalEvent(actionID, status, userID, reason)

	return map[string]interface{}{
		"action_id":    actionID,
		"status":       status,
		"action_name":  action.ActionName,
		"target":       action.Target,
		"processed_by": userID,
		"timestamp":    time.Now().Unix(),
	}, nil
}

// executeApprovedAction executes the original action that was approved
func (ge *GovernorEndpoints) executeApprovedAction(_ context.Context, action *database.GovernorAction) error {
	// TODO: Implement actual action execution based on action type
	// This would involve calling the appropriate module/function based on action.ActionName
	log.Printf("GovernorEndpoints: Executing approved action '%s' on target '%s'", action.ActionName, action.Target)

	// For now, just log the execution
	// In a real implementation, this would:
	// 1. Parse the action metadata to determine what to execute
	// 2. Call the appropriate function/module
	// 3. Handle execution results and errors
	// 4. Update the action with execution results

	return nil
}

// publishApprovalEvent publishes WebSocket event for approval status change
func (ge *GovernorEndpoints) publishApprovalEvent(actionID, status, userID, reason string) {
	if ge.wsManager == nil {
		log.Printf("GovernorEndpoints: WebSocket manager not available, skipping approval event broadcast")
		return
	}

	// Create approval status update message
	approvalMessage := websocket.WebSocketMessage{
		Data: map[string]interface{}{
			"action_id":    actionID,
			"status":       status,
			"processed_by": userID,
			"reason":       reason,
			"timestamp":    time.Now().Unix(),
		},
	}

	// Broadcast to all connected clients
	ge.wsManager.BroadcastUpdate("governor_approval", "governor_actions", approvalMessage.Data)
	log.Printf("GovernorEndpoints: Published approval event for action %s: %s by %s", actionID, status, userID)
}

// GetRouter returns the router for governor endpoints
func (ge *GovernorEndpoints) GetRouter() *http.ServeMux {
	return ge.router
}

// Start starts the safety governor monitoring
func (ge *GovernorEndpoints) Start() {
	ge.safetyGovernor.Start()
}

// Stop stops the safety governor monitoring
func (ge *GovernorEndpoints) Stop() {
	ge.safetyGovernor.Stop()
}
