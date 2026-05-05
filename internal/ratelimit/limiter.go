package ratelimit

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter manages advanced rate limiting and DDoS protection
type RateLimiter struct {
	limits          map[string]*LimitConfig
	buckets         map[string]*TokenBucket
	clients         map[string]*ClientState
	ddosDetector    *DDoSDetector
	whitelist       map[string]bool
	blacklist       map[string]bool
	mu              sync.RWMutex
	enabled         bool
	cleanupInterval time.Duration
	stats           *RateLimitStats
}

// LimitConfig represents rate limit configuration
type LimitConfig struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Requests    int           `json:"requests"`
	Window      time.Duration `json:"window"`
	Burst       int           `json:"burst"`
	Priority    int           `json:"priority"`
	Enabled     bool          `json:"enabled"`
	Conditions  []Condition   `json:"conditions"`
	Actions     []Action      `json:"actions"`
	Created     time.Time     `json:"created"`
	Updated     time.Time     `json:"updated"`
}

// Condition represents a rate limit condition
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Type     string      `json:"type"` // "ip", "user", "path", "method", "header"
}

// Action represents a rate limit action
type Action struct {
	Type       string                 `json:"type"` // "block", "throttle", "alert", "redirect"
	Parameters map[string]interface{} `json:"parameters"`
	Enabled    bool                   `json:"enabled"`
}

// TokenBucket implements token bucket algorithm
type TokenBucket struct {
	capacity   int
	tokens     int
	refillRate int
	lastRefill time.Time
	mu         sync.Mutex
}

// ClientState tracks client state for rate limiting
type ClientState struct {
	IPAddress  string                  `json:"ip_address"`
	UserID     string                  `json:"user_id"`
	SessionID  string                  `json:"session_id"`
	Requests   int                     `json:"requests"`
	FirstSeen  time.Time               `json:"first_seen"`
	LastSeen   time.Time               `json:"last_seen"`
	Violations int                     `json:"violations"`
	Blocked    bool                    `json:"blocked"`
	BlockUntil time.Time               `json:"block_until"`
	Buckets    map[string]*TokenBucket `json:"buckets"`
	Metadata   map[string]interface{}  `json:"metadata"`
}

// DDoSDetector detects DDoS attacks
type DDoSDetector struct {
	thresholds      map[string]*DDoSThreshold
	attacks         map[string]*DDoSAttack
	mu              sync.RWMutex
	enabled         bool
	detectionWindow time.Duration
}

// DDoSThreshold represents DDoS detection thresholds
type DDoSThreshold struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Type      string        `json:"type"` // "requests_per_second", "connections_per_ip", "bandwidth"
	Threshold int           `json:"threshold"`
	Window    time.Duration `json:"window"`
	Severity  string        `json:"severity"` // "low", "medium", "high", "critical"
	Enabled   bool          `json:"enabled"`
	Actions   []Action      `json:"actions"`
}

// DDoSAttack represents a detected DDoS attack
type DDoSAttack struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Source     string                 `json:"source"`
	Target     string                 `json:"target"`
	Severity   string                 `json:"severity"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	Active     bool                   `json:"active"`
	Metrics    map[string]interface{} `json:"metrics"`
	Mitigation []Action               `json:"mitigation"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// RateLimitRequest represents a rate limit request
type RateLimitRequest struct {
	IPAddress string                 `json:"ip_address"`
	UserID    string                 `json:"user_id"`
	SessionID string                 `json:"session_id"`
	Path      string                 `json:"path"`
	Method    string                 `json:"method"`
	Headers   map[string]string      `json:"headers"`
	UserAgent string                 `json:"user_agent"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// RateLimitResult represents the result of rate limiting
type RateLimitResult struct {
	Allowed   bool                   `json:"allowed"`
	Limited   bool                   `json:"limited"`
	Blocked   bool                   `json:"blocked"`
	Reason    string                 `json:"reason"`
	Remaining int                    `json:"remaining"`
	ResetTime time.Time              `json:"reset_time"`
	Actions   []Action               `json:"actions"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// RateLimitStats represents rate limiting statistics
type RateLimitStats struct {
	TotalRequests   int64 `json:"total_requests"`
	AllowedRequests int64 `json:"allowed_requests"`
	LimitedRequests int64 `json:"limited_requests"`
	BlockedRequests int64 `json:"blocked_requests"`
	ActiveClients   int   `json:"active_clients"`
	ActiveAttacks   int   `json:"active_attacks"`
	DDoSDetections  int64 `json:"ddos_detections"`
	WhitelistedIPs  int   `json:"whitelisted_ips"`
	BlacklistedIPs  int   `json:"blacklisted_ips"`
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limits:          make(map[string]*LimitConfig),
		buckets:         make(map[string]*TokenBucket),
		clients:         make(map[string]*ClientState),
		whitelist:       make(map[string]bool),
		blacklist:       make(map[string]bool),
		enabled:         true,
		cleanupInterval: 5 * time.Minute,
		stats:           &RateLimitStats{},
		ddosDetector:    NewDDoSDetector(),
	}

	// Initialize default rate limits
	rl.initializeDefaultLimits()

	return rl
}

// initializeDefaultLimits initializes default rate limit configurations
func (rl *RateLimiter) initializeDefaultLimits() {
	defaultLimits := []*LimitConfig{
		{
			ID:          "global_api",
			Name:        "Global API Rate Limit",
			Description: "Global rate limit for all API requests",
			Requests:    1000,
			Window:      1 * time.Minute,
			Burst:       100,
			Priority:    1,
			Enabled:     true,
			Conditions: []Condition{
				{
					Field:    "path",
					Operator: "regex",
					Value:    "^/api/",
					Type:     "path",
				},
			},
			Actions: []Action{
				{
					Type: "throttle",
					Parameters: map[string]interface{}{
						"delay": "100ms",
					},
					Enabled: true,
				},
			},
			Created: time.Now(),
			Updated: time.Now(),
		},
		{
			ID:          "auth_endpoints",
			Name:        "Authentication Endpoints",
			Description: "Rate limit for authentication endpoints",
			Requests:    10,
			Window:      1 * time.Minute,
			Burst:       5,
			Priority:    2,
			Enabled:     true,
			Conditions: []Condition{
				{
					Field:    "path",
					Operator: "regex",
					Value:    "^/api/.*/auth/",
					Type:     "path",
				},
			},
			Actions: []Action{
				{
					Type: "block",
					Parameters: map[string]interface{}{
						"duration": "15m",
					},
					Enabled: true,
				},
			},
			Created: time.Now(),
			Updated: time.Now(),
		},
		{
			ID:          "sensitive_operations",
			Name:        "Sensitive Operations",
			Description: "Rate limit for sensitive operations",
			Requests:    5,
			Window:      1 * time.Minute,
			Burst:       2,
			Priority:    3,
			Enabled:     true,
			Conditions: []Condition{
				{
					Field:    "path",
					Operator: "regex",
					Value:    "^/api/.*/(delete|admin|config)",
					Type:     "path",
				},
			},
			Actions: []Action{
				{
					Type: "block",
					Parameters: map[string]interface{}{
						"duration": "1h",
					},
					Enabled: true,
				},
				{
					Type: "alert",
					Parameters: map[string]interface{}{
						"priority": "high",
					},
					Enabled: true,
				},
			},
			Created: time.Now(),
			Updated: time.Now(),
		},
		{
			ID:          "per_ip_limit",
			Name:        "Per IP Rate Limit",
			Description: "Rate limit per IP address",
			Requests:    100,
			Window:      1 * time.Minute,
			Burst:       20,
			Priority:    4,
			Enabled:     true,
			Conditions: []Condition{
				{
					Field:    "ip",
					Operator: "exists",
					Value:    "",
					Type:     "ip",
				},
			},
			Actions: []Action{
				{
					Type: "throttle",
					Parameters: map[string]interface{}{
						"delay": "500ms",
					},
					Enabled: true,
				},
			},
			Created: time.Now(),
			Updated: time.Now(),
		},
	}

	for _, limit := range defaultLimits {
		rl.limits[limit.ID] = limit
	}
}

// CheckRateLimit checks if a request should be rate limited
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, req RateLimitRequest) (*RateLimitResult, error) {
	if !rl.enabled {
		return &RateLimitResult{
			Allowed:   true,
			Limited:   false,
			Blocked:   false,
			Reason:    "rate limiting disabled",
			Remaining: -1,
		}, nil
	}

	rl.stats.TotalRequests++

	// Check whitelist and blacklist
	if rl.isWhitelisted(req.IPAddress) {
		rl.stats.AllowedRequests++
		return &RateLimitResult{
			Allowed:   true,
			Limited:   false,
			Blocked:   false,
			Reason:    "whitelisted",
			Remaining: -1,
		}, nil
	}

	if rl.isBlacklisted(req.IPAddress) {
		rl.stats.BlockedRequests++
		return &RateLimitResult{
			Allowed:   false,
			Limited:   false,
			Blocked:   true,
			Reason:    "blacklisted",
			Remaining: 0,
		}, nil
	}

	// Get or create client state
	client := rl.getOrCreateClient(req)

	// Check if client is blocked
	if client.Blocked && time.Now().Before(client.BlockUntil) {
		rl.stats.BlockedRequests++
		return &RateLimitResult{
			Allowed:   false,
			Limited:   false,
			Blocked:   true,
			Reason:    "client blocked",
			Remaining: 0,
			ResetTime: client.BlockUntil,
		}, nil
	}

	// Check DDoS detection
	if rl.ddosDetector.enabled {
		attack := rl.ddosDetector.DetectAttack(req)
		if attack != nil {
			rl.stats.DDoSDetections++
			rl.handleDDoSAttack(ctx, attack, req)
			return &RateLimitResult{
				Allowed:   false,
				Limited:   false,
				Blocked:   true,
				Reason:    "ddos attack detected",
				Remaining: 0,
			}, nil
		}
	}

	// Check rate limits
	for _, limit := range rl.limits {
		if !limit.Enabled {
			continue
		}

		if rl.matchesConditions(limit.Conditions, req) {
			result := rl.checkLimit(ctx, limit, req, client)
			if !result.Allowed {
				rl.stats.LimitedRequests++
				return result, nil
			}
		}
	}

	rl.stats.AllowedRequests++
	return &RateLimitResult{
		Allowed:   true,
		Limited:   false,
		Blocked:   false,
		Reason:    "allowed",
		Remaining: -1,
	}, nil
}

// matchesConditions checks if request matches limit conditions
func (rl *RateLimiter) matchesConditions(conditions []Condition, req RateLimitRequest) bool {
	for _, condition := range conditions {
		if !rl.matchesCondition(condition, req) {
			return false
		}
	}
	return true
}

// matchesCondition checks if request matches a single condition
func (rl *RateLimiter) matchesCondition(condition Condition, req RateLimitRequest) bool {
	var fieldValue interface{}

	switch condition.Type {
	case "ip":
		fieldValue = req.IPAddress
	case "user":
		fieldValue = req.UserID
	case "path":
		fieldValue = req.Path
	case "method":
		fieldValue = req.Method
	case "header":
		if value, exists := req.Headers[condition.Field]; exists {
			fieldValue = value
		} else {
			return false
		}
	default:
		return false
	}

	return rl.evaluateCondition(condition.Operator, fieldValue, condition.Value)
}

// evaluateCondition evaluates a condition
func (rl *RateLimiter) evaluateCondition(operator string, fieldValue, conditionValue interface{}) bool {
	switch operator {
	case "equals":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", conditionValue)
	case "exists":
		return fieldValue != nil && fieldValue != ""
	case "regex":
		// Simplified regex matching - in production, use actual regex
		fieldStr := fmt.Sprintf("%v", fieldValue)
		conditionStr := fmt.Sprintf("%v", conditionValue)
		return len(fieldStr) > 0 && len(conditionStr) > 0
	case "contains":
		fieldStr := fmt.Sprintf("%v", fieldValue)
		conditionStr := fmt.Sprintf("%v", conditionValue)
		return len(fieldStr) > 0 && len(conditionStr) > 0
	default:
		return false
	}
}

// checkLimit checks a specific rate limit
func (rl *RateLimiter) checkLimit(ctx context.Context, limit *LimitConfig, req RateLimitRequest, client *ClientState) *RateLimitResult {
	// Get or create token bucket for this limit
	bucketKey := fmt.Sprintf("%s_%s", limit.ID, req.IPAddress)
	bucket := rl.getOrCreateBucket(bucketKey, limit)

	// Check if request is allowed
	if bucket.consume(1) {
		return &RateLimitResult{
			Allowed:   true,
			Limited:   false,
			Blocked:   false,
			Reason:    "within limit",
			Remaining: bucket.tokens,
		}
	}

	// Rate limit exceeded - execute actions
	for _, action := range limit.Actions {
		if action.Enabled {
			rl.executeAction(ctx, action, req, client)
		}
	}

	return &RateLimitResult{
		Allowed:   false,
		Limited:   true,
		Blocked:   false,
		Reason:    fmt.Sprintf("rate limit exceeded: %s", limit.Name),
		Remaining: 0,
		ResetTime: time.Now().Add(limit.Window),
		Actions:   limit.Actions,
	}
}

// executeAction executes a rate limit action
func (rl *RateLimiter) executeAction(ctx context.Context, action Action, req RateLimitRequest, client *ClientState) {
	switch action.Type {
	case "block":
		rl.executeBlockAction(ctx, action, req, client)
	case "throttle":
		rl.executeThrottleAction(ctx, action, req, client)
	case "alert":
		rl.executeAlertAction(ctx, action, req, client)
	case "redirect":
		rl.executeRedirectAction(ctx, action, req, client)
	}
}

// executeBlockAction executes a block action
func (rl *RateLimiter) executeBlockAction(ctx context.Context, action Action, req RateLimitRequest, client *ClientState) {
	duration := 15 * time.Minute // Default duration
	if d, ok := action.Parameters["duration"].(string); ok {
		if parsed, err := time.ParseDuration(d); err == nil {
			duration = parsed
		}
	}

	client.Blocked = true
	client.BlockUntil = time.Now().Add(duration)
	client.Violations++

	log.Printf("Blocked client %s for %v due to rate limit violation", req.IPAddress, duration)
}

// executeThrottleAction executes a throttle action
func (rl *RateLimiter) executeThrottleAction(ctx context.Context, action Action, req RateLimitRequest, client *ClientState) {
	delay := 100 * time.Millisecond // Default delay
	if d, ok := action.Parameters["delay"].(string); ok {
		if parsed, err := time.ParseDuration(d); err == nil {
			delay = parsed
		}
	}

	log.Printf("Throttling client %s with %v delay", req.IPAddress, delay)
}

// executeAlertAction executes an alert action
func (rl *RateLimiter) executeAlertAction(ctx context.Context, action Action, req RateLimitRequest, client *ClientState) {
	priority := "medium"
	if p, ok := action.Parameters["priority"].(string); ok {
		priority = p
	}

	log.Printf("ALERT: Rate limit violation from %s (priority: %s)", req.IPAddress, priority)
}

// executeRedirectAction executes a redirect action
func (rl *RateLimiter) executeRedirectAction(ctx context.Context, action Action, req RateLimitRequest, client *ClientState) {
	redirectURL := "/rate-limited"
	if url, ok := action.Parameters["url"].(string); ok {
		redirectURL = url
	}

	log.Printf("Redirecting client %s to %s", req.IPAddress, redirectURL)
}

// getOrCreateClient gets or creates a client state
func (rl *RateLimiter) getOrCreateClient(req RateLimitRequest) *ClientState {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	clientKey := req.IPAddress
	if req.UserID != "" {
		clientKey = req.UserID
	}

	client, exists := rl.clients[clientKey]
	if !exists {
		client = &ClientState{
			IPAddress: req.IPAddress,
			UserID:    req.UserID,
			SessionID: req.SessionID,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
			Buckets:   make(map[string]*TokenBucket),
			Metadata:  make(map[string]interface{}),
		}
		rl.clients[clientKey] = client
	}

	client.LastSeen = time.Now()
	client.Requests++

	return client
}

// getOrCreateBucket gets or creates a token bucket
func (rl *RateLimiter) getOrCreateBucket(key string, limit *LimitConfig) *TokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			capacity:   limit.Requests,
			tokens:     limit.Requests,
			refillRate: limit.Requests / int(limit.Window.Seconds()),
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}

	return bucket
}

// consume consumes tokens from the bucket
func (tb *TokenBucket) consume(tokens int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}

	// Check if enough tokens are available
	if tb.tokens >= tokens {
		tb.tokens -= tokens
		return true
	}

	return false
}

// handleDDoSAttack handles a detected DDoS attack
func (rl *RateLimiter) handleDDoSAttack(ctx context.Context, attack *DDoSAttack, req RateLimitRequest) {
	log.Printf("DDoS attack detected: %s from %s", attack.Type, attack.Source)

	// Block the source IP
	rl.blacklist[req.IPAddress] = true

	// Execute mitigation actions
	for _, action := range attack.Mitigation {
		if action.Enabled {
			rl.executeAction(ctx, action, req, rl.getOrCreateClient(req))
		}
	}
}

// isWhitelisted checks if an IP is whitelisted
func (rl *RateLimiter) isWhitelisted(ip string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	// Check exact match
	if rl.whitelist[ip] {
		return true
	}

	// Check CIDR ranges (simplified)
	// In production, implement proper CIDR matching
	return false
}

// isBlacklisted checks if an IP is blacklisted
func (rl *RateLimiter) isBlacklisted(ip string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.blacklist[ip]
}

// AddWhitelist adds an IP to the whitelist
func (rl *RateLimiter) AddWhitelist(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.whitelist[ip] = true
}

// AddBlacklist adds an IP to the blacklist
func (rl *RateLimiter) AddBlacklist(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.blacklist[ip] = true
}

// RemoveWhitelist removes an IP from the whitelist
func (rl *RateLimiter) RemoveWhitelist(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.whitelist, ip)
}

// RemoveBlacklist removes an IP from the blacklist
func (rl *RateLimiter) RemoveBlacklist(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.blacklist, ip)
}

// GetStats returns rate limiting statistics
func (rl *RateLimiter) GetStats() *RateLimitStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	rl.stats.ActiveClients = len(rl.clients)
	rl.stats.WhitelistedIPs = len(rl.whitelist)
	rl.stats.BlacklistedIPs = len(rl.blacklist)
	rl.stats.ActiveAttacks = rl.ddosDetector.GetActiveAttackCount()

	return rl.stats
}

// Cleanup performs periodic cleanup of old data
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Hour) // Remove clients older than 1 hour

	for key, client := range rl.clients {
		if client.LastSeen.Before(cutoff) {
			delete(rl.clients, key)
		}
	}

	// Clean up old token buckets
	for key, bucket := range rl.buckets {
		if now.Sub(bucket.lastRefill) > 1*time.Hour {
			delete(rl.buckets, key)
		}
	}
}

// NewDDoSDetector creates a new DDoS detector
func NewDDoSDetector() *DDoSDetector {
	dd := &DDoSDetector{
		thresholds:      make(map[string]*DDoSThreshold),
		attacks:         make(map[string]*DDoSAttack),
		enabled:         true,
		detectionWindow: 1 * time.Minute,
	}

	// Initialize default thresholds
	dd.initializeDefaultThresholds()

	return dd
}

// initializeDefaultThresholds initializes default DDoS detection thresholds
func (dd *DDoSDetector) initializeDefaultThresholds() {
	defaultThresholds := []*DDoSThreshold{
		{
			ID:        "requests_per_second",
			Name:      "Requests Per Second",
			Type:      "requests_per_second",
			Threshold: 1000,
			Window:    1 * time.Minute,
			Severity:  "high",
			Enabled:   true,
			Actions: []Action{
				{
					Type: "block",
					Parameters: map[string]interface{}{
						"duration": "1h",
					},
					Enabled: true,
				},
			},
		},
		{
			ID:        "connections_per_ip",
			Name:      "Connections Per IP",
			Type:      "connections_per_ip",
			Threshold: 100,
			Window:    1 * time.Minute,
			Severity:  "medium",
			Enabled:   true,
			Actions: []Action{
				{
					Type: "throttle",
					Parameters: map[string]interface{}{
						"delay": "1s",
					},
					Enabled: true,
				},
			},
		},
	}

	for _, threshold := range defaultThresholds {
		dd.thresholds[threshold.ID] = threshold
	}
}

// DetectAttack detects if a request is part of a DDoS attack
func (dd *DDoSDetector) DetectAttack(req RateLimitRequest) *DDoSAttack {
	if !dd.enabled {
		return nil
	}

	dd.mu.RLock()
	defer dd.mu.RUnlock()

	// Simplified DDoS detection
	// In production, implement sophisticated detection algorithms
	for _, threshold := range dd.thresholds {
		if !threshold.Enabled {
			continue
		}

		if dd.checkThreshold(threshold, req) {
			attack := &DDoSAttack{
				ID:         fmt.Sprintf("ddos_%d", time.Now().UnixNano()),
				Type:       threshold.Type,
				Source:     req.IPAddress,
				Target:     req.Path,
				Severity:   threshold.Severity,
				StartTime:  time.Now(),
				Active:     true,
				Metrics:    make(map[string]interface{}),
				Mitigation: threshold.Actions,
				Metadata:   make(map[string]interface{}),
			}

			dd.attacks[attack.ID] = attack
			return attack
		}
	}

	return nil
}

// checkThreshold checks if a threshold is exceeded
func (dd *DDoSDetector) checkThreshold(threshold *DDoSThreshold, req RateLimitRequest) bool {
	// Simplified threshold checking
	// In production, implement proper threshold checking with sliding windows
	switch threshold.Type {
	case "requests_per_second":
		// Mock implementation - in production, track actual request rate
		return false
	case "connections_per_ip":
		// Mock implementation - in production, track actual connections per IP
		return false
	default:
		return false
	}
}

// GetActiveAttackCount returns the number of active DDoS attacks
func (dd *DDoSDetector) GetActiveAttackCount() int {
	dd.mu.RLock()
	defer dd.mu.RUnlock()

	count := 0
	for _, attack := range dd.attacks {
		if attack.Active {
			count++
		}
	}

	return count
}

// HTTPMiddleware creates HTTP middleware for rate limiting
func (rl *RateLimiter) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := RateLimitRequest{
			IPAddress: getClientIP(r),
			Path:      r.URL.Path,
			Method:    r.Method,
			Headers:   make(map[string]string),
			UserAgent: r.UserAgent(),
			Timestamp: time.Now(),
		}

		// Copy headers
		for key, values := range r.Header {
			if len(values) > 0 {
				req.Headers[key] = values[0]
			}
		}

		// Check rate limit
		result, err := rl.CheckRateLimit(context.Background(), req)
		if err != nil {
			log.Printf("Rate limiting error: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		// Handle rate limiting response
		if !result.Allowed {
			if result.Blocked {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Request blocked due to rate limiting"))
			} else if result.Limited {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Request rate limited"))
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
