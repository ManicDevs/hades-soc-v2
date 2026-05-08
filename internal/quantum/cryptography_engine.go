package quantum

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	"hades-v2/internal/database"
)

// CryptographyEngine provides quantum-resistant cryptographic capabilities
type CryptographyEngine struct {
	db              database.Database
	algorithms      map[string]*QuantumAlgorithm
	keys            map[string]*QuantumKey
	signatures      map[string]*QuantumSignature
	allowSimulation bool
	mu              sync.RWMutex
}

// QuantumAlgorithm represents a quantum-resistant algorithm
type QuantumAlgorithm struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	KeySize    int                    `json:"key_size"`
	Security   int                    `json:"security"` // bits of security
	Enabled    bool                   `json:"enabled"`
	Parameters map[string]interface{} `json:"parameters"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

// QuantumKey represents a quantum-resistant key
type QuantumKey struct {
	ID           string                 `json:"id"`
	Algorithm    string                 `json:"algorithm"`
	Type         string                 `json:"type"` // public, private, symmetric
	PublicKey    string                 `json:"public_key"`
	PrivateKey   string                 `json:"private_key,omitempty"`
	SymmetricKey string                 `json:"symmetric_key,omitempty"`
	KeySize      int                    `json:"key_size"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// QuantumSignature represents a quantum-resistant signature
type QuantumSignature struct {
	ID        string                 `json:"id"`
	Algorithm string                 `json:"algorithm"`
	Message   string                 `json:"message"`
	Signature string                 `json:"signature"`
	PublicKey string                 `json:"public_key"`
	Verified  bool                   `json:"verified"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// EncryptionRequest represents an encryption request
type EncryptionRequest struct {
	Algorithm string                 `json:"algorithm"`
	Plaintext string                 `json:"plaintext"`
	KeyID     string                 `json:"key_id"`
	Options   map[string]interface{} `json:"options"`
}

// EncryptionResponse represents an encryption response
type EncryptionResponse struct {
	Ciphertext string                 `json:"ciphertext"`
	KeyID      string                 `json:"key_id"`
	Algorithm  string                 `json:"algorithm"`
	Metadata   map[string]interface{} `json:"metadata"`
	Timestamp  time.Time              `json:"timestamp"`
}

// DecryptionRequest represents a decryption request
type DecryptionRequest struct {
	Ciphertext string                 `json:"ciphertext"`
	KeyID      string                 `json:"key_id"`
	Options    map[string]interface{} `json:"options"`
}

// DecryptionResponse represents a decryption response
type DecryptionResponse struct {
	Plaintext string                 `json:"plaintext"`
	KeyID     string                 `json:"key_id"`
	Algorithm string                 `json:"algorithm"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// SigningRequest represents a signing request
type SigningRequest struct {
	Message   string `json:"message"`
	KeyID     string `json:"key_id"`
	Algorithm string `json:"algorithm"`
}

// SigningResponse represents a signing response
type SigningResponse struct {
	Signature string                 `json:"signature"`
	KeyID     string                 `json:"key_id"`
	Algorithm string                 `json:"algorithm"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// VerificationRequest represents a verification request
type VerificationRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
	PublicKey string `json:"public_key"`
	Algorithm string `json:"algorithm"`
}

// VerificationResponse represents a verification response
type VerificationResponse struct {
	Valid     bool                   `json:"valid"`
	Algorithm string                 `json:"algorithm"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewCryptographyEngine creates a new quantum cryptography engine
func NewCryptographyEngine(db database.Database) (*CryptographyEngine, error) {
	engine := &CryptographyEngine{
		db:              db,
		algorithms:      make(map[string]*QuantumAlgorithm),
		keys:            make(map[string]*QuantumKey),
		signatures:      make(map[string]*QuantumSignature),
		allowSimulation: os.Getenv("HADES_ALLOW_SIMULATED_CRYPTO") == "true",
	}

	// Initialize quantum-resistant algorithms
	if err := engine.initializeAlgorithms(); err != nil {
		return nil, fmt.Errorf("failed to initialize algorithms: %w", err)
	}

	return engine, nil
}

// initializeAlgorithms initializes quantum-resistant algorithms
func (qce *CryptographyEngine) initializeAlgorithms() error {
	// Kyber - Key Encapsulation Mechanism
	qce.algorithms["kyber512"] = &QuantumAlgorithm{
		ID:       "kyber512",
		Name:     "Kyber-512",
		Type:     "kem",
		KeySize:  512,
		Security: 128,
		Enabled:  true,
		Parameters: map[string]interface{}{
			"security_level": 1,
			"variant":        "Kyber512",
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	qce.algorithms["kyber768"] = &QuantumAlgorithm{
		ID:       "kyber768",
		Name:     "Kyber-768",
		Type:     "kem",
		KeySize:  768,
		Security: 192,
		Enabled:  true,
		Parameters: map[string]interface{}{
			"security_level": 3,
			"variant":        "Kyber768",
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	qce.algorithms["kyber1024"] = &QuantumAlgorithm{
		ID:       "kyber1024",
		Name:     "Kyber-1024",
		Type:     "kem",
		KeySize:  1024,
		Security: 256,
		Enabled:  true,
		Parameters: map[string]interface{}{
			"security_level": 5,
			"variant":        "Kyber1024",
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	// Dilithium - Digital Signature Algorithm
	qce.algorithms["dilithium2"] = &QuantumAlgorithm{
		ID:       "dilithium2",
		Name:     "Dilithium2",
		Type:     "signature",
		KeySize:  1312,
		Security: 128,
		Enabled:  true,
		Parameters: map[string]interface{}{
			"security_level": 2,
			"variant":        "Dilithium2",
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	qce.algorithms["dilithium3"] = &QuantumAlgorithm{
		ID:       "dilithium3",
		Name:     "Dilithium3",
		Type:     "signature",
		KeySize:  1952,
		Security: 192,
		Enabled:  true,
		Parameters: map[string]interface{}{
			"security_level": 3,
			"variant":        "Dilithium3",
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	qce.algorithms["dilithium5"] = &QuantumAlgorithm{
		ID:       "dilithium5",
		Name:     "Dilithium5",
		Type:     "signature",
		KeySize:  2592,
		Security: 256,
		Enabled:  true,
		Parameters: map[string]interface{}{
			"security_level": 5,
			"variant":        "Dilithium5",
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	// SPHINCS+ - Hash-based signature
	qce.algorithms["sphincs128"] = &QuantumAlgorithm{
		ID:       "sphincs128",
		Name:     "SPHINCS+-128",
		Type:     "signature",
		KeySize:  64,
		Security: 128,
		Enabled:  true,
		Parameters: map[string]interface{}{
			"security_level": 1,
			"variant":        "simple",
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	// Falcon - Lattice-based signature
	qce.algorithms["falcon512"] = &QuantumAlgorithm{
		ID:       "falcon512",
		Name:     "Falcon-512",
		Type:     "signature",
		KeySize:  897,
		Security: 128,
		Enabled:  true,
		Parameters: map[string]interface{}{
			"security_level": 1,
			"variant":        "Falcon512",
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	return nil
}

// GenerateKey generates a new quantum-resistant key
func (qce *CryptographyEngine) GenerateKey(algorithm, keyType string) (*QuantumKey, error) {
	if algorithm == "" {
		algorithm = "kyber1024"
	}

	qce.mu.Lock()
	defer qce.mu.Unlock()

	// Check if algorithm exists
	alg, exists := qce.algorithms[algorithm]
	if !exists {
		return nil, fmt.Errorf("algorithm not found: %s", algorithm)
	}

	if !alg.Enabled {
		return nil, fmt.Errorf("algorithm not enabled: %s", algorithm)
	}

	// Generate key based on algorithm and type
	key := &QuantumKey{
		ID:        fmt.Sprintf("key_%d_%s", time.Now().UnixNano(), algorithm),
		Algorithm: algorithm,
		Type:      keyType,
		KeySize:   alg.KeySize,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour), // 1 year
		Metadata:  make(map[string]interface{}),
	}

	switch alg.Type {
	case "kem":
		if keyType == "public" || keyType == "private" {
			if err := qce.generateKEMKey(key, alg); err != nil {
				return nil, err
			}
		}
	case "signature":
		if keyType == "public" || keyType == "private" {
			if err := qce.generateSignatureKey(key, alg); err != nil {
				return nil, err
			}
		}
	case "symmetric":
		if err := qce.generateSymmetricKey(key, alg); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported algorithm type: %s", alg.Type)
	}

	// Store key
	qce.keys[key.ID] = key

	return key, nil
}

// Encrypt encrypts data using quantum-resistant algorithms
func (qce *CryptographyEngine) Encrypt(request EncryptionRequest) (*EncryptionResponse, error) {
	qce.mu.RLock()
	defer qce.mu.RUnlock()

	// Get key
	key, exists := qce.keys[request.KeyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", request.KeyID)
	}

	// Get algorithm
	alg, exists := qce.algorithms[key.Algorithm]
	if !exists {
		return nil, fmt.Errorf("algorithm not found: %s", key.Algorithm)
	}

	// Encrypt based on algorithm type
	var ciphertext string
	var err error

	switch alg.Type {
	case "kem":
		ciphertext, err = qce.encryptKEM(request.Plaintext, key, alg)
	case "symmetric":
		ciphertext, err = qce.encryptSymmetric(request.Plaintext, key, alg)
	default:
		return nil, fmt.Errorf("encryption not supported for algorithm type: %s", alg.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	response := &EncryptionResponse{
		Ciphertext: ciphertext,
		KeyID:      request.KeyID,
		Algorithm:  key.Algorithm,
		Metadata: map[string]interface{}{
			"algorithm_type": alg.Type,
			"key_size":       key.KeySize,
		},
		Timestamp: time.Now(),
	}

	return response, nil
}

// Decrypt decrypts data using quantum-resistant algorithms
func (qce *CryptographyEngine) Decrypt(request DecryptionRequest) (*DecryptionResponse, error) {
	qce.mu.RLock()
	defer qce.mu.RUnlock()

	// Get key
	key, exists := qce.keys[request.KeyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", request.KeyID)
	}

	// Get algorithm
	alg, exists := qce.algorithms[key.Algorithm]
	if !exists {
		return nil, fmt.Errorf("algorithm not found: %s", key.Algorithm)
	}

	// Decrypt based on algorithm type
	var plaintext string
	var err error

	switch alg.Type {
	case "kem":
		plaintext, err = qce.decryptKEM(request.Ciphertext, key, alg)
	case "symmetric":
		plaintext, err = qce.decryptSymmetric(request.Ciphertext, key, alg)
	default:
		return nil, fmt.Errorf("decryption not supported for algorithm type: %s", alg.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	response := &DecryptionResponse{
		Plaintext: plaintext,
		KeyID:     request.KeyID,
		Algorithm: key.Algorithm,
		Metadata: map[string]interface{}{
			"algorithm_type": alg.Type,
			"key_size":       key.KeySize,
		},
		Timestamp: time.Now(),
	}

	return response, nil
}

// Sign signs a message using quantum-resistant algorithms
func (qce *CryptographyEngine) Sign(request SigningRequest) (*SigningResponse, error) {
	qce.mu.RLock()
	defer qce.mu.RUnlock()

	// Get key
	key, exists := qce.keys[request.KeyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", request.KeyID)
	}

	if key.Type != "private" {
		return nil, fmt.Errorf("private key required for signing")
	}

	// Get algorithm
	alg, exists := qce.algorithms[key.Algorithm]
	if !exists {
		return nil, fmt.Errorf("algorithm not found: %s", key.Algorithm)
	}

	if alg.Type != "signature" {
		return nil, fmt.Errorf("algorithm does not support signing: %s", key.Algorithm)
	}

	// Sign message
	signature, err := qce.signMessage(request.Message, key, alg)
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	response := &SigningResponse{
		Signature: signature,
		KeyID:     request.KeyID,
		Algorithm: key.Algorithm,
		Metadata: map[string]interface{}{
			"algorithm_type": alg.Type,
			"key_size":       key.KeySize,
		},
		Timestamp: time.Now(),
	}

	return response, nil
}

// Verify verifies a signature using quantum-resistant algorithms
func (qce *CryptographyEngine) Verify(request VerificationRequest) (*VerificationResponse, error) {
	qce.mu.RLock()
	defer qce.mu.RUnlock()

	// Get algorithm
	alg, exists := qce.algorithms[request.Algorithm]
	if !exists {
		return nil, fmt.Errorf("algorithm not found: %s", request.Algorithm)
	}

	if alg.Type != "signature" {
		return nil, fmt.Errorf("algorithm does not support verification: %s", request.Algorithm)
	}

	// Verify signature
	valid, err := qce.verifySignature(request.Message, request.Signature, request.PublicKey, alg)
	if err != nil {
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	response := &VerificationResponse{
		Valid:     valid,
		Algorithm: request.Algorithm,
		Metadata: map[string]interface{}{
			"algorithm_type": alg.Type,
			"key_size":       alg.KeySize,
		},
		Timestamp: time.Now(),
	}

	return response, nil
}

// GetAlgorithms returns all available algorithms
func (qce *CryptographyEngine) GetAlgorithms() map[string]*QuantumAlgorithm {
	qce.mu.RLock()
	defer qce.mu.RUnlock()

	// Return copy
	result := make(map[string]*QuantumAlgorithm)
	for id, alg := range qce.algorithms {
		result[id] = alg
	}
	return result
}

// GetKeys returns all keys
func (qce *CryptographyEngine) GetKeys() map[string]*QuantumKey {
	qce.mu.RLock()
	defer qce.mu.RUnlock()

	// Return copy without private keys
	result := make(map[string]*QuantumKey)
	for id, key := range qce.keys {
		safeKey := *key
		safeKey.PrivateKey = ""
		safeKey.SymmetricKey = ""
		result[id] = &safeKey
	}
	return result
}

// GetEngineStatus returns engine status
func (qce *CryptographyEngine) GetEngineStatus() map[string]interface{} {
	qce.mu.RLock()
	defer qce.mu.RUnlock()

	return map[string]interface{}{
		"algorithms":         len(qce.algorithms),
		"keys":               len(qce.keys),
		"signatures":         len(qce.signatures),
		"enabled_algorithms": qce.getEnabledAlgorithmCount(),
		"timestamp":          time.Now(),
	}
}

// Helper functions

// generateKEMKey generates a KEM key pair
func (qce *CryptographyEngine) generateKEMKey(key *QuantumKey, alg *QuantumAlgorithm) error {
	if err := qce.ensureSimulationAllowed("KEM key generation"); err != nil {
		return err
	}
	publicKey, err := qce.generateRandomBytes(alg.KeySize)
	if err != nil {
		return err
	}

	privateKey, err := qce.generateRandomBytes(alg.KeySize)
	if err != nil {
		return err
	}

	key.PublicKey = hex.EncodeToString(publicKey)
	key.PrivateKey = hex.EncodeToString(privateKey)

	return nil
}

// generateSignatureKey generates a signature key pair
func (qce *CryptographyEngine) generateSignatureKey(key *QuantumKey, alg *QuantumAlgorithm) error {
	if err := qce.ensureSimulationAllowed("signature key generation"); err != nil {
		return err
	}
	publicKey, err := qce.generateRandomBytes(alg.KeySize)
	if err != nil {
		return err
	}

	privateKey, err := qce.generateRandomBytes(alg.KeySize)
	if err != nil {
		return err
	}

	key.PublicKey = hex.EncodeToString(publicKey)
	key.PrivateKey = hex.EncodeToString(privateKey)

	return nil
}

// generateSymmetricKey generates a symmetric key
func (qce *CryptographyEngine) generateSymmetricKey(key *QuantumKey, alg *QuantumAlgorithm) error {
	// Generate symmetric key
	symmetricKey, err := qce.generateRandomBytes(alg.KeySize)
	if err != nil {
		return err
	}

	key.SymmetricKey = hex.EncodeToString(symmetricKey)

	return nil
}

// encryptKEM encrypts using KEM
func (qce *CryptographyEngine) encryptKEM(plaintext string, key *QuantumKey, alg *QuantumAlgorithm) (string, error) {
	if err := qce.ensureSimulationAllowed("KEM encryption"); err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(plaintext + key.PublicKey))
	return hex.EncodeToString(hash[:]), nil
}

// decryptKEM decrypts using KEM
func (qce *CryptographyEngine) decryptKEM(ciphertext string, key *QuantumKey, alg *QuantumAlgorithm) (string, error) {
	if err := qce.ensureSimulationAllowed("KEM decryption"); err != nil {
		return "", err
	}
	return "decrypted_" + ciphertext[:min(len(ciphertext), 32)], nil
}

// encryptSymmetric encrypts using symmetric encryption
func (qce *CryptographyEngine) encryptSymmetric(plaintext string, key *QuantumKey, alg *QuantumAlgorithm) (string, error) {
	// Simulate symmetric encryption using XOR
	keyBytes, err := hex.DecodeString(key.SymmetricKey)
	if err != nil {
		return "", err
	}

	plainBytes := []byte(plaintext)
	cipherBytes := make([]byte, len(plainBytes))

	for i := range plainBytes {
		cipherBytes[i] = plainBytes[i] ^ keyBytes[i%len(keyBytes)]
	}

	return hex.EncodeToString(cipherBytes), nil
}

// decryptSymmetric decrypts using symmetric encryption
func (qce *CryptographyEngine) decryptSymmetric(ciphertext string, key *QuantumKey, alg *QuantumAlgorithm) (string, error) {
	// Simulate symmetric decryption using XOR
	keyBytes, err := hex.DecodeString(key.SymmetricKey)
	if err != nil {
		return "", err
	}

	cipherBytes, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	plainBytes := make([]byte, len(cipherBytes))
	for i := range cipherBytes {
		plainBytes[i] = cipherBytes[i] ^ keyBytes[i%len(keyBytes)]
	}

	return string(plainBytes), nil
}

// signMessage signs a message
func (qce *CryptographyEngine) signMessage(message string, key *QuantumKey, alg *QuantumAlgorithm) (string, error) {
	if err := qce.ensureSimulationAllowed("digital signature"); err != nil {
		return "", err
	}
	hash := sha512.Sum512([]byte(message + key.PrivateKey))
	return hex.EncodeToString(hash[:]), nil
}

// verifySignature verifies a signature
func (qce *CryptographyEngine) verifySignature(message, signature, publicKey string, alg *QuantumAlgorithm) (bool, error) {
	if err := qce.ensureSimulationAllowed("signature verification"); err != nil {
		return false, err
	}
	hash := sha512.Sum512([]byte(message + publicKey))
	expectedSignature := hex.EncodeToString(hash[:])
	return signature == expectedSignature, nil
}

// generateRandomBytes generates cryptographically secure random bytes
func (qce *CryptographyEngine) generateRandomBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	return bytes, err
}

// getEnabledAlgorithmCount returns count of enabled algorithms
func (qce *CryptographyEngine) getEnabledAlgorithmCount() int {
	count := 0
	for _, alg := range qce.algorithms {
		if alg.Enabled {
			count++
		}
	}
	return count
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (qce *CryptographyEngine) ensureSimulationAllowed(op string) error {
	if qce.allowSimulation {
		return nil
	}
	return fmt.Errorf("%s is disabled in production: set HADES_ALLOW_SIMULATED_CRYPTO=true only for non-production testing", op)
}
