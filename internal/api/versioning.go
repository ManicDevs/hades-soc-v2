package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Custom context key types to avoid staticcheck SA1029
type versioningContextKey string

const (
	versioningAPIVersionKey versioningContextKey = "api_version"
)

// APIVersion represents an API version with metadata
type APIVersion struct {
	Version     string     `json:"version"`
	Status      string     `json:"status"` // active, deprecated, sunset
	Released    time.Time  `json:"released"`
	Deprecation *time.Time `json:"deprecation,omitempty"`
	Sunset      *time.Time `json:"sunset,omitempty"`
	Endpoints   []string   `json:"endpoints"`
	Features    []string   `json:"features"`
}

// VersioningConfig holds versioning configuration
type VersioningConfig struct {
	DefaultVersion    string            `json:"default_version"`
	SupportedVersions []string          `json:"supported_versions"`
	VersionHeaders    map[string]string `json:"version_headers"`
	DeprecationPolicy string            `json:"deprecation_policy"`
}

// VersioningMiddleware handles API versioning
type VersioningMiddleware struct {
	config   VersioningConfig
	versions map[string]*APIVersion
	router   *http.ServeMux
}

// NewVersioningMiddleware creates a new versioning middleware
func NewVersioningMiddleware() *VersioningMiddleware {
	vm := &VersioningMiddleware{
		config: VersioningConfig{
			DefaultVersion:    "v2",
			SupportedVersions: []string{"v1", "v2"},
			VersionHeaders: map[string]string{
				"API-Version": "X-API-Version",
				"Deprecated":  "X-API-Deprecated",
				"Sunset":      "X-API-Sunset",
				"Supported":   "X-API-Supported-Versions",
			},
			DeprecationPolicy: "6-month deprecation, 12-month sunset",
		},
		versions: make(map[string]*APIVersion),
		router:   http.NewServeMux(),
	}

	// Initialize versions
	vm.initializeVersions()
	return vm
}

// initializeVersions sets up all API versions
func (vm *VersioningMiddleware) initializeVersions() {
	// API v1 - Current stable
	vm.versions["v1"] = &APIVersion{
		Version:   "v1",
		Status:    "active",
		Released:  time.Now().Add(-365 * 24 * time.Hour), // Released 1 year ago
		Endpoints: []string{"/auth", "/dashboard", "/threats", "/users", "/security"},
		Features: []string{
			"JWT Authentication",
			"Basic CRUD operations",
			"Mock data responses",
			"Simple error handling",
		},
	}

	// API v2 - Latest with enhanced features
	vm.versions["v2"] = &APIVersion{
		Version:   "v2",
		Status:    "active",
		Released:  time.Now().Add(-30 * 24 * time.Hour), // Released 30 days ago
		Endpoints: []string{"/auth", "/dashboard", "/threats", "/users", "/security", "/analytics", "/webhooks"},
		Features: []string{
			"Enhanced JWT with refresh tokens",
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
	}

	// API v3 - Beta (future)
	vm.versions["v3"] = &APIVersion{
		Version:   "v3",
		Status:    "beta",
		Released:  time.Now().Add(-7 * 24 * time.Hour), // Released 7 days ago
		Endpoints: []string{"/auth", "/dashboard", "/threats", "/users", "/security", "/analytics", "/webhooks", "/ml", "/automation"},
		Features: []string{
			"All v2 features",
			"Machine learning threat detection",
			"Automated response workflows",
			"Advanced analytics dashboard",
			"Multi-tenant support",
			"GraphQL subscriptions",
			"Event-driven architecture",
		},
	}
}

// VersionHandler handles version-specific routing
func (vm *VersioningMiddleware) VersionHandler(version string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add version headers
		w.Header().Set(vm.config.VersionHeaders["API-Version"], version)
		w.Header().Set(vm.config.VersionHeaders["Supported"], strings.Join(vm.config.SupportedVersions, ", "))

		// Check deprecation status
		if apiVersion, exists := vm.versions[version]; exists {
			if apiVersion.Status == "deprecated" && apiVersion.Deprecation != nil {
				w.Header().Set(vm.config.VersionHeaders["Deprecated"], "true")
				w.Header().Set(vm.config.VersionHeaders["Sunset"], apiVersion.Sunset.Format(time.RFC3339))
			}
		}

		// Add version context to request
		ctx := context.WithValue(r.Context(), versioningAPIVersionKey, version)
		handler(w, r.WithContext(ctx))
	}
}

// VersionFromRequest extracts API version from request
func (vm *VersioningMiddleware) VersionFromRequest(r *http.Request) string {
	// Method 1: Header-based versioning
	if version := r.Header.Get("API-Version"); version != "" {
		if vm.isVersionSupported(version) {
			return version
		}
	}

	// Method 2: URL path versioning (/api/v1/users)
	if parts := strings.Split(r.URL.Path, "/"); len(parts) >= 3 && parts[1] == "api" {
		version := parts[2]
		if vm.isVersionSupported(version) {
			return version
		}
	}

	// Method 3: Query parameter versioning (?version=v1)
	if version := r.URL.Query().Get("version"); version != "" {
		if vm.isVersionSupported(version) {
			return version
		}
	}

	// Method 4: Content-Type versioning (application/vnd.hades.v2+json)
	if contentType := r.Header.Get("Content-Type"); strings.Contains(contentType, "vnd.hades.") {
		parts := strings.Split(contentType, "vnd.hades.")
		if len(parts) > 1 {
			version := strings.Split(parts[1], "+")[0]
			if vm.isVersionSupported(version) {
				return version
			}
		}
	}

	// Fallback to default version
	return vm.config.DefaultVersion
}

// isVersionSupported checks if a version is supported
func (vm *VersioningMiddleware) isVersionSupported(version string) bool {
	for _, supported := range vm.config.SupportedVersions {
		if supported == version {
			return true
		}
	}
	return false
}

// GetVersionInfo returns information about all versions
func (vm *VersioningMiddleware) GetVersionInfo() map[string]*APIVersion {
	return vm.versions
}

// GetVersion returns specific version info
func (vm *VersioningMiddleware) GetVersion(version string) (*APIVersion, error) {
	if v, exists := vm.versions[version]; exists {
		return v, nil
	}
	return nil, fmt.Errorf("version %s not found", version)
}

// Advanced Versioning Patterns:

// 1. Semantic Versioning (semver)
type SemanticVersion struct {
	Major int    `json:"major"`
	Minor int    `json:"minor"`
	Patch int    `json:"patch"`
	Pre   string `json:"pre,omitempty"`
}

// 2. Date-based versioning (YYYY-MM-DD)
type DateVersion struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

// 3. Feature-based versioning
type FeatureVersion struct {
	Version  string    `json:"version"`
	Features []string  `json:"features"`
	Added    time.Time `json:"added"`
}

// 4. Environment-based versioning
type EnvironmentVersion struct {
	Environment string `json:"environment"`
	Version     string `json:"version"`
	Stable      bool   `json:"stable"`
}

// VersionNegotiator handles content negotiation
type VersionNegotiator struct {
	supportedVersions []string
	defaultVersion    string
}

// NegotiateVersion performs content negotiation for API version
func (vn *VersionNegotiator) NegotiateVersion(r *http.Request) string {
	// Check Accept header for version preferences
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/vnd.hades") {
		// Parse vendor-specific media type
		parts := strings.Split(accept, "application/vnd.hades.")
		if len(parts) > 1 {
			versionPart := strings.Split(parts[1], ";")[0]
			version := strings.Split(versionPart, "+")[0]
			for _, supported := range vn.supportedVersions {
				if strings.Contains(version, supported) {
					return supported
				}
			}
		}
	}

	return vn.defaultVersion
}

// DeprecationManager handles API deprecation lifecycle
type DeprecationManager struct {
	policies map[string]DeprecationPolicy
}

type DeprecationPolicy struct {
	Name         string        `json:"name"`
	Deprecation  time.Duration `json:"deprecation"`  // How long before deprecation
	Sunset       time.Duration `json:"sunset"`       // How long before sunset
	Notification string        `json:"notification"` // How to notify users
}

// NewDeprecationManager creates a new deprecation manager
func NewDeprecationManager() *DeprecationManager {
	return &DeprecationManager{
		policies: map[string]DeprecationPolicy{
			"standard": {
				Name:         "Standard Policy",
				Deprecation:  6 * 30 * 24 * time.Hour,  // 6 months
				Sunset:       12 * 30 * 24 * time.Hour, // 12 months
				Notification: "email, banner, api-header",
			},
			"security": {
				Name:         "Security Policy",
				Deprecation:  3 * 30 * 24 * time.Hour, // 3 months
				Sunset:       6 * 30 * 24 * time.Hour, // 6 months
				Notification: "email, banner, api-header, security-advisory",
			},
			"experimental": {
				Name:         "Experimental Policy",
				Deprecation:  1 * 30 * 24 * time.Hour, // 1 month
				Sunset:       3 * 30 * 24 * time.Hour, // 3 months
				Notification: "api-header, documentation",
			},
		},
	}
}

// CheckDeprecation checks if a version is deprecated
func (dm *DeprecationManager) CheckDeprecation(version *APIVersion) (bool, *DeprecationPolicy) {
	if version.Status == "deprecated" && version.Deprecation != nil {
		policy := dm.policies["standard"]
		return true, &policy
	}
	return false, nil
}
