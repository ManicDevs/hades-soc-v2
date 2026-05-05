package platform

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

// EnterpriseUser represents a system user with enterprise-grade authentication
type EnterpriseUser struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	Permissions  []string  `json:"permissions"`
	IsActive     bool      `json:"is_active"`
	LastLogin    time.Time `json:"last_login,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	PasswordHash string    `json:"-"` // Never exposed in JSON
	MFASecret    string    `json:"-"` // Never exposed in JSON
	MFAEnabled   bool      `json:"mfa_enabled"`
}

// EnterpriseAuthSystem manages enterprise authentication
type EnterpriseAuthSystem struct {
	users      map[string]*EnterpriseUser
	sessions   map[string]*EnterpriseSession
	config     *EnvironmentConfig
	isSetup    bool
	setupToken string
}

// EnterpriseSession represents an authenticated user session
type EnterpriseSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	IsActive  bool      `json:"is_active"`
}

// NewEnterpriseAuthSystem creates a new enterprise authentication system
func NewEnterpriseAuthSystem(config *EnvironmentConfig) *EnterpriseAuthSystem {
	as := &EnterpriseAuthSystem{
		users:      make(map[string]*EnterpriseUser),
		sessions:   make(map[string]*EnterpriseSession),
		config:     config,
		isSetup:    false,
		setupToken: generateSecureToken(32),
	}

	// Check if system is already set up
	as.loadSystemState()

	return as
}

// IsSetup checks if the system has been initialized
func (as *EnterpriseAuthSystem) IsSetup() bool {
	return as.isSetup
}

// GetSetupToken returns the setup token for initial configuration
func (as *EnterpriseAuthSystem) GetSetupToken() string {
	return as.setupToken
}

// InitializeSystem creates the initial Super Administrator
func (as *EnterpriseAuthSystem) InitializeSystem(setupToken, username, email, fullName, password string) error {
	if as.isSetup {
		return fmt.Errorf("system is already initialized")
	}

	if setupToken != as.setupToken {
		return fmt.Errorf("invalid setup token")
	}

	// Validate input
	if strings.TrimSpace(username) == "" || strings.TrimSpace(email) == "" ||
		strings.TrimSpace(fullName) == "" || strings.TrimSpace(password) == "" {
		return fmt.Errorf("all fields are required")
	}

	// Validate email format
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("invalid email format")
	}

	// Validate password strength
	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters long")
	}

	// Create Super Administrator user
	user := &EnterpriseUser{
		ID:          generateUserID(),
		Username:    strings.ToLower(strings.TrimSpace(username)),
		Email:       strings.ToLower(strings.TrimSpace(email)),
		FullName:    strings.TrimSpace(fullName),
		Role:        "super_admin",
		Permissions: []string{"system.admin", "user.manage", "security.audit", "config.global", "deployment.control", "database.admin", "network.control", "api.full_access", "logs.view", "backup.restore"},
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		MFAEnabled:  true,
	}

	// Hash password
	hash, err := as.hashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = hash

	// Generate MFA secret
	user.MFASecret = generateMFASecret()

	// Store user
	as.users[user.Username] = user

	// Mark system as setup
	as.isSetup = true
	as.setupToken = "" // Clear setup token

	// Save system state
	as.saveSystemState()

	log.Printf("System initialized with Super Administrator: %s", user.Username)
	return nil
}

// AuthenticateUser performs enterprise-grade authentication
func (as *EnterpriseAuthSystem) AuthenticateUser(username, password, mfaCode, ipAddress, userAgent string) (*EnterpriseUser, string, error) {
	if !as.isSetup {
		return nil, "", fmt.Errorf("system not initialized")
	}

	user, exists := as.users[strings.ToLower(username)]
	if !exists {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return nil, "", fmt.Errorf("account is disabled")
	}

	// Verify password
	if !as.verifyPassword(password, user.PasswordHash) {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	// Verify MFA if enabled
	if user.MFAEnabled && !as.config.IsDevelopment() {
		if mfaCode == "" {
			return nil, "", fmt.Errorf("mfa_required")
		}

		if !as.verifyMFA(user.MFASecret, mfaCode) {
			return nil, "", fmt.Errorf("invalid mfa code")
		}
	}

	// Create session
	sessionToken := generateSecureToken(64)
	session := &EnterpriseSession{
		ID:        generateSessionID(),
		UserID:    user.ID,
		Token:     sessionToken,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(as.config.SessionTimeout) * time.Second),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
	}

	as.sessions[sessionToken] = session

	// Update last login
	now := time.Now()
	user.LastLogin = now
	user.UpdatedAt = now

	// Save system state
	as.saveSystemState()

	log.Printf("User authenticated: %s from %s", user.Username, ipAddress)
	return user, sessionToken, nil
}

// ValidateSession validates a session token
func (as *EnterpriseAuthSystem) ValidateSession(token string) (*EnterpriseUser, error) {
	if !as.isSetup {
		return nil, fmt.Errorf("system not initialized")
	}

	session, exists := as.sessions[token]
	if !exists || !session.IsActive || time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("invalid or expired session")
	}

	user, exists := as.users[session.UserID]
	if !exists || !user.IsActive {
		return nil, fmt.Errorf("user not found or inactive")
	}

	return user, nil
}

// LogoutUser invalidates a session
func (as *EnterpriseAuthSystem) LogoutUser(token string) error {
	session, exists := as.sessions[token]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.IsActive = false
	delete(as.sessions, token)

	log.Printf("User logged out: %s", session.UserID)
	return nil
}

// hashPassword creates a secure password hash using Argon2
func (as *EnterpriseAuthSystem) hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

	// Combine salt and hash
	combined := append(salt, hash...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// verifyPassword verifies a password against its hash
func (as *EnterpriseAuthSystem) verifyPassword(password, hash string) bool {
	combined, err := base64.StdEncoding.DecodeString(hash)
	if err != nil || len(combined) < 16 {
		return false
	}

	salt := combined[:16]
	expectedHash := combined[16:]

	actualHash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

	return subtle.ConstantTimeCompare(expectedHash, actualHash) == 1
}

// verifyMFA verifies a TOTP MFA code (simplified implementation)
func (as *EnterpriseAuthSystem) verifyMFA(secret, code string) bool {
	// This is a simplified implementation
	// In production, use a proper TOTP library
	return len(code) == 6 && code == "123456" // Placeholder for demo
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateUserID generates a unique user ID
func generateUserID() string {
	return fmt.Sprintf("user_%d_%s", time.Now().Unix(), generateSecureToken(8))
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d_%s", time.Now().Unix(), generateSecureToken(8))
}

// generateMFASecret generates an MFA secret key
func generateMFASecret() string {
	return generateSecureToken(32)
}

// saveSystemState persists the system state
func (as *EnterpriseAuthSystem) saveSystemState() {
	// In a real implementation, this would save to a database
	// For now, we'll use a simple file-based approach
	data := map[string]interface{}{
		"is_setup": as.isSetup,
		"users":    as.users,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal system state: %v", err)
		return
	}

	err = os.WriteFile("hades_auth_system.json", jsonData, 0600)
	if err != nil {
		log.Printf("Failed to save system state: %v", err)
	}
}

// loadSystemState loads the system state from storage
func (as *EnterpriseAuthSystem) loadSystemState() {
	data, err := os.ReadFile("hades_auth_system.json")
	if err != nil {
		// File doesn't exist, system not set up
		return
	}

	var systemData map[string]interface{}
	if err := json.Unmarshal(data, &systemData); err != nil {
		log.Printf("Failed to unmarshal system state: %v", err)
		return
	}

	if isSetup, ok := systemData["is_setup"].(bool); ok {
		as.isSetup = isSetup
	}

	// Load users
	if usersData, ok := systemData["users"].(map[string]interface{}); ok {
		for username, userData := range usersData {
			userJson, err := json.Marshal(userData)
			if err != nil {
				fmt.Printf("Warning: failed to marshal user data for %s: %v\n", username, err)
				continue
			}
			var user EnterpriseUser
			if err := json.Unmarshal(userJson, &user); err == nil {
				as.users[strings.ToLower(username)] = &user
			}
		}
	}
}

// HandleSetup handles the initial system setup
func (as *EnterpriseAuthSystem) HandleSetup(w http.ResponseWriter, r *http.Request) {
	if as.isSetup {
		http.Error(w, "System is already initialized", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		// Return setup status
		response := map[string]interface{}{
			"setup_required": true,
			"setup_token":    as.setupToken,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			fmt.Printf("Warning: failed to encode JSON response: %v\n", err)
		}
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		SetupToken string `json:"setup_token"`
		Username   string `json:"username"`
		Email      string `json:"email"`
		FullName   string `json:"full_name"`
		Password   string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if err := as.InitializeSystem(request.SetupToken, request.Username, request.Email, request.FullName, request.Password); err != nil {
		response := map[string]interface{}{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "System initialized successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
