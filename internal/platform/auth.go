package platform

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

// UserRole represents user permission levels
type UserRole string

const (
	RoleViewer   UserRole = "viewer"
	RoleOperator UserRole = "operator"
	RoleAdmin    UserRole = "admin"
	RoleRoot     UserRole = "root"
)

// User represents an authenticated user
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Role         UserRole  `json:"role"`
	LastLogin    time.Time `json:"last_login"`
	CreatedAt    time.Time `json:"created_at"`
	IsActive     bool      `json:"is_active"`
	PasswordHash string    `json:"-"`
}

// Session represents an active user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	SessionTimeout    time.Duration `json:"session_timeout"`
	MaxFailedAttempts int           `json:"max_failed_attempts"`
	LockoutDuration   time.Duration `json:"lockout_duration"`
	RequireMFA        bool          `json:"require_mfa"`
}

// DefaultAuthConfig returns sensible authentication defaults
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		SessionTimeout:    24 * time.Hour,
		MaxFailedAttempts: 5,
		LockoutDuration:   15 * time.Minute,
		RequireMFA:        false,
	}
}

// AuthManager provides authentication and authorization services
type AuthManager struct {
	config         *AuthConfig
	users          map[string]*User
	sessions       map[string]*Session
	failedAttempts map[string]int
	lockedUsers    map[string]time.Time
	mu             sync.RWMutex
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(config *AuthConfig) *AuthManager {
	if config == nil {
		config = DefaultAuthConfig()
	}

	return &AuthManager{
		config:         config,
		users:          make(map[string]*User),
		sessions:       make(map[string]*Session),
		failedAttempts: make(map[string]int),
		lockedUsers:    make(map[string]time.Time),
	}
}

// CreateUser adds a new user to the system
func (am *AuthManager) CreateUser(ctx context.Context, username, email, password string, role UserRole) (*User, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if username == "" || email == "" || password == "" {
		return nil, fmt.Errorf("hades.platform.auth: username, email, and password required")
	}

	if _, exists := am.users[username]; exists {
		return nil, fmt.Errorf("hades.platform.auth: user %s already exists", username)
	}

	hash, err := am.hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.auth: %w", err)
	}

	user := &User{
		ID:           am.generateID(),
		Username:     username,
		Email:        email,
		Role:         role,
		LastLogin:    time.Time{},
		CreatedAt:    time.Now(),
		IsActive:     true,
		PasswordHash: hash,
	}

	am.users[username] = user
	return user, nil
}

// Authenticate validates user credentials and creates a session
func (am *AuthManager) Authenticate(ctx context.Context, username, password, ipAddress, userAgent string) (*Session, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.isLocked(username) {
		return nil, fmt.Errorf("hades.platform.auth: user %s is locked", username)
	}

	user, exists := am.users[username]
	if !exists {
		am.recordFailedAttempt(username)
		return nil, fmt.Errorf("hades.platform.auth: invalid credentials")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("hades.platform.auth: user account is inactive")
	}

	if !am.verifyPassword(password, user.PasswordHash) {
		am.recordFailedAttempt(username)
		return nil, fmt.Errorf("hades.platform.auth: invalid credentials")
	}

	am.clearFailedAttempts(username)
	user.LastLogin = time.Now()

	session := &Session{
		ID:        am.generateID(),
		UserID:    user.ID,
		Token:     am.generateToken(),
		ExpiresAt: time.Now().Add(am.config.SessionTimeout),
		CreatedAt: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	am.sessions[session.Token] = session
	return session, nil
}

// ValidateSession checks if a session token is valid
func (am *AuthManager) ValidateSession(ctx context.Context, token string) (*User, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	session, exists := am.sessions[token]
	if !exists {
		return nil, fmt.Errorf("hades.platform.auth: invalid session token")
	}

	if time.Now().After(session.ExpiresAt) {
		delete(am.sessions, token)
		return nil, fmt.Errorf("hades.platform.auth: session expired")
	}

	user, exists := am.users[session.UserID]
	if !exists || !user.IsActive {
		return nil, fmt.Errorf("hades.platform.auth: user not found or inactive")
	}

	return user, nil
}

// Authorize checks if user has permission for the requested action
func (am *AuthManager) Authorize(user *User, requiredRole UserRole) bool {
	roleHierarchy := map[UserRole]int{
		RoleViewer:   1,
		RoleOperator: 2,
		RoleAdmin:    3,
		RoleRoot:     4,
	}

	userLevel, userExists := roleHierarchy[user.Role]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

// InvalidateSession removes a session token
func (am *AuthManager) InvalidateSession(ctx context.Context, token string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.sessions[token]; !exists {
		return fmt.Errorf("hades.platform.auth: session not found")
	}

	delete(am.sessions, token)
	return nil
}

// hashPassword creates a secure password hash using Argon2
func (am *AuthManager) hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("hades.platform.auth: failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	combined := append(salt, hash...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// verifyPassword checks if a password matches its hash
func (am *AuthManager) verifyPassword(password, hash string) bool {
	combined, err := base64.StdEncoding.DecodeString(hash)
	if err != nil || len(combined) < 16 {
		return false
	}

	salt := combined[:16]
	storedHash := combined[16:]

	computedHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	return subtle.ConstantTimeCompare(storedHash, computedHash) == 1
}

// generateToken creates a secure random session token
func (am *AuthManager) generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateID creates a unique identifier
func (am *AuthManager) generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// isLocked checks if a user account is locked
func (am *AuthManager) isLocked(username string) bool {
	lockTime, exists := am.lockedUsers[username]
	if !exists {
		return false
	}

	if time.Now().After(lockTime) {
		delete(am.lockedUsers, username)
		return false
	}

	return true
}

// recordFailedAttempt increments failed login attempts
func (am *AuthManager) recordFailedAttempt(username string) {
	am.failedAttempts[username]++

	if am.failedAttempts[username] >= am.config.MaxFailedAttempts {
		am.lockedUsers[username] = time.Now().Add(am.config.LockoutDuration)
	}
}

// clearFailedAttempts resets failed login counter
func (am *AuthManager) clearFailedAttempts(username string) {
	delete(am.failedAttempts, username)
	delete(am.lockedUsers, username)
}

// GetUser retrieves a user by ID
func (am *AuthManager) GetUser(ctx context.Context, userID string) (*User, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, user := range am.users {
		if user.ID == userID {
			return user, nil
		}
	}

	return nil, fmt.Errorf("hades.platform.auth: user not found")
}

// ListUsers returns all users (admin only)
func (am *AuthManager) ListUsers(ctx context.Context) ([]*User, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	users := make([]*User, 0, len(am.users))
	for _, user := range am.users {
		users = append(users, user)
	}

	return users, nil
}

// CleanupExpiredSessions removes expired sessions
func (am *AuthManager) CleanupExpiredSessions(ctx context.Context) {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	for token, session := range am.sessions {
		if now.After(session.ExpiresAt) {
			delete(am.sessions, token)
		}
	}
}
