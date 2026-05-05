package api

import (
	"encoding/json"
	"fmt"
	"hades-v2/internal/api/versioning"
	"hades-v2/internal/database"
	"hades-v2/internal/websocket"
	"hades-v2/internal/workers"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	router                 *mux.Router
	port                   int
	versionMgr             *versioning.VersionManager
	versionInt             *versioning.ServerIntegration
	database               database.Database
	workerPool             *workers.WorkerPool
	aiEndpoints            *AIEndpoints
	analyticsEndpoints     *AnalyticsEndpoints
	threatHuntingEndpoints *ThreatHuntingEndpoints
	blockchainEndpoints    *BlockchainEndpoints
	zeroTrustEndpoints     *ZeroTrustEndpoints
	quantumEndpoints       *QuantumEndpoints
	siemEndpoints          *SIEMEndpoints
	incidentEndpoints      *IncidentEndpoints
	threatEndpoints        *ThreatEndpoints
	kubernetesEndpoints    *KubernetesEndpoints
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type User struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	LastLogin   time.Time `json:"lastLogin"`
	Permissions []string  `json:"permissions"`
}

type Threat struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

type SecurityPolicy struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type Vulnerability struct {
	ID       int    `json:"id"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Affected string `json:"affected"`
	Status   string `json:"status"`
}

type DashboardMetrics struct {
	SecurityScore  int `json:"securityScore"`
	ActiveThreats  int `json:"activeThreats"`
	BlockedAttacks int `json:"blockedAttacks"`
	SystemHealth   int `json:"systemHealth"`
	ActiveUsers    int `json:"activeUsers"`
}

type Activity struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	Time     string `json:"time"`
	Severity string `json:"severity"`
}

func NewServer(port int) *Server {
	// Create database connection (default to SQLite for development)
	dbConfig := database.DatabaseConfig{
		Type:     database.SQLite,
		Database: "./hades_toolkit.db",
	}

	db := database.NewDatabase(database.SQLite)
	if err := db.Connect(dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	migrator := database.NewMigrator(db)
	if err := migrator.RunMigrations("./migrations"); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	// Create version manager with hierarchy
	versionConfig := versioning.DefaultConfig()
	versionMgr := versioning.NewVersionManager(versionConfig)
	versionInt := versioning.NewServerIntegration(versionMgr)

	// Create worker pool
	workerPool := workers.NewWorkerPool(5, db)

	// Create AI endpoints
	aiEndpoints, err := NewAIEndpoints()
	if err != nil {
		log.Printf("Warning: Failed to create AI endpoints: %v", err)
	}

	// Create analytics endpoints
	analyticsEndpoints, err := NewAnalyticsEndpoints(db)
	if err != nil {
		log.Printf("Warning: Failed to create analytics endpoints: %v", err)
	}

	// Create threat hunting endpoints
	var threatHuntingEndpoints *ThreatHuntingEndpoints
	if aiEndpoints != nil && analyticsEndpoints != nil {
		threatHuntingEndpoints, err = NewThreatHuntingEndpoints(aiEndpoints.GetThreatEngine(), analyticsEndpoints.GetAnalyticsEngine())
		if err != nil {
			log.Printf("Warning: Failed to create threat hunting endpoints: %v", err)
		}
	}

	// Create blockchain endpoints
	blockchainEndpoints, err := NewBlockchainEndpoints(db)
	if err != nil {
		log.Printf("Warning: Failed to create blockchain endpoints: %v", err)
	}

	// Create zero-trust endpoints
	zeroTrustEndpoints, err := NewZeroTrustEndpoints(db)
	if err != nil {
		log.Printf("Warning: Failed to create zero-trust endpoints: %v", err)
	}

	// Create quantum endpoints
	quantumEndpoints, err := NewQuantumEndpoints(db)
	if err != nil {
		log.Printf("Warning: Failed to create quantum endpoints: %v", err)
	}

	// Create SIEM endpoints
	siemEndpoints, err := NewSIEMEndpoints(db)
	if err != nil {
		log.Printf("Warning: Failed to create SIEM endpoints: %v", err)
	}

	// Create incident response endpoints
	incidentEndpoints, err := NewIncidentEndpoints(db)
	if err != nil {
		log.Printf("Warning: Failed to create incident endpoints: %v", err)
	}

	// Create threat modeling endpoints
	threatEndpoints, err := NewThreatEndpoints(db)
	if err != nil {
		log.Printf("Warning: Failed to create threat endpoints: %v", err)
	}

	// Create Kubernetes endpoints
	kubernetesEndpoints, err := NewKubernetesEndpoints(db)
	if err != nil {
		log.Printf("Warning: Failed to create Kubernetes endpoints: %v", err)
	}

	s := &Server{
		router:                 mux.NewRouter(),
		port:                   port,
		versionMgr:             versionMgr,
		versionInt:             versionInt,
		database:               db,
		workerPool:             workerPool,
		aiEndpoints:            aiEndpoints,
		analyticsEndpoints:     analyticsEndpoints,
		threatHuntingEndpoints: threatHuntingEndpoints,
		blockchainEndpoints:    blockchainEndpoints,
		zeroTrustEndpoints:     zeroTrustEndpoints,
		quantumEndpoints:       quantumEndpoints,
		siemEndpoints:          siemEndpoints,
		incidentEndpoints:      incidentEndpoints,
		threatEndpoints:        threatEndpoints,
		kubernetesEndpoints:    kubernetesEndpoints,
	}
	s.setupRoutes()

	// Start worker pool
	workerPool.Start()

	return s
}

func (s *Server) setupRoutes() {
	// Apply versioning middleware
	s.router.Use(s.versionInt.Middleware)

	// Root and version discovery routes
	s.router.HandleFunc("/", s.handleRoot).Methods("GET")
	s.router.HandleFunc("/api", s.handleAPIInfo).Methods("GET")

	// Version discovery endpoints
	s.router.HandleFunc("/api/versions", s.versionMgr.VersionInfoHandler).Methods("GET")
	s.router.HandleFunc("/api/version", s.versionMgr.VersionHandler).Methods("GET")
	s.router.HandleFunc("/api/migration", s.versionMgr.MigrationHandler).Methods("GET")

	// WebSocket endpoint for real-time updates
	s.router.HandleFunc("/ws", s.handleWebSocket).Methods("GET")

	// Events WebSocket endpoint for agent activity monitoring
	s.router.HandleFunc("/api/v2/ws/events", s.handleEventsWebSocket).Methods("GET")

	// Agent Stream WebSocket endpoint for thought stream and actions
	s.router.HandleFunc("/ws/agent-stream", s.handleAgentStreamWebSocket).Methods("GET")

	// API v1 routes (existing)
	s.setupV1Routes()

	// API v2 routes (enhanced)
	s.setupV2Routes()

	// AI-powered endpoints
	if s.aiEndpoints != nil {
		// Mount AI endpoints router
		s.router.PathPrefix("/api/v2/ai").Handler(s.aiEndpoints.GetRouter())
	}

	// Advanced analytics endpoints
	if s.analyticsEndpoints != nil {
		// Mount analytics endpoints router
		s.router.PathPrefix("/api/v2/analytics").Handler(s.analyticsEndpoints.GetRouter())
	}

	// Threat hunting endpoints
	if s.threatHuntingEndpoints != nil {
		// Mount threat hunting endpoints router
		s.router.PathPrefix("/api/v2/threat-hunting").Handler(s.threatHuntingEndpoints.GetRouter())
	}

	// Blockchain audit endpoints
	if s.blockchainEndpoints != nil {
		// Mount blockchain endpoints router
		s.router.PathPrefix("/api/v2/blockchain").Handler(s.blockchainEndpoints.GetRouter())
	}

	// Zero-trust network endpoints
	if s.zeroTrustEndpoints != nil {
		// Mount zero-trust endpoints router
		s.router.PathPrefix("/api/v2/zerotrust").Handler(s.zeroTrustEndpoints.GetRouter())
	}

	// Quantum cryptography endpoints
	if s.quantumEndpoints != nil {
		// Mount quantum endpoints router
		s.router.PathPrefix("/api/v2/quantum").Handler(s.quantumEndpoints.GetRouter())
	}

	// SIEM integration endpoints
	if s.siemEndpoints != nil {
		// Mount SIEM endpoints router
		s.router.PathPrefix("/api/v2/siem").Handler(s.siemEndpoints.GetRouter())
	}

	// Incident response endpoints
	if s.incidentEndpoints != nil {
		// Mount incident endpoints router
		s.router.PathPrefix("/api/v2/incident").Handler(s.incidentEndpoints.GetRouter())
	}

	// Threat modeling endpoints
	if s.threatEndpoints != nil {
		// Mount threat endpoints router
		s.router.PathPrefix("/api/v2/threat").Handler(s.threatEndpoints.GetRouter())
	}

	// Kubernetes endpoints
	if s.kubernetesEndpoints != nil {
		// Mount Kubernetes endpoints router
		s.router.PathPrefix("/api/v2/kubernetes").Handler(s.kubernetesEndpoints.GetRouter())
	}

	// API v3 routes (beta) - commented out for now
	// s.setupV3Routes()
}

// handleWebSocket handles WebSocket connections for real-time updates
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Create WebSocket manager
	wsManager := websocket.NewWebSocketManager(s.database)

	// Handle WebSocket connection
	wsManager.HandleWebSocket(w, r)
}

func (s *Server) setupV1Routes() {
	// Authentication routes
	s.router.HandleFunc("/api/v1/auth/login", s.handleLogin).Methods("POST")
	s.router.HandleFunc("/api/v1/auth/logout", s.handleLogout).Methods("POST")
	s.router.HandleFunc("/api/v1/auth/refresh", s.handleRefreshToken).Methods("POST")
	s.router.HandleFunc("/api/v1/auth/me", s.handleGetCurrentUser).Methods("GET")

	// Dashboard routes
	s.router.HandleFunc("/api/v1/dashboard/metrics", s.handleGetMetrics).Methods("GET")
	s.router.HandleFunc("/api/v1/dashboard/activity", s.handleGetActivity).Methods("GET")
	s.router.HandleFunc("/api/v1/dashboard/status", s.handleGetSystemStatus).Methods("GET")
	s.router.HandleFunc("/api/v1/dashboard/security", s.handleGetSecurityOverview).Methods("GET")

	// Threats routes
	s.router.HandleFunc("/api/v1/threats", s.handleGetThreats).Methods("GET")
	s.router.HandleFunc("/api/v1/threats/{id}", s.handleGetThreat).Methods("GET")
	s.router.HandleFunc("/api/v1/threats/{id}/status", s.handleUpdateThreatStatus).Methods("PATCH")
	s.router.HandleFunc("/api/v1/threats/stats", s.handleGetThreatStats).Methods("GET")
	s.router.HandleFunc("/api/v1/threats/feed", s.handleGetThreatFeed).Methods("GET")

	// Users routes
	s.router.HandleFunc("/api/v1/users", s.handleGetUsers).Methods("GET")
	s.router.HandleFunc("/api/v1/users/{id}", s.handleGetUser).Methods("GET")
	s.router.HandleFunc("/api/v1/users", s.handleCreateUser).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}", s.handleUpdateUser).Methods("PUT")
	s.router.HandleFunc("/api/v1/users/{id}", s.handleDeleteUser).Methods("DELETE")
	s.router.HandleFunc("/api/v1/users/stats", s.handleGetUserStats).Methods("GET")
	s.router.HandleFunc("/api/v1/users/roles", s.handleGetUserRoles).Methods("GET")

	// Security routes
	s.router.HandleFunc("/api/v1/security/policies", s.handleGetPolicies).Methods("GET")
	s.router.HandleFunc("/api/v1/security/policies/{id}", s.handleUpdatePolicy).Methods("PUT")
	s.router.HandleFunc("/api/v1/security/vulnerabilities", s.handleGetVulnerabilities).Methods("GET")
	s.router.HandleFunc("/api/v1/security/vulnerabilities/{id}", s.handleUpdateVulnerability).Methods("PATCH")
	s.router.HandleFunc("/api/v1/security/score", s.handleGetSecurityScore).Methods("GET")
	s.router.HandleFunc("/api/v1/security/scan", s.handleRunSecurityScan).Methods("POST")
	s.router.HandleFunc("/api/v1/security/audit-logs", s.handleGetAuditLogs).Methods("GET")

	// Health check
	s.router.HandleFunc("/api/v1/health", s.handleHealthCheck).Methods("GET")

	// Reports endpoints
	s.router.HandleFunc("/api/v1/reports", s.handleGetReports).Methods("GET")
	s.router.HandleFunc("/api/v1/reports/{filename}", s.handleGetReportContent).Methods("GET")
}

func (s *Server) setupV2Routes() {
	// Enhanced v2 endpoints
	v2 := s.router.PathPrefix("/api/v2")

	// Enhanced authentication
	v2.PathPrefix("/api/v2").Subrouter().HandleFunc("/auth/login", s.handleLogin).Methods("POST")
	v2.PathPrefix("/api/v2").Subrouter().HandleFunc("/auth/logout", s.handleLogout).Methods("POST")
	v2.PathPrefix("/api/v2").Subrouter().HandleFunc("/auth/refresh", s.handleRefreshToken).Methods("POST")
	v2.PathPrefix("/api/v2").Subrouter().HandleFunc("/auth/me", s.handleGetCurrentUser).Methods("GET")
	s.router.HandleFunc("/api/v2/auth/refresh", s.handleRefreshToken).Methods("POST")
	s.router.HandleFunc("/api/v2/auth/me", s.handleGetCurrentUser).Methods("GET")

	// Dashboard routes v2
	s.router.HandleFunc("/api/v2/dashboard/metrics", s.handleV2DashboardMetrics).Methods("GET")
	s.router.HandleFunc("/api/v2/dashboard/activity", s.handleGetActivity).Methods("GET")
	s.router.HandleFunc("/api/v2/dashboard/status", s.handleGetSystemStatus).Methods("GET")
	s.router.HandleFunc("/api/v2/dashboard/security", s.handleGetSecurityOverview).Methods("GET")

	// Threats routes v2
	s.router.HandleFunc("/api/v2/threats", s.handleV2Threats).Methods("GET")
	s.router.HandleFunc("/api/v2/threats/{id}", s.handleGetThreat).Methods("GET")
	s.router.HandleFunc("/api/v2/threats/{id}/status", s.handleUpdateThreatStatus).Methods("PATCH")
	s.router.HandleFunc("/api/v2/threats/stats", s.handleGetThreatStats).Methods("GET")
	s.router.HandleFunc("/api/v2/threats/feed", s.handleGetThreatFeed).Methods("GET")

	// Users routes v2
	s.router.HandleFunc("/api/v2/users", s.handleGetUsers).Methods("GET")
	s.router.HandleFunc("/api/v2/users/{id}", s.handleGetUser).Methods("GET")
	s.router.HandleFunc("/api/v2/users", s.handleCreateUser).Methods("POST")
	s.router.HandleFunc("/api/v2/users/{id}", s.handleUpdateUser).Methods("PUT")
	s.router.HandleFunc("/api/v2/users/{id}", s.handleDeleteUser).Methods("DELETE")
	s.router.HandleFunc("/api/v2/users/stats", s.handleGetUserStats).Methods("GET")
	s.router.HandleFunc("/api/v2/users/roles", s.handleGetUserRoles).Methods("GET")

	// Security routes v2
	s.router.HandleFunc("/api/v2/security/policies", s.handleGetPolicies).Methods("GET")
	s.router.HandleFunc("/api/v2/security/policies/{id}", s.handleUpdatePolicy).Methods("PUT")
	s.router.HandleFunc("/api/v2/security/vulnerabilities", s.handleGetVulnerabilities).Methods("GET")
	s.router.HandleFunc("/api/v2/security/vulnerabilities/{id}", s.handleUpdateVulnerability).Methods("PATCH")
	s.router.HandleFunc("/api/v2/security/score", s.handleGetSecurityScore).Methods("GET")
	s.router.HandleFunc("/api/v2/security/scan", s.handleRunSecurityScan).Methods("POST")
	s.router.HandleFunc("/api/v2/security/audit-logs", s.handleGetAuditLogs).Methods("GET")

	// New v2 endpoints
	s.router.HandleFunc("/api/v2/analytics", s.handleV2Analytics).Methods("GET")
	s.router.HandleFunc("/api/v2/analytics/summary", s.handleAnalyticsSummary).Methods("GET")
	s.router.HandleFunc("/api/v2/webhooks", s.handleV2Webhooks).Methods("GET", "POST")
	s.router.HandleFunc("/api/v2/health", s.handleHealthCheck).Methods("GET")

	// Threat detection endpoints
	s.router.HandleFunc("/api/v2/threat/alerts", s.handleThreatAlerts).Methods("GET", "POST")
	s.router.HandleFunc("/api/v2/threat/detect", s.handleThreatDetection).Methods("POST")
	s.router.HandleFunc("/api/v2/threat/intel", s.handleThreatIntelligence).Methods("GET")
	s.router.HandleFunc("/api/v2/threat/status", s.handleThreatStatus).Methods("GET")

	// Backup and system management endpoints
	s.router.HandleFunc("/api/v2/backup/status", s.handleBackupStatus).Methods("GET")
	s.router.HandleFunc("/api/v2/workers/status", s.handleWorkerStatus).Methods("GET")
	s.router.HandleFunc("/api/v2/websocket/status", s.handleWebSocketStatus).Methods("GET")
}

func (s *Server) Start() error {
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(s.router)

	log.Printf("Starting Hades API Server on port %d", s.port)
	log.Printf("API endpoints available at http://localhost:%d/api/v2/ (Preferred)", s.port)
	log.Printf("Version hierarchy: v1 (Legacy) | v2 (Preferred) | v3 (Beta)")

	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), handler)
}

// handleGetReports returns a list of available daily reports
func (s *Server) handleGetReports(w http.ResponseWriter, r *http.Request) {
	reports := []map[string]interface{}{}

	// Read reports directory
	entries, err := os.ReadDir("reports")
	if err != nil {
		// If directory doesn't exist, return empty list
		s.writeSuccess(w, map[string]interface{}{
			"reports": []interface{}{},
		})
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Extract date from filename (daily_report_YYYYMMDD.md)
		date := ""
		if strings.HasPrefix(entry.Name(), "daily_report_") && len(entry.Name()) >= 26 {
			dateStr := entry.Name()[13:21] // Extract YYYYMMDD
			if t, err := time.Parse("20060102", dateStr); err == nil {
				date = t.Format("2006-01-02")
			}
		}

		reportType := "daily"
		if strings.HasSuffix(entry.Name(), "_latest.md") {
			reportType = "latest"
			date = time.Now().Format("2006-01-02")
		}

		reports = append(reports, map[string]interface{}{
			"filename": entry.Name(),
			"date":     date,
			"type":     reportType,
			"size":     info.Size(),
			"modified": info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	s.writeSuccess(w, map[string]interface{}{
		"reports": reports,
		"count":   len(reports),
	})
}

// handleGetReportContent returns the content of a specific report
func (s *Server) handleGetReportContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Sanitize filename to prevent directory traversal
	filename = filepath.Clean(filename)
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		s.writeError(w, http.StatusBadRequest, "Invalid filename")
		return
	}

	// Ensure file is in reports directory
	path := filepath.Join("reports", filename)

	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.writeError(w, http.StatusNotFound, "Report not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, "Failed to read report")
		return
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to read report content")
		return
	}

	s.writeSuccess(w, map[string]interface{}{
		"filename": filename,
		"size":     info.Size(),
		"content":  string(content),
		"modified": info.ModTime().Format("2006-01-02 15:04:05"),
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleEventsWebSocket(w http.ResponseWriter, r *http.Request) {
	wsManager := websocket.NewWebSocketManager(s.database)
	wsManager.HandleWebSocket(w, r)

	log.Printf("Events WebSocket handler started")
}

func (s *Server) handleAgentStreamWebSocket(w http.ResponseWriter, r *http.Request) {
	wsManager := websocket.NewAgentStreamWebSocketManager(s.database)
	wsManager.HandleWebSocket(w, r)

	log.Printf("Agent Stream WebSocket handler started")
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, Response{
		Success: false,
		Error:   message,
	})
}

func (s *Server) writeSuccess(w http.ResponseWriter, data interface{}) {
	s.writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}
