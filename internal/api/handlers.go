package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

// Mock data for demonstration
var mockUsers = []User{
	{ID: 1, Username: "admin", Email: "admin@hades-toolkit.com", Role: "Administrator", Status: "active", LastLogin: time.Now().Add(-1 * time.Hour), Permissions: []string{"read", "write", "admin"}},
	{ID: 2, Username: "dev", Email: "dev@hades-toolkit.com", Role: "Developer", Status: "active", LastLogin: time.Now().Add(-1 * time.Hour), Permissions: []string{"read", "write", "dev", "user_management"}},
	{ID: 3, Username: "jsmith", Email: "jsmith@hades-toolkit.com", Role: "Security Analyst", Status: "active", LastLogin: time.Now().Add(-1 * time.Hour), Permissions: []string{"read", "write"}},
	{ID: 4, Username: "mrodriguez", Email: "mrodriguez@hades-toolkit.com", Role: "Security Engineer", Status: "active", LastLogin: time.Now().Add(-24 * time.Hour), Permissions: []string{"read", "write"}},
	{ID: 4, Username: "sbrown", Email: "sbrown@hades-toolkit.com", Role: "Auditor", Status: "inactive", LastLogin: time.Now().Add(-7 * 24 * time.Hour), Permissions: []string{"read"}},
}

var mockThreats = []Threat{
	{ID: 1, Type: "malware", Severity: "critical", Title: "Trojan.Dropper Detected", Source: "192.168.1.105", Status: "blocked", Timestamp: time.Now().Add(-2 * time.Hour), Description: "Malicious payload detected and blocked at network perimeter"},
	{ID: 2, Type: "phishing", Severity: "high", Title: "Suspicious Email Campaign", Source: "external", Status: "monitoring", Timestamp: time.Now().Add(-3 * time.Hour), Description: "Phishing attempt targeting corporate email accounts"},
	{ID: 3, Type: "brute-force", Severity: "medium", Title: "SSH Brute Force Attack", Source: "203.0.113.45", Status: "blocked", Timestamp: time.Now().Add(-4 * time.Hour), Description: "Multiple failed login attempts detected on SSH server"},
	{ID: 4, Type: "ddos", Severity: "low", Title: "DDoS Attack Mitigated", Source: "distributed", Status: "resolved", Timestamp: time.Now().Add(-5 * time.Hour), Description: "Volume-based DDoS attack successfully mitigated"},
}

var mockPolicies = []SecurityPolicy{
	{ID: 1, Name: "Password Policy", Status: "active", LastUpdated: time.Now().Add(-2 * time.Hour)},
	{ID: 2, Name: "Access Control", Status: "active", LastUpdated: time.Now().Add(-24 * time.Hour)},
	{ID: 3, Name: "Encryption Standards", Status: "active", LastUpdated: time.Now().Add(-72 * time.Hour)},
	{ID: 4, Name: "Audit Logging", Status: "warning", LastUpdated: time.Now().Add(-120 * time.Hour)},
}

var mockVulnerabilities = []Vulnerability{
	{ID: 1, Severity: "critical", Title: "Outdated SSL Certificate", Affected: "web-server", Status: "open"},
	{ID: 2, Severity: "high", Title: "Weak Password Policy", Affected: "auth-system", Status: "in-progress"},
	{ID: 3, Severity: "medium", Title: "Missing Security Headers", Affected: "api-endpoints", Status: "resolved"},
	{ID: 4, Severity: "low", Title: "Information Disclosure", Affected: "error-pages", Status: "open"},
}

var mockActivity = []Activity{
	{ID: 1, Type: "threat", Message: "Malware attack blocked", Time: "2 min ago", Severity: "high"},
	{ID: 2, Type: "user", Message: "New user registered", Time: "15 min ago", Severity: "low"},
	{ID: 3, Type: "system", Message: "Security scan completed", Time: "1 hour ago", Severity: "medium"},
	{ID: 4, Type: "threat", Message: "Suspicious login attempt", Time: "2 hours ago", Severity: "high"},
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// JWT Secret Key
var jwtSecret = []byte("hades-toolkit-secret-key")

// Environment-based authentication
var mockPasswords = map[string]string{
	// Development environment credentials only
	"dev":   "dev123",   // Dev access credentials for development
	"admin": "admin123", // Admin access credentials for testing
	// Production users - no default passwords
	"jsmith":     "secureUserPass456!", // This should be a bcrypt hash in production
	"mrodriguez": "secureUserPass789!", // This should be a bcrypt hash in production
}

// Authentication Handlers
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate credentials - demo credentials allowed for testing
	if !s.validateCredentials(req.Username, req.Password) {
		s.writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Get user information
	user, exists := s.getUserByUsername(req.Username)
	if !exists {
		s.writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Update last login
	user.LastLogin = time.Now()

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := LoginResponse{
		Token: tokenString,
		User:  user,
	}

	s.writeSuccess(w, response)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.writeSuccess(w, map[string]string{"message": "Logged out successfully"})
}

func (s *Server) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Mock token refresh
	user := User{
		ID:       1,
		Username: "admin",
		Role:     "Administrator",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to refresh token")
		return
	}

	s.writeSuccess(w, map[string]string{"token": tokenString})
}

func (s *Server) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := User{
		ID:          1,
		Username:    "admin",
		Email:       "admin@hades-toolkit.com",
		Role:        "Administrator",
		Status:      "active",
		LastLogin:   time.Now().Add(-2 * time.Hour),
		Permissions: []string{"read", "write", "admin"},
	}
	s.writeSuccess(w, user)
}

// Dashboard Handlers
func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := DashboardMetrics{
		SecurityScore:  98 + rand.Intn(3),
		ActiveThreats:  1 + rand.Intn(5),
		BlockedAttacks: 1200 + rand.Intn(100),
		SystemHealth:   95 + rand.Intn(5),
		ActiveUsers:    20 + rand.Intn(10),
	}
	s.writeSuccess(w, metrics)
}

func (s *Server) handleGetActivity(w http.ResponseWriter, r *http.Request) {
	s.writeSuccess(w, mockActivity)
}

func (s *Server) handleGetSystemStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"database":        "operational",
		"api_server":      "running",
		"security_engine": "active",
		"backup_service":  "scheduled",
		"uptime":          "99.9%",
		"last_restart":    time.Now().Add(-24 * time.Hour),
	}
	s.writeSuccess(w, status)
}

func (s *Server) handleGetSecurityOverview(w http.ResponseWriter, r *http.Request) {
	overview := map[string]interface{}{
		"score":           98,
		"threats_blocked": 1247,
		"scans_completed": 156,
		"policies_active": 4,
		"last_scan":       time.Now().Add(-2 * time.Hour),
	}
	s.writeSuccess(w, overview)
}

// Threat Handlers
func (s *Server) handleGetThreats(w http.ResponseWriter, r *http.Request) {
	s.writeSuccess(w, mockThreats)
}

func (s *Server) handleGetThreat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid threat ID")
		return
	}

	for _, threat := range mockThreats {
		if threat.ID == id {
			s.writeSuccess(w, threat)
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "Threat not found")
}

func (s *Server) handleUpdateThreatStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid threat ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Update mock threat status
	for i, threat := range mockThreats {
		if threat.ID == id {
			mockThreats[i].Status = req.Status
			s.writeSuccess(w, map[string]string{"status": "updated"})
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "Threat not found")
}

func (s *Server) handleGetThreatStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"total_threats":     len(mockThreats),
		"blocked_threats":   3,
		"active_threats":    1,
		"critical_threats":  1,
		"high_threats":      1,
		"medium_threats":    1,
		"low_threats":       1,
		"threats_today":     12,
		"threats_this_week": 47,
	}
	s.writeSuccess(w, stats)
}

func (s *Server) handleGetThreatFeed(w http.ResponseWriter, r *http.Request) {
	feed := []map[string]interface{}{
		{
			"id":          1,
			"type":        "malware",
			"severity":    "critical",
			"title":       "New malware variant detected",
			"timestamp":   time.Now().Add(-10 * time.Minute),
			"source":      "internal",
			"description": "Unknown malware variant detected in network traffic",
		},
		{
			"id":          2,
			"type":        "phishing",
			"severity":    "high",
			"title":       "Phishing campaign targeting finance",
			"timestamp":   time.Now().Add(-30 * time.Minute),
			"source":      "external",
			"description": "Phishing emails targeting financial department detected",
		},
	}
	s.writeSuccess(w, feed)
}

// User Handlers
func (s *Server) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	s.writeSuccess(w, mockUsers)
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	for _, user := range mockUsers {
		if user.ID == id {
			s.writeSuccess(w, user)
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "User not found")
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Assign new ID
	user.ID = len(mockUsers) + 1
	user.Status = "active"
	user.LastLogin = time.Now()

	mockUsers = append(mockUsers, user)
	s.writeSuccess(w, user)
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	for i, existingUser := range mockUsers {
		if existingUser.ID == id {
			mockUsers[i] = user
			mockUsers[i].ID = id
			s.writeSuccess(w, user)
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "User not found")
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	for i, user := range mockUsers {
		if user.ID == id {
			mockUsers = append(mockUsers[:i], mockUsers[i+1:]...)
			s.writeSuccess(w, map[string]string{"message": "User deleted"})
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "User not found")
}

func (s *Server) handleGetUserStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"total_users":     len(mockUsers),
		"active_users":    3,
		"inactive_users":  1,
		"admin_users":     1,
		"users_today":     2,
		"users_this_week": 5,
	}
	s.writeSuccess(w, stats)
}

func (s *Server) handleGetUserRoles(w http.ResponseWriter, r *http.Request) {
	roles := []map[string]interface{}{
		{"id": 1, "name": "Administrator", "permissions": []string{"read", "write", "admin"}},
		{"id": 2, "name": "Security Analyst", "permissions": []string{"read", "write"}},
		{"id": 3, "name": "Security Engineer", "permissions": []string{"read", "write"}},
		{"id": 4, "name": "Auditor", "permissions": []string{"read"}},
	}
	s.writeSuccess(w, roles)
}

// Security Handlers
func (s *Server) handleGetPolicies(w http.ResponseWriter, r *http.Request) {
	s.writeSuccess(w, mockPolicies)
}

func (s *Server) handleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid policy ID")
		return
	}

	var policy SecurityPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	for i, existingPolicy := range mockPolicies {
		if existingPolicy.ID == id {
			mockPolicies[i] = policy
			mockPolicies[i].ID = id
			mockPolicies[i].LastUpdated = time.Now()
			s.writeSuccess(w, policy)
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "Policy not found")
}

func (s *Server) handleGetVulnerabilities(w http.ResponseWriter, r *http.Request) {
	s.writeSuccess(w, mockVulnerabilities)
}

func (s *Server) handleUpdateVulnerability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid vulnerability ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	for i, vuln := range mockVulnerabilities {
		if vuln.ID == id {
			mockVulnerabilities[i].Status = req.Status
			s.writeSuccess(w, map[string]string{"status": "updated"})
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "Vulnerability not found")
}

func (s *Server) handleGetSecurityScore(w http.ResponseWriter, r *http.Request) {
	score := map[string]interface{}{
		"overall_score": 98,
		"categories": map[string]int{
			"authentication": 95,
			"authorization":  98,
			"encryption":     100,
			"monitoring":     96,
			"compliance":     99,
		},
		"last_calculated": time.Now(),
		"trend":           "improving",
	}
	s.writeSuccess(w, score)
}

func (s *Server) handleRunSecurityScan(w http.ResponseWriter, r *http.Request) {
	scan := map[string]interface{}{
		"scan_id":              fmt.Sprintf("scan-%d", time.Now().Unix()),
		"status":               "started",
		"scan_type":            "comprehensive",
		"started_at":           time.Now(),
		"estimated_completion": time.Now().Add(5 * time.Minute),
	}
	s.writeSuccess(w, scan)
}

func (s *Server) handleGetAuditLogs(w http.ResponseWriter, r *http.Request) {
	logs := []map[string]interface{}{
		{
			"id":        1,
			"user":      "admin",
			"action":    "login",
			"resource":  "system",
			"timestamp": time.Now().Add(-2 * time.Hour),
			"ip":        "192.168.1.100",
			"status":    "success",
		},
		{
			"id":        2,
			"user":      "jsmith",
			"action":    "view",
			"resource":  "threats",
			"timestamp": time.Now().Add(-1 * time.Hour),
			"ip":        "192.168.1.105",
			"status":    "success",
		},
	}
	s.writeSuccess(w, logs)
}

// Backup status handler
func (s *Server) handleBackupStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"backup_count":     7,
		"backup_size":      "2.4GB",
		"compression":      "enabled",
		"encryption":       "AES-256",
		"last_backup":      "2026-05-04T19:32:00.625248936+01:00",
		"next_backup":      "2026-05-05T19:32:00.625250276+01:00",
		"retention_days":   30,
		"status":           "healthy",
		"storage_location": "/backups/hades/",
	}

	s.writeSuccess(w, response)
}

// handleAnalyticsSummary handles analytics summary requests
func (s *Server) handleAnalyticsSummary(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"total_events":       1247,
		"threats_detected":   127,
		"incidents_resolved": 89,
		"scan_results":       45,
		"risk_score":         6.5,
		"compliance_score":   87.3,
		"uptime":             "99.9%",
		"response_time":      "0.08s",
	}

	s.writeSuccess(w, response)
}

// handleThreatAlerts handles threat alert requests
func (s *Server) handleThreatAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		alerts := []map[string]interface{}{
			{
				"id":        "alert_001",
				"type":      "malware",
				"severity":  "high",
				"source_ip": "192.168.1.100",
				"timestamp": "2026-05-04T21:30:00Z",
				"status":    "active",
			},
			{
				"id":        "alert_002",
				"type":      "suspicious_activity",
				"severity":  "medium",
				"source_ip": "10.0.0.50",
				"timestamp": "2026-05-04T21:25:00Z",
				"status":    "investigating",
			},
		}
		s.writeSuccess(w, map[string]interface{}{"alerts": alerts, "count": len(alerts)})
	} else {
		// Handle POST request for creating alerts
		s.writeSuccess(w, map[string]interface{}{"success": true, "alert_id": "alert_003"})
	}
}

// handleThreatDetection handles threat detection requests
func (s *Server) handleThreatDetection(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"analysis_id":     "analysis_001",
		"status":          "completed",
		"threats_found":   3,
		"confidence":      0.94,
		"processing_time": "2.3s",
		"recommendations": []string{"Block IP 192.168.1.100", "Update firewall rules", "Scan affected systems"},
	}

	s.writeSuccess(w, response)
}

// handleThreatIntelligence handles threat intelligence requests
func (s *Server) handleThreatIntelligence(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"feeds_active":         12,
		"indicators_processed": 4567,
		"last_update":          "2026-05-04T21:00:00Z",
		"reputation_sources":   []string{"VirusTotal", "AbuseIPDB", "OTX"},
		"threat_level":         "medium",
	}

	s.writeSuccess(w, response)
}

// handleThreatStatus handles threat status requests
func (s *Server) handleThreatStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"detector_status":     "active",
		"ml_models_loaded":    5,
		"processing_queue":    23,
		"detection_rate":      97.6,
		"false_positive_rate": 2.4,
		"uptime":              "24h",
	}

	s.writeSuccess(w, response)
}

// Worker status handler
func (s *Server) handleWorkerStatus(w http.ResponseWriter, r *http.Request) {
	workers := []map[string]interface{}{
		{
			"id":              "worker-1",
			"status":          "active",
			"tasks_completed": 1247,
			"current_task":    "threat_analysis",
			"uptime":          "4h32m",
			"memory_usage":    "45MB",
			"cpu_usage":       "12%",
		},
		{
			"id":              "worker-2",
			"status":          "active",
			"tasks_completed": 892,
			"current_task":    "scan_processing",
			"uptime":          "4h32m",
			"memory_usage":    "38MB",
			"cpu_usage":       "8%",
		},
		{
			"id":              "worker-3",
			"status":          "idle",
			"tasks_completed": 756,
			"current_task":    "none",
			"uptime":          "4h32m",
			"memory_usage":    "32MB",
			"cpu_usage":       "2%",
		},
		{
			"id":              "worker-4",
			"status":          "active",
			"tasks_completed": 623,
			"current_task":    "data_processing",
			"uptime":          "4h32m",
			"memory_usage":    "41MB",
			"cpu_usage":       "15%",
		},
		{
			"id":              "worker-5",
			"status":          "active",
			"tasks_completed": 445,
			"current_task":    "report_generation",
			"uptime":          "4h32m",
			"memory_usage":    "29MB",
			"cpu_usage":       "6%",
		},
	}
	s.writeSuccess(w, workers)
}

// WebSocket status handler
func (s *Server) handleWebSocketStatus(w http.ResponseWriter, r *http.Request) {
	wsStatus := map[string]interface{}{
		"status":            "active",
		"connected_clients": 24,
		"total_connections": 156,
		"messages_sent":     8923,
		"messages_received": 1247,
		"uptime":            "4h32m",
		"protocol":          "WebSocket",
		"version":           "13",
	}
	s.writeSuccess(w, wsStatus)
	logs := []map[string]interface{}{
		{
			"id":        1,
			"user":      "admin",
			"action":    "login",
			"resource":  "system",
			"timestamp": time.Now().Add(-2 * time.Hour),
			"ip":        "192.168.1.100",
			"status":    "success",
		},
		{
			"id":        2,
			"user":      "jsmith",
			"action":    "view",
			"resource":  "threats",
			"timestamp": time.Now().Add(-1 * time.Hour),
			"ip":        "192.168.1.105",
			"status":    "success",
		},
	}
	s.writeSuccess(w, logs)
}

// Root route
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Hades Toolkit API Server",
		"version": "2.0.0",
		"status":  "running",
		"versions": map[string]interface{}{
			"v1": map[string]interface{}{
				"status":      "legacy",
				"description": "Deprecated version - migrate to v2",
				"endpoints": map[string]string{
					"health":    "/api/v1/health",
					"auth":      "/api/v1/auth/*",
					"dashboard": "/api/v1/dashboard/*",
					"threats":   "/api/v1/threats/*",
					"users":     "/api/v1/users/*",
					"security":  "/api/v1/security/*",
				},
			},
			"v2": map[string]interface{}{
				"status":      "preferred",
				"description": "Current stable version with enhanced features",
				"endpoints": map[string]string{
					"health":    "/api/v2/health",
					"auth":      "/api/v2/auth/*",
					"dashboard": "/api/v2/dashboard/*",
					"analytics": "/api/v2/analytics",
					"webhooks":  "/api/v2/webhooks",
					"threats":   "/api/v2/threats/*",
					"users":     "/api/v2/users/*",
					"security":  "/api/v2/security/*",
				},
			},
			"v3": map[string]interface{}{
				"status":      "beta",
				"description": "Future development version with ML features",
				"endpoints": map[string]string{
					"health":     "/api/v3/health",
					"auth":       "/api/v3/auth/*",
					"dashboard":  "/api/v3/dashboard/*",
					"analytics":  "/api/v3/analytics",
					"ml":         "/api/v3/ml/*",
					"automation": "/api/v3/automation/*",
				},
			},
		},
		"default_version":     "v2",
		"recommended_version": "v2",
		"version_discovery":   "/api/versions",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// API Info Handler
func (s *Server) handleAPIInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"name":        "Hades Toolkit API",
		"description": "Enterprise Security Platform API",
		"version":     "2.0.0",
		"base_url":    "/api",
		"documentation": map[string]string{
			"openapi": "/api/docs",
			"swagger": "/api/swagger",
		},
		"contact": map[string]string{
			"email":   "api@hades-toolkit.com",
			"support": "https://support.hades-toolkit.com",
		},
		"license": map[string]string{
			"name": "MIT",
			"url":  "https://opensource.org/licenses/MIT",
		},
	}
	s.writeSuccess(w, info)
}

// Version Discovery Handler - now handled by the new versioning system
// This functionality is provided by versionMgr.VersionInfoHandler

// V2 Analytics Handler
func (s *Server) handleV2Analytics(w http.ResponseWriter, r *http.Request) {
	analytics := map[string]interface{}{
		"api_metrics": map[string]interface{}{
			"requests_per_second":   145.7,
			"average_response_time": "120ms",
			"error_rate":            0.02,
			"uptime":                "99.9%",
		},
		"top_endpoints": []map[string]interface{}{
			{"path": "/api/v2/dashboard/metrics", "requests": 1247, "avg_response": 45.2},
			{"path": "/api/v2/threats", "requests": 892, "avg_response": 67.8},
			{"path": "/api/v2/users", "requests": 456, "avg_response": 34.1},
		},
		"user_analytics": map[string]interface{}{
			"active_users":         24,
			"total_sessions":       156,
			"avg_session_duration": "45m",
		},
		"security_metrics": map[string]interface{}{
			"blocked_requests":       1247,
			"failed_authentications": 23,
			"suspicious_activities":  5,
		},
		"performance": map[string]interface{}{
			"cpu_usage":    45.2,
			"memory_usage": 67.8,
			"disk_usage":   23.4,
			"network_io":   12.1,
		},
	}
	s.writeSuccess(w, analytics)
}

// V2 Webhooks Handler
func (s *Server) handleV2Webhooks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		webhooks := []map[string]interface{}{
			{
				"id":             "webhook-001",
				"name":           "Threat Alert Webhook",
				"url":            "https://customer.example.com/webhooks/threats",
				"events":         []string{"threat.created", "threat.updated", "threat.resolved"},
				"active":         true,
				"created_at":     time.Now().Add(-30 * 24 * time.Hour),
				"last_triggered": time.Now().Add(-2 * time.Hour),
			},
			{
				"id":             "webhook-002",
				"name":           "Security Score Webhook",
				"url":            "https://monitoring.example.com/webhooks/security",
				"events":         []string{"security.score.changed"},
				"active":         true,
				"created_at":     time.Now().Add(-15 * 24 * time.Hour),
				"last_triggered": time.Now().Add(-1 * time.Hour),
			},
		}
		s.writeSuccess(w, webhooks)

	case "POST":
		var webhook struct {
			Name   string   `json:"name"`
			URL    string   `json:"url"`
			Events []string `json:"events"`
		}

		if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		newWebhook := map[string]interface{}{
			"id":             "webhook-" + fmt.Sprintf("%03d", rand.Intn(1000)),
			"name":           webhook.Name,
			"url":            webhook.URL,
			"events":         webhook.Events,
			"active":         true,
			"created_at":     time.Now(),
			"last_triggered": nil,
		}

		s.writeSuccess(w, newWebhook)
	}
}

// Helper methods for authentication
func (s *Server) validateCredentials(username, password string) bool {
	// Check if user exists and password matches
	expectedPassword, exists := mockPasswords[username]
	if !exists {
		return false
	}

	// In production, this would use bcrypt to compare hashed passwords
	return password == expectedPassword
}

func (s *Server) getUserByUsername(username string) (User, bool) {
	// Find user in mockUsers slice
	for _, user := range mockUsers {
		if user.Username == username {
			return user, true
		}
	}
	return User{}, false
}

// Health Check
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "2.0.0",
		"uptime":    time.Since(time.Now().Add(-24 * time.Hour)).String(),
		"api_versions": map[string]interface{}{
			"v1": map[string]string{"status": "stable"},
			"v2": map[string]string{"status": "stable"},
			"v3": map[string]string{"status": "beta"},
		},
	}
	s.writeSuccess(w, health)
}
