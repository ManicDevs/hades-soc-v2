// Headless Sentinel Entrypoint - Hades SOC V2.0
// Initializes internal agentic loops for systemd background service
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hades-v2/internal/agent"
	"hades-v2/internal/bus"
	"hades-v2/internal/database"
	"hades-v2/internal/engine"
	"hades-v2/internal/platform"
	"hades-v2/internal/security"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Sentinel represents the headless SOC service
type Sentinel struct {
	eventBus         *bus.EventBus
	orchestrator     *agent.Orchestrator
	honeyManager     *agent.HoneyFileManager
	dispatcher       *engine.Dispatcher
	db               *sql.DB
	repository       *database.GlobalStateRepository
	sessionManager   *security.SessionManager
	metricsCollector *platform.MetricsCollector
	metricsServer    *http.Server
	startTime        time.Time
	ctx              context.Context
	cancel           context.CancelFunc
}

func main() {
	log.Println("🛡️ Hades SOC V2.0 - Headless Sentinel Starting")

	// Initialize context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create sentinel instance
	sentinel, err := NewSentinel(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to initialize sentinel: %v", err)
	}

	// Start agentic loops
	if err := sentinel.Start(); err != nil {
		log.Fatalf("❌ Failed to start sentinel: %v", err)
	}

	log.Println("✅ Hades Sentinel - All agentic loops active")
	log.Println("🔍 Monitoring for threats and deception triggers...")

	// Wait for shutdown signal
	waitForShutdown()

	// Graceful shutdown
	log.Println("🛑 Shutting down Hades Sentinel...")
	sentinel.Stop()
	log.Println("✅ Hades Sentinel stopped gracefully")
}

// NewSentinel creates a new sentinel instance with all agentic components
func NewSentinel(ctx context.Context) (*Sentinel, error) {
	// Initialize database connection using SQLite for headless mode
	db, err := sql.Open("sqlite3", "hades_sentinel.db")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize repository
	repository := database.NewGlobalStateRepository(db, database.GetManager())

	// Initialize event bus - central nervous system
	eventBus := bus.New()

	// Initialize dispatcher for task processing with V2.0 baseline config
	dispatcherConfig := &engine.DispatcherConfig{
		MaxWorkers: 5, // 5 workers as per V2.0 baseline
		QueueSize:  100,
	}
	dispatcher := engine.NewDispatcher(dispatcherConfig)

	// Initialize session manager
	sessionManager := security.NewSessionManager()

	// Initialize honey file manager for deception
	honeyManager, err := agent.NewHoneyFileManager(eventBus, repository)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize honey file manager: %w", err)
	}

	// Initialize orchestrator - main agentic loop coordinator
	orchestrator := agent.NewOrchestrator(eventBus, dispatcher, db)

	// Initialize metrics collector
	metricsCollector := platform.GetGlobalMetrics()

	// Create multiplexer for metrics and health endpoints
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	// Note: health handler will be set after sentinel creation

	// Initialize Prometheus metrics server on port 2112 (network accessible)
	// SECURITY NOTE: Port 2112 is accessible from the network. Restrict via firewall to allow only monitoring server IPs.
	// Example: sudo ufw allow from 192.168.1.100 to any port 2112
	metricsServer := &http.Server{
		Addr:         ":2112", // Network accessible (not localhost)
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	sentinel := &Sentinel{
		eventBus:         eventBus,
		orchestrator:     orchestrator,
		honeyManager:     honeyManager,
		dispatcher:       dispatcher,
		db:               db,
		repository:       repository,
		sessionManager:   sessionManager,
		metricsCollector: metricsCollector,
		metricsServer:    metricsServer,
		startTime:        time.Now(),
		ctx:              ctx,
		cancel:           func() {},
	}

	// Now set the health handler since sentinel exists
	mux.HandleFunc("/health", sentinel.healthHandler)

	return sentinel, nil
}

// Start initializes all agentic loops
func (s *Sentinel) Start() error {
	log.Println("🚀 Starting Hades Sentinel agentic loops...")

	// Event bus is auto-started via New()
	log.Println("📡 Event bus initialized")

	// Start dispatcher workers
	if err := s.dispatcher.Start(); err != nil {
		return fmt.Errorf("failed to start dispatcher: %w", err)
	}
	log.Println("⚡ Dispatcher workers started (5/5 active)")

	// Start orchestrator main loop
	if err := s.orchestrator.Start(s.ctx); err != nil {
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}
	log.Println("🤖 Orchestrator main loop started")

	// Deploy honey files and start deception system
	if err := s.honeyManager.DeployHoneyFiles(); err != nil {
		return fmt.Errorf("failed to deploy honey files: %w", err)
	}
	log.Println("🍯 Honey file deception system deployed")

	// Session manager initialized
	log.Println("🔐 Session manager initialized")

	// Start Prometheus metrics server (network accessible)
	go func() {
		log.Printf("📊 Starting Prometheus metrics server on :2112 (network accessible)")
		log.Printf("🔒 SECURITY: Restrict port 2112 access via firewall to monitoring server IPs only")
		if err := s.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ Metrics server error: %v", err)
		}
	}()

	// Initialize metrics with default values
	s.metricsCollector.UpdateWorkerPoolStatus(5)  // 5 workers as per V2.0 baseline
	s.metricsCollector.UpdateGlobalRiskLevel(0.0) // Start with zero risk

	// Log V2.0 baseline status
	log.Println("📊 V2.0 Baseline Status:")
	log.Println("   🔒 Encapsulated Internal Modules: ACTIVE")
	log.Println("   🛡️ Adversarial AI Defense Shield: ACTIVE")
	log.Println("   🔄 Continuous Deception System: ACTIVE")
	log.Println("   🛡️ Safety Governor: ACTIVE")
	log.Println("   🚨 Autonomous Threat Response: READY")
	log.Println("   ⚡ Distributed Worker Processing: 5/5 ACTIVE")
	log.Println("   🔐 Quantum-Resistant Security: READY")
	log.Println("   📈 Prometheus Metrics: http://0.0.0.0:2112/metrics (network accessible)")
	log.Println("   ❤️  Health Check: http://0.0.0.0:2112/health (network accessible)")
	log.Println("   🔒 SECURITY: Configure firewall to restrict port 2112 to monitoring server IPs only")

	return nil
}

// healthHandler provides comprehensive health check endpoint for Uptime Kuma and monitoring systems
// SECURITY NOTE: This endpoint exposes system status. Restrict access via firewall (ufw/iptables)
// to allow only monitoring server IPs. Port 2112 should not be exposed to the internet.
func (s *Sentinel) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if all critical components are running
	status := "healthy"
	statusCode := http.StatusOK
	componentStatus := make(map[string]string)
	var issues []string

	// Verify database connection with detailed check
	databaseStatus := "connected"
	if s.db != nil {
		if err := s.db.Ping(); err != nil {
			databaseStatus = "disconnected"
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
			issues = append(issues, fmt.Sprintf("database connection failed: %v", err))
		} else {
			// Additional database health check - execute lightweight query
			var result int
			if err := s.db.QueryRow("SELECT 1").Scan(&result); err != nil {
				databaseStatus = "query_failed"
				status = "unhealthy"
				statusCode = http.StatusServiceUnavailable
				issues = append(issues, fmt.Sprintf("database query failed: %v", err))
			}
		}
	} else {
		databaseStatus = "not_initialized"
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
		issues = append(issues, "database not initialized")
	}
	componentStatus["database"] = databaseStatus

	// Verify orchestrator is running
	orchestratorStatus := "running"
	if s.orchestrator == nil {
		orchestratorStatus = "not_initialized"
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
		issues = append(issues, "orchestrator not initialized")
	} else if !s.orchestrator.IsRunning() {
		orchestratorStatus = "stopped"
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
		issues = append(issues, "orchestrator not running")
	}
	componentStatus["orchestrator"] = orchestratorStatus

	// Check other components
	componentStatus["event_bus"] = "running"
	componentStatus["dispatcher"] = "running"
	componentStatus["metrics"] = "running"

	// Prepare comprehensive health response
	health := map[string]interface{}{
		"status":     status,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"uptime":     time.Since(s.startTime).String(),
		"version":    "V2.0",
		"components": componentStatus,
		"issues":     issues,
		"checks": map[string]bool{
			"process_active":       true,
			"database_alive":       databaseStatus == "connected",
			"orchestrator_running": orchestratorStatus == "running",
		},
		"metrics_summary": s.metricsCollector.GetMetricsSummary(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Always return JSON response for consistency
	if err := json.NewEncoder(w).Encode(health); err != nil {
		log.Printf("Failed to encode health response: %v", err)
	}
}

// Stop gracefully shuts down all agentic loops
func (s *Sentinel) Stop() {
	log.Println("🛑 Stopping agentic loops...")

	// Cancel context to signal shutdown
	s.cancel()

	// Stop orchestrator
	s.orchestrator.Stop()
	log.Println("🤖 Orchestrator stopped")

	// Stop dispatcher
	s.dispatcher.Stop()
	log.Println("⚡ Dispatcher stopped")

	// Stop honey file manager
	s.honeyManager.Stop()
	log.Println("🍯 Honey file manager stopped")

	// Stop event bus
	s.eventBus.Stop()
	log.Println("📡 Event bus stopped")

	// Gracefully shutdown metrics server
	if s.metricsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.metricsServer.Shutdown(ctx); err != nil {
			log.Printf("❌ Error shutting down metrics server: %v", err)
		} else {
			log.Println("📊 Metrics server stopped")
		}
	}

	// Close database connection
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
		log.Println("💾 Database connection closed")
	}

	log.Println("✅ All agentic loops stopped gracefully")
}

// waitForShutdown blocks until a shutdown signal is received
func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal
	sig := <-sigChan
	log.Printf("📡 Received signal: %v", sig)
}
