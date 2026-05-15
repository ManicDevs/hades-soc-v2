package auxiliary

import (
	"context"
	"fmt"
	"sync"
	"time"

	"hades-v2/pkg/sdk"
)

// Node represents a distributed scanning node
type Node struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Address  string                 `json:"address"`
	Status   string                 `json:"status"`
	Capacity int                    `json:"capacity"`
	Load     int                    `json:"load"`
	LastSeen time.Time              `json:"last_seen"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ScanTask represents a distributed scan task
type ScanTask struct {
	ID          string                 `json:"id"`
	ModuleName  string                 `json:"module_name"`
	Target      string                 `json:"target"`
	Config      map[string]interface{} `json:"config"`
	Priority    int                    `json:"priority"`
	Status      string                 `json:"status"`
	AssignedTo  string                 `json:"assigned_to"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	Result      map[string]interface{} `json:"result"`
	Error       string                 `json:"error"`
}

// DistributedScanner provides distributed scanning capabilities
type DistributedScanner struct {
	*sdk.BaseModule
	nodes      map[string]*Node
	tasks      map[string]*ScanTask
	taskQueue  chan *ScanTask
	mu         sync.RWMutex
	nodeStatus map[string]time.Time
}

// NewDistributedScanner creates a new distributed scanner instance
func NewDistributedScanner() *DistributedScanner {
	return &DistributedScanner{
		BaseModule: sdk.NewBaseModule(
			"distributed_scanner",
			"Perform distributed scanning across multiple nodes",
			sdk.CategoryReporting,
		),
		nodes:      make(map[string]*Node),
		tasks:      make(map[string]*ScanTask),
		taskQueue:  make(chan *ScanTask, 1000),
		nodeStatus: make(map[string]time.Time),
	}
}

// Execute starts the distributed scanner
func (ds *DistributedScanner) Execute(ctx context.Context) error {
	ds.SetStatus(sdk.StatusRunning)
	defer ds.SetStatus(sdk.StatusIdle)

	// Start task scheduler
	go ds.taskScheduler(ctx)

	// Start node health monitor
	go ds.nodeHealthMonitor(ctx)

	// Process tasks
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			// Continue processing
		}
	}
}

// RegisterNode adds a new scanning node
func (ds *DistributedScanner) RegisterNode(node *Node) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if node.ID == "" {
		return fmt.Errorf("hades.auxiliary.distributed_scanner: node ID cannot be empty")
	}

	if node.Address == "" {
		return fmt.Errorf("hades.auxiliary.distributed_scanner: node address cannot be empty")
	}

	node.Status = "active"
	node.LastSeen = time.Now()

	ds.nodes[node.ID] = node
	ds.nodeStatus[node.ID] = node.LastSeen

	return nil
}

// UnregisterNode removes a scanning node
func (ds *DistributedScanner) UnregisterNode(nodeID string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if _, exists := ds.nodes[nodeID]; !exists {
		return fmt.Errorf("hades.auxiliary.distributed_scanner: node not found: %s", nodeID)
	}

	delete(ds.nodes, nodeID)
	delete(ds.nodeStatus, nodeID)

	return nil
}

// SubmitTask submits a new scan task
func (ds *DistributedScanner) SubmitTask(ctx context.Context, moduleName, target string, config map[string]interface{}, priority int) (*ScanTask, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	task := &ScanTask{
		ID:         ds.generateTaskID(),
		ModuleName: moduleName,
		Target:     target,
		Config:     config,
		Priority:   priority,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}

	ds.tasks[task.ID] = task

	select {
	case ds.taskQueue <- task:
		return task, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, fmt.Errorf("hades.auxiliary.distributed_scanner: task queue is full")
	}
}

// GetTask retrieves a task by ID
func (ds *DistributedScanner) GetTask(taskID string) (*ScanTask, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	task, exists := ds.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("hades.auxiliary.distributed_scanner: task not found: %s", taskID)
	}

	return task, nil
}

// ListTasks returns all tasks with optional filtering
func (ds *DistributedScanner) ListTasks(status string, limit int) []*ScanTask {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var tasks []*ScanTask
	for _, task := range ds.tasks {
		if status == "" || task.Status == status {
			tasks = append(tasks, task)
		}
	}

	// Apply limit
	if limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}

	return tasks
}

// ListNodes returns all active nodes
func (ds *DistributedScanner) ListNodes() []*Node {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var nodes []*Node
	for _, node := range ds.nodes {
		if node.Status == "active" {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// GetNodeStatistics returns scanning statistics
func (ds *DistributedScanner) GetNodeStatistics() map[string]interface{} {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	stats := map[string]interface{}{
		"total_nodes":     len(ds.nodes),
		"active_nodes":    0,
		"total_tasks":     len(ds.tasks),
		"pending_tasks":   0,
		"running_tasks":   0,
		"completed_tasks": 0,
		"failed_tasks":    0,
		"average_load":    0,
	}

	totalLoad := 0
	for _, node := range ds.nodes {
		if node.Status == "active" {
			stats["active_nodes"] = stats["active_nodes"].(int) + 1
			totalLoad += node.Load
		}
	}

	if stats["active_nodes"].(int) > 0 {
		stats["average_load"] = totalLoad / stats["active_nodes"].(int)
	}

	for _, task := range ds.tasks {
		switch task.Status {
		case "pending":
			stats["pending_tasks"] = stats["pending_tasks"].(int) + 1
		case "running":
			stats["running_tasks"] = stats["running_tasks"].(int) + 1
		case "completed":
			stats["completed_tasks"] = stats["completed_tasks"].(int) + 1
		case "failed":
			stats["failed_tasks"] = stats["failed_tasks"].(int) + 1
		}
	}

	return stats
}

// taskScheduler distributes tasks to available nodes
func (ds *DistributedScanner) taskScheduler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-ds.taskQueue:
			ds.assignTaskToNode(ctx, task)
		}
	}
}

// assignTaskToNode assigns a task to the best available node
func (ds *DistributedScanner) assignTaskToNode(ctx context.Context, task *ScanTask) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Find best node (least load)
	var bestNode *Node
	minLoad := 999999

	for _, node := range ds.nodes {
		if node.Status == "active" && node.Load < node.Capacity && node.Load < minLoad {
			bestNode = node
			minLoad = node.Load
		}
	}

	if bestNode == nil {
		// No available nodes, requeue task
		go func() {
			select {
			case ds.taskQueue <- task:
			case <-ctx.Done():
			case <-time.After(5 * time.Second):
				// Give up after timeout
			}
		}()
		return
	}

	// Assign task to node
	task.Status = "assigned"
	task.AssignedTo = bestNode.ID
	task.StartedAt = time.Now()

	bestNode.Load++
	ds.tasks[task.ID] = task

	// Simulate task execution
	go ds.executeTaskOnNode(ctx, task, bestNode)
}

// executeTaskOnNode simulates task execution on a node
func (ds *DistributedScanner) executeTaskOnNode(ctx context.Context, task *ScanTask, node *Node) {
	// Simulate execution time
	executionTime := time.Duration(5+task.Priority) * time.Second

	select {
	case <-time.After(executionTime):
		ds.mu.Lock()
		defer ds.mu.Unlock()

		task.Status = "completed"
		task.CompletedAt = time.Now()
		task.Result = map[string]interface{}{
			"node_id":        node.ID,
			"target":         task.Target,
			"module_name":    task.ModuleName,
			"execution_time": executionTime.Seconds(),
		}

		// Update node load
		if node, exists := ds.nodes[node.ID]; exists {
			node.Load--
			node.LastSeen = time.Now()
		}

	case <-ctx.Done():
		ds.mu.Lock()
		defer ds.mu.Unlock()

		task.Status = "cancelled"
		task.Error = "Context cancelled"

		if node, exists := ds.nodes[node.ID]; exists {
			node.Load--
		}
	}
}

// nodeHealthMonitor checks node health and removes inactive nodes
func (ds *DistributedScanner) nodeHealthMonitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ds.checkNodeHealth()
		}
	}
}

// checkNodeHealth checks if nodes are still responsive
func (ds *DistributedScanner) checkNodeHealth() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	now := time.Now()
	threshold := 2 * time.Minute

	for nodeID, lastSeen := range ds.nodeStatus {
		if now.Sub(lastSeen) > threshold {
			// Mark node as inactive
			if node, exists := ds.nodes[nodeID]; exists {
				node.Status = "inactive"
			}
		}
	}
}

// generateTaskID creates a unique task ID
func (ds *DistributedScanner) generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// GetResult returns scanner status and statistics
func (ds *DistributedScanner) GetResult() string {
	stats := ds.GetNodeStatistics()

	return fmt.Sprintf("Distributed Scanner Status:\n"+
		"Nodes: %d active / %d total\n"+
		"Tasks: %d pending, %d running, %d completed, %d failed\n"+
		"Average Load: %.1f%%",
		stats["active_nodes"], stats["total_nodes"],
		stats["pending_tasks"], stats["running_tasks"],
		stats["completed_tasks"], stats["failed_tasks"],
		stats["average_load"])
}
