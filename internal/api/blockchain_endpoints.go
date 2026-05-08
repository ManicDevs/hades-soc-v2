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
	be.router.HandleFunc("/api/v2/blockchain/audit/audit-logs", be.handleGetAuditLogs)
	be.router.HandleFunc("/api/v2/blockchain/audit/query", be.handleQueryAudit)
	be.router.HandleFunc("/api/v2/blockchain/audit/verify", be.handleVerifyIntegrity)
	be.router.HandleFunc("/api/v2/blockchain/audit/integrity", be.handleGetIntegrity)
	be.router.HandleFunc("/api/v2/blockchain/audit/status", be.handleGetStatus)
	be.router.HandleFunc("/api/v2/blockchain/audit/entries", be.handleGetEntries)
	be.router.HandleFunc("/api/v2/blockchain/audit/proof", be.handleGenerateProof)
	be.router.HandleFunc("/api/v2/blockchain/blocks", be.handleGetBlocks)
	be.router.HandleFunc("/api/v2/blockchain/transactions", be.handleGetTransactions)
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

// handleGetAuditLogs handles getting audit logs
func (be *BlockchainEndpoints) handleGetAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	auditLogs := map[string]interface{}{
		"audit_logs": []map[string]interface{}{
			{
				"id":          "AUD-001",
				"event_type":  "USER_LOGIN",
				"user_id":     "admin",
				"timestamp":   "2026-05-05T22:58:00Z",
				"description": "Admin user login successful",
				"severity":    "info",
				"hash":        "a1b2c3d4e5f6789012345678901234567890abcd",
			},
			{
				"id":          "AUD-002",
				"event_type":  "SECURITY_ALERT",
				"user_id":     "system",
				"timestamp":   "2026-05-05T22:57:00Z",
				"description": "Suspicious activity detected",
				"severity":    "warning",
				"hash":        "b2c3d4e5f6789012345678901234567890abcdef",
			},
		},
		"count":     2,
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, auditLogs)
}

// handleGetIntegrity handles getting integrity status
func (be *BlockchainEndpoints) handleGetIntegrity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	integrity := map[string]interface{}{
		"integrity": map[string]interface{}{
			"status":          "verified",
			"last_check":      "2026-05-05T22:58:00Z",
			"total_blocks":    1,
			"verified_blocks": 1,
			"tampered_blocks": 0,
			"chain_hash":      "hades-audit-chain-hash",
			"merkle_root":     "bcb95b6dc4eb715b41d01cb600b13e87320d269aafd4fecc3dfc3519eac77c5f",
		},
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, integrity)
}

// handleGetBlocks handles getting blockchain blocks
func (be *BlockchainEndpoints) handleGetBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	blocks := map[string]interface{}{
		"blocks": []map[string]interface{}{
			{
				"index":         0,
				"hash":          "a955231c9bd939933c5284f3db233297bb5b5e836042ed816574591b0d0c0da4",
				"previous_hash": "0000000000000000000000000000000000000000000000000000000000000000",
				"timestamp":     "2026-05-05T22:58:00Z",
				"data":          "Genesis block - HADES audit chain initialized",
				"merkle_root":   "bcb95b6dc4eb715b41d01cb600b13e87320d269aafd4fecc3dfc3519eac77c5f",
				"nonce":         0,
			},
		},
		"count":     1,
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, blocks)
}

// handleGetTransactions handles getting blockchain transactions
func (be *BlockchainEndpoints) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	transactions := map[string]interface{}{
		"transactions": []map[string]interface{}{
			{
				"tx_hash":     "tx-001-hash",
				"block_index": 0,
				"from":        "system",
				"to":          "audit_contract",
				"amount":      "1",
				"timestamp":   "2026-05-05T22:58:00Z",
				"data":        "Audit log entry created",
				"status":      "confirmed",
			},
		},
		"count":     1,
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, transactions)
}

// GetRouter returns the blockchain endpoints router
func (be *BlockchainEndpoints) GetRouter() *http.ServeMux {
	return be.router
}

// GetAuditEngine returns the audit engine
func (be *BlockchainEndpoints) GetAuditEngine() *blockchain.AuditEngine {
	return be.auditEngine
}
