package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"hades-v2/internal/api/versioning"
	"hades-v2/internal/database"
	"hades-v2/internal/websocket"
	"hades-v2/internal/workers"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type Server struct {
	router      *mux.Router
	port        int
	versionMgr  *versioning.VersionManager
	versionInt  *versioning.ServerIntegration
	database    database.Database
	workerPool  *workers.WorkerPool
	wsManager   *websocket.WebSocketManager
	aiEndpoints *AIEndpoints
	healer      *AutoHealer
}

func NewServer(port int) *Server {
	godotenv.Load()

	srv := &Server{
		router:      mux.NewRouter(),
		port:        port,
		versionMgr:  versioning.NewVersionManager(versioning.ManagerConfig{}),
		versionInt:  versioning.NewServerIntegration(versioning.NewVersionManager(versioning.ManagerConfig{})),
		database:    database.NewDatabase(database.SQLite),
		workerPool:  workers.NewWorkerPool(10, database.NewDatabase(database.SQLite)),
		wsManager:   websocket.NewWebSocketManager(database.NewDatabase(database.SQLite)),
		aiEndpoints: &AIEndpoints{},
		healer:      initAutoHealer(30),
	}
	return srv
}

func (s *Server) registerRoutes() {
	// Root endpoint - API info
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"Hades Toolkit","version":"2.0","status":"running","modules":["tor","hotplug","peer-network","anti-analysis","siem","governor","incident","quantum","zerotrust","blockchain","threat-hunting","analytics","ai","threat","kubernetes","llm","auto-heal"]}`))
	})

	// Register all API routes
	RegisterAllRoutes(s.router)
}

func (s *Server) Start() error {
	// Register routes
	s.registerRoutes()

	// Start worker pool
	s.workerPool.Start()

	// Start auto-healer
	s.healer.Start()

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Printf("\nShutting down API server...\n")
		s.healer.Stop()
		os.Exit(0)
	}()

	// Start HTTP server with CORS
	handler := cors.Default().Handler(s.router)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: handler,
	}

	log.Printf("Starting Hades Toolkit API Server\n")
	fmt.Printf("=================================\n")
	fmt.Printf("Port: %d\n", s.port)
	fmt.Printf("API: http://localhost:%d/api/v2/ (Preferred)\n", s.port)
	fmt.Printf("LLM: http://localhost:%d/api/v2/llm/query\n", s.port)
	fmt.Printf("Self-Heal: http://localhost:%d/api/v2/self-heal/health\n", s.port)
	fmt.Printf("Versions: v1 (Legacy) | v2 (Preferred) | v3 (Beta)\n")
	fmt.Printf("Discovery: http://localhost:%d/api/versions\n", s.port)
	log.Printf("Starting server on port %d\n", s.port)
	return server.ListenAndServe()
}
