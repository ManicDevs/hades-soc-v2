package versioning

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Custom context key types to avoid staticcheck SA1029
type contextKey string

const (
	apiVersionKey  contextKey = "api_version"
	versionInfoKey contextKey = "version_info"
)

// Version Status Constants
const (
	StatusLegacy     = "legacy"
	StatusStable     = "stable"
	StatusPreferred  = "preferred"
	StatusBeta       = "beta"
	StatusDeprecated = "deprecated"
)

// Version represents an API version with metadata
type Version struct {
	ID          string                 `json:"id"`
	Number      string                 `json:"number"`
	Status      string                 `json:"status"`
	Released    time.Time              `json:"released"`
	Deprecation *time.Time             `json:"deprecation,omitempty"`
	Sunset      *time.Time             `json:"sunset,omitempty"`
	Features    []string               `json:"features"`
	Endpoints   []string               `json:"endpoints"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// VersionManager manages API versions with hierarchy support
type VersionManager struct {
	versions         map[string]*Version
	defaultVersion   string
	preferredVersion string
	config           ManagerConfig
}

// ManagerConfig holds version manager configuration
type ManagerConfig struct {
	DefaultVersion    string            `json:"default_version"`
	PreferredVersion  string            `json:"preferred_version"`
	SupportedVersions []string          `json:"supported_versions"`
	DeprecationPolicy DeprecationPolicy `json:"deprecation_policy"`
	VersionHeaders    map[string]string `json:"version_headers"`
}

// DeprecationPolicy defines how version deprecation is handled
type DeprecationPolicy struct {
	Name         string        `json:"name"`
	Deprecation  time.Duration `json:"deprecation"`
	Sunset       time.Duration `json:"sunset"`
	Notification string        `json:"notification"`
}

// NewVersionManager creates a new version manager with hierarchy
func NewVersionManager(config ManagerConfig) *VersionManager {
	vm := &VersionManager{
		versions:         make(map[string]*Version),
		defaultVersion:   config.DefaultVersion,
		preferredVersion: config.PreferredVersion,
		config:           config,
	}

	vm.initializeVersions()
	return vm
}

// initializeVersions sets up the version hierarchy
func (vm *VersionManager) initializeVersions() {
	now := time.Now()

	// v1 - Legacy (deprecated but still supported)
	vm.versions["v1"] = &Version{
		ID:          "v1",
		Number:      "1.0.0",
		Status:      StatusLegacy,
		Released:    now.Add(-365 * 24 * time.Hour),                 // Released 1 year ago
		Deprecation: &[]time.Time{now.Add(30 * 24 * time.Hour)}[0],  // Deprecated in 30 days
		Sunset:      &[]time.Time{now.Add(180 * 24 * time.Hour)}[0], // Sunset in 6 months
		Features: []string{
			"Basic authentication",
			"Simple CRUD operations",
			"Basic error handling",
			"Mock data responses",
		},
		Endpoints: []string{
			"/auth/login",
			"/auth/logout",
			"/dashboard/metrics",
			"/threats",
			"/users",
			"/security/policies",
		},
		Metadata: map[string]interface{}{
			"compatibility":   "legacy",
			"support_level":   "limited",
			"migration_guide": "/api/v1/migration-guide",
		},
	}

	// v2 - Preferred (current stable version)
	vm.versions["v2"] = &Version{
		ID:       "v2",
		Number:   "2.0.0",
		Status:   StatusPreferred,
		Released: now.Add(-30 * 24 * time.Hour), // Released 30 days ago
		Features: []string{
			"Enhanced JWT authentication",
			"Advanced filtering and pagination",
			"Real-time WebSocket support",
			"Rate limiting and quotas",
			"Advanced error responses",
			"Request/response compression",
			"API analytics and metrics",
			"Webhook integrations",
			"GraphQL endpoints",
			"OpenAPI 3.0 documentation",
		},
		Endpoints: []string{
			"/auth/login",
			"/auth/logout",
			"/auth/refresh",
			"/dashboard/metrics",
			"/dashboard/analytics",
			"/threats",
			"/threats/feed",
			"/users",
			"/users/sessions",
			"/security/policies",
			"/security/vulnerabilities",
			"/analytics",
			"/webhooks",
		},
		Metadata: map[string]interface{}{
			"compatibility": "current",
			"support_level": "full",
			"documentation": "/api/v2/docs",
			"recommended":   true,
		},
	}

	// v3 - Beta (future development)
	vm.versions["v3"] = &Version{
		ID:       "v3",
		Number:   "3.0.0-beta",
		Status:   StatusBeta,
		Released: now.Add(-7 * 24 * time.Hour), // Released 7 days ago
		Features: []string{
			"All v2 features",
			"Machine learning threat detection",
			"Automated response workflows",
			"Advanced analytics dashboard",
			"Multi-tenant support",
			"GraphQL subscriptions",
			"Event-driven architecture",
			"Advanced RBAC system",
			"Real-time collaboration",
			"AI-powered recommendations",
		},
		Endpoints: []string{
			"/auth/login",
			"/auth/logout",
			"/auth/refresh",
			"/dashboard/metrics",
			"/dashboard/analytics",
			"/dashboard/ai-insights",
			"/threats",
			"/threats/predictive",
			"/threats/automated-response",
			"/users",
			"/users/collaboration",
			"/security/policies",
			"/security/ai-recommendations",
			"/analytics",
			"/analytics/ml",
			"/webhooks",
			"/ml/models",
			"/automation/workflows",
		},
		Metadata: map[string]interface{}{
			"compatibility":    "beta",
			"support_level":    "preview",
			"documentation":    "/api/v3/docs",
			"feedback_channel": "/api/v3/feedback",
			"stability":        "beta",
		},
	}
}

// GetVersion returns version information
func (vm *VersionManager) GetVersion(version string) (*Version, error) {
	if v, exists := vm.versions[version]; exists {
		return v, nil
	}
	return nil, fmt.Errorf("version %s not found", version)
}

// GetDefaultVersion returns the default version
func (vm *VersionManager) GetDefaultVersion() *Version {
	return vm.versions[vm.defaultVersion]
}

// GetPreferredVersion returns the preferred version
func (vm *VersionManager) GetPreferredVersion() *Version {
	return vm.versions[vm.preferredVersion]
}

// GetAllVersions returns all versions
func (vm *VersionManager) GetAllVersions() map[string]*Version {
	return vm.versions
}

// GetSupportedVersions returns supported versions only
func (vm *VersionManager) GetSupportedVersions() []string {
	var supported []string
	for _, version := range vm.versions {
		if version.Status != StatusDeprecated {
			supported = append(supported, version.Number)
		}
	}
	return supported
}

// IsVersionSupported checks if a version is supported
func (vm *VersionManager) IsVersionSupported(version string) bool {
	if v, exists := vm.versions[version]; exists {
		return v.Status != StatusDeprecated
	}
	return false
}

// GetVersionStatus returns the status of a version
func (vm *VersionManager) GetVersionStatus(version string) string {
	if v, exists := vm.versions[version]; exists {
		return v.Status
	}
	return "unknown"
}

// ApplyVersionHeaders applies version-specific headers
func (vm *VersionManager) ApplyVersionHeaders(w http.ResponseWriter, version string) {
	v, exists := vm.versions[version]
	if !exists {
		return
	}

	// Standard headers
	w.Header().Set("API-Version", v.Number)
	w.Header().Set("X-API-Version-ID", v.ID)
	w.Header().Set("X-API-Version-Status", v.Status)
	w.Header().Set("X-API-Supported-Versions", strings.Join(vm.GetSupportedVersions(), ", "))

	// Status-specific headers
	switch v.Status {
	case StatusLegacy:
		w.Header().Set("X-API-Legacy", "true")
		w.Header().Set("X-API-Deprecation-Warning", "This version is deprecated. Please migrate to v2.")
		if v.Deprecation != nil {
			w.Header().Set("X-API-Deprecation-Date", v.Deprecation.Format(time.RFC3339))
		}
		if v.Sunset != nil {
			w.Header().Set("X-API-Sunset-Date", v.Sunset.Format(time.RFC3339))
		}
		w.Header().Set("X-API-Migration-Guide", "/api/v1/migration-guide")

	case StatusPreferred:
		w.Header().Set("X-API-Preferred", "true")
		w.Header().Set("X-API-Stable", "true")
		w.Header().Set("X-API-Recommended", "true")

	case StatusBeta:
		w.Header().Set("X-API-Beta", "true")
		w.Header().Set("X-API-Stability", "beta")
		w.Header().Set("X-API-Feedback-Channel", "/api/v3/feedback")
		w.Header().Set("X-API-Preview", "true")
	}

	// Feature headers
	if len(v.Features) > 0 {
		w.Header().Set("X-API-Features", strings.Join(v.Features, ","))
	}

	// Metadata headers
	for key, value := range v.Metadata {
		if headerKey, exists := vm.config.VersionHeaders[key]; exists {
			if strValue, ok := value.(string); ok {
				w.Header().Set(headerKey, strValue)
			}
		}
	}
}

// Middleware creates versioning middleware
func (vm *VersionManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract version from request
		version := vm.extractVersion(r)

		// Validate version
		if !vm.IsVersionSupported(version) {
			// Fall back to preferred version
			version = vm.preferredVersion
		}

		// Apply version headers
		vm.ApplyVersionHeaders(w, version)

		// Add version to context
		ctx := context.WithValue(r.Context(), apiVersionKey, version)
		ctx = context.WithValue(ctx, versionInfoKey, vm.versions[version])

		// Record metrics
		vm.recordVersionUsage(version, r)

		// Call next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractVersion extracts version from request using multiple strategies
func (vm *VersionManager) extractVersion(r *http.Request) string {
	// Strategy 1: URL path versioning (/api/v1/users, /api/v2/users)
	path := r.URL.Path
	if strings.Contains(path, "/api/") {
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "api" && i+1 < len(parts) {
				version := parts[i+1]
				if vm.IsVersionSupported(version) {
					return version
				}
			}
		}
	}

	// Strategy 2: Header versioning (API-Version: v2)
	if version := r.Header.Get("API-Version"); version != "" {
		if vm.IsVersionSupported(version) {
			return version
		}
	}

	// Strategy 3: Query parameter versioning (?version=v2)
	if version := r.URL.Query().Get("version"); version != "" {
		if vm.IsVersionSupported(version) {
			return version
		}
	}

	// Strategy 4: Content-Type versioning (application/vnd.hades.v2+json)
	if contentType := r.Header.Get("Content-Type"); strings.Contains(contentType, "vnd.hades.") {
		parts := strings.Split(contentType, "vnd.hades.")
		if len(parts) > 1 {
			version := strings.Split(parts[1], "+")[0]
			if vm.IsVersionSupported(version) {
				return version
			}
		}
	}

	// Strategy 5: Accept header versioning
	if accept := r.Header.Get("Accept"); strings.Contains(accept, "vnd.hades.") {
		parts := strings.Split(accept, "vnd.hades.")
		if len(parts) > 1 {
			version := strings.Split(parts[1], "+")[0]
			if vm.IsVersionSupported(version) {
				return version
			}
		}
	}

	// Fallback to preferred version
	return vm.preferredVersion
}

// recordVersionUsage records version usage metrics
func (vm *VersionManager) recordVersionUsage(version string, r *http.Request) {
	// This would integrate with your metrics system
	// For now, we'll just log the usage
	fmt.Printf("Version usage: %s requested %s\n", version, r.URL.Path)
}

// GetMigrationPath returns the recommended migration path
func (vm *VersionManager) GetMigrationPath(fromVersion, toVersion string) (*MigrationPath, error) {
	from, err := vm.GetVersion(fromVersion)
	if err != nil {
		return nil, fmt.Errorf("source version not found: %v", err)
	}

	to, err := vm.GetVersion(toVersion)
	if err != nil {
		return nil, fmt.Errorf("target version not found: %v", err)
	}

	return &MigrationPath{
		From:          from.Number,
		To:            to.Number,
		Complexity:    vm.calculateMigrationComplexity(from, to),
		Breaking:      vm.hasBreakingChanges(from, to),
		Guide:         fmt.Sprintf("/api/migration/%s-to-%s", from.Number, to.Number),
		Automated:     from.ID == "v1" && to.ID == "v2", // Only v1->v2 is automated
		EstimatedTime: vm.estimateMigrationTime(from, to),
	}, nil
}

// MigrationPath represents a migration path between versions
type MigrationPath struct {
	From          string        `json:"from"`
	To            string        `json:"to"`
	Complexity    string        `json:"complexity"`
	Breaking      bool          `json:"breaking"`
	Guide         string        `json:"guide"`
	Automated     bool          `json:"automated"`
	EstimatedTime time.Duration `json:"estimated_time"`
}

// Helper functions
func (vm *VersionManager) calculateMigrationComplexity(from, to *Version) string {
	if from.ID == "v1" && to.ID == "v2" {
		return "medium"
	}
	if from.ID == "v2" && to.ID == "v3" {
		return "high"
	}
	return "unknown"
}

func (vm *VersionManager) hasBreakingChanges(from, to *Version) bool {
	// v1->v2 has breaking changes
	// v2->v3 has breaking changes (beta)
	return true
}

func (vm *VersionManager) estimateMigrationTime(from, to *Version) time.Duration {
	if from.ID == "v1" && to.ID == "v2" {
		return 2 * time.Hour
	}
	if from.ID == "v2" && to.ID == "v3" {
		return 4 * time.Hour
	}
	return 1 * time.Hour
}
