package cluster

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"hades-v2/internal/database"
)

// ClusterManager handles database clustering and replication
type ClusterManager struct {
	db            database.Database
	config        ClusterConfig
	nodes         map[string]*ClusterNode
	mu            sync.RWMutex
	healthChecker *HealthChecker
}

// ClusterConfig represents cluster configuration
type ClusterConfig struct {
	ClusterID       string        `json:"cluster_id"`
	NodeID          string        `json:"node_id"`
	IsPrimary       bool          `json:"is_primary"`
	ReplicationMode string        `json:"replication_mode"` // master-slave, multi-master
	SyncInterval    time.Duration `json:"sync_interval"`
	HealthInterval  time.Duration `json:"health_interval"`
	MaxRetries      int           `json:"max_retries"`
	BackupInterval  time.Duration `json:"backup_interval"`
}

// ClusterNode represents a database cluster node
type ClusterNode struct {
	ID           string        `json:"id"`
	Address      string        `json:"address"`
	Port         int           `json:"port"`
	Status       string        `json:"status"` // online, offline, syncing, recovering
	Role         string        `json:"role"`   // primary, secondary, arbiter
	LastSync     time.Time     `json:"last_sync"`
	DatabaseSize int64         `json:"database_size"`
	Connections  int           `json:"connections"`
	Lag          time.Duration `json:"replication_lag"`
	LoadAverage  float64       `json:"load_average"`
	HealthScore  float64       `json:"health_score"`
	Timestamp    time.Time     `json:"timestamp"`
}

// ReplicationStatus represents replication status
type ReplicationStatus struct {
	PrimaryNode    string        `json:"primary_node"`
	SecondaryNodes []string      `json:"secondary_nodes"`
	SyncMode       string        `json:"sync_mode"` // synchronous, asynchronous
	LagTime        time.Duration `json:"lag_time"`
	Throughput     float64       `json:"throughput"`
	ErrorRate      float64       `json:"error_rate"`
	Status         string        `json:"status"` // healthy, degraded, failed
	Timestamp      time.Time     `json:"timestamp"`
}

// HealthChecker monitors cluster health
type HealthChecker struct {
	cluster  *ClusterManager
	stopChan chan bool
	ticker   *time.Ticker
}

// NewClusterManager creates a new cluster manager
func NewClusterManager(db database.Database, config ClusterConfig) *ClusterManager {
	cm := &ClusterManager{
		db:     db,
		config: config,
		nodes:  make(map[string]*ClusterNode),
	}

	cm.healthChecker = NewHealthChecker(cm)
	return cm
}

// InitializeCluster initializes the database cluster
func (cm *ClusterManager) InitializeCluster() error {
	log.Printf("Initializing database cluster %s", cm.config.ClusterID)

	// Create cluster configuration table
	err := cm.createClusterConfigTable()
	if err != nil {
		return fmt.Errorf("failed to create cluster config table: %w", err)
	}

	// Register this node
	err = cm.registerNode()
	if err != nil {
		return fmt.Errorf("failed to register node: %w", err)
	}

	// Start health checking
	go cm.healthChecker.Start()

	log.Printf("Cluster %s initialized successfully", cm.config.ClusterID)
	return nil
}

// createClusterConfigTable creates the cluster configuration table
func (cm *ClusterManager) createClusterConfigTable() error {
	sqlDB, ok := cm.db.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	query := `
		CREATE TABLE IF NOT EXISTS cluster_config (
			cluster_id VARCHAR(255) PRIMARY KEY,
			node_id VARCHAR(255) PRIMARY KEY,
			cluster_role VARCHAR(50) NOT NULL,
			is_primary BOOLEAN DEFAULT FALSE,
			replication_mode VARCHAR(50) DEFAULT 'master-slave',
			sync_interval INTEGER DEFAULT 30,
			health_interval INTEGER DEFAULT 10,
			max_retries INTEGER DEFAULT 3,
			backup_interval INTEGER DEFAULT 3600,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := sqlDB.Exec(query)
	return err
}

// registerNode registers this node in the cluster
func (cm *ClusterManager) registerNode() error {
	sqlDB, ok := cm.db.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	query := `
		INSERT INTO cluster_config (cluster_id, node_id, cluster_role, is_primary, replication_mode, sync_interval, health_interval, max_retries, backup_interval)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (cluster_id, node_id) 
		DO UPDATE SET 
			updated_at = CURRENT_TIMESTAMP
		WHERE cluster_id = $1 AND node_id = $2
	`

	_, err := sqlDB.Exec(query, cm.config.ClusterID, cm.config.NodeID,
		cm.getRole(), cm.config.IsPrimary, cm.config.ReplicationMode,
		cm.config.SyncInterval.Seconds(), cm.config.HealthInterval.Seconds(),
		cm.config.MaxRetries, cm.config.BackupInterval.Seconds())

	return err
}

// getRole determines the role of this node
func (cm *ClusterManager) getRole() string {
	if cm.config.IsPrimary {
		return "primary"
	}
	return "secondary"
}

// AddNode adds a new node to the cluster
func (cm *ClusterManager) AddNode(node ClusterNode) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.nodes[node.ID] = &node

	log.Printf("Added node %s to cluster %s", node.ID, cm.config.ClusterID)
	return nil
}

// RemoveNode removes a node from the cluster
func (cm *ClusterManager) RemoveNode(nodeID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.nodes, nodeID)

	// Remove from database
	sqlDB, ok := cm.db.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	query := "DELETE FROM cluster_config WHERE cluster_id = $1 AND node_id = $2"
	_, err := sqlDB.Exec(query, cm.config.ClusterID, nodeID)
	if err != nil {
		return fmt.Errorf("failed to remove node: %w", err)
	}

	log.Printf("Removed node %s from cluster %s", nodeID, cm.config.ClusterID)
	return nil
}

// GetNodes retrieves all nodes in the cluster
func (cm *ClusterManager) GetNodes() []ClusterNode {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	nodes := make([]ClusterNode, 0, len(cm.nodes))
	for _, node := range cm.nodes {
		nodes = append(nodes, *node)
	}

	return nodes
}

// GetReplicationStatus gets the current replication status
func (cm *ClusterManager) GetReplicationStatus() (*ReplicationStatus, error) {
	sqlDB, ok := cm.db.GetConnection().(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT 
			cluster_role,
			is_primary,
			replication_mode,
			updated_at
		FROM cluster_config 
		WHERE cluster_id = $1
		ORDER BY is_primary DESC, updated_at DESC
	`

	rows, err := sqlDB.Query(query, cm.config.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get replication status: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var primaryNode string
	var secondaryNodes []string
	var replicationMode string
	var lastUpdate time.Time

	for rows.Next() {
		var isPrimary bool
		var nodeID string
		err := rows.Scan(&primaryNode, &secondaryNodes, &replicationMode, &lastUpdate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replication row: %w", err)
		}

		if !isPrimary {
			secondaryNodes = append(secondaryNodes, nodeID)
		} else {
			primaryNode = nodeID
		}
	}

	status := &ReplicationStatus{
		PrimaryNode:    primaryNode,
		SecondaryNodes: secondaryNodes,
		SyncMode:       replicationMode,
		Status:         "healthy",
		Timestamp:      time.Now(),
	}

	return status, nil
}

// PromoteToPrimary promotes a secondary node to primary
func (cm *ClusterManager) PromoteToPrimary(nodeID string) error {
	log.Printf("Promoting node %s to primary in cluster %s", nodeID, cm.config.ClusterID)

	sqlDB, ok := cm.db.GetConnection().(*sql.DB)
	if !ok {
		return fmt.Errorf("database is not an SQL database")
	}

	// Update node role
	query := `
		UPDATE cluster_config 
		SET 
			is_primary = TRUE,
			cluster_role = 'primary',
			updated_at = CURRENT_TIMESTAMP
		WHERE cluster_id = $1 AND node_id = $2
	`

	_, err := sqlDB.Exec(query, cm.config.ClusterID, nodeID)
	if err != nil {
		return fmt.Errorf("failed to promote node: %w", err)
	}

	// Update other nodes to secondary
	updateQuery := `
		UPDATE cluster_config 
		SET 
			is_primary = FALSE,
			cluster_role = 'secondary',
			updated_at = CURRENT_TIMESTAMP
		WHERE cluster_id = $1 AND node_id != $2 AND is_primary = TRUE
	`

	_, err = sqlDB.Exec(updateQuery, cm.config.ClusterID, nodeID)
	if err != nil {
		return fmt.Errorf("failed to update other nodes: %w", err)
	}

	log.Printf("Node %s promoted to primary successfully", nodeID)
	return nil
}

// Failover performs automatic failover to a secondary node
func (cm *ClusterManager) Failover() error {
	log.Printf("Initiating failover for cluster %s", cm.config.ClusterID)

	nodes := cm.GetNodes()
	var primaryNode *ClusterNode
	var secondaryNodes []*ClusterNode

	for _, node := range nodes {
		if node.Status == "online" {
			if node.Role == "primary" {
				primaryNode = &node
			} else {
				secondaryNodes = append(secondaryNodes, &node)
			}
		}
	}

	if primaryNode == nil {
		return fmt.Errorf("no primary node found for failover")
	}

	if len(secondaryNodes) == 0 {
		return fmt.Errorf("no secondary nodes available for failover")
	}

	// Select the best secondary node (highest health score)
	var bestNode *ClusterNode
	for _, node := range secondaryNodes {
		if bestNode == nil || node.HealthScore > bestNode.HealthScore {
			bestNode = node
		}
	}

	if bestNode == nil {
		return fmt.Errorf("no suitable secondary node found for failover")
	}

	// Promote the best secondary node to primary
	err := cm.PromoteToPrimary(bestNode.ID)
	if err != nil {
		return fmt.Errorf("failed to promote node during failover: %w", err)
	}

	log.Printf("Failover completed: %s promoted to primary", bestNode.ID)
	return nil
}

// SyncData synchronizes data across cluster nodes
func (cm *ClusterManager) SyncData() error {
	log.Printf("Synchronizing data across cluster %s", cm.config.ClusterID)

	nodes := cm.GetNodes()

	for _, node := range nodes {
		if node.Status == "online" && node.ID != cm.config.NodeID {
			err := cm.syncWithNode(node)
			if err != nil {
				log.Printf("Failed to sync with node %s: %v", node.ID, err)
			} else {
				log.Printf("Successfully synced with node %s", node.ID)
			}
		}
	}

	return nil
}

// syncWithNode synchronizes data with a specific node
func (cm *ClusterManager) syncWithNode(node ClusterNode) error {
	// This would implement actual data synchronization logic
	// For demonstration, just log the sync operation

	log.Printf("Syncing data with node %s at %s", node.ID, node.Address)

	// Simulate sync operation
	time.Sleep(100 * time.Millisecond)

	return nil
}

// GetClusterMetrics retrieves comprehensive cluster metrics
func (cm *ClusterManager) GetClusterMetrics() (map[string]interface{}, error) {
	nodes := cm.GetNodes()

	metrics := map[string]interface{}{
		"total_nodes":      len(nodes),
		"online_nodes":     cm.countOnlineNodes(nodes),
		"primary_nodes":    cm.countPrimaryNodes(nodes),
		"secondary_nodes":  cm.countSecondaryNodes(nodes),
		"average_health":   cm.calculateAverageHealth(nodes),
		"cluster_id":       cm.config.ClusterID,
		"replication_mode": cm.config.ReplicationMode,
		"timestamp":        time.Now(),
	}

	return metrics, nil
}

// Helper methods
func (cm *ClusterManager) countOnlineNodes(nodes []ClusterNode) int {
	count := 0
	for _, node := range nodes {
		if node.Status == "online" {
			count++
		}
	}
	return count
}

func (cm *ClusterManager) countPrimaryNodes(nodes []ClusterNode) int {
	count := 0
	for _, node := range nodes {
		if node.Role == "primary" {
			count++
		}
	}
	return count
}

func (cm *ClusterManager) countSecondaryNodes(nodes []ClusterNode) int {
	count := 0
	for _, node := range nodes {
		if node.Role == "secondary" {
			count++
		}
	}
	return count
}

func (cm *ClusterManager) calculateAverageHealth(nodes []ClusterNode) float64 {
	if len(nodes) == 0 {
		return 0
	}

	totalHealth := 0.0
	for _, node := range nodes {
		totalHealth += node.HealthScore
	}

	return totalHealth / float64(len(nodes))
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(cluster *ClusterManager) *HealthChecker {
	return &HealthChecker{
		cluster:  cluster,
		stopChan: make(chan bool),
	}
}

// Start begins the health checking process
func (hc *HealthChecker) Start() {
	hc.ticker = time.NewTicker(hc.cluster.config.HealthInterval)

	go func() {
		for {
			select {
			case <-hc.stopChan:
				hc.ticker.Stop()
				return
			case <-hc.ticker.C:
				hc.checkHealth()
			}
		}
	}()
}

// Stop stops the health checking process
func (hc *HealthChecker) Stop() {
	close(hc.stopChan)
}

// checkHealth performs health checks on all cluster nodes
func (hc *HealthChecker) checkHealth() {
	nodes := hc.cluster.GetNodes()

	for _, node := range nodes {
		go hc.checkNodeHealth(node)
	}
}

// checkNodeHealth performs health check on a specific node
func (hc *HealthChecker) checkNodeHealth(node ClusterNode) {
	// Simulate health check
	// In a real implementation, this would:
	// - Check database connectivity
	// - Check replication lag
	// - Check system resources
	// - Check network connectivity

	healthScore := hc.calculateNodeHealth(node)
	node.HealthScore = healthScore
	node.Timestamp = time.Now()

	// Update node status
	if healthScore > 80 {
		node.Status = "online"
	} else if healthScore > 50 {
		node.Status = "degraded"
	} else {
		node.Status = "offline"
	}

	log.Printf("Node %s health status: %s (score: %.1f)", node.ID, node.Status, healthScore)
}

// calculateNodeHealth calculates the health score for a node
func (hc *HealthChecker) calculateNodeHealth(node ClusterNode) float64 {
	// Simulate health calculation
	// In a real implementation, this would consider:
	// - Database response time
	// - Replication lag
	// - System load
	// - Network latency
	// - Disk usage
	// - Memory usage

	baseScore := 100.0

	// Deduct for offline status
	if node.Status == "offline" {
		baseScore -= 50
	}

	// Deduct for high load
	if node.LoadAverage > 80 {
		baseScore -= 20
	}

	// Deduct for high replication lag
	if node.Lag > time.Second*5 {
		baseScore -= 15
	}

	// Deduct for degraded status
	if node.Status == "degraded" {
		baseScore -= 10
	}

	return baseScore
}
