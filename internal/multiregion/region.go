package multiregion

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Region represents a geographic deployment region
type Region struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Country    string    `json:"country"`
	City       string    `json:"city"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Status     string    `json:"status"` // "active", "standby", "maintenance"
	Load       int       `json:"load"`
	Capacity   int       `json:"capacity"`
	LastHealth time.Time `json:"last_health"`
	Endpoint   string    `json:"endpoint"`
	Priority   int       `json:"priority"` // 1=highest, 5=lowest
}

// RegionManager manages multi-region deployments
type RegionManager struct {
	regions       map[string]*Region
	regionsMutex  sync.RWMutex
	config        *MultiRegionConfig
	healthChecker *HealthChecker
	loadBalancer  *LoadBalancer
}

// MultiRegionConfig holds configuration for multi-region deployment
type MultiRegionConfig struct {
	Enabled             bool          `yaml:"enabled"`
	DefaultRegion       string        `yaml:"default_region"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	FailoverThreshold   int           `yaml:"failover_threshold"`
	LoadBalanceStrategy string        `yaml:"load_balance_strategy"` // "round_robin", "weighted", "geographic"
}

// HealthChecker monitors region health
type HealthChecker struct {
	interval time.Duration
	regions  *map[string]*Region
	mutex    *sync.RWMutex
}

// LoadBalancer distributes load across regions
type LoadBalancer struct {
	strategy        string
	regions         *map[string]*Region
	mutex           *sync.RWMutex
	roundRobinIndex int
}

// NewRegionManager creates a new region manager
func NewRegionManager(config *MultiRegionConfig) *RegionManager {
	rm := &RegionManager{
		regions: make(map[string]*Region),
		config:  config,
	}

	if config.Enabled {
		rm.healthChecker = &HealthChecker{
			interval: config.HealthCheckInterval,
			regions:  &rm.regions,
			mutex:    &rm.regionsMutex,
		}

		rm.loadBalancer = &LoadBalancer{
			strategy: config.LoadBalanceStrategy,
			regions:  &rm.regions,
			mutex:    &rm.regionsMutex,
		}

		// Initialize default regions for simulation
		rm.initializeDefaultRegions()
		go rm.healthChecker.Start()
	}

	return rm
}

// initializeDefaultRegions creates simulated regions for testing
func (rm *RegionManager) initializeDefaultRegions() {
	defaultRegions := []*Region{
		{
			ID:        "us-east-1",
			Name:      "US East (N. Virginia)",
			Country:   "United States",
			City:      "Ashburn, VA",
			Latitude:  39.0438,
			Longitude: -77.4874,
			Status:    "active",
			Load:      45,
			Capacity:  100,
			Endpoint:  "https://us-east-1.hades.local",
			Priority:  1,
		},
		{
			ID:        "us-west-2",
			Name:      "US West (Oregon)",
			Country:   "United States",
			City:      "Portland, OR",
			Latitude:  45.5152,
			Longitude: -122.6784,
			Status:    "active",
			Load:      32,
			Capacity:  100,
			Endpoint:  "https://us-west-2.hades.local",
			Priority:  2,
		},
		{
			ID:        "eu-west-1",
			Name:      "EU West (Ireland)",
			Country:   "Ireland",
			City:      "Dublin",
			Latitude:  53.4084,
			Longitude: -8.2439,
			Status:    "active",
			Load:      28,
			Capacity:  100,
			Endpoint:  "https://eu-west-1.hades.local",
			Priority:  2,
		},
		{
			ID:        "ap-southeast-1",
			Name:      "Asia Pacific (Singapore)",
			Country:   "Singapore",
			City:      "Singapore",
			Latitude:  1.3521,
			Longitude: 103.8198,
			Status:    "standby",
			Load:      15,
			Capacity:  100,
			Endpoint:  "https://ap-southeast-1.hades.local",
			Priority:  3,
		},
		{
			ID:        "ap-northeast-1",
			Name:      "Asia Pacific (Tokyo)",
			Country:   "Japan",
			City:      "Tokyo",
			Latitude:  35.6762,
			Longitude: 139.6503,
			Status:    "standby",
			Load:      12,
			Capacity:  100,
			Endpoint:  "https://ap-northeast-1.hades.local",
			Priority:  3,
		},
	}

	rm.regionsMutex.Lock()
	defer rm.regionsMutex.Unlock()

	for _, region := range defaultRegions {
		region.LastHealth = time.Now()
		rm.regions[region.ID] = region
	}
}

// AddRegion adds a new region to the manager
func (rm *RegionManager) AddRegion(region *Region) error {
	rm.regionsMutex.Lock()
	defer rm.regionsMutex.Unlock()

	if _, exists := rm.regions[region.ID]; exists {
		return fmt.Errorf("region %s already exists", region.ID)
	}

	region.LastHealth = time.Now()
	rm.regions[region.ID] = region
	return nil
}

// GetRegion returns a region by ID
func (rm *RegionManager) GetRegion(id string) (*Region, error) {
	rm.regionsMutex.RLock()
	defer rm.regionsMutex.RUnlock()

	region, exists := rm.regions[id]
	if !exists {
		return nil, fmt.Errorf("region %s not found", id)
	}

	return region, nil
}

// ListRegions returns all regions
func (rm *RegionManager) ListRegions() []*Region {
	rm.regionsMutex.RLock()
	defer rm.regionsMutex.RUnlock()

	regions := make([]*Region, 0, len(rm.regions))
	for _, region := range rm.regions {
		regions = append(regions, region)
	}

	return regions
}

// GetOptimalRegion returns the best region for a request based on strategy
func (rm *RegionManager) GetOptimalRegion(clientLocation *ClientLocation) (*Region, error) {
	if !rm.config.Enabled {
		return nil, fmt.Errorf("multi-region is disabled")
	}

	rm.regionsMutex.RLock()
	defer rm.regionsMutex.RUnlock()

	if rm.loadBalancer == nil {
		return nil, fmt.Errorf("load balancer not initialized")
	}

	return rm.loadBalancer.SelectRegion(clientLocation)
}

// Start begins the region manager operations
func (rm *RegionManager) Start(ctx context.Context) error {
	if !rm.config.Enabled {
		return fmt.Errorf("multi-region is disabled")
	}

	// Start health checking
	if rm.healthChecker != nil {
		go rm.healthChecker.Start()
	}

	return nil
}

// Stop shuts down the region manager
func (rm *RegionManager) Stop() {
	// Cleanup resources
}

// ClientLocation represents client geographic information
type ClientLocation struct {
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Start begins health checking for all regions
func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for range ticker.C {
		hc.checkAllRegions()
	}
}

// checkAllRegions performs health checks on all regions
func (hc *HealthChecker) checkAllRegions() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	for id, region := range *hc.regions {
		// Simulate health check
		healthy := hc.simulateHealthCheck(region)

		if healthy {
			region.Status = "active"
			region.LastHealth = time.Now()
		} else {
			region.Status = "unhealthy"
		}

		(*hc.regions)[id] = region
	}
}

// simulateHealthCheck simulates a health check for a region
func (hc *HealthChecker) simulateHealthCheck(region *Region) bool {
	// Simulate 95% success rate for active regions, 70% for standby
	successRate := 0.95
	if region.Status == "standby" {
		successRate = 0.7
	}

	// Simulate random health check
	return int(time.Now().UnixNano()%100) < int(successRate*100)
}

// SelectRegion chooses the best region based on the configured strategy
func (lb *LoadBalancer) SelectRegion(clientLocation *ClientLocation) (*Region, error) {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	activeRegions := make([]*Region, 0)
	for _, region := range *lb.regions {
		if region.Status == "active" && region.Load < region.Capacity {
			activeRegions = append(activeRegions, region)
		}
	}

	if len(activeRegions) == 0 {
		return nil, fmt.Errorf("no active regions available")
	}

	switch lb.strategy {
	case "geographic":
		return lb.selectGeographicRegion(activeRegions, clientLocation)
	case "weighted":
		return lb.selectWeightedRegion(activeRegions)
	default: // round_robin
		return lb.selectRoundRobinRegion(activeRegions)
	}
}

// selectGeographicRegion selects the closest region geographically
func (lb *LoadBalancer) selectGeographicRegion(regions []*Region, clientLocation *ClientLocation) (*Region, error) {
	if clientLocation == nil {
		// Fall back to round robin if no location info
		return lb.selectRoundRobinRegion(regions)
	}

	bestRegion := regions[0]
	minDistance := calculateDistance(
		clientLocation.Latitude, clientLocation.Longitude,
		bestRegion.Latitude, bestRegion.Longitude,
	)

	for _, region := range regions[1:] {
		distance := calculateDistance(
			clientLocation.Latitude, clientLocation.Longitude,
			region.Latitude, region.Longitude,
		)
		if distance < minDistance {
			minDistance = distance
			bestRegion = region
		}
	}

	return bestRegion, nil
}

// selectWeightedRegion selects region based on load and priority
func (lb *LoadBalancer) selectWeightedRegion(regions []*Region) (*Region, error) {
	// Calculate weight based on available capacity and priority
	bestRegion := regions[0]
	bestWeight := float64(bestRegion.Capacity-bestRegion.Load) / float64(bestRegion.Priority)

	for _, region := range regions[1:] {
		weight := float64(region.Capacity-region.Load) / float64(region.Priority)
		if weight > bestWeight {
			bestWeight = weight
			bestRegion = region
		}
	}

	return bestRegion, nil
}

// selectRoundRobinRegion selects region using round-robin
func (lb *LoadBalancer) selectRoundRobinRegion(regions []*Region) (*Region, error) {
	region := regions[lb.roundRobinIndex%len(regions)]
	lb.roundRobinIndex++
	return region, nil
}

// calculateDistance calculates distance between two geographic points (simplified)
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Simplified distance calculation
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	return dlat*dlat + dlon*dlon
}

// GetRegionStats returns statistics for all regions
func (rm *RegionManager) GetRegionStats() map[string]interface{} {
	rm.regionsMutex.RLock()
	defer rm.regionsMutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_regions"] = len(rm.regions)
	stats["active_regions"] = 0
	stats["standby_regions"] = 0
	stats["total_capacity"] = 0
	stats["total_load"] = 0
	stats["average_load"] = 0.0

	for _, region := range rm.regions {
		switch region.Status {
		case "active":
			stats["active_regions"] = stats["active_regions"].(int) + 1
		case "standby":
			stats["standby_regions"] = stats["standby_regions"].(int) + 1
		}

		stats["total_capacity"] = stats["total_capacity"].(int) + region.Capacity
		stats["total_load"] = stats["total_load"].(int) + region.Load
	}

	if len(rm.regions) > 0 {
		stats["average_load"] = float64(stats["total_load"].(int)) / float64(stats["total_capacity"].(int)) * 100
	}

	return stats
}
