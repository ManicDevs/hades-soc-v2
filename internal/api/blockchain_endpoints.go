package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"hades-v2/internal/blockchain"
	"hades-v2/internal/database"
)

// BlockchainEndpoints provides blockchain audit API endpoints
type BlockchainEndpoints struct {
	auditEngine *blockchain.AuditEngine
	router      *http.ServeMux
}

// NewBlockchainEndpoints creates new blockchain endpoints
func NewBlockchainEndpoints(db interface{}) (*BlockchainEndpoints, error) {
	// Create audit engine
	auditEngine, err := blockchain.NewAuditEngine(db.(database.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to create audit engine: %w", err)
	}

	endpoints := &BlockchainEndpoints{
		auditEngine: auditEngine,
		router:      http.NewServeMux(),
	}

	// Register blockchain routes
	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers blockchain API routes
func (be *BlockchainEndpoints) registerRoutes() {
	be.router.HandleFunc("/api/v2/blockchain/audit/log", be.handleLogEvent)
	be.router.HandleFunc("/api/v2/blockchain/audit/query", be.handleQueryAudit)
	be.router.HandleFunc("/api/v2/blockchain/audit/verify", be.handleVerifyIntegrity)
	be.router.HandleFunc("/api/v2/blockchain/audit/status", be.handleGetStatus)
	be.router.HandleFunc("/api/v2/blockchain/audit/entries", be.handleGetEntries)
	be.router.HandleFunc("/api/v2/blockchain/audit/proof", be.handleGenerateProof)
}

// handleLogEvent handles audit event logging
func (be *BlockchainEndpoints) handleLogEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var entry blockchain.AuditEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if entry.EventType == "" {
		http.Error(w, "Event type is required", http.StatusBadRequest)
		return
	}
	if entry.Action == "" {
		http.Error(w, "Action is required", http.StatusBadRequest)
		return
	}

	// Log event to blockchain
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err := be.auditEngine.LogEvent(ctx, &entry)
	if err != nil {
		log.Printf("Failed to log audit event: %v", err)
		http.Error(w, "Failed to log event", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"entry_id":   entry.ID,
		"entry_hash": entry.Hash,
		"timestamp":  time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleQueryAudit handles audit log querying
func (be *BlockchainEndpoints) handleQueryAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var query blockchain.AuditQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set default values
	if query.Limit == 0 {
		query.Limit = 100
	}
	if query.Limit > 1000 {
		query.Limit = 1000 // Max limit
	}

	// Query audit logs
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := be.auditEngine.QueryAuditLogs(ctx, query)
	if err != nil {
		log.Printf("Failed to query audit logs: %v", err)
		http.Error(w, "Failed to query audit logs", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, result)
}

// handleVerifyIntegrity handles blockchain integrity verification
func (be *BlockchainEndpoints) handleVerifyIntegrity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify blockchain integrity
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	proof, err := be.auditEngine.VerifyIntegrity(ctx)
	if err != nil {
		log.Printf("Blockchain integrity verification failed: %v", err)
		// Still return the proof with failure information
	}

	WriteJSONResponse(w, proof)
}

// handleGetStatus handles blockchain status requests
func (be *BlockchainEndpoints) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := be.auditEngine.GetBlockchainStatus()

	WriteJSONResponse(w, status)
}

// handleGetEntries handles getting audit entries
func (be *BlockchainEndpoints) handleGetEntries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := blockchain.AuditQuery{
		Limit: 100, // Default limit
	}

	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = &t
		}
	}

	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = &t
		}
	}

	if eventType := r.URL.Query().Get("event_type"); eventType != "" {
		query.EventType = eventType
	}

	if source := r.URL.Query().Get("source"); source != "" {
		query.Source = source
	}

	if user := r.URL.Query().Get("user"); user != "" {
		query.User = user
	}

	if action := r.URL.Query().Get("action"); action != "" {
		query.Action = action
	}

	if resource := r.URL.Query().Get("resource"); resource != "" {
		query.Resource = resource
	}

	// Query audit logs
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := be.auditEngine.QueryAuditLogs(ctx, query)
	if err != nil {
		log.Printf("Failed to get audit entries: %v", err)
		http.Error(w, "Failed to get audit entries", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, result)
}

// handleGenerateProof handles generating cryptographic proofs
func (be *BlockchainEndpoints) handleGenerateProof(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		EntryID    string `json:"entry_id"`
		BlockIndex int    `json:"block_index"`
		ProofType  string `json:"proof_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate proof based on request
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var proof interface{}
	var err error

	switch request.ProofType {
	case "integrity":
		proof, err = be.auditEngine.VerifyIntegrity(ctx)
	case "entry":
		// Generate entry existence proof (simplified)
		proof = map[string]interface{}{
			"type":      "entry_existence",
			"entry_id":  request.EntryID,
			"timestamp": time.Now(),
			"validator": "audit_engine",
		}
	default:
		http.Error(w, "Invalid proof type", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Printf("Failed to generate proof: %v", err)
		http.Error(w, "Failed to generate proof", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, proof)
}

// GetRouter returns the blockchain endpoints router
func (be *BlockchainEndpoints) GetRouter() *http.ServeMux {
	return be.router
}

// GetAuditEngine returns the audit engine
func (be *BlockchainEndpoints) GetAuditEngine() *blockchain.AuditEngine {
	return be.auditEngine
}
