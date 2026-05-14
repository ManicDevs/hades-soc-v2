package platform

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// DistributedDatabaseConfig holds distributed database configuration
type DistributedDatabaseConfig struct {
	Type                string         `json:"type"`             // "postgresql", "mysql", "sqlite"
	Primary             DatabaseNode   `json:"primary"`          // Primary database node
	Replicas            []DatabaseNode `json:"replicas"`         // Replica database nodes
	FailoverEnabled     bool           `json:"failover_enabled"` // Enable automatic failover
	HealthCheckInterval time.Duration  `json:"health_check_interval"`
	MaxRetries          int            `json:"max_retries"` // Maximum connection retries
	ConnectionTimeout   time.Duration  `json:"connection_timeout"`
	LoadBalancing       string         `json:"load_balancing"` // "round_robin", "least_connections", "random"
}

// DatabaseNode represents a database node in the distributed system
type DatabaseNode struct {
	ID       string `json:"id"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
	Weight   int    `json:"weight"`  // For load balancing
	Healthy  bool   `json:"healthy"` // Health status
}

// DistributedDatabase provides distributed database functionality
type DistributedDatabase struct {
	config     *DistributedDatabaseConfig
	primary    *sql.DB
	replicas   []*sql.DB
	currentIdx int
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	healthChan chan DatabaseNode
}

// DefaultDistributedDatabaseConfig returns sensible distributed database defaults
func DefaultDistributedDatabaseConfig() *DistributedDatabaseConfig {
	return &DistributedDatabaseConfig{
		Type: "postgresql",
		Primary: DatabaseNode{
			ID:       "primary-1",
			Host:     "localhost",
			Port:     5432,
			Database: "hades_toolkit",
			Username: "hades",
			Password: "hades_password",
			SSLMode:  "disable",
			Weight:   100,
			Healthy:  true,
		},
		Replicas: []DatabaseNode{
			{
				ID:       "replica-1",
				Host:     "localhost",
				Port:     5433,
				Database: "hades_toolkit_replica",
				Username: "hades",
				Password: "hades_password",
				SSLMode:  "disable",
				Weight:   80,
				Healthy:  true,
			},
			{
				ID:       "replica-2",
				Host:     "localhost",
				Port:     5434,
				Database: "hades_toolkit_replica2",
				Username: "hades",
				Password: "hades_password",
				SSLMode:  "disable",
				Weight:   60,
				Healthy:  true,
			},
		},
		FailoverEnabled:     true,
		HealthCheckInterval: 30 * time.Second,
		MaxRetries:          3,
		ConnectionTimeout:   10 * time.Second,
		LoadBalancing:       "least_connections",
	}
}

// NewDistributedDatabase creates a new distributed database instance
func NewDistributedDatabase(config *DistributedDatabaseConfig) (*DistributedDatabase, error) {
	if config == nil {
		config = DefaultDistributedDatabaseConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	db := &DistributedDatabase{
		config:     config,
		currentIdx: 0,
		ctx:        ctx,
		cancel:     cancel,
		healthChan: make(chan DatabaseNode, 100),
	}

	// Initialize primary database
	if err := db.initPrimary(); err != nil {
		return nil, fmt.Errorf("failed to initialize primary database: %w", err)
	}

	// Initialize replica databases
	if err := db.initReplicas(); err != nil {
		log.Printf("Warning: Failed to initialize some replicas: %v", err)
	}

	// Start health monitoring
	go db.healthMonitor()

	return db, nil
}

// initPrimary initializes the primary database connection
func (dd *DistributedDatabase) initPrimary() error {
	connStr := dd.buildConnectionString(dd.config.Primary)

	db, err := sql.Open(dd.config.Type, connStr)
	if err != nil {
		return fmt.Errorf("failed to open primary database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(dd.ctx, dd.config.ConnectionTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping primary database: %w", err)
	}

	dd.primary = db
	return nil
}

// initReplicas initializes replica database connections
func (dd *DistributedDatabase) initReplicas() error {
	for i, replica := range dd.config.Replicas {
		connStr := dd.buildConnectionString(replica)

		db, err := sql.Open(dd.config.Type, connStr)
		if err != nil {
			log.Printf("Failed to open replica %s: %v", replica.ID, err)
			dd.config.Replicas[i].Healthy = false
			continue
		}

		// Test connection
		ctx, cancel := context.WithTimeout(dd.ctx, dd.config.ConnectionTimeout)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Printf("Failed to ping replica %s: %v", replica.ID, err)
			dd.config.Replicas[i].Healthy = false
			if err := db.Close(); err != nil {
				log.Printf("Warning: failed to close database connection: %v", err)
			}
			continue
		}

		dd.replicas = append(dd.replicas, db)
	}

	return nil
}

// buildConnectionString builds a database connection string
func (dd *DistributedDatabase) buildConnectionString(node DatabaseNode) string {
	switch dd.config.Type {
	case "postgresql":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			node.Host, node.Port, node.Username, node.Password, node.Database, node.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			node.Username, node.Password, node.Host, node.Port, node.Database)
	case "sqlite":
		return node.Database
	default:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			node.Host, node.Port, node.Username, node.Password, node.Database, node.SSLMode)
	}
}

// GetConnection returns a database connection based on load balancing strategy
func (dd *DistributedDatabase) GetConnection() *sql.DB {
	dd.mu.RLock()
	defer dd.mu.RUnlock()

	// If primary is healthy, use it for writes
	if dd.config.Primary.Healthy && dd.primary != nil {
		return dd.primary
	}

	// Use replicas for reads or if primary is down
	return dd.getHealthyReplica()
}

// getHealthyReplica returns a healthy replica based on load balancing
func (dd *DistributedDatabase) getHealthyReplica() *sql.DB {
	healthyReplicas := dd.getHealthyReplicas()
	if len(healthyReplicas) == 0 {
		return dd.primary // Fallback to primary even if unhealthy
	}

	switch dd.config.LoadBalancing {
	case "round_robin":
		return dd.roundRobinReplica(healthyReplicas)
	case "least_connections":
		return dd.leastConnectionsReplica(healthyReplicas)
	case "random":
		return dd.randomReplica(healthyReplicas)
	default:
		return healthyReplicas[0]
	}
}

// getHealthyReplicas returns a list of healthy replicas
func (dd *DistributedDatabase) getHealthyReplicas() []*sql.DB {
	var healthy []*sql.DB
	for i, replica := range dd.replicas {
		if dd.config.Replicas[i].Healthy {
			healthy = append(healthy, replica)
		}
	}
	return healthy
}

// roundRobinReplica implements round-robin load balancing
func (dd *DistributedDatabase) roundRobinReplica(replicas []*sql.DB) *sql.DB {
	if len(replicas) == 0 {
		return nil
	}

	dd.currentIdx = (dd.currentIdx + 1) % len(replicas)
	return replicas[dd.currentIdx]
}

// leastConnectionsReplica implements least connections load balancing
func (dd *DistributedDatabase) leastConnectionsReplica(replicas []*sql.DB) *sql.DB {
	if len(replicas) == 0 {
		return nil
	}

	var bestReplica *sql.DB
	minConns := int(^uint(0) >> 1) // Max int

	for _, replica := range replicas {
		stats := replica.Stats()
		if stats.OpenConnections < minConns {
			minConns = stats.OpenConnections
			bestReplica = replica
		}
	}

	return bestReplica
}

// randomReplica implements random load balancing
func (dd *DistributedDatabase) randomReplica(replicas []*sql.DB) *sql.DB {
	if len(replicas) == 0 {
		return nil
	}

	// Simple random selection based on time
	idx := time.Now().UnixNano() % int64(len(replicas))
	return replicas[idx]
}

// healthMonitor monitors the health of database nodes
func (dd *DistributedDatabase) healthMonitor() {
	ticker := time.NewTicker(dd.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dd.ctx.Done():
			return
		case <-ticker.C:
			dd.checkHealth()
		}
	}
}

// checkHealth checks the health of all database nodes
func (dd *DistributedDatabase) checkHealth() {
	// Check primary
	if dd.primary != nil {
		ctx, cancel := context.WithTimeout(dd.ctx, 5*time.Second)
		if err := dd.primary.PingContext(ctx); err != nil {
			dd.config.Primary.Healthy = false
			log.Printf("Primary database unhealthy: %v", err)
		} else {
			dd.config.Primary.Healthy = true
		}
		cancel()
	}

	// Check replicas
	for i, rep := range dd.replicas {
		ctx, cancel := context.WithTimeout(dd.ctx, 5*time.Second)
		if err := rep.PingContext(ctx); err != nil {
			dd.config.Replicas[i].Healthy = false
			log.Printf("Replica %s unhealthy: %v", dd.config.Replicas[i].ID, err)
		} else {
			dd.config.Replicas[i].Healthy = true
		}
		cancel()
	}
}

// HotSwapDatabase allows hot-swapping database configurations
func (dd *DistributedDatabase) HotSwapDatabase(newConfig *DistributedDatabaseConfig) error {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	// Create new distributed database instance
	newDB, err := NewDistributedDatabase(newConfig)
	if err != nil {
		return fmt.Errorf("failed to create new database configuration: %w", err)
	}

	// Close old connections
	if dd.primary != nil {
		if err := dd.primary.Close(); err != nil {
			log.Printf("Error closing primary database connection: %v", err)
		}
	}
	for _, replica := range dd.replicas {
		if err := replica.Close(); err != nil {
			log.Printf("Error closing replica connection: %v", err)
		}
	}

	// Swap configurations
	dd.config = newConfig
	dd.primary = newDB.primary
	dd.replicas = newDB.replicas

	log.Println("Database configuration hot-swapped successfully")
	return nil
}

// GetStatus returns the current status of the distributed database
func (dd *DistributedDatabase) GetStatus() map[string]interface{} {
	dd.mu.RLock()
	defer dd.mu.RUnlock()

	status := map[string]interface{}{
		"type": dd.config.Type,
		"primary": map[string]interface{}{
			"id":      dd.config.Primary.ID,
			"healthy": dd.config.Primary.Healthy,
		},
		"replicas":       []map[string]interface{}{},
		"load_balancing": dd.config.LoadBalancing,
	}

	for i := range dd.replicas {
		if i < len(dd.config.Replicas) {
			status["replicas"] = append(status["replicas"].([]map[string]interface{}), map[string]interface{}{
				"id":      dd.config.Replicas[i].ID,
				"healthy": dd.config.Replicas[i].Healthy,
			})
		}
	}

	return status
}

// Close closes all database connections
func (dd *DistributedDatabase) Close() error {
	dd.cancel()

	var errs []error

	if dd.primary != nil {
		if err := dd.primary.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	for _, replica := range dd.replicas {
		if err := replica.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errs)
	}

	return nil
}
