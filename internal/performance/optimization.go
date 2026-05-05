package performance

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ConnectionPool manages a pool of connections
type ConnectionPool struct {
	mu          sync.RWMutex
	connections []interface{}
	active      map[interface{}]bool
	maxSize     int
	currentSize int
	factory     func() (interface{}, error)
	closer      func(interface{}) error
	validator   func(interface{}) bool
}

// PoolConfig holds connection pool configuration
type PoolConfig struct {
	MaxSize     int                         `json:"max_size"`
	InitSize    int                         `json:"init_size"`
	Factory     func() (interface{}, error) `json:"-"`
	Closer      func(interface{}) error     `json:"-"`
	Validator   func(interface{}) bool      `json:"-"`
	IdleTimeout time.Duration               `json:"idle_timeout"`
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config *PoolConfig) *ConnectionPool {
	pool := &ConnectionPool{
		connections: make([]interface{}, 0, config.MaxSize),
		active:      make(map[interface{}]bool),
		maxSize:     config.MaxSize,
		factory:     config.Factory,
		closer:      config.Closer,
		validator:   config.Validator,
	}

	// Initialize pool with initial connections
	for i := 0; i < config.InitSize; i++ {
		if conn, err := pool.factory(); err == nil {
			pool.connections = append(pool.connections, conn)
			pool.currentSize++
		}
	}

	return pool
}

// Get gets a connection from the pool
func (cp *ConnectionPool) Get() (interface{}, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Find an available connection
	for i, conn := range cp.connections {
		if !cp.active[conn] {
			if cp.validator == nil || cp.validator(conn) {
				cp.active[conn] = true
				return conn, nil
			} else {
				// Remove invalid connection
				cp.removeConnection(i)
			}
		}
	}

	// Try to create a new connection if under max size
	if cp.currentSize < cp.maxSize {
		conn, err := cp.factory()
		if err != nil {
			return nil, err
		}
		cp.connections = append(cp.connections, conn)
		cp.active[conn] = true
		cp.currentSize++
		return conn, nil
	}

	return nil, fmt.Errorf("connection pool exhausted")
}

// Put returns a connection to the pool
func (cp *ConnectionPool) Put(conn interface{}) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.active[conn] {
		cp.active[conn] = false
		return nil
	}

	return fmt.Errorf("connection not from this pool")
}

// Remove removes a connection from the pool
func (cp *ConnectionPool) Remove(conn interface{}) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for i, c := range cp.connections {
		if c == conn {
			cp.removeConnection(i)
			if cp.closer != nil {
				return cp.closer(conn)
			}
			return nil
		}
	}

	return fmt.Errorf("connection not found in pool")
}

// removeConnection removes a connection at index (must be called with lock held)
func (cp *ConnectionPool) removeConnection(index int) {
	conn := cp.connections[index]
	delete(cp.active, conn)
	cp.connections = append(cp.connections[:index], cp.connections[index+1:]...)
	cp.currentSize--
}

// Close closes all connections in the pool
func (cp *ConnectionPool) Close() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	var lastErr error
	for _, conn := range cp.connections {
		if cp.closer != nil {
			if err := cp.closer(conn); err != nil {
				lastErr = err
			}
		}
	}

	cp.connections = nil
	cp.active = nil
	cp.currentSize = 0

	return lastErr
}

// Stats returns pool statistics
func (cp *ConnectionPool) Stats() map[string]interface{} {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	activeCount := 0
	for _, isActive := range cp.active {
		if isActive {
			activeCount++
		}
	}

	return map[string]interface{}{
		"total_connections":  cp.currentSize,
		"active_connections": activeCount,
		"idle_connections":   cp.currentSize - activeCount,
		"max_size":           cp.maxSize,
	}
}

// CacheManager manages in-memory caching
type CacheManager struct {
	mu      sync.RWMutex
	data    map[string]*CacheItem
	maxSize int
	ttl     time.Duration
	onEvict func(key string, value interface{})
}

// CacheItem represents a cached item
type CacheItem struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	Accessed  time.Time   `json:"accessed"`
	Hits      uint64      `json:"hits"`
	Size      int         `json:"size"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	MaxSize int                       `json:"max_size"`
	TTL     time.Duration             `json:"ttl"`
	OnEvict func(string, interface{}) `json:"-"`
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config *CacheConfig) *CacheManager {
	return &CacheManager{
		data:    make(map[string]*CacheItem),
		maxSize: config.MaxSize,
		ttl:     config.TTL,
		onEvict: config.OnEvict,
	}
}

// Set sets a value in the cache
func (cm *CacheManager) Set(key string, value interface{}) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if we need to evict items
	if len(cm.data) >= cm.maxSize {
		cm.evictLRU()
	}

	item := &CacheItem{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(cm.ttl),
		Accessed:  time.Now(),
		Hits:      0,
	}

	cm.data[key] = item
	return nil
}

// Get gets a value from the cache
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	item, exists := cm.data[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.ExpiresAt) {
		go cm.Delete(key) // Async delete
		return nil, false
	}

	// Update access info
	item.Accessed = time.Now()
	item.Hits++

	return item.Value, true
}

// Delete deletes a value from the cache
func (cm *CacheManager) Delete(key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if item, exists := cm.data[key]; exists {
		delete(cm.data, key)
		if cm.onEvict != nil {
			cm.onEvict(key, item.Value)
		}
	}
}

// Clear clears all items from the cache
func (cm *CacheManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.onEvict != nil {
		for key, item := range cm.data {
			cm.onEvict(key, item.Value)
		}
	}

	cm.data = make(map[string]*CacheItem)
}

// evictLRU evicts the least recently used item
func (cm *CacheManager) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range cm.data {
		if oldestKey == "" || item.Accessed.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.Accessed
		}
	}

	if oldestKey != "" {
		cm.Delete(oldestKey)
	}
}

// Stats returns cache statistics
func (cm *CacheManager) Stats() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	totalHits := uint64(0)
	expiredCount := 0
	now := time.Now()

	for _, item := range cm.data {
		totalHits += item.Hits
		if now.After(item.ExpiresAt) {
			expiredCount++
		}
	}

	hitRate := float64(0)
	if totalHits > 0 {
		hitRate = float64(totalHits) / float64(len(cm.data))
	}

	return map[string]interface{}{
		"total_items":     len(cm.data),
		"max_size":        cm.maxSize,
		"total_hits":      totalHits,
		"hit_rate":        hitRate,
		"expired_items":   expiredCount,
		"memory_usage_mb": cm.estimateMemoryUsage(),
	}
}

// estimateMemoryUsage estimates memory usage in MB
func (cm *CacheManager) estimateMemoryUsage() float64 {
	// Rough estimation - in a real implementation you'd use more sophisticated methods
	return float64(len(cm.data)) * 0.001 // Assume 1KB per item
}

// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
	workers    int
	jobQueue   chan Job
	workerPool chan chan Job
	quit       chan bool
	wg         sync.WaitGroup
}

// Job represents a job to be processed
type Job struct {
	ID       string                  `json:"id"`
	Type     string                  `json:"type"`
	Data     interface{}             `json:"data"`
	Function func(interface{}) error `json:"-"`
	Timeout  time.Duration           `json:"timeout"`
}

// Worker represents a worker goroutine
type Worker struct {
	id         int
	jobChannel chan Job
	workerPool chan chan Job
	quit       chan bool
	wg         *sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	return &WorkerPool{
		workers:    workers,
		jobQueue:   make(chan Job, queueSize),
		workerPool: make(chan chan Job, workers),
		quit:       make(chan bool),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	// Start workers
	for i := 0; i < wp.workers; i++ {
		worker := &Worker{
			id:         i,
			jobChannel: make(chan Job),
			workerPool: wp.workerPool,
			quit:       make(chan bool),
			wg:         &wp.wg,
		}
		worker.Start()
	}

	// Start dispatcher
	go wp.dispatch()
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	close(wp.quit)
	wp.wg.Wait()
}

// Submit submits a job to the pool
func (wp *WorkerPool) Submit(job Job) error {
	select {
	case wp.jobQueue <- job:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

// SubmitWithTimeout submits a job with timeout
func (wp *WorkerPool) SubmitWithTimeout(job Job, timeout time.Duration) error {
	select {
	case wp.jobQueue <- job:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("job submission timed out")
	}
}

// dispatch dispatches jobs to workers
func (wp *WorkerPool) dispatch() {
	for {
		select {
		case job := <-wp.jobQueue:
			go func() {
				jobChannel := <-wp.workerPool
				jobChannel <- job
			}()
		case <-wp.quit:
			return
		}
	}
}

// Start starts a worker
func (w *Worker) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			w.workerPool <- w.jobChannel
			select {
			case job := <-w.jobChannel:
				w.processJob(job)
			case <-w.quit:
				return
			}
		}
	}()
}

// processJob processes a job
func (w *Worker) processJob(job Job) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker panic recovered: %v", r)
		}
	}()

	if job.Function != nil {
		if job.Timeout > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- job.Function(job.Data)
			}()

			select {
			case <-ctx.Done():
				// Job timed out
			case err := <-done:
				// Job completed
				if err != nil {
					log.Printf("Error executing job function: %v", err)
				}
			}
		} else {
			if err := job.Function(job.Data); err != nil {
				log.Printf("Error executing job function: %v", err)
			}
		}
	}
}

// PerformanceOptimizer combines all performance optimizations
type PerformanceOptimizer struct {
	connectionPools map[string]*ConnectionPool
	cacheManagers   map[string]*CacheManager
	workerPools     map[string]*WorkerPool
	metrics         *MetricsCollector
}

// MetricsCollector collects performance metrics
type MetricsCollector struct {
	connectionPoolMetrics map[string]interface{}
	cacheMetrics          map[string]interface{}
	workerPoolMetrics     map[string]interface{}
}

// NewPerformanceOptimizer creates a new performance optimizer
func NewPerformanceOptimizer() *PerformanceOptimizer {
	return &PerformanceOptimizer{
		connectionPools: make(map[string]*ConnectionPool),
		cacheManagers:   make(map[string]*CacheManager),
		workerPools:     make(map[string]*WorkerPool),
		metrics: &MetricsCollector{
			connectionPoolMetrics: make(map[string]interface{}),
			cacheMetrics:          make(map[string]interface{}),
			workerPoolMetrics:     make(map[string]interface{}),
		},
	}
}

// AddConnectionPool adds a connection pool
func (po *PerformanceOptimizer) AddConnectionPool(name string, pool *ConnectionPool) {
	po.connectionPools[name] = pool
}

// AddCacheManager adds a cache manager
func (po *PerformanceOptimizer) AddCacheManager(name string, cache *CacheManager) {
	po.cacheManagers[name] = cache
}

// AddWorkerPool adds a worker pool
func (po *PerformanceOptimizer) AddWorkerPool(name string, pool *WorkerPool) {
	po.workerPools[name] = pool
}

// GetConnectionPool gets a connection pool by name
func (po *PerformanceOptimizer) GetConnectionPool(name string) *ConnectionPool {
	return po.connectionPools[name]
}

// GetCacheManager gets a cache manager by name
func (po *PerformanceOptimizer) GetCacheManager(name string) *CacheManager {
	return po.cacheManagers[name]
}

// GetWorkerPool gets a worker pool by name
func (po *PerformanceOptimizer) GetWorkerPool(name string) *WorkerPool {
	return po.workerPools[name]
}

// CollectMetrics collects performance metrics
func (po *PerformanceOptimizer) CollectMetrics() map[string]interface{} {
	po.metrics.connectionPoolMetrics = make(map[string]interface{})
	po.metrics.cacheMetrics = make(map[string]interface{})
	po.metrics.workerPoolMetrics = make(map[string]interface{})

	// Collect connection pool metrics
	for name, pool := range po.connectionPools {
		po.metrics.connectionPoolMetrics[name] = pool.Stats()
	}

	// Collect cache metrics
	for name, cache := range po.cacheManagers {
		po.metrics.cacheMetrics[name] = cache.Stats()
	}

	// Collect worker pool metrics
	for name, pool := range po.workerPools {
		po.metrics.workerPoolMetrics[name] = map[string]interface{}{
			"workers":    pool.workers,
			"queue_size": len(pool.jobQueue),
		}
	}

	return map[string]interface{}{
		"connection_pools": po.metrics.connectionPoolMetrics,
		"cache_managers":   po.metrics.cacheMetrics,
		"worker_pools":     po.metrics.workerPoolMetrics,
		"timestamp":        time.Now(),
	}
}

// OptimizeSystem applies system-wide optimizations
func (po *PerformanceOptimizer) OptimizeSystem() {
	// Clear expired cache items
	for _, cache := range po.cacheManagers {
		// This would be enhanced with better cleanup logic
		_ = cache // Use cache to avoid unused variable warning
	}

	// Optimize connection pools
	for _, pool := range po.connectionPools {
		// This would be enhanced with pool optimization logic
		_ = pool // Use pool to avoid unused variable warning
	}
}

// Shutdown gracefully shuts down all components
func (po *PerformanceOptimizer) Shutdown() error {
	var lastErr error

	// Close connection pools
	for _, pool := range po.connectionPools {
		if err := pool.Close(); err != nil {
			lastErr = err
		}
	}

	// Clear caches
	for _, cache := range po.cacheManagers {
		cache.Clear()
	}

	// Stop worker pools
	for _, pool := range po.workerPools {
		pool.Stop()
	}

	return lastErr
}

// CreateDatabasePool creates a database connection pool
func CreateDatabasePool(maxConnections int) *ConnectionPool {
	config := &PoolConfig{
		MaxSize:  maxConnections,
		InitSize: maxConnections / 2,
		Factory: func() (interface{}, error) {
			// In a real implementation, this would create a database connection
			return fmt.Sprintf("db_conn_%d", time.Now().UnixNano()), nil
		},
		Closer: func(conn interface{}) error {
			// In a real implementation, this would close the database connection
			return nil
		},
		Validator: func(conn interface{}) bool {
			// In a real implementation, this would validate the connection
			return conn != nil
		},
	}

	return NewConnectionPool(config)
}

// CreateMemoryCache creates an in-memory cache
func CreateMemoryCache(maxSize int, ttl time.Duration) *CacheManager {
	config := &CacheConfig{
		MaxSize: maxSize,
		TTL:     ttl,
		OnEvict: func(key string, value interface{}) {
			// Handle eviction
		},
	}

	return NewCacheManager(config)
}

// CreateWorkerPool creates a worker pool
func CreateWorkerPool(workers, queueSize int) *WorkerPool {
	pool := NewWorkerPool(workers, queueSize)
	pool.Start()
	return pool
}
