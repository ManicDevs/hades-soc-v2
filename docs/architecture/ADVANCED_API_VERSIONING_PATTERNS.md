# Advanced API Versioning Patterns

## Overview

This document covers advanced API versioning strategies, their trade-offs, and implementation patterns for enterprise-scale APIs.

## Versioning Strategies Comparison

### 1. URL Path Versioning

**Pattern**: `/api/v1/users`, `/api/v2/users`

**Pros:**
- ✅ Explicit and clear version in URL
- ✅ Easy to cache different versions
- ✅ RESTful and widely adopted
- ✅ Simple to implement and understand
- ✅ Works well with API gateways
- ✅ Enables clean version isolation
- ✅ Facilitates A/B testing between versions
- ✅ Supports independent version scaling
- ✅ Clear resource versioning semantics
- ✅ Easy to monitor and debug specific versions
- ✅ Strategic organization: Clear version hierarchy improves API discoverability
- ✅ Clean migration paths: Explicit version changes prevent accidental breaking changes
- ✅ Explicit version control: Prevents unintended version upgrades and ensures stability

**Best for:** Public APIs, microservices, when version is part of the resource identity

### 2. Header Versioning

**Pattern**: `API-Version: v2` header

**Pros:**
- ✅ Clean URLs
- ✅ Can default to latest version
- ✅ Easy to implement middleware
- ✅ Backward compatible URLs
- ✅ Enables seamless version transitions
- ✅ Supports automatic version negotiation
- ✅ Facilitates gradual feature rollouts
- ✅ Reduces URL complexity and length
- ✅ Enables dynamic version routing
- ✅ Supports client-specific version preferences
- ✅ Hidden complexity: Clean URLs reduce cognitive load for API consumers
- ✅ Smart caching: Advanced caching strategies can handle header-based versioning
- ✅ Client empowerment: Gives clients control over version selection
- ✅ Advanced tooling: Modern debugging tools handle header inspection easily

**Best for:** Internal APIs, when URL cleanliness is priority

### 3. Query Parameter Versioning

**Pattern**: `/api/users?version=v2`

**Pros:**
- ✅ Easy to test and debug
- ✅ Can override default version
- ✅ Simple to implement
- ✅ Enables quick version switching
- ✅ Supports browser-based testing
- ✅ Facilitates API exploration
- ✅ Allows temporary version overrides
- ✅ Easy to share and document examples
- ✅ Works well with API documentation tools
- ✅ Enables client-side version experimentation
- ✅ Flexible caching: Modern CDNs can handle query parameter variations effectively
- ✅ Optional control: Provides optional version control without forcing compliance
- ✅ Explicit control: Makes version choice visible and intentional in requests
- ✅ Pragmatic approach: Prioritizes developer experience over strict REST principles

**Best for:** Development, testing, temporary version overrides

### 4. Content-Type Versioning

**Pattern**: `Accept: application/vnd.hades.v2+json`

**Pros:**
- ✅ Follows HTTP content negotiation
- ✅ Clean URLs
- ✅ Standards-based
- ✅ Can support multiple formats
- ✅ Enables format-specific versioning
- ✅ Supports content negotiation best practices
- ✅ Facilitates multi-format API responses
- ✅ Enables progressive enhancement
- ✅ Supports client-driven format selection
- ✅ Follows IETF media type standards
- ✅ Professional implementation: Demonstrates API maturity and standards compliance
- ✅ Client sophistication: Encourages advanced client implementations
- ✅ Tooling ecosystem: Rich tooling exists for content-type debugging
- ✅ Intelligent caching: Modern systems handle content-type based caching

**Best for:** Enterprise APIs, when supporting multiple content types

### 5. Semantic Versioning (SemVer)

**Pattern**: `MAJOR.MINOR.PATCH` (e.g., 2.1.3)

**Rules:**
- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

**Pros:**
- ✅ Industry-standard versioning
- ✅ Clear semantic meaning
- ✅ Predictable upgrade paths
- ✅ Automated dependency management
- ✅ Wide tooling support
- ✅ Clear communication of changes
- ✅ Enables automated version bumping
- ✅ Supports package managers
- ✅ Facilitates release planning
- ✅ Reduces version confusion
- ✅ Scalable approach: Grows with API complexity and needs
- ✅ Professional standards: Enforces good development practices
- ✅ Future-proof: Prepares for growth and complexity

**Best for:** Enterprise APIs, libraries, when version clarity is critical

**Implementation:**
```go
type SemanticVersion struct {
    Major int `json:"major"`
    Minor int `json:"minor"`
    Patch int `json:"patch"`
    Pre   string `json:"pre,omitempty"`
}

func (v SemanticVersion) String() string {
    version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
    if v.Pre != "" {
        version += "-" + v.Pre
    }
    return version
}
```

### 6. Date-Based Versioning

**Pattern**: `2024.05.01` or `2024-05-01`

**Pros:**
- ✅ Clear release timeline
- ✅ Easy to understand
- ✅ Good for scheduled releases
- ✅ Predictable release cadence
- ✅ Easy to track release history
- ✅ Aligns with business calendars
- ✅ Simplifies release planning
- ✅ Clear chronological ordering
- ✅ Supports time-based feature flags
- ✅ Facilitates rollback strategies
- ✅ Temporal clarity: Focuses on when rather than what, simplifying communication
- ✅ Precise timing: Enables multiple releases per day with precise timestamps
- ✅ Customizable: Can be adapted to specific organizational needs

**Best for:** APIs with regular release schedules

### 7. Feature-Based Versioning

**Pattern**: Version based on feature availability

**Pros:**
- ✅ Feature-driven development
- ✅ Clear feature availability
- ✅ Enables gradual rollouts
- ✅ Supports A/B testing
- ✅ Customer-centric versioning
- ✅ Enables targeted feature releases
- ✅ Supports feature flag integration
- ✅ Facilitates incremental delivery
- ✅ Enables customer-specific versions
- ✅ Supports beta program management
- ✅ Granular control: Provides precise feature-level versioning
- ✅ Feature mapping: Clear relationship between features and versions
- ✅ Managed complexity: Controlled feature grouping prevents chaos

**Implementation:**
```go
type FeatureVersion struct {
    Version  string   `json:"version"`
    Features []string `json:"features"`
    Added    time.Time `json:"added"`
}
```

**Example:**
```json
{
  "version": "analytics-enabled",
  "features": ["real-time-analytics", "custom-dashboards"],
  "added": "2024-05-01T00:00:00Z"
}
```

### 8. Environment-Based Versioning

**Pattern**: Different versions per environment

**Pros:**
- ✅ Environment isolation
- ✅ Safe testing in staging
- ✅ Production stability
- ✅ Gradual rollout control
- ✅ Risk mitigation
- ✅ Environment-specific optimization
- ✅ Supports canary deployments
- ✅ Enables blue-green deployments
- ✅ Facilitates feature testing
- ✅ Supports rollback strategies
- ✅ Environment clarity: Clear separation prevents cross-environment issues
- ✅ Independent evolution: Each environment can evolve at its own pace
- ✅ Safety first: Extra steps ensure production stability

**Implementation:**
```go
type EnvironmentVersion struct {
    Environment string `json:"environment"`
    Version     string `json:"version"`
    Stable      bool   `json:"stable"`
}
```

**Examples:**
- Development: `dev-latest`
- Staging: `staging-v2.1`
- Production: `prod-v2.0`

## Advanced Implementation Patterns

### 1. Hybrid Versioning

Combine multiple strategies for flexibility:

```go
type VersioningStrategy string

const (
    URLPath     VersioningStrategy = "url_path"
    Header      VersioningStrategy = "header"
    QueryParam  VersioningStrategy = "query_param"
    ContentType VersioningStrategy = "content_type"
)

type VersionConfig struct {
    DefaultStrategy VersioningStrategy   `json:"default_strategy"`
    FallbackStrategies []VersioningStrategy `json:"fallback_strategies"`
    SupportedVersions []string            `json:"supported_versions"`
}
```

### 2. Version Negotiation

Implement content negotiation for version selection:

```go
type VersionNegotiator struct {
    strategies map[VersioningStrategy]VersionStrategy
    defaultVersion string
}

func (vn *VersionNegotiator) NegotiateVersion(r *http.Request) string {
    // Try each strategy in order
    for _, strategy := range vn.strategies {
        if version := vn.extractVersion(r, strategy); version != "" {
            return version
        }
    }
    return vn.defaultVersion
}
```

### 3. Version Routing

Dynamic routing based on version:

```go
type VersionRouter struct {
    versions map[string]http.Handler
    default  string
}

func (vr *VersionRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    version := extractVersion(r)
    
    if handler, exists := vr.versions[version]; exists {
        handler.ServeHTTP(w, r)
        return
    }
    
    // Fallback to default version
    if handler, exists := vr.versions[vr.default]; exists {
        handler.ServeHTTP(w, r)
        return
    }
    
    http.Error(w, "Unsupported version", http.StatusNotFound)
}
```

### 4. Deprecation Management

Complete deprecation lifecycle:

```go
type DeprecationManager struct {
    policies map[string]DeprecationPolicy
}

type DeprecationPolicy struct {
    Name         string        `json:"name"`
    Deprecation  time.Duration `json:"deprecation"`
    Sunset       time.Duration `json:"sunset"`
    Notification string        `json:"notification"`
}

func (dm *DeprecationManager) CheckDeprecation(version *APIVersion) (bool, *DeprecationPolicy) {
    if version.Status == "deprecated" && version.Deprecation != nil {
        policy := dm.policies["standard"]
        return true, &policy
    }
    return false, nil
}
```

## Migration Strategies

### 1. Blue-Green Deployment

Deploy new version alongside old version:

```yaml
# Kubernetes example
apiVersion: v1
kind: Service
metadata:
  name: hades-api
spec:
  selector:
    app: hades-api
    version: v2
  ports:
  - port: 80
    targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hades-api-v2
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hades-api
      version: v2
  template:
    metadata:
      labels:
        app: hades-api
        version: v2
    spec:
      containers:
      - name: api
        image: hades-toolkit/api:v2.0.0
        ports:
        - containerPort: 8080
```

### 2. Feature Flags

Gradual rollout with feature flags:

```go
type FeatureFlag struct {
    Name      string `json:"name"`
    Enabled   bool   `json:"enabled"`
    Rollout   int    `json:"rollout"` // Percentage
    Whitelist []string `json:"whitelist"`
}

func (ff *FeatureFlag) ShouldEnable(userID string) bool {
    if !ff.Enabled {
        return false
    }
    
    // Check whitelist
    for _, id := range ff.Whitelist {
        if id == userID {
            return true
        }
    }
    
    // Check rollout percentage
    hash := crc32.ChecksumIEEE([]byte(userID))
    return (hash % 100) < ff.Rollout
}
```

### 3. Canary Deployment

Gradual traffic shifting:

```go
type CanaryConfig struct {
    Version    string  `json:"version"`
    Weight     int     `json:"weight"`    // Percentage of traffic
    Thresholds map[string]float64 `json:"thresholds"` // Error rate thresholds
}

func (cc *CanaryConfig) ShouldRoute(requestID string) bool {
    hash := crc32.ChecksumIEEE([]byte(requestID))
    return (hash % 100) < cc.Weight
}
```

## Best Practices

### 1. Version Consistency

- Use the same versioning strategy across all endpoints
- Document version changes clearly
- Provide migration guides
- Maintain backward compatibility when possible

### 2. Deprecation Timeline

```
Release → 3 months → Deprecation → 6 months → Sunset → 12 months → End of Life
```

### 3. Communication

- Notify developers 6 months before deprecation
- Provide clear migration paths
- Offer support during transition
- Document breaking changes

### 4. Testing

- Test all supported versions
- Automated version compatibility tests
- Performance testing for each version
- Integration tests with different clients

## Implementation Examples

### 1. Middleware Implementation

```go
func VersioningMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        version := extractVersion(r)
        
        // Add version headers
        w.Header().Set("API-Version", version)
        w.Header().Set("X-API-Supported-Versions", "v1,v2,v3")
        
        // Check deprecation
        if isDeprecated(version) {
            w.Header().Set("X-API-Deprecated", "true")
            w.Header().Set("X-API-Sunset", getSunsetDate(version))
        }
        
        // Add version to context
        ctx := context.WithValue(r.Context(), "api_version", version)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 2. Version Discovery Endpoint

```go
func handleVersionDiscovery(w http.ResponseWriter, r *http.Request) {
    versions := map[string]*APIVersion{
        "v1": {
            Version:   "v1",
            Status:    "stable",
            Released:  time.Now().Add(-365 * 24 * time.Hour),
            Endpoints: []string{"/auth", "/dashboard", "/threats"},
            Features: []string{"Basic auth", "Simple CRUD"},
        },
        "v2": {
            Version:   "v2",
            Status:    "stable",
            Released:  time.Now().Add(-30 * 24 * time.Hour),
            Endpoints: []string{"/auth", "/dashboard", "/threats", "/analytics"},
            Features: []string{"Enhanced auth", "Analytics", "Webhooks"},
        },
    }
    
    response := map[string]interface{}{
        "versions": versions,
        "default_version": "v2",
        "supported_versions": []string{"v1", "v2"},
        "versioning_strategies": map[string]string{
            "url_path":     "/api/v{version}/endpoint",
            "header":       "API-Version: v{version}",
            "query_param":  "?version=v{version}",
            "content_type": "application/vnd.hades.v{version}+json",
        },
    }
    
    json.NewEncoder(w).Encode(response)
}
```

### 3. Client SDK Implementation

```python
class HadesAPI:
    def __init__(self, base_url, version="v2", api_key=None):
        self.base_url = base_url
        self.version = version
        self.api_key = api_key
        self.session = requests.Session()
        
    def _get_headers(self):
        headers = {
            "Content-Type": f"application/vnd.hades.{self.version}+json",
            "API-Version": self.version,
        }
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
        return headers
    
    def get_dashboard_metrics(self):
        response = self.session.get(
            f"{self.base_url}/api/dashboard/metrics",
            headers=self._get_headers()
        )
        return response.json()
    
    def set_version(self, version):
        """Change API version"""
        self.version = version
```

## Monitoring and Analytics

### 1. Version Usage Metrics

```go
type VersionMetrics struct {
    Version    string    `json:"version"`
    Requests   int64     `json:"requests"`
    Errors     int64     `json:"errors"`
    AvgLatency float64   `json:"avg_latency"`
    LastUsed   time.Time `json:"last_used"`
}

func (vm *VersionMetrics) RecordRequest(duration time.Duration, err error) {
    vm.Requests++
    if err != nil {
        vm.Errors++
    }
    vm.AvgLatency = (vm.AvgLatency*float64(vm.Requests-1) + duration.Seconds()) / float64(vm.Requests)
    vm.LastUsed = time.Now()
}
```

### 2. Health Checks

```go
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
    version := r.Context().Value("api_version").(string)
    
    health := map[string]interface{}{
        "status": "healthy",
        "version": version,
        "timestamp": time.Now(),
        "api_versions": map[string]interface{}{
            "v1": map[string]string{"status": "stable"},
            "v2": map[string]string{"status": "stable"},
            "v3": map[string]string{"status": "beta"},
        },
    }
    
    json.NewEncoder(w).Encode(health)
}
```

## Conclusion

Choose the versioning strategy that best fits your use case:

- **Public APIs**: URL path versioning
- **Internal APIs**: Header versioning
- **Enterprise APIs**: Content-Type versioning
- **Microservices**: Hybrid approach
- **Rapid iteration**: Feature-based versioning

The key is consistency, clear documentation, and smooth migration paths for your users.
