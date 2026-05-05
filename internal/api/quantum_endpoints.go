package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hades-v2/internal/database"
	"hades-v2/internal/quantum"
)

// QuantumEndpoints provides quantum cryptography API endpoints
type QuantumEndpoints struct {
	cryptographyEngine *quantum.CryptographyEngine
	router             *http.ServeMux
}

// NewQuantumEndpoints creates new quantum endpoints
func NewQuantumEndpoints(db interface{}) (*QuantumEndpoints, error) {
	// Create cryptography engine
	cryptographyEngine, err := quantum.NewCryptographyEngine(db.(database.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to create cryptography engine: %w", err)
	}

	endpoints := &QuantumEndpoints{
		cryptographyEngine: cryptographyEngine,
		router:             http.NewServeMux(),
	}

	// Register quantum routes
	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers quantum API routes
func (qe *QuantumEndpoints) registerRoutes() {
	qe.router.HandleFunc("/api/v2/quantum/algorithms", qe.handleGetAlgorithms)
	qe.router.HandleFunc("/api/v2/quantum/keys/generate", qe.handleGenerateKey)
	qe.router.HandleFunc("/api/v2/quantum/keys", qe.handleGetKeys)
	qe.router.HandleFunc("/api/v2/quantum/encrypt", qe.handleEncrypt)
	qe.router.HandleFunc("/api/v2/quantum/decrypt", qe.handleDecrypt)
	qe.router.HandleFunc("/api/v2/quantum/sign", qe.handleSign)
	qe.router.HandleFunc("/api/v2/quantum/verify", qe.handleVerify)
	qe.router.HandleFunc("/api/v2/quantum/status", qe.handleGetStatus)
}

// handleGetAlgorithms handles getting available algorithms
func (qe *QuantumEndpoints) handleGetAlgorithms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	algorithms := qe.cryptographyEngine.GetAlgorithms()

	response := map[string]interface{}{
		"algorithms": algorithms,
		"count":      len(algorithms),
		"timestamp":  time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGenerateKey handles key generation
func (qe *QuantumEndpoints) handleGenerateKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Algorithm string `json:"algorithm"`
		KeyType   string `json:"key_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Algorithm == "" {
		http.Error(w, "Algorithm is required", http.StatusBadRequest)
		return
	}
	if request.KeyType == "" {
		http.Error(w, "Key type is required", http.StatusBadRequest)
		return
	}

	// Generate key
	key, err := qe.cryptographyEngine.GenerateKey(request.Algorithm, request.KeyType)
	if err != nil {
		http.Error(w, "Failed to generate key", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"key":       key,
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetKeys handles getting keys
func (qe *QuantumEndpoints) handleGetKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keys := qe.cryptographyEngine.GetKeys()

	response := map[string]interface{}{
		"keys":      keys,
		"count":     len(keys),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleEncrypt handles encryption requests
func (qe *QuantumEndpoints) handleEncrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request quantum.EncryptionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Algorithm == "" {
		http.Error(w, "Algorithm is required", http.StatusBadRequest)
		return
	}
	if request.Plaintext == "" {
		http.Error(w, "Plaintext is required", http.StatusBadRequest)
		return
	}
	if request.KeyID == "" {
		http.Error(w, "Key ID is required", http.StatusBadRequest)
		return
	}

	// Encrypt
	response, err := qe.cryptographyEngine.Encrypt(request)
	if err != nil {
		http.Error(w, "Encryption failed", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, response)
}

// handleDecrypt handles decryption requests
func (qe *QuantumEndpoints) handleDecrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request quantum.DecryptionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Ciphertext == "" {
		http.Error(w, "Ciphertext is required", http.StatusBadRequest)
		return
	}
	if request.KeyID == "" {
		http.Error(w, "Key ID is required", http.StatusBadRequest)
		return
	}

	// Decrypt
	response, err := qe.cryptographyEngine.Decrypt(request)
	if err != nil {
		http.Error(w, "Decryption failed", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, response)
}

// handleSign handles signing requests
func (qe *QuantumEndpoints) handleSign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request quantum.SigningRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}
	if request.KeyID == "" {
		http.Error(w, "Key ID is required", http.StatusBadRequest)
		return
	}

	// Sign
	response, err := qe.cryptographyEngine.Sign(request)
	if err != nil {
		http.Error(w, "Signing failed", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, response)
}

// handleVerify handles verification requests
func (qe *QuantumEndpoints) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request quantum.VerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}
	if request.Signature == "" {
		http.Error(w, "Signature is required", http.StatusBadRequest)
		return
	}
	if request.PublicKey == "" {
		http.Error(w, "Public key is required", http.StatusBadRequest)
		return
	}
	if request.Algorithm == "" {
		http.Error(w, "Algorithm is required", http.StatusBadRequest)
		return
	}

	// Verify
	response, err := qe.cryptographyEngine.Verify(request)
	if err != nil {
		http.Error(w, "Verification failed", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, response)
}

// handleGetStatus handles getting engine status
func (qe *QuantumEndpoints) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := qe.cryptographyEngine.GetEngineStatus()

	WriteJSONResponse(w, status)
}

// GetRouter returns the quantum endpoints router
func (qe *QuantumEndpoints) GetRouter() *http.ServeMux {
	return qe.router
}

// GetCryptographyEngine returns the cryptography engine
func (qe *QuantumEndpoints) GetCryptographyEngine() *quantum.CryptographyEngine {
	return qe.cryptographyEngine
}
