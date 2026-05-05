package security

import (
	"crypto/sha256"
	"fmt"
	"log"
	"sync"
	"time"
)

// RBACManager provides Role-Based Access Control functionality
type RBACManager struct {
	roles       map[string]*Role
	permissions map[string]*Permission
	userRoles   map[string][]string
	rolePerms   map[string][]string
}

// Role represents a role in the RBAC system
type Role struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Permission represents a permission in the RBAC system
type Permission struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Resource    string            `json:"resource"`
	Action      string            `json:"action"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
}

// UserPermissions represents a user's effective permissions
type UserPermissions struct {
	UserID       string            `json:"user_id"`
	Roles        []string          `json:"roles"`
	Permissions  []string          `json:"permissions"`
	Effective    map[string]bool   `json:"effective"`
	Metadata     map[string]string `json:"metadata"`
	CalculatedAt time.Time         `json:"calculated_at"`
}

// AccessRequest represents an access control request
type AccessRequest struct {
	UserID   string                 `json:"user_id"`
	Resource string                 `json:"resource"`
	Action   string                 `json:"action"`
	Context  map[string]interface{} `json:"context"`
}

// AccessResult represents the result of an access check
type AccessResult struct {
	Allowed   bool                   `json:"allowed"`
	UserID    string                 `json:"user_id"`
	Resource  string                 `json:"resource"`
	Action    string                 `json:"action"`
	Reason    string                 `json:"reason"`
	Roles     []string               `json:"roles"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager() *RBACManager {
	rbac := &RBACManager{
		roles:       make(map[string]*Role),
		permissions: make(map[string]*Permission),
		userRoles:   make(map[string][]string),
		rolePerms:   make(map[string][]string),
	}

	// Initialize default roles and permissions
	rbac.initializeDefaults()
	return rbac
}

// initializeDefaults sets up default roles and permissions
func (r *RBACManager) initializeDefaults() {
	// Define default permissions
	defaultPermissions := []*Permission{
		{ID: "user.read", Name: "Read Users", Description: "Read user information", Resource: "users", Action: "read"},
		{ID: "user.write", Name: "Write Users", Description: "Create and modify users", Resource: "users", Action: "write"},
		{ID: "user.delete", Name: "Delete Users", Description: "Delete users", Resource: "users", Action: "delete"},
		{ID: "module.read", Name: "Read Modules", Description: "Read module information", Resource: "modules", Action: "read"},
		{ID: "module.execute", Name: "Execute Modules", Description: "Execute security modules", Resource: "modules", Action: "execute"},
		{ID: "module.configure", Name: "Configure Modules", Description: "Configure module settings", Resource: "modules", Action: "configure"},
		{ID: "exploit.read", Name: "Read Exploits", Description: "Read exploit database", Resource: "exploits", Action: "read"},
		{ID: "exploit.write", Name: "Write Exploits", Description: "Add and modify exploits", Resource: "exploits", Action: "write"},
		{ID: "exploit.execute", Name: "Execute Exploits", Description: "Execute exploits", Resource: "exploits", Action: "execute"},
		{ID: "system.read", Name: "Read System", Description: "Read system information", Resource: "system", Action: "read"},
		{ID: "system.configure", Name: "Configure System", Description: "Configure system settings", Resource: "system", Action: "configure"},
		{ID: "audit.read", Name: "Read Audit Logs", Description: "Read audit logs", Resource: "audit", Action: "read"},
		{ID: "config.read", Name: "Read Configuration", Description: "Read configuration", Resource: "config", Action: "read"},
		{ID: "config.write", Name: "Write Configuration", Description: "Modify configuration", Resource: "config", Action: "write"},
	}

	for _, perm := range defaultPermissions {
		perm.CreatedAt = time.Now()
		r.permissions[perm.ID] = perm
	}

	// Define default roles
	defaultRoles := []*Role{
		{
			ID:          "admin",
			Name:        "Administrator",
			Description: "Full system access",
			Permissions: []string{"user.read", "user.write", "user.delete", "module.read", "module.execute", "module.configure",
				"exploit.read", "exploit.write", "exploit.execute", "system.read", "system.configure", "audit.read", "config.read", "config.write"},
			Metadata: map[string]string{"level": "system", "category": "administrative"},
		},
		{
			ID:          "operator",
			Name:        "Operator",
			Description: "Operational access to modules and exploits",
			Permissions: []string{"user.read", "module.read", "module.execute", "exploit.read", "exploit.execute", "system.read", "audit.read", "config.read"},
			Metadata:    map[string]string{"level": "operational", "category": "user"},
		},
		{
			ID:          "analyst",
			Name:        "Security Analyst",
			Description: "Read-only access to security data",
			Permissions: []string{"user.read", "module.read", "exploit.read", "system.read", "audit.read", "config.read"},
			Metadata:    map[string]string{"level": "analyst", "category": "readonly"},
		},
		{
			ID:          "viewer",
			Name:        "Viewer",
			Description: "Limited read access",
			Permissions: []string{"module.read", "exploit.read", "system.read"},
			Metadata:    map[string]string{"level": "viewer", "category": "limited"},
		},
	}

	for _, role := range defaultRoles {
		role.CreatedAt = time.Now()
		role.UpdatedAt = time.Now()
		r.roles[role.ID] = role
		r.rolePerms[role.ID] = role.Permissions
	}
}

// CreateRole creates a new role
func (r *RBACManager) CreateRole(id, name, description string, permissions []string, metadata map[string]string) error {
	if _, exists := r.roles[id]; exists {
		return fmt.Errorf("role %s already exists", id)
	}

	// Validate permissions
	for _, permID := range permissions {
		if _, exists := r.permissions[permID]; !exists {
			return fmt.Errorf("permission %s does not exist", permID)
		}
	}

	role := &Role{
		ID:          id,
		Name:        name,
		Description: description,
		Permissions: permissions,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	r.roles[id] = role
	r.rolePerms[id] = permissions
	return nil
}

// CreatePermission creates a new permission
func (r *RBACManager) CreatePermission(id, name, description, resource, action string, metadata map[string]string) error {
	if _, exists := r.permissions[id]; exists {
		return fmt.Errorf("permission %s already exists", id)
	}

	permission := &Permission{
		ID:          id,
		Name:        name,
		Description: description,
		Resource:    resource,
		Action:      action,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
	}

	r.permissions[id] = permission
	return nil
}

// AssignRole assigns a role to a user
func (r *RBACManager) AssignRole(userID, roleID string) error {
	if _, exists := r.roles[roleID]; !exists {
		return fmt.Errorf("role %s does not exist", roleID)
	}

	if r.userRoles[userID] == nil {
		r.userRoles[userID] = []string{}
	}

	// Check if user already has this role
	for _, role := range r.userRoles[userID] {
		if role == roleID {
			return fmt.Errorf("user %s already has role %s", userID, roleID)
		}
	}

	r.userRoles[userID] = append(r.userRoles[userID], roleID)
	return nil
}

// RemoveRole removes a role from a user
func (r *RBACManager) RemoveRole(userID, roleID string) error {
	if r.userRoles[userID] == nil {
		return fmt.Errorf("user %s has no roles", userID)
	}

	for i, role := range r.userRoles[userID] {
		if role == roleID {
			r.userRoles[userID] = append(r.userRoles[userID][:i], r.userRoles[userID][i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("user %s does not have role %s", userID, roleID)
}

// GetUserRoles returns all roles for a user
func (r *RBACManager) GetUserRoles(userID string) []string {
	if roles, exists := r.userRoles[userID]; exists {
		return roles
	}
	return []string{}
}

// GetRolePermissions returns all permissions for a role
func (r *RBACManager) GetRolePermissions(roleID string) []string {
	if perms, exists := r.rolePerms[roleID]; exists {
		return perms
	}
	return []string{}
}

// GetUserPermissions calculates all effective permissions for a user
func (r *RBACManager) GetUserPermissions(userID string) *UserPermissions {
	userRoles := r.GetUserRoles(userID)
	permissionSet := make(map[string]bool)

	// Collect all permissions from user's roles
	for _, roleID := range userRoles {
		rolePerms := r.GetRolePermissions(roleID)
		for _, permID := range rolePerms {
			permissionSet[permID] = true
		}
	}

	// Convert to slice
	permissions := make([]string, 0, len(permissionSet))
	for permID := range permissionSet {
		permissions = append(permissions, permID)
	}

	return &UserPermissions{
		UserID:       userID,
		Roles:        userRoles,
		Permissions:  permissions,
		Effective:    permissionSet,
		CalculatedAt: time.Now(),
	}
}

// CheckAccess checks if a user has permission to perform an action on a resource
func (r *RBACManager) CheckAccess(userID, resource, action string, context map[string]interface{}) *AccessResult {
	userPerms := r.GetUserPermissions(userID)

	// Check for exact permission match
	permissionID := fmt.Sprintf("%s.%s", resource, action)
	allowed := userPerms.Effective[permissionID]

	reason := "Access denied"
	if allowed {
		reason = "Access granted"
	}

	// Check for wildcard permissions if not explicitly allowed
	if !allowed {
		// Check resource wildcard
		wildcardPerm := fmt.Sprintf("%s.*", resource)
		if userPerms.Effective[wildcardPerm] {
			allowed = true
			reason = "Access granted via resource wildcard"
		}

		// Check action wildcard
		if !allowed {
			actionWildcard := fmt.Sprintf("*.%s", action)
			if userPerms.Effective[actionWildcard] {
				allowed = true
				reason = "Access granted via action wildcard"
			}
		}

		// Check global wildcard
		if !allowed && userPerms.Effective["*.*"] {
			allowed = true
			reason = "Access granted via global wildcard"
		}
	}

	return &AccessResult{
		Allowed:   allowed,
		UserID:    userID,
		Resource:  resource,
		Action:    action,
		Reason:    reason,
		Roles:     userPerms.Roles,
		Timestamp: time.Now(),
		Context:   context,
	}
}

// HasPermission checks if a user has a specific permission
func (r *RBACManager) HasPermission(userID, permissionID string) bool {
	userPerms := r.GetUserPermissions(userID)
	return userPerms.Effective[permissionID]
}

// GetRole returns a role by ID
func (r *RBACManager) GetRole(roleID string) (*Role, error) {
	if role, exists := r.roles[roleID]; exists {
		return role, nil
	}
	return nil, fmt.Errorf("role %s not found", roleID)
}

// GetPermission returns a permission by ID
func (r *RBACManager) GetPermission(permissionID string) (*Permission, error) {
	if permission, exists := r.permissions[permissionID]; exists {
		return permission, nil
	}
	return nil, fmt.Errorf("permission %s not found", permissionID)
}

// ListRoles returns all roles
func (r *RBACManager) ListRoles() []*Role {
	roles := make([]*Role, 0, len(r.roles))
	for _, role := range r.roles {
		roles = append(roles, role)
	}
	return roles
}

// ListPermissions returns all permissions
func (r *RBACManager) ListPermissions() []*Permission {
	permissions := make([]*Permission, 0, len(r.permissions))
	for _, permission := range r.permissions {
		permissions = append(permissions, permission)
	}
	return permissions
}

// UpdateRole updates an existing role
func (r *RBACManager) UpdateRole(roleID, name, description string, permissions []string, metadata map[string]string) error {
	role, exists := r.roles[roleID]
	if !exists {
		return fmt.Errorf("role %s not found", roleID)
	}

	// Validate permissions
	for _, permID := range permissions {
		if _, exists := r.permissions[permID]; !exists {
			return fmt.Errorf("permission %s does not exist", permID)
		}
	}

	role.Name = name
	role.Description = description
	role.Permissions = permissions
	role.Metadata = metadata
	role.UpdatedAt = time.Now()

	r.rolePerms[roleID] = permissions
	return nil
}

// DeleteRole deletes a role
func (r *RBACManager) DeleteRole(roleID string) error {
	if _, exists := r.roles[roleID]; !exists {
		return fmt.Errorf("role %s not found", roleID)
	}

	delete(r.roles, roleID)
	delete(r.rolePerms, roleID)

	// Remove role from all users
	for userID, roles := range r.userRoles {
		newRoles := []string{}
		for _, role := range roles {
			if role != roleID {
				newRoles = append(newRoles, role)
			}
		}
		r.userRoles[userID] = newRoles
	}

	return nil
}

// DeletePermission deletes a permission
func (r *RBACManager) DeletePermission(permissionID string) error {
	if _, exists := r.permissions[permissionID]; !exists {
		return fmt.Errorf("permission %s not found", permissionID)
	}

	delete(r.permissions, permissionID)

	// Remove permission from all roles
	for _, role := range r.roles {
		newPerms := []string{}
		for _, perm := range role.Permissions {
			if perm != permissionID {
				newPerms = append(newPerms, perm)
			}
		}
		role.Permissions = newPerms
		r.rolePerms[role.ID] = newPerms
	}

	return nil
}

// GetUsersWithRole returns all users who have a specific role
func (r *RBACManager) GetUsersWithRole(roleID string) []string {
	users := []string{}
	for userID, roles := range r.userRoles {
		for _, role := range roles {
			if role == roleID {
				users = append(users, userID)
				break
			}
		}
	}
	return users
}

// ExportRBAC exports the entire RBAC configuration
func (r *RBACManager) ExportRBAC() map[string]interface{} {
	return map[string]interface{}{
		"roles":       r.roles,
		"permissions": r.permissions,
		"user_roles":  r.userRoles,
		"role_perms":  r.rolePerms,
		"exported_at": time.Now(),
	}
}

// ImportRBAC imports RBAC configuration
func (r *RBACManager) ImportRBAC(data map[string]interface{}) error {
	// This would need proper JSON unmarshalling in a real implementation
	// For now, just return success
	return nil
}

// ValidateRBAC validates the RBAC configuration
func (r *RBACManager) ValidateRBAC() []string {
	errors := []string{}

	// Check for orphaned permissions in roles
	for roleID, role := range r.roles {
		for _, permID := range role.Permissions {
			if _, exists := r.permissions[permID]; !exists {
				errors = append(errors, fmt.Sprintf("Role %s references non-existent permission %s", roleID, permID))
			}
		}
	}

	// Check for orphaned roles in user assignments
	for userID, roles := range r.userRoles {
		for _, roleID := range roles {
			if _, exists := r.roles[roleID]; !exists {
				errors = append(errors, fmt.Sprintf("User %s assigned to non-existent role %s", userID, roleID))
			}
		}
	}

	return errors
}

// GetPermissionSummary returns a summary of permissions by resource
func (r *RBACManager) GetPermissionSummary() map[string][]string {
	summary := make(map[string][]string)

	for _, permission := range r.permissions {
		resource := permission.Resource
		summary[resource] = append(summary[resource], permission.Action)
	}

	return summary
}

// GetUserCount returns the number of users with each role
func (r *RBACManager) GetUserCount() map[string]int {
	count := make(map[string]int)

	for _, roles := range r.userRoles {
		for _, roleID := range roles {
			count[roleID]++
		}
	}

	return count
}

// Session represents a user session for RBAC tracking
type Session struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	Token          string                 `json:"token"`
	CreatedAt      time.Time              `json:"created_at"`
	ExpiresAt      time.Time              `json:"expires_at"`
	LastActivity   time.Time              `json:"last_activity"`
	RequiresReauth bool                   `json:"requires_reauth"`
	ReauthReason   string                 `json:"reauth_reason,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// SessionManager handles session lifecycle and re-authentication
type SessionManager struct {
	sessions map[string]*Session // sessionID -> Session
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// HoneyToken represents a bait credential for detecting attackers
type HoneyToken struct {
	ID           string       `json:"id"`
	Username     string       `json:"username"`
	PasswordHash string       `json:"password_hash"`
	Role         string       `json:"role"`
	CreatedAt    time.Time    `json:"created_at"`
	IsTriggered  bool         `json:"is_triggered"`
	TriggerInfo  *TriggerInfo `json:"trigger_info,omitempty"`
}

// TriggerInfo contains details when honey token is accessed
type TriggerInfo struct {
	SourceIP     string    `json:"source_ip"`
	Fingerprint  string    `json:"fingerprint"`
	Timestamp    time.Time `json:"timestamp"`
	PasswordUsed string    `json:"password_used"`
}

// HoneyTokenManager manages bait credentials
type HoneyTokenManager struct {
	tokens map[string]*HoneyToken // username -> HoneyToken
	mu     sync.RWMutex
}

// NewHoneyTokenManager creates a new honey token manager
func NewHoneyTokenManager() *HoneyTokenManager {
	return &HoneyTokenManager{
		tokens: make(map[string]*HoneyToken),
	}
}

// GenerateHoneyToken creates a new bait user with plausible credentials
func (htm *HoneyTokenManager) GenerateHoneyToken(username string) *HoneyToken {
	htm.mu.Lock()
	defer htm.mu.Unlock()

	// Generate unique but plausible password hash
	uniqueSeed := fmt.Sprintf("%s_%d_%s", username, time.Now().UnixNano(), generateRandomString(16))
	passwordHash := hashString(uniqueSeed)

	token := &HoneyToken{
		ID:           fmt.Sprintf("honey_%d", time.Now().UnixNano()),
		Username:     username,
		PasswordHash: passwordHash,
		Role:         "admin", // Bait role to attract attackers
		CreatedAt:    time.Now(),
		IsTriggered:  false,
	}

	htm.tokens[username] = token

	log.Printf("HoneyToken: Generated bait user '%s' with role '%s'", username, token.Role)
	return token
}

// IsHoneyToken checks if a username is a honey token
func (htm *HoneyTokenManager) IsHoneyToken(username string) bool {
	htm.mu.RLock()
	defer htm.mu.RUnlock()
	_, exists := htm.tokens[username]
	return exists
}

// TriggerHoneyToken marks a honey token as triggered and returns trigger info
func (htm *HoneyTokenManager) TriggerHoneyToken(username, sourceIP, fingerprint, passwordUsed string) *HoneyToken {
	htm.mu.Lock()
	defer htm.mu.Unlock()

	token, exists := htm.tokens[username]
	if !exists {
		return nil
	}

	token.IsTriggered = true
	token.TriggerInfo = &TriggerInfo{
		SourceIP:     sourceIP,
		Fingerprint:  fingerprint,
		Timestamp:    time.Now(),
		PasswordUsed: passwordUsed,
	}

	log.Printf("HoneyToken: ALERT - Bait user '%s' accessed from %s (fingerprint: %s)",
		username, sourceIP, fingerprint)

	return token
}

// GetAllHoneyTokens returns all honey tokens (for admin review)
func (htm *HoneyTokenManager) GetAllHoneyTokens() []*HoneyToken {
	htm.mu.RLock()
	defer htm.mu.RUnlock()

	tokens := make([]*HoneyToken, 0, len(htm.tokens))
	for _, token := range htm.tokens {
		tokens = append(tokens, token)
	}
	return tokens
}

// Helper functions
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func hashString(s string) string {
	// Simple hash for demo - in production use bcrypt or Argon2
	return fmt.Sprintf("$2a$10$%x", sha256.Sum256([]byte(s)))[:60]
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(userID, token string, duration time.Duration) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	session := &Session{
		ID:             generateSessionID(),
		UserID:         userID,
		Token:          token,
		CreatedAt:      now,
		ExpiresAt:      now.Add(duration),
		LastActivity:   now,
		RequiresReauth: false,
		Metadata:       make(map[string]interface{}),
	}

	sm.sessions[session.ID] = session
	return session
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	return session, exists
}

// ForceReauthentication marks a session as requiring re-authentication
// This is called by the Quantum Shield when a threat is detected
func (sm *SessionManager) ForceReauthentication(sessionID, reason string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.RequiresReauth = true
	session.ReauthReason = reason
	session.LastActivity = time.Now()

	log.Printf("RBAC: Forced re-authentication for session %s (user: %s) - Reason: %s",
		sessionID, session.UserID, reason)

	return nil
}

// ForceReauthenticationByUser forces all sessions for a user to re-authenticate
func (sm *SessionManager) ForceReauthenticationByUser(userID, reason string) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	count := 0
	for _, session := range sm.sessions {
		if session.UserID == userID {
			session.RequiresReauth = true
			session.ReauthReason = reason
			count++
		}
	}

	if count > 0 {
		log.Printf("RBAC: Forced re-authentication for %d sessions of user %s - Reason: %s",
			count, userID, reason)
	}

	return count
}

// ValidateSession checks if a session is valid and doesn't require re-authentication
func (sm *SessionManager) ValidateSession(sessionID string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	if session.RequiresReauth {
		return nil, fmt.Errorf("re-authentication required: %s", session.ReauthReason)
	}

	return session, nil
}

// UpdateSessionActivity updates the last activity timestamp
func (sm *SessionManager) UpdateSessionActivity(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.LastActivity = time.Now()
	}
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, sessionID)
}

// GetActiveSessions returns all active sessions for a user
func (sm *SessionManager) GetActiveSessions(userID string) []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var userSessions []*Session
	for _, session := range sm.sessions {
		if session.UserID == userID && time.Now().Before(session.ExpiresAt) {
			userSessions = append(userSessions, session)
		}
	}

	return userSessions
}

// CleanupExpiredSessions removes expired sessions
func (sm *SessionManager) CleanupExpiredSessions() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	count := 0
	now := time.Now()
	for id, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, id)
			count++
		}
	}

	return count
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("sess_%d_%s", time.Now().UnixNano(), randomString(8))
}

// randomString generates a random string
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
