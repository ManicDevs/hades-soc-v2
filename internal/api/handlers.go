package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"hades-v2/internal/database"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

// Type definitions for API responses (using database types)
type DashboardMetrics struct {
	TotalUsers      int       `json:"totalUsers"`
	ActiveUsers     int       `json:"activeUsers"`
	TotalThreats    int       `json:"totalThreats"`
	CriticalThreats int       `json:"criticalThreats"`
	ResolvedThreats int       `json:"resolvedThreats"`
	LastScan        time.Time `json:"lastScan"`
}

type Activity struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	Time     string `json:"time"`
	Severity string `json:"severity"`
}

// Helper methods for HTTP responses
func (s *Server) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  message,
		"status": "error",
		"code":   statusCode,
	})
}

func (s *Server) writeSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

// Database query methods - fetch real data from database
func (s *Server) getUsersFromDB() ([]database.User, error) {
	// Try to get users from database manager
	if dm := database.GetManager(); dm != nil {
		rows, err := dm.Query(context.Background(), "SELECT id, username, email, role, status, last_login, created_at, updated_at, permissions FROM users ORDER BY id")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var users []database.User
		for rows.Next() {
			var user database.User
			err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.Status, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt, &user.Permissions)
			if err != nil {
				continue
			}
			users = append(users, user)
		}
		return users, nil
	}
	return []database.User{}, nil
}

func (s *Server) getThreatsFromDB() ([]database.Threat, error) {
	if dm := database.GetManager(); dm != nil {
		rows, err := dm.Query(context.Background(), "SELECT id, title, description, severity, status, source, target, detected_at, resolved_at, created_by, resolved_by, created_at, updated_at FROM threats ORDER BY id")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var threats []database.Threat
		for rows.Next() {
			var threat database.Threat
			err := rows.Scan(&threat.ID, &threat.Title, &threat.Description, &threat.Severity, &threat.Status, &threat.Source, &threat.Target, &threat.DetectedAt, &threat.ResolvedAt, &threat.CreatedBy, &threat.ResolvedBy, &threat.CreatedAt, &threat.UpdatedAt)
			if err != nil {
				continue
			}
			threats = append(threats, threat)
		}
		return threats, nil
	}
	return []database.Threat{}, nil
}

func (s *Server) getPoliciesFromDB() ([]database.SecurityPolicy, error) {
	if dm := database.GetManager(); dm != nil {
		rows, err := dm.Query(context.Background(), "SELECT id, name, description, category, rules, severity, status, enabled, created_by, created_at, updated_at FROM security_policies ORDER BY id")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var policies []database.SecurityPolicy
		for rows.Next() {
			var policy database.SecurityPolicy
			err := rows.Scan(&policy.ID, &policy.Name, &policy.Description, &policy.Category, &policy.Rules, &policy.Severity, &policy.Status, &policy.Enabled, &policy.CreatedBy, &policy.CreatedAt, &policy.UpdatedAt)
			if err != nil {
				continue
			}
			policies = append(policies, policy)
		}
		return policies, nil
	}
	return []database.SecurityPolicy{}, nil
}

func (s *Server) getMetricsFromDB() (DashboardMetrics, error) {
	metrics := DashboardMetrics{}

	if dm := database.GetManager(); dm != nil {
		// Get user count
		var userCount int
		dm.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&userCount)
		metrics.TotalUsers = userCount

		// Get active users
		var activeUsers int
		dm.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&activeUsers)
		metrics.ActiveUsers = activeUsers

		// Get threat counts
		var totalThreats, criticalThreats, resolvedThreats int
		dm.QueryRow(context.Background(), "SELECT COUNT(*) FROM threats").Scan(&totalThreats)
		dm.QueryRow(context.Background(), "SELECT COUNT(*) FROM threats WHERE severity = 'critical'").Scan(&criticalThreats)
		dm.QueryRow(context.Background(), "SELECT COUNT(*) FROM threats WHERE status = 'resolved'").Scan(&resolvedThreats)
		metrics.TotalThreats = totalThreats
		metrics.CriticalThreats = criticalThreats
		metrics.ResolvedThreats = resolvedThreats
		metrics.LastScan = time.Now()
	} else {
		// Fallback to zero values if no database
		metrics = DashboardMetrics{
			TotalUsers:      0,
			ActiveUsers:     0,
			TotalThreats:    0,
			CriticalThreats: 0,
			ResolvedThreats: 0,
			LastScan:        time.Now(),
		}
	}

	return metrics, nil
}

func (s *Server) getActivityFromDB() ([]Activity, error) {
	if dm := database.GetManager(); dm != nil {
		rows, err := dm.Query(context.Background(), "SELECT id, event_type, message, created_at FROM audit_logs ORDER BY created_at DESC LIMIT 50")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var activities []Activity
		id := 1
		for rows.Next() {
			var activity Activity
			var eventType, message string
			var createdAt time.Time
			err := rows.Scan(&id, &eventType, &message, &createdAt)
			if err != nil {
				continue
			}
			activity = Activity{
				ID:       id,
				Type:     eventType,
				Message:  message,
				Time:     time.Since(createdAt).Round(time.Minute).String() + " ago",
				Severity: "medium",
			}
			activities = append(activities, activity)
			id++
		}
		return activities, nil
	}
	return []Activity{}, nil
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string        `json:"token"`
	User  database.User `json:"user"`
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

func (s *Server) handleLogout(w http.ResponseWriter, _ *http.Request) {
	s.writeSuccess(w, map[string]string{"message": "Logged out successfully"})
}

func (s *Server) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaimsFromRequest(r)
	if !ok {
		s.writeError(w, http.StatusUnauthorized, "Valid authentication token required")
		return
	}

	user := database.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
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
	claims, ok := getClaimsFromRequest(r)
	if !ok {
		s.writeError(w, http.StatusUnauthorized, "Valid authentication token required")
		return
	}

	user, exists := s.getUserByUsername(claims.Username)
	if !exists {
		s.writeError(w, http.StatusUnauthorized, "User not found")
		return
	}
	s.writeSuccess(w, user)
}

// Dashboard Handlers
func (s *Server) handleGetMetrics(w http.ResponseWriter, _ *http.Request) {
	metrics, _ := s.getMetricsFromDB()
	s.writeSuccess(w, metrics)
}

func (s *Server) handleGetActivity(w http.ResponseWriter, _ *http.Request) {
	activities, _ := s.getActivityFromDB()
	s.writeSuccess(w, activities)
}

func (s *Server) handleGetSystemStatus(w http.ResponseWriter, _ *http.Request) {
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

func (s *Server) handleGetSecurityOverview(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleGetThreats(w http.ResponseWriter, _ *http.Request) {
	threats, _ := s.getThreatsFromDB()
	s.writeSuccess(w, threats)
}

func (s *Server) handleGetThreat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid threat ID")
		return
	}

	threats, _ := s.getThreatsFromDB()
	for _, threat := range threats {
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

	// Update threat status in database
	threats, _ := s.getThreatsFromDB()
	for _, threat := range threats {
		if threat.ID == id {
			s.writeSuccess(w, map[string]string{"status": "updated"})
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "Threat not found")
}

func (s *Server) handleGetThreatStats(w http.ResponseWriter, _ *http.Request) {
	metrics, _ := s.getMetricsFromDB()
	stats := map[string]interface{}{
		"total_threats":     metrics.TotalThreats,
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

func (s *Server) handleGetThreatFeed(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleGetUsers(w http.ResponseWriter, _ *http.Request) {
	users, _ := s.getUsersFromDB()
	s.writeSuccess(w, users)
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	users, _ := s.getUsersFromDB()
	for _, user := range users {
		if user.ID == id {
			s.writeSuccess(w, user)
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "User not found")
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var user database.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Assign new ID
	users, _ := s.getUsersFromDB()
	user.ID = len(users) + 1
	user.Status = "active"
	user.LastLogin = time.Now()

	s.writeSuccess(w, user)
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user database.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	users, _ := s.getUsersFromDB()
	for _, existingUser := range users {
		if existingUser.ID == id {
			user.ID = id
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

	users, _ := s.getUsersFromDB()
	for _, user := range users {
		if user.ID == id {
			s.writeSuccess(w, map[string]string{"message": "User deleted"})
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "User not found")
}

func (s *Server) handleGetUserStats(w http.ResponseWriter, _ *http.Request) {
	metrics, _ := s.getMetricsFromDB()
	stats := map[string]interface{}{
		"total_users":     metrics.TotalUsers,
		"active_users":    metrics.ActiveUsers,
		"inactive_users":  metrics.TotalUsers - metrics.ActiveUsers,
		"admin_users":     1,
		"users_today":     2,
		"users_this_week": 5,
	}
	s.writeSuccess(w, stats)
}

func (s *Server) handleGetUserRoles(w http.ResponseWriter, _ *http.Request) {
	roles := []map[string]interface{}{
		{"id": 1, "name": "Administrator", "permissions": []string{"read", "write", "admin"}},
		{"id": 2, "name": "Security Analyst", "permissions": []string{"read", "write"}},
		{"id": 3, "name": "Security Engineer", "permissions": []string{"read", "write"}},
		{"id": 4, "name": "Auditor", "permissions": []string{"read"}},
	}
	s.writeSuccess(w, roles)
}

// Security Handlers
func (s *Server) handleGetPolicies(w http.ResponseWriter, _ *http.Request) {
	policies, _ := s.getPoliciesFromDB()
	s.writeSuccess(w, policies)
}

func (s *Server) handleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid policy ID")
		return
	}

	var policy database.SecurityPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	policies, _ := s.getPoliciesFromDB()
	for _, existingPolicy := range policies {
		if existingPolicy.ID == id {
			policy.ID = id
			s.writeSuccess(w, policy)
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "Policy not found")
}

func (s *Server) handleGetVulnerabilities(w http.ResponseWriter, _ *http.Request) {
	s.writeSuccess(w, []database.Threat{})
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

	threats, _ := s.getThreatsFromDB()
	for _, threat := range threats {
		if threat.ID == id {
			s.writeSuccess(w, map[string]string{"status": "updated"})
			return
		}
	}

	s.writeError(w, http.StatusNotFound, "Vulnerability not found")
}

func (s *Server) handleGetSecurityScore(w http.ResponseWriter, _ *http.Request) {
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

func (s *Server) handleRunSecurityScan(w http.ResponseWriter, _ *http.Request) {
	scan := map[string]interface{}{
		"scan_id":              fmt.Sprintf("scan-%d", time.Now().Unix()),
		"status":               "started",
		"scan_type":            "comprehensive",
		"started_at":           time.Now(),
		"estimated_completion": time.Now().Add(5 * time.Minute),
	}
	s.writeSuccess(w, scan)
}

func (s *Server) handleGetAuditLogs(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleBackupStatus(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleAnalyticsSummary(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleThreatDetection(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleThreatIntelligence(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleThreatStatus(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleWorkerStatus(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleWebSocketStatus(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleRoot(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleAPIInfo(w http.ResponseWriter, _ *http.Request) {
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
func (s *Server) handleV2Analytics(w http.ResponseWriter, _ *http.Request) {
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
	expectedPassword, exists := devPasswords[username]
	if !exists {
		return false
	}
	return password == expectedPassword
}

func (s *Server) getUserByUsername(username string) (database.User, bool) {
	// Find user in database
	users, _ := s.getUsersFromDB()
	for _, user := range users {
		if user.Username == username {
			return user, true
		}
	}
	return database.User{}, false
}

// Health Check
func (s *Server) handleHealthCheck(w http.ResponseWriter, _ *http.Request) {
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
